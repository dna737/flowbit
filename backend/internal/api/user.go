package api

import (
	"net/http"

	"flowbit/backend/internal/session"
)

var defaultSessions = session.NewManager()

func (s *Server) withSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}

		userID := s.sessionManager().Ensure(w, r)
		next.ServeHTTP(w, r.WithContext(session.ContextWithUserID(r.Context(), userID)))
	})
}

func (s *Server) requireUserID(w http.ResponseWriter, r *http.Request) (string, bool) {
	if userID, ok := session.UserIDFromContext(r.Context()); ok {
		return userID, true
	}
	return s.sessionManager().Ensure(w, r), true
}

func (s *Server) sessionManager() *session.Manager {
	if s != nil && s.Sessions != nil {
		return s.Sessions
	}
	return defaultSessions
}
