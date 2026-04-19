package repo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"flowbit/backend/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var ErrJobNotFound = errors.New("job not found")

// dbPool is the subset of *pgxpool.Pool used by JobsRepo (mockable in tests).
type dbPool interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

type JobsRepo struct {
	pool dbPool
}

// NewJobsRepo builds a repo backed by a pgx pool or any dbPool implementation (e.g. tests).
func NewJobsRepo(pool dbPool) *JobsRepo {
	return &JobsRepo{pool: pool}
}

// CreateJob persists a new job owned by userID.
func (r *JobsRepo) CreateJob(ctx context.Context, userID, jobType string, parameters map[string]any, status string) (models.Job, error) {
	payload, err := json.Marshal(parameters)
	if err != nil {
		return models.Job{}, fmt.Errorf("marshal parameters: %w", err)
	}

	row := r.pool.QueryRow(ctx, `
		INSERT INTO jobs (user_id, job_type, parameters, status)
		VALUES ($1, $2, $3::jsonb, $4)
		RETURNING id::text, job_type, parameters, status, attempts, last_error, created_at, updated_at
	`, userID, jobType, string(payload), status)

	out, err := scanJob(row)
	if err != nil {
		return models.Job{}, fmt.Errorf("insert job: %w", err)
	}
	return out, nil
}

// GetJobByID fetches a job by id without scoping. Internal use only (e.g. realtime listener);
// API handlers must use GetJobByUserAndID to prevent cross-user reads.
func (r *JobsRepo) GetJobByID(ctx context.Context, id string) (models.Job, error) {
	if _, err := uuid.Parse(id); err != nil {
		return models.Job{}, ErrJobNotFound
	}

	row := r.pool.QueryRow(ctx, `
		SELECT id::text, job_type, parameters, status, attempts, last_error, created_at, updated_at
		FROM jobs
		WHERE id = $1
	`, id)

	out, err := scanJob(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Job{}, ErrJobNotFound
		}
		return models.Job{}, fmt.Errorf("query job: %w", err)
	}
	return out, nil
}

// GetJobByUserAndID fetches a job only if it belongs to userID. Returns ErrJobNotFound
// for both "no such id" and "id belongs to another user" so we don't leak existence.
func (r *JobsRepo) GetJobByUserAndID(ctx context.Context, userID, id string) (models.Job, error) {
	if _, err := uuid.Parse(id); err != nil {
		return models.Job{}, ErrJobNotFound
	}

	row := r.pool.QueryRow(ctx, `
		SELECT id::text, job_type, parameters, status, attempts, last_error, created_at, updated_at
		FROM jobs
		WHERE id = $1 AND user_id = $2
	`, id, userID)

	out, err := scanJob(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Job{}, ErrJobNotFound
		}
		return models.Job{}, fmt.Errorf("query job: %w", err)
	}
	return out, nil
}

// ListJobsByUser returns the most recent jobs owned by userID, capped at limit.
func (r *JobsRepo) ListJobsByUser(ctx context.Context, userID string, limit int) ([]models.Job, error) {
	if limit <= 0 {
		limit = 100
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id::text, job_type, parameters, status, attempts, last_error, created_at, updated_at
		FROM jobs
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("list jobs: %w", err)
	}
	defer rows.Close()

	var jobs []models.Job
	for rows.Next() {
		job, err := scanJob(rows)
		if err != nil {
			return nil, fmt.Errorf("scan listed job: %w", err)
		}
		jobs = append(jobs, job)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate listed jobs: %w", err)
	}

	return jobs, nil
}

func scanJob(row pgx.Row) (models.Job, error) {
	var out models.Job
	var rawParams []byte
	if err := row.Scan(&out.ID, &out.JobType, &rawParams, &out.Status, &out.Attempts, &out.LastError, &out.CreatedAt, &out.UpdatedAt); err != nil {
		return models.Job{}, err
	}
	if err := json.Unmarshal(rawParams, &out.Parameters); err != nil {
		return models.Job{}, fmt.Errorf("decode parameters: %w", err)
	}
	return out, nil
}

// UpdateJobStatus updates a job's status and emits a small NOTIFY payload (`{id,user_id}`)
// so the realtime listener can fan out to that user's WS clients. Sending only the id
// keeps NOTIFY well under Postgres' ~8000-byte payload cap regardless of parameter size.
func (r *JobsRepo) UpdateJobStatus(ctx context.Context, id string, status string, lastError *string) error {
	row := r.pool.QueryRow(ctx, `
		WITH updated AS (
			UPDATE jobs
			SET status     = $2,
			    last_error = $3,
			    attempts   = CASE WHEN $2 = 'running' THEN attempts + 1 ELSE attempts END,
			    updated_at = NOW()
			WHERE id = $1
			RETURNING id::text AS id, user_id
		)
		SELECT updated.id
		FROM updated, LATERAL pg_notify(
			'job_status',
			json_build_object('id', updated.id, 'user_id', updated.user_id)::text
		)
	`, id, status, lastError)

	var updatedID string
	if err := row.Scan(&updatedID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrJobNotFound
		}
		return fmt.Errorf("update job status: %w", err)
	}
	return nil
}

func (r *JobsRepo) WriteToDLQ(ctx context.Context, jobID, jobType string, payload []byte, errMsg string) error {
	if _, err := uuid.Parse(jobID); err != nil {
		return fmt.Errorf("invalid job_id for dlq: %w", err)
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO dead_letter_queue (job_id, job_type, payload, error_message)
		VALUES ($1, $2, $3::jsonb, $4)
	`, jobID, jobType, string(payload), errMsg)
	if err != nil {
		return fmt.Errorf("write to dlq: %w", err)
	}
	return nil
}
