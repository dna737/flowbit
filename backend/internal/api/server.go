package api

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"slices"
	"strings"
	"time"

	"flowbit/backend/internal/dispatcher"
	"flowbit/backend/internal/models"
	"flowbit/backend/internal/queue"
	"flowbit/backend/internal/realtime"
	"flowbit/backend/internal/repo"
)

// JobStore is the persistence surface needed by the HTTP API.
type JobStore interface {
	CreateJob(ctx context.Context, jobType string, parameters map[string]any, status string) (models.Job, error)
	GetJobByID(ctx context.Context, id string) (models.Job, error)
	UpdateJobStatus(ctx context.Context, id string, status string, lastError *string) error
}

// JobPublisher publishes a job message after it is persisted.
type JobPublisher interface {
	PublishJob(ctx context.Context, msg queue.JobMessage) error
}

// AIDispatcher translates a plain-English prompt into a structured job payload.
// If nil, POST /dispatch returns 501 Not Implemented.
type AIDispatcher interface {
	Dispatch(ctx context.Context, prompt string, categories []string, jobTypes []string) (dispatcher.DispatchResult, error)
}

// AllowedJobTypesSource returns the canonical job_type strings allowed for AI dispatch (from dispatcher_config).
type AllowedJobTypesSource interface {
	GetAllowedJobTypes(ctx context.Context) ([]string, error)
}

type Hub interface {
	Register(*realtime.Client)
	Unregister(*realtime.Client)
	Broadcast([]byte)
}

// Server wires HTTP handlers to a store and publisher.
type Server struct {
	Store          JobStore
	Publisher      JobPublisher
	AIDispatcher   AIDispatcher
	Categories     CategoryStore
	JobTypes       AllowedJobTypesSource
	Hub            Hub
	Lister         realtime.JobLister
	AllowedOrigins []string
}

type createJobRequest struct {
	JobType    string         `json:"job_type"`
	Parameters map[string]any `json:"parameters"`
}

// Mount registers routes on mux (Go 1.22+ patterns).
func (s *Server) Mount(mux *http.ServeMux) {
	mux.HandleFunc("GET /healthz", s.HandleHealthz)
	mux.HandleFunc("POST /jobs", s.HandleCreateJob)
	mux.HandleFunc("GET /jobs/{id}", s.HandleGetJob)
	mux.HandleFunc("POST /dispatch", s.HandleDispatch)
	mux.HandleFunc("GET /settings/dispatch-categories", s.HandleGetDispatchCategories)
	mux.HandleFunc("PUT /settings/dispatch-categories", s.HandlePutDispatchCategories)
	if s.Hub != nil && s.Lister != nil {
		mux.Handle("GET /ws", realtime.Handler(s.Hub, s.Lister, s.AllowedOrigins))
	}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	s.Mount(mux)
	return s.withCORS(mux)
}

func (s *Server) HandleHealthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) HandleCreateJob(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req createJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			writeJSON(w, http.StatusRequestEntityTooLarge, map[string]string{"error": "request body too large"})
			return
		}
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}
	if strings.TrimSpace(req.JobType) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "job_type is required"})
		return
	}
	if req.Parameters == nil {
		req.Parameters = map[string]any{}
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	job, err := s.Store.CreateJob(ctx, req.JobType, req.Parameters, models.JobStatusPending)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to persist job"})
		return
	}

	msg := queue.JobMessage{
		JobID:      job.ID,
		JobType:    job.JobType,
		Parameters: job.Parameters,
	}
	if err := s.Publisher.PublishJob(ctx, msg); err != nil {
		failMsg := "kafka publish failed: " + err.Error()
		_ = s.Store.UpdateJobStatus(ctx, job.ID, models.JobStatusFailed, &failMsg)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to publish job"})
		return
	}

	writeJSON(w, http.StatusCreated, job)
}

func (s *Server) HandleGetJob(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("id"))
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "job id is required"})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	job, err := s.Store.GetJobByID(ctx, id)
	if err != nil {
		if errors.Is(err, repo.ErrJobNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "job not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch job"})
		return
	}
	writeJSON(w, http.StatusOK, job)
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Printf("json encode error: %v", err)
	}
}

func (s *Server) withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := strings.TrimSpace(r.Header.Get("Origin"))
		if origin != "" && slices.Contains(s.AllowedOrigins, origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-User-Id")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
