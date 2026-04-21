package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"flowbit/backend/internal/models"
	"flowbit/backend/internal/queue"
)

type dispatchRequest struct {
	Prompt string `json:"prompt"`
}

// HandleDispatch accepts a plain-English prompt, calls the AI dispatcher to
// extract a job_type + parameters, then enqueues the job via the normal path.
//
// The user's dispatch_categories list is the single source of truth for the
// allowed job_type labels passed to Gemini — the same list the Settings
// dialog edits.
func (s *Server) HandleDispatch(w http.ResponseWriter, r *http.Request) {
	if s.AIDispatcher == nil {
		writeJSON(w, http.StatusNotImplemented, map[string]string{"error": "ai dispatcher not configured"})
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req dispatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}
	if strings.TrimSpace(req.Prompt) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "prompt is required"})
		return
	}

	userID, ok := RequireUserID(w, r)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	if s.Categories == nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "category store not configured"})
		return
	}
	rawJobTypes, err := s.Categories.GetCategories(ctx, userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load allowed job types"})
		return
	}
	jobTypes := normalizeCategoryList(rawJobTypes)
	if len(jobTypes) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "no job type labels configured — open Settings to add at least one"})
		return
	}

	result, err := s.AIDispatcher.Dispatch(ctx, req.Prompt, jobTypes)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "ai dispatch failed: " + err.Error()})
		return
	}

	if result.Parameters == nil {
		result.Parameters = map[string]any{}
	}

	job, err := s.Store.CreateJob(ctx, userID, result.JobType, result.Parameters, models.JobStatusPending)
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
