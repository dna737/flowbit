package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"flowbit/backend/internal/models"
	"flowbit/backend/internal/queue"
	"flowbit/backend/internal/repo"
)

type fakeStore struct {
	createJob func(ctx context.Context, jobType string, parameters map[string]any, status string) (models.Job, error)
	getJob    func(ctx context.Context, id string) (models.Job, error)
	update    func(ctx context.Context, id string, status string, lastError *string) error
}

func (f *fakeStore) CreateJob(ctx context.Context, jobType string, parameters map[string]any, status string) (models.Job, error) {
	if f.createJob != nil {
		return f.createJob(ctx, jobType, parameters, status)
	}
	return models.Job{}, errors.New("not implemented")
}

func (f *fakeStore) GetJobByID(ctx context.Context, id string) (models.Job, error) {
	if f.getJob != nil {
		return f.getJob(ctx, id)
	}
	return models.Job{}, repo.ErrJobNotFound
}

func (f *fakeStore) UpdateJobStatus(ctx context.Context, id string, status string, lastError *string) error {
	if f.update != nil {
		return f.update(ctx, id, status, lastError)
	}
	return nil
}

type fakePublisher struct {
	publish func(ctx context.Context, msg queue.JobMessage) error
}

func (f *fakePublisher) PublishJob(ctx context.Context, msg queue.JobMessage) error {
	if f.publish != nil {
		return f.publish(ctx, msg)
	}
	return nil
}

func TestHandleHealthz(t *testing.T) {
	s := &Server{}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	s.HandleHealthz(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d", rr.Code)
	}
}

func TestHandleCreateJob_invalidJSON(t *testing.T) {
	s := &Server{Store: &fakeStore{}, Publisher: &fakePublisher{}}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/jobs", bytes.NewBufferString(`{`))
	s.HandleCreateJob(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("want 400 got %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestHandleCreateJob_missingJobType(t *testing.T) {
	s := &Server{Store: &fakeStore{}, Publisher: &fakePublisher{}}
	body := `{"parameters":{}}`
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/jobs", bytes.NewBufferString(body))
	s.HandleCreateJob(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("want 400 got %d", rr.Code)
	}
}

func TestHandleCreateJob_success(t *testing.T) {
	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	job := models.Job{
		ID:         "550e8400-e29b-41d4-a716-446655440000",
		JobType:    "echo",
		Parameters: map[string]any{"k": "v"},
		Status:     models.JobStatusPending,
		Attempts:   0,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	var published queue.JobMessage
	s := &Server{
		Store: &fakeStore{
			createJob: func(_ context.Context, jobType string, parameters map[string]any, status string) (models.Job, error) {
				if jobType != "echo" || status != models.JobStatusPending {
					t.Fatalf("unexpected create args")
				}
				return job, nil
			},
		},
		Publisher: &fakePublisher{
			publish: func(_ context.Context, msg queue.JobMessage) error {
				published = msg
				return nil
			},
		},
	}
	body := `{"job_type":"echo","parameters":{"k":"v"}}`
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/jobs", bytes.NewBufferString(body))
	s.HandleCreateJob(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("want 201 got %d: %s", rr.Code, rr.Body.String())
	}
	if published.JobID != job.ID || published.JobType != "echo" {
		t.Fatalf("unexpected publish: %+v", published)
	}
}

func TestHandleCreateJob_publishFails_marksFailed(t *testing.T) {
	job := models.Job{ID: "550e8400-e29b-41d4-a716-446655440000", JobType: "echo", Parameters: map[string]any{}, Status: models.JobStatusPending}
	var updatedID, updatedStatus string
	s := &Server{
		Store: &fakeStore{
			createJob: func(context.Context, string, map[string]any, string) (models.Job, error) {
				return job, nil
			},
			update: func(_ context.Context, id string, status string, lastError *string) error {
				updatedID, updatedStatus = id, status
				if lastError == nil || *lastError == "" {
					t.Fatal("expected last_error")
				}
				return nil
			},
		},
		Publisher: &fakePublisher{
			publish: func(context.Context, queue.JobMessage) error {
				return errors.New("broker down")
			},
		},
	}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/jobs", bytes.NewBufferString(`{"job_type":"echo"}`))
	s.HandleCreateJob(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("want 500 got %d", rr.Code)
	}
	if updatedID != job.ID || updatedStatus != models.JobStatusFailed {
		t.Fatalf("update: id=%s status=%s", updatedID, updatedStatus)
	}
}

func TestHandleGetJob_notFound(t *testing.T) {
	s := &Server{
		Store: &fakeStore{
			getJob: func(context.Context, string) (models.Job, error) {
				return models.Job{}, repo.ErrJobNotFound
			},
		},
	}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/jobs/x", nil)
	req.SetPathValue("id", "550e8400-e29b-41d4-a716-446655440001")
	s.HandleGetJob(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("want 404 got %d", rr.Code)
	}
}

func TestHandleGetJob_ok(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	want := models.Job{
		ID:        "550e8400-e29b-41d4-a716-446655440002",
		JobType:   "echo",
		Status:    models.JobStatusSucceeded,
		Attempts:  1,
		CreatedAt: now,
		UpdatedAt: now,
	}
	s := &Server{
		Store: &fakeStore{
			getJob: func(_ context.Context, id string) (models.Job, error) {
				if id != want.ID {
					t.Fatalf("id %q", id)
				}
				return want, nil
			},
		},
	}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/jobs/"+want.ID, nil)
	req.SetPathValue("id", want.ID)
	s.HandleGetJob(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("want 200 got %d: %s", rr.Code, rr.Body.String())
	}
	var got models.Job
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.ID != want.ID || got.Status != want.Status {
		t.Fatalf("got %+v", got)
	}
}
