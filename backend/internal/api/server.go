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
	"flowbit/backend/internal/session"
)

// JobStore is the persistence surface needed by the HTTP API.
// Reads and writes are scoped to a userID so callers cannot touch other users' jobs.
type JobStore interface {
	CreateJob(ctx context.Context, userID, jobType string, parameters map[string]any, status string) (models.Job, error)
	GetJobByUserAndID(ctx context.Context, userID, id string) (models.Job, error)
	UpdateJobStatus(ctx context.Context, id string, status string, lastError *string) error
}

// JobPublisher publishes a job message after it is persisted.
type JobPublisher interface {
	PublishJob(ctx context.Context, msg queue.JobMessage) error
}

// AIDispatcher translates a plain-English prompt into a structured job payload.
// If nil, POST /dispatch returns 501 Not Implemented.
//
// jobTypes is the user's single list of allowed job_type labels (sourced from
// the user's dispatch_categories Settings list) — the AI must pick exactly one.
type AIDispatcher interface {
	Dispatch(ctx context.Context, prompt string, jobTypes []string) (dispatcher.DispatchResult, error)
}

type Hub interface {
	Register(*realtime.Client)
	Unregister(*realtime.Client)
	BroadcastToUser(userID string, payload []byte)
}

// Server wires HTTP handlers to a store and publisher.
type Server struct {
	Store          JobStore
	Publisher      JobPublisher
	AIDispatcher   AIDispatcher
	Categories     CategoryStore
	Hub            Hub
	Lister         realtime.JobLister
	AllowedOrigins []string
	Sessions       *session.Manager
	// PostgresPing checks database connectivity (e.g. pgxpool.Ping). Used by GET /readyz
	// to wake idle serverless computes. If nil, /readyz returns 503.
	PostgresPing func(context.Context) error
}

type createJobRequest struct {
	JobType    string         `json:"job_type"`
	Parameters map[string]any `json:"parameters"`
}

// Mount registers routes on mux (Go 1.22+ patterns).
func (s *Server) Mount(mux *http.ServeMux) {
	mux.HandleFunc("GET /healthz", s.HandleHealthz)
	mux.HandleFunc("GET /readyz", s.HandleReadyz)
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
    apiHandler := http.StripPrefix("/api", mux)
    mainMux := http.NewServeMux()

    // Main requests go in here
    mainMux.Handle("/api/", apiHandler)

    // This keeps Cloud Run happy
    mainMux.HandleFunc("GET /healthz", s.HandleHealthz)
    return s.withCORS(s.withSession(mainMux))
}

func (s *Server) HandleHealthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) HandleReadyz(w http.ResponseWriter, r *http.Request) {
	if s.PostgresPing == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "postgres ping not configured"})
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	if err := s.PostgresPing(ctx); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "postgres unavailable"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) HandleCreateJob(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.requireUserID(w, r)
	if !ok {
		return
	}

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
	jobType := strings.TrimSpace(req.JobType)
	if jobType == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "job_type is required"})
		return
	}
	if req.Parameters == nil {
		req.Parameters = map[string]any{}
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// Validate job_type against the user's own dispatch_categories list — the
	// single source of truth for what labels this user's jobs may carry. This
	// keeps direct POST /jobs callers and AI-dispatched jobs on the same rail.
	if s.Categories == nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "category store not configured"})
		return
	}
	allowed, err := s.Categories.GetCategories(ctx, userID)
	if err != nil || len(allowed) == 0 {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load allowed job types"})
		return
	}
	canonical, ok := matchAllowed(jobType, allowed)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "job_type not allowed"})
		return
	}
	jobType = canonical

	job, err := s.Store.CreateJob(ctx, userID, jobType, req.Parameters, models.JobStatusPending)
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
	userID, ok := s.requireUserID(w, r)
	if !ok {
		return
	}

	id := strings.TrimSpace(r.PathValue("id"))
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "job id is required"})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	job, err := s.Store.GetJobByUserAndID(ctx, userID, id)
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

// matchAllowed returns the canonical (case-preserved) entry from allowed equal to s,
// case-insensitively, plus ok=true. ok=false when nothing matches.
func matchAllowed(s string, allowed []string) (string, bool) {
	for _, a := range allowed {
		if strings.EqualFold(s, strings.TrimSpace(a)) {
			return a, true
		}
	}
	return "", false
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
		// 1. Always set Vary: Origin to prevent CDN caching issues
		w.Header().Set("Vary", "Origin")

		origin := strings.TrimSpace(r.Header.Get("Origin"))

		// If there is no origin, it's a same-origin or non-browser request.
		if origin == "" {
			next.ServeHTTP(w, r)
			return
		}

		// 2. Validate Origin
		if !slices.Contains(s.AllowedOrigins, origin) {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "origin not allowed"})
			return
		}

		// 3. Set CORS Headers
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true") // Required if you use cookies/sessions later

		// 4. Handle Preflight correctly
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent) // 204 is perfect for OPTIONS
			return
		}

		next.ServeHTTP(w, r)
	})
}
