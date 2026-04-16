package worker

import (
	"context"
	"errors"
	"testing"
	"time"

	"flowbit/backend/internal/models"
	"flowbit/backend/internal/queue"
)

// spyStore records UpdateJobStatus calls and DLQ writes.
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

// spyPublisher records re-publish calls.
type spyPublisher struct {
	published []queue.JobMessage
	failAt    int // attempt number that triggers a publish error (-1 = never)
}

func newSpy() *spyPublisher { return &spyPublisher{failAt: -1} }

func (p *spyPublisher) PublishJob(_ context.Context, msg queue.JobMessage) error {
	p.published = append(p.published, msg)
	if p.failAt >= 0 && msg.Attempt == p.failAt {
		return errors.New("publish err")
	}
	return nil
}

// --- success path ---

func TestHandleJob_echo_succeeds(t *testing.T) {
	var st spyStore
	pub := newSpy()
	id := "550e8400-e29b-41d4-a716-446655440010"
	HandleJob(context.Background(), &st, pub, queue.JobMessage{JobID: id, JobType: "echo", Parameters: map[string]any{}}, nil)
	if len(st.updates) != 2 {
		t.Fatalf("updates: %+v", st.updates)
	}
	if st.updates[0][1] != models.JobStatusRunning || st.updates[1][1] != models.JobStatusSucceeded {
		t.Fatalf("want running then succeeded: %+v", st.updates)
	}
	if st.dlqWrites != 0 {
		t.Fatalf("dlq writes: %d", st.dlqWrites)
	}
	if len(pub.published) != 0 {
		t.Fatalf("no re-publishes expected: %+v", pub.published)
	}
}

// --- retry path ---

func TestHandleJob_fail_firstAttempt_retries(t *testing.T) {
	var st spyStore
	pub := newSpy()
	id := "550e8400-e29b-41d4-a716-446655440020"
	HandleJob(context.Background(), &st, pub,
		queue.JobMessage{JobID: id, JobType: "fail", Attempt: 0}, nil)

	// running → retrying
	if len(st.updates) != 2 {
		t.Fatalf("updates: %+v", st.updates)
	}
	if st.updates[0][1] != models.JobStatusRunning {
		t.Fatalf("want running first, got %q", st.updates[0][1])
	}
	if st.updates[1][1] != models.JobStatusRetrying {
		t.Fatalf("want retrying second, got %q", st.updates[1][1])
	}
	if st.dlqWrites != 0 {
		t.Fatalf("no DLQ expected yet, got %d", st.dlqWrites)
	}
	// re-published with attempt incremented
	if len(pub.published) != 1 || pub.published[0].Attempt != 1 {
		t.Fatalf("want re-publish attempt=1, got %+v", pub.published)
	}
}

func TestHandleJob_fail_secondAttempt_retries(t *testing.T) {
	var st spyStore
	pub := newSpy()
	id := "550e8400-e29b-41d4-a716-446655440021"
	HandleJob(context.Background(), &st, pub,
		queue.JobMessage{JobID: id, JobType: "fail", Attempt: 1}, nil)

	if st.updates[1][1] != models.JobStatusRetrying {
		t.Fatalf("want retrying, got %+v", st.updates)
	}
	if len(pub.published) != 1 || pub.published[0].Attempt != 2 {
		t.Fatalf("want re-publish attempt=2, got %+v", pub.published)
	}
	if st.dlqWrites != 0 {
		t.Fatalf("no DLQ expected yet")
	}
}

// --- DLQ path ---

func TestHandleJob_fail_finalAttempt_DLQ(t *testing.T) {
	var st spyStore
	pub := newSpy()
	id := "550e8400-e29b-41d4-a716-446655440022"
	HandleJob(context.Background(), &st, pub,
		queue.JobMessage{JobID: id, JobType: "fail", Attempt: maxAttempts - 1}, nil)

	// running → failed
	if len(st.updates) != 2 {
		t.Fatalf("updates: %+v", st.updates)
	}
	if st.updates[1][1] != models.JobStatusFailed {
		t.Fatalf("want failed, got %q", st.updates[1][1])
	}
	if st.dlqWrites != 1 {
		t.Fatalf("want 1 DLQ write, got %d", st.dlqWrites)
	}
	if len(pub.published) != 0 {
		t.Fatalf("no re-publish expected on final attempt")
	}
}

func TestHandleJob_unsupported_finalAttempt_DLQ(t *testing.T) {
	var st spyStore
	pub := newSpy()
	id := "550e8400-e29b-41d4-a716-446655440011"
	HandleJob(context.Background(), &st, pub,
		queue.JobMessage{JobID: id, JobType: "unknown", Parameters: map[string]any{"x": 1}, Attempt: maxAttempts - 1}, nil)
	if len(st.updates) != 2 || st.updates[1][1] != models.JobStatusFailed {
		t.Fatalf("updates: %+v", st.updates)
	}
	if st.dlqWrites != 1 {
		t.Fatalf("want 1 dlq, got %d", st.dlqWrites)
	}
}

// --- error path: mark-running fails ---

func TestHandleJob_runningFails_shortCircuits(t *testing.T) {
	var st spyStore
	st.failRun = true
	pub := newSpy()
	HandleJob(context.Background(), &st, pub,
		queue.JobMessage{JobID: "550e8400-e29b-41d4-a716-446655440012", JobType: "echo"}, nil)
	if len(st.updates) != 1 || st.updates[0][1] != models.JobStatusRunning {
		t.Fatalf("updates: %+v", st.updates)
	}
	if len(pub.published) != 0 {
		t.Fatal("no re-publish expected when mark-running fails")
	}
}

// --- backoff helper ---

func TestReadBackoff(t *testing.T) {
	if d := ReadBackoff(0); d != 200*time.Millisecond {
		t.Fatalf("n=0: %v", d)
	}
	if d := ReadBackoff(20); d != 30*time.Second {
		t.Fatalf("n=20 should cap: %v", d)
	}
}
