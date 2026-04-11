package worker

import (
	"context"
	"errors"
	"testing"
	"time"

	"flowbit/backend/internal/models"
	"flowbit/backend/internal/queue"
)

type spyStore struct {
	updates   [][3]string // id, status, lastErr string (empty = nil)
	dlqWrites int
	failRun   bool
	failFail  bool
	failDLQ   bool
	failOK    bool
}

func (s *spyStore) UpdateJobStatus(_ context.Context, id string, status string, lastError *string) error {
	le := ""
	if lastError != nil {
		le = *lastError
	}
	s.updates = append(s.updates, [3]string{id, status, le})
	switch status {
	case models.JobStatusRunning:
		if s.failRun {
			return errors.New("run err")
		}
	case models.JobStatusFailed:
		if s.failFail {
			return errors.New("fail err")
		}
	case models.JobStatusSucceeded:
		if s.failOK {
			return errors.New("ok err")
		}
	}
	return nil
}

func (s *spyStore) WriteToDLQ(context.Context, string, string, []byte, string) error {
	s.dlqWrites++
	if s.failDLQ {
		return errors.New("dlq err")
	}
	return nil
}

func TestHandleJob_echo_succeeds(t *testing.T) {
	var st spyStore
	id := "550e8400-e29b-41d4-a716-446655440010"
	HandleJob(context.Background(), &st, queue.JobMessage{JobID: id, JobType: "echo", Parameters: map[string]any{}}, nil)
	if len(st.updates) != 2 {
		t.Fatalf("updates: %+v", st.updates)
	}
	if st.updates[0][1] != models.JobStatusRunning || st.updates[1][1] != models.JobStatusSucceeded {
		t.Fatalf("want running then succeeded: %+v", st.updates)
	}
	if st.dlqWrites != 0 {
		t.Fatalf("dlq: %d", st.dlqWrites)
	}
}

func TestHandleJob_unsupported_failedAndDLQ(t *testing.T) {
	var st spyStore
	id := "550e8400-e29b-41d4-a716-446655440011"
	HandleJob(context.Background(), &st, queue.JobMessage{JobID: id, JobType: "unknown", Parameters: map[string]any{"x": 1}}, nil)
	if len(st.updates) != 2 || st.updates[1][1] != models.JobStatusFailed {
		t.Fatalf("updates: %+v", st.updates)
	}
	if st.dlqWrites != 1 {
		t.Fatalf("want 1 dlq, got %d", st.dlqWrites)
	}
}

func TestHandleJob_runningFails_shortCircuits(t *testing.T) {
	var st spyStore
	st.failRun = true
	HandleJob(context.Background(), &st, queue.JobMessage{JobID: "550e8400-e29b-41d4-a716-446655440012", JobType: "echo"}, nil)
	if len(st.updates) != 1 || st.updates[0][1] != models.JobStatusRunning {
		t.Fatalf("updates: %+v", st.updates)
	}
}

func TestReadBackoff(t *testing.T) {
	if d := ReadBackoff(0); d != 200*time.Millisecond {
		t.Fatalf("n=0: %v", d)
	}
	if d := ReadBackoff(20); d != 30*time.Second {
		t.Fatalf("n=20 should cap: %v", d)
	}
}
