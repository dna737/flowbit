package models

import "time"

const (
	JobStatusPending   = "pending"
	JobStatusRunning   = "running"
	JobStatusSucceeded = "succeeded"
	JobStatusFailed    = "failed"
)

type Job struct {
	ID         string                 `json:"id"`
	JobType    string                 `json:"job_type"`
	Parameters map[string]any         `json:"parameters"`
	Status     string                 `json:"status"`
	Attempts   int                    `json:"attempts"`
	LastError  *string                `json:"last_error,omitempty"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
}
