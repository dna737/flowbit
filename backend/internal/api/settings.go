package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// MaxDispatchCategories is the maximum number of distinct category labels per user.
const MaxDispatchCategories = 10

type categoriesRequest struct {
	Categories []string `json:"categories"`
}

// CategoryStore loads and persists per-user dispatch category labels for Gemini.
type CategoryStore interface {
	GetCategories(ctx context.Context, userID string) ([]string, error)
	SetCategories(ctx context.Context, userID string, categories []string) error
}

// HandleGetDispatchCategories returns the stored category list (may be empty).
func (s *Server) HandleGetDispatchCategories(w http.ResponseWriter, r *http.Request) {
	if s.Categories == nil {
		writeJSON(w, http.StatusNotImplemented, map[string]string{"error": "category store not configured"})
		return
	}
	userID, ok := s.requireUserID(w, r)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	list, err := s.Categories.GetCategories(ctx, userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load categories"})
		return
	}
	if list == nil {
		list = []string{}
	}
	writeJSON(w, http.StatusOK, map[string][]string{"categories": list})
}

// HandlePutDispatchCategories replaces the full category list.
func (s *Server) HandlePutDispatchCategories(w http.ResponseWriter, r *http.Request) {
	if s.Categories == nil {
		writeJSON(w, http.StatusNotImplemented, map[string]string{"error": "category store not configured"})
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req categoriesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}

	normalized := normalizeCategoryList(req.Categories)
	if len(normalized) > MaxDispatchCategories {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": fmt.Sprintf("at most %d categories allowed", MaxDispatchCategories),
		})
		return
	}

	userID, ok := s.requireUserID(w, r)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	if err := s.Categories.SetCategories(ctx, userID, normalized); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to save categories"})
		return
	}

	writeJSON(w, http.StatusOK, map[string][]string{"categories": normalized})
}

func normalizeCategoryList(in []string) []string {
	if len(in) == 0 {
		return []string{}
	}
	seen := make(map[string]struct{})
	out := make([]string, 0, len(in))
	for _, s := range in {
		t := strings.TrimSpace(s)
		if t == "" {
			continue
		}
		key := strings.ToLower(t)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, t)
	}
	return out
}
