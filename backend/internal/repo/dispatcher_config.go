package repo

import (
	"context"
	"encoding/json"
	"fmt"
)

const dispatcherConfigSingletonID = 1

// DispatcherConfigRepo loads global dispatcher settings (allowed job types for Gemini).
type DispatcherConfigRepo struct {
	pool dbPool
}

func NewDispatcherConfigRepo(pool dbPool) *DispatcherConfigRepo {
	return &DispatcherConfigRepo{pool: pool}
}

// GetAllowedJobTypes returns the canonical list of job_type strings (order preserved as in JSON).
func (r *DispatcherConfigRepo) GetAllowedJobTypes(ctx context.Context) ([]string, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT allowed_job_types
		FROM dispatcher_config
		WHERE id = $1
	`, dispatcherConfigSingletonID)

	var raw []byte
	if err := row.Scan(&raw); err != nil {
		return nil, fmt.Errorf("get dispatcher_config: %w", err)
	}

	var out []string
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("decode allowed_job_types: %w", err)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("allowed_job_types is empty")
	}
	return out, nil
}
