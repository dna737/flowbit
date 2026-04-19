package repo

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// UsersRepo reads and writes per-user rows (dispatch_categories for Gemini).
type UsersRepo struct {
	pool            dbPool
	jobTypeDefaults *DispatcherConfigRepo // when set, seeds new users from allowed_job_types
}

func NewUsersRepo(pool dbPool, jobTypeDefaults *DispatcherConfigRepo) *UsersRepo {
	return &UsersRepo{pool: pool, jobTypeDefaults: jobTypeDefaults}
}

// GetCategories returns stored category labels for userID (may be empty).
// If there is no users row yet, seeds dispatch_categories from dispatcher_config (same as default job types) when jobTypeDefaults is configured.
func (r *UsersRepo) GetCategories(ctx context.Context, userID string) ([]string, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT dispatch_categories
		FROM users
		WHERE id = $1
	`, userID)

	var raw []byte
	if err := row.Scan(&raw); err != nil {
		if err == pgx.ErrNoRows {
			if r.jobTypeDefaults == nil {
				return []string{}, nil
			}
			defaults, derr := r.jobTypeDefaults.GetAllowedJobTypes(ctx)
			if derr != nil {
				return nil, fmt.Errorf("load default categories: %w", derr)
			}
			if err := r.SetCategories(ctx, userID, defaults); err != nil {
				return nil, err
			}
			return defaults, nil
		}
		return nil, fmt.Errorf("get users: %w", err)
	}

	var out []string
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("decode dispatch_categories: %w", err)
	}
	return out, nil
}

// SetCategories replaces the full category list for userID.
func (r *UsersRepo) SetCategories(ctx context.Context, userID string, categories []string) error {
	if categories == nil {
		categories = []string{}
	}
	payload, err := json.Marshal(categories)
	if err != nil {
		return fmt.Errorf("marshal categories: %w", err)
	}

	_, err = r.pool.Exec(ctx, `
		INSERT INTO users (id, dispatch_categories)
		VALUES ($1, $2::jsonb)
		ON CONFLICT (id) DO UPDATE SET dispatch_categories = EXCLUDED.dispatch_categories
	`, userID, payload)
	if err != nil {
		return fmt.Errorf("set users: %w", err)
	}
	return nil
}
