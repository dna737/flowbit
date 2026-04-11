package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"flowbit/backend/internal/models"
	"flowbit/backend/internal/queue"
)

// Store is the persistence surface the worker needs for job execution.
type Store interface {
	UpdateJobStatus(ctx context.Context, id string, status string, lastError *string) error
	WriteToDLQ(ctx context.Context, jobID, jobType string, payload []byte, errMsg string) error
}

// Logf logs a formatted message (e.g. log.Printf); may be nil to skip info logs.
type Logf func(format string, args ...any)

// ReadBackoff returns sleep duration after Kafka read errors (exponential cap).
func ReadBackoff(n int) time.Duration {
	const maxBackoff = 30 * time.Second
	d := time.Duration(1<<uint(n)) * 200 * time.Millisecond
	if d > maxBackoff || d <= 0 {
		return maxBackoff
	}
	return d
}

// HandleJob transitions job status: running -> succeeded, or failed + DLQ on error.
func HandleJob(ctx context.Context, store Store, msg queue.JobMessage, logf Logf) {
	if err := store.UpdateJobStatus(ctx, msg.JobID, models.JobStatusRunning, nil); err != nil {
		if logf != nil {
			logf("mark running failed for %s: %v", msg.JobID, err)
		}
		return
	}

	var execErr error
	switch msg.JobType {
	case "echo":
		if logf != nil {
			logf("processed echo job %s params=%v", msg.JobID, msg.Parameters)
		}
	default:
		execErr = fmt.Errorf("unsupported job_type: %s", msg.JobType)
	}

	if execErr != nil {
		lastError := execErr.Error()
		if err := store.UpdateJobStatus(ctx, msg.JobID, models.JobStatusFailed, &lastError); err != nil {
			if logf != nil {
				logf("mark failed for %s: %v", msg.JobID, err)
			}
		}
		payload, _ := json.Marshal(msg.Parameters)
		if err := store.WriteToDLQ(ctx, msg.JobID, msg.JobType, payload, lastError); err != nil {
			if logf != nil {
				logf("dlq write failed for %s: %v", msg.JobID, err)
			}
		}
		return
	}

	if err := store.UpdateJobStatus(ctx, msg.JobID, models.JobStatusSucceeded, nil); err != nil {
		if logf != nil {
			logf("mark succeeded failed for %s: %v", msg.JobID, err)
		}
	}
}
