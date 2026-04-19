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
	"flowbit/backend/internal/realtime"
	"flowbit/backend/internal/repo"
)

type fakeStore struct {
	createJob func(ctx context.Context, userID, jobType string, parameters map[string]any, status string) (models.Job, error)
	getJob    func(ctx context.Context, userID, id string) (models.Job, error)
	update    func(ctx context.Context, id string, status string, lastError *string) error
}

func (f *fakeStore) CreateJob(ctx context.Context, userID, jobType string, parameters map[string]any, status string) (models.Job, error) {
	if f.createJob != nil {
		return f.createJob(ctx, userID, jobType, parameters, status)
	}
	return models.Job{}, errors.New("not implemented")
}

func (f *fakeStore) GetJobByUserAndID(ctx context.Context, userID, id string) (models.Job, error) {
	if f.getJob != nil {
		return f.getJob(ctx, userID, id)
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

type fakeLister struct {
	listJobs func(ctx context.Context, userID string, limit int) ([]models.Job, error)
}

func (f *fakeLister) ListJobsByUser(ctx context.Context, userID string, limit int) ([]models.Job, error) {
	if f.listJobs != nil {
		return f.listJobs(ctx, userID, limit)
	}
	return nil, nil
}

type fakeHub struct{}

func (fakeHub) Register(*realtime.Client)        {}
func (fakeHub) Unregister(*realtime.Client)      {}
func (fakeHub) BroadcastToUser(string, []byte)   {}

func TestHandleHealthz(t *testing.T) {
	s := &Server{}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	s.HandleHealthz(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d", rr.Code)
	}
}

func TestHandleCreateJob_missingUserID(t *testing.T) {
	s := &Server{Store: &fakeStore{}, Publisher: &fakePublisher{}, JobTypes: &fakeJobTypes{}}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/jobs", bytes.NewBufferString(`{"job_type":"echo"}`))
	s.HandleCreateJob(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("want 400 got %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestHandleCreateJob_invalidJSON(t *testing.T) {
	s := &Server{Store: &fakeStore{}, Publisher: &fakePublisher{}, JobTypes: &fakeJobTypes{}}
	rr := httptest.NewRecorder()
	req := withUserID(httptest.NewRequest(http.MethodPost, "/jobs", bytes.NewBufferString(`{`)))
	s.HandleCreateJob(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("want 400 got %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestHandleCreateJob_missingJobType(t *testing.T) {
	s := &Server{Store: &fakeStore{}, Publisher: &fakePublisher{}, JobTypes: &fakeJobTypes{}}
	body := `{"parameters":{}}`
	rr := httptest.NewRecorder()
	req := withUserID(httptest.NewRequest(http.MethodPost, "/jobs", bytes.NewBufferString(body)))
	s.HandleCreateJob(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("want 400 got %d", rr.Code)
	}
}

func TestHandleCreateJob_disallowedJobType(t *testing.T) {
	s := &Server{Store: &fakeStore{}, Publisher: &fakePublisher{}, JobTypes: &fakeJobTypes{}}
	rr := httptest.NewRecorder()
	req := withUserID(httptest.NewRequest(http.MethodPost, "/jobs", bytes.NewBufferString(`{"job_type":"rm-rf"}`)))
	s.HandleCreateJob(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("want 400 got %d body=%s", rr.Code, rr.Body.String())
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
			createJob: func(_ context.Context, userID, jobType string, parameters map[string]any, status string) (models.Job, error) {
				if userID != "test-user" {
					t.Fatalf("unexpected userID %q", userID)
				}
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
		JobTypes: &fakeJobTypes{},
	}
	body := `{"job_type":"echo","parameters":{"k":"v"}}`
	rr := httptest.NewRecorder()
	req := withUserID(httptest.NewRequest(http.MethodPost, "/jobs", bytes.NewBufferString(body)))
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
			createJob: func(context.Context, string, string, map[string]any, string) (models.Job, error) {
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
		JobTypes: &fakeJobTypes{},
	}
	rr := httptest.NewRecorder()
	req := withUserID(httptest.NewRequest(http.MethodPost, "/jobs", bytes.NewBufferString(`{"job_type":"echo"}`)))
	s.HandleCreateJob(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("want 500 got %d", rr.Code)
	}
	if updatedID != job.ID || updatedStatus != models.JobStatusFailed {
		t.Fatalf("update: id=%s status=%s", updatedID, updatedStatus)
	}
}

func TestHandleGetJob_missingUserID(t *testing.T) {
	s := &Server{Store: &fakeStore{}}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/jobs/x", nil)
	req.SetPathValue("id", "550e8400-e29b-41d4-a716-446655440001")
	s.HandleGetJob(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("want 400 got %d", rr.Code)
	}
}

func TestHandleGetJob_notFound(t *testing.T) {
	s := &Server{
		Store: &fakeStore{
			getJob: func(context.Context, string, string) (models.Job, error) {
				return models.Job{}, repo.ErrJobNotFound
			},
		},
	}
	rr := httptest.NewRecorder()
	req := withUserID(httptest.NewRequest(http.MethodGet, "/jobs/x", nil))
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
			getJob: func(_ context.Context, userID, id string) (models.Job, error) {
				if userID != "test-user" {
					t.Fatalf("userID %q", userID)
				}
				if id != want.ID {
					t.Fatalf("id %q", id)
				}
				return want, nil
			},
		},
	}
	rr := httptest.NewRecorder()
	req := withUserID(httptest.NewRequest(http.MethodGet, "/jobs/"+want.ID, nil))
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

func TestServerHandler_blocksDisallowedOrigin(t *testing.T) {
	s := &Server{
		AllowedOrigins: []string{"http://localhost:5173"},
	}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set("Origin", "https://evil.example.com")

	s.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("want 403 got %d", rr.Code)
	}
}

func TestServerHandler_setsCORSHeadersForAllowedOrigin(t *testing.T) {
	s := &Server{
		AllowedOrigins: []string{"http://localhost:5173"},
	}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set("Origin", "http://localhost:5173")

	s.Handler().ServeHTTP(rr, req)

	if rr.Header().Get("Access-Control-Allow-Origin") != "http://localhost:5173" {
		t.Fatalf("unexpected allow origin header: %q", rr.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestServerHandler_handlesPreflight(t *testing.T) {
	s := &Server{
		AllowedOrigins: []string{"http://localhost:5173"},
	}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/jobs", nil)
	req.Header.Set("Origin", "http://localhost:5173")

	s.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("want 204 got %d", rr.Code)
	}
}

func TestServerMount_omitsWebSocketRouteWithoutRealtimeDeps(t *testing.T) {
	s := &Server{}
	mux := http.NewServeMux()
	s.Mount(mux)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("want 404 got %d", rr.Code)
	}
}

func TestServerMount_registersWebSocketRouteWithRealtimeDeps(t *testing.T) {
	s := &Server{
		Hub:            fakeHub{},
		Lister:         &fakeLister{},
		AllowedOrigins: []string{"http://localhost:5173"},
	}
	mux := http.NewServeMux()
	s.Mount(mux)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Sec-WebSocket-Version", "13")
	req.Header.Set("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==")
	mux.ServeHTTP(rr, req)

	if rr.Code == http.StatusNotFound {
		t.Fatal("expected websocket route to be registered")
	}
}
