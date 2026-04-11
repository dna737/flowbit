package repo

import (
	"context"
	"errors"
	"testing"
	"time"

	"flowbit/backend/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
)

func TestJobsRepo_CreateJob(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	t0 := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)
	rows := mock.NewRows([]string{"id", "job_type", "parameters", "status", "attempts", "last_error", "created_at", "updated_at"}).
		AddRow("550e8400-e29b-41d4-a716-446655440000", "echo", []byte(`{"k":"v"}`), "pending", 0, nil, t0, t0)

	mock.ExpectQuery(`INSERT INTO jobs`).
		WithArgs("echo", pgxmock.AnyArg(), models.JobStatusPending).
		WillReturnRows(rows)

	r := NewJobsRepo(mock)
	job, err := r.CreateJob(context.Background(), "echo", map[string]any{"k": "v"}, models.JobStatusPending)
	if err != nil {
		t.Fatal(err)
	}
	if job.ID != "550e8400-e29b-41d4-a716-446655440000" || job.JobType != "echo" || job.Status != "pending" {
		t.Fatalf("unexpected job: %+v", job)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestJobsRepo_GetJobByID_invalidUUID(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	r := NewJobsRepo(mock)
	_, err = r.GetJobByID(context.Background(), "not-a-uuid")
	if !errors.Is(err, ErrJobNotFound) {
		t.Fatalf("want ErrJobNotFound, got %v", err)
	}
}

func TestJobsRepo_GetJobByID_notFound(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	mock.ExpectQuery(`SELECT id::text`).
		WithArgs("550e8400-e29b-41d4-a716-446655440001").
		WillReturnError(pgx.ErrNoRows)

	r := NewJobsRepo(mock)
	_, err = r.GetJobByID(context.Background(), "550e8400-e29b-41d4-a716-446655440001")
	if !errors.Is(err, ErrJobNotFound) {
		t.Fatalf("want ErrJobNotFound, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestJobsRepo_UpdateJobStatus_noRows(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	mock.ExpectExec(`UPDATE jobs`).
		WithArgs("550e8400-e29b-41d4-a716-446655440002", models.JobStatusRunning, pgxmock.AnyArg()).
		WillReturnResult(pgxmock.NewResult("UPDATE", 0))

	r := NewJobsRepo(mock)
	err = r.UpdateJobStatus(context.Background(), "550e8400-e29b-41d4-a716-446655440002", models.JobStatusRunning, nil)
	if !errors.Is(err, ErrJobNotFound) {
		t.Fatalf("want ErrJobNotFound, got %v", err)
	}
}

func TestJobsRepo_WriteToDLQ_invalidJobID(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	r := NewJobsRepo(mock)
	err = r.WriteToDLQ(context.Background(), "bad-id", "echo", []byte(`{}`), "oops")
	if err == nil {
		t.Fatal("expected error")
	}
}
