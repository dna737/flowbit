package api

import (
	"context"
	"net/http"
	"strings"
	"time"

	"flowbit/backend/internal/auth"
)

func (s *Server) withAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}
		if isPublicAPIPath(r) {
			next.ServeHTTP(w, r)
			return
		}
		if _, ok := auth.ClaimsFromContext(r.Context()); ok {
			next.ServeHTTP(w, r)
			return
		}
		if s.Auth == nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "auth not configured"})
			return
		}

		claims, err := s.Auth.Verify(r.Context(), auth.TokenFromRequest(r))
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		next.ServeHTTP(w, r.WithContext(auth.ContextWithClaims(r.Context(), claims)))
	})
}

func (s *Server) requireUserID(w http.ResponseWriter, r *http.Request) (string, bool) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return "", false
	}
	if s.Users != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		if err := s.Users.UpsertUser(ctx, claims); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to sync user"})
			return "", false
		}
	}
	return claims.Subject, true
}

func isPublicAPIPath(r *http.Request) bool {
	path := strings.TrimRight(r.URL.Path, "/")
	switch path {
	case "/api/healthz", "/api/readyz", "/healthz", "/readyz":
		return true
	default:
		return false
	}
}
