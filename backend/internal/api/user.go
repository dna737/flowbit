package api

import (
	"net/http"
	"strings"
)

// RequireUserID returns the X-User-Id value or writes 400 and false if missing/blank.
func RequireUserID(w http.ResponseWriter, r *http.Request) (string, bool) {
	u := strings.TrimSpace(r.Header.Get("X-User-Id"))
	if u == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "X-User-Id header is required"})
		return "", false
	}
	return u, true
}
