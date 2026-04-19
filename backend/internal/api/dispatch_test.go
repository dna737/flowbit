package api

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"flowbit/backend/internal/dispatcher"
	"flowbit/backend/internal/models"
)

type fakeDispatcher struct {
	result dispatcher.DispatchResult
	err    error
}

func (f *fakeDispatcher) Dispatch(_ context.Context, _ string, _ []string, _ []string) (dispatcher.DispatchResult, error) {
	return f.result, f.err
}

type fakeJobTypes struct{}

func (*fakeJobTypes) GetAllowedJobTypes(context.Context) ([]string, error) {
	return []string{"echo", "email", "image_resize", "url_scrape", "fail"}, nil
}

type fakeCategoryStore struct{}

func (f *fakeCategoryStore) GetCategories(_ context.Context, _ string) ([]string, error) {
	return []string{}, nil
}

func (f *fakeCategoryStore) SetCategories(_ context.Context, _ string, _ []string) error {
	return nil
}

func withUserID(req *http.Request) *http.Request {
	req.Header.Set("X-User-Id", "test-user")
	return req
}

func TestHandleDispatch_nilDispatcher(t *testing.T) {
	s := &Server{Store: &fakeStore{}, Publisher: &fakePublisher{}}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/dispatch", bytes.NewBufferString(`{"prompt":"anything"}`))
	s.HandleDispatch(rr, req)
	if rr.Code != http.StatusNotImplemented {
		t.Fatalf("want 501 got %d", rr.Code)
	}
}

func TestHandleDispatch_emptyPrompt(t *testing.T) {
	s := &Server{
		Store:        &fakeStore{},
		Publisher:    &fakePublisher{},
		AIDispatcher: &fakeDispatcher{},
		Categories:   &fakeCategoryStore{},
		JobTypes:     &fakeJobTypes{},
	}
	rr := httptest.NewRecorder()
	req := withUserID(httptest.NewRequest(http.MethodPost, "/dispatch", bytes.NewBufferString(`{"prompt":"  "}`)))
	s.HandleDispatch(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("want 400 got %d", rr.Code)
	}
}

func TestHandleDispatch_dispatcherError(t *testing.T) {
	s := &Server{
		Store:     &fakeStore{},
		Publisher: &fakePublisher{},
		AIDispatcher: &fakeDispatcher{
			err: errors.New("api down"),
		},
		Categories: &fakeCategoryStore{},
		JobTypes:   &fakeJobTypes{},
	}
	rr := httptest.NewRecorder()
	req := withUserID(httptest.NewRequest(http.MethodPost, "/dispatch", bytes.NewBufferString(`{"prompt":"send email"}`)))
	s.HandleDispatch(rr, req)
	if rr.Code != http.StatusBadGateway {
		t.Fatalf("want 502 got %d", rr.Code)
	}
}

func TestHandleDispatch_success(t *testing.T) {
	now := time.Now().UTC()
	job := models.Job{
		ID:         "550e8400-e29b-41d4-a716-446655440000",
		JobType:    "email",
		Parameters: map[string]any{"to": "bob@example.com"},
		Status:     models.JobStatusPending,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	s := &Server{
		Store: &fakeStore{
			createJob: func(_ context.Context, jobType string, _ map[string]any, _ string) (models.Job, error) {
				if jobType != "email" {
					t.Fatalf("unexpected job_type: %s", jobType)
				}
				return job, nil
			},
		},
		Publisher: &fakePublisher{},
		AIDispatcher: &fakeDispatcher{
			result: dispatcher.DispatchResult{
				JobType:    "email",
				Parameters: map[string]any{"to": "bob@example.com"},
			},
		},
		Categories: &fakeCategoryStore{},
		JobTypes:   &fakeJobTypes{},
	}
	rr := httptest.NewRecorder()
	req := withUserID(httptest.NewRequest(http.MethodPost, "/dispatch", bytes.NewBufferString(`{"prompt":"send email to bob@example.com"}`)))
	s.HandleDispatch(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("want 201 got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandleDispatch_invalidJSON(t *testing.T) {
	s := &Server{
		Store:        &fakeStore{},
		Publisher:    &fakePublisher{},
		AIDispatcher: &fakeDispatcher{},
	}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/dispatch", bytes.NewBufferString(`{`))
	s.HandleDispatch(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("want 400 got %d", rr.Code)
	}
}
