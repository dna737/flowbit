package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"flowbit/backend/internal/models"
	"flowbit/backend/internal/queue"
)

// maxAttempts is the total number of delivery attempts before a job is sent to
// the dead-letter queue. Attempts are 0-based; attempt maxAttempts-1 is the last.
const maxAttempts = 3

// Store is the persistence surface the worker needs for job execution.
type Store interface {
	UpdateJobStatus(ctx context.Context, id string, status string, lastError *string) error
	WriteToDLQ(ctx context.Context, jobID, jobType string, payload []byte, errMsg string) error
}

// Publisher re-publishes a job message to the queue (used for retries).
type Publisher interface {
	PublishJob(ctx context.Context, msg queue.JobMessage) error
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

// HandleJob processes a single job message:
//   - marks it running, executes it, and marks succeeded on success
//   - on failure: re-publishes with attempt+1 if attempt < maxAttempts-1
//   - on the final attempt (attempt == maxAttempts-1): marks failed and writes to DLQ
func HandleJob(ctx context.Context, store Store, pub Publisher, msg queue.JobMessage, logf Logf) {
	if err := store.UpdateJobStatus(ctx, msg.JobID, models.JobStatusRunning, nil); err != nil {
		if logf != nil {
			logf("mark running failed for %s: %v", msg.JobID, err)
		}
		return
	}

	execErr := execute(msg, logf)

	if execErr != nil {
		lastError := execErr.Error()
		if msg.Attempt < maxAttempts-1 {
			// Not yet exhausted — update status to retrying and re-publish.
			if err := store.UpdateJobStatus(ctx, msg.JobID, models.JobStatusRetrying, &lastError); err != nil {
				if logf != nil {
					logf("mark retrying failed for %s: %v", msg.JobID, err)
				}
			}
			retry := msg
			retry.Attempt = msg.Attempt + 1
			if err := pub.PublishJob(ctx, retry); err != nil {
				if logf != nil {
					logf("re-publish failed for %s (attempt %d): %v", msg.JobID, retry.Attempt, err)
				}
			} else if logf != nil {
				logf("job %s failed (attempt %d/%d); re-queued for attempt %d",
					msg.JobID, msg.Attempt+1, maxAttempts, retry.Attempt+1)
			}
			return
		}

		// Final attempt exhausted — mark failed and write to DLQ.
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

// execute runs the job type's business logic. Returns non-nil on failure.
// The "fail" type always fails and is used to drive the retry/DLQ path in tests and demos.
func execute(msg queue.JobMessage, logf Logf) error {
	switch msg.JobType {
	case "echo":
		if logf != nil {
			logf("processed echo job %s params=%v", msg.JobID, msg.Parameters)
		}
		return nil
	case "email":
		if logf != nil {
			logf("processed email job %s to=%v subject=%v", msg.JobID, msg.Parameters["to"], msg.Parameters["subject"])
		}
		return nil
	case "image_resize":
		if logf != nil {
			logf("processed image_resize job %s width=%v height=%v", msg.JobID, msg.Parameters["width"], msg.Parameters["height"])
		}
		return nil
	case "url_scrape":
		if logf != nil {
			logf("processed url_scrape job %s url=%v", msg.JobID, msg.Parameters["url"])
		}
		return nil
	case "fail":
		return fmt.Errorf("deliberate failure (job %s attempt %d)", msg.JobID, msg.Attempt)
	default:
		return fmt.Errorf("unsupported job_type: %s", msg.JobType)
	}
}
