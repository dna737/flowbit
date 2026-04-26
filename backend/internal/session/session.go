package session

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	defaultCookieName = "flowbit_session"
	defaultTTL        = 30 * 24 * time.Hour
)

type contextKey string

const userIDContextKey contextKey = "flowbit.session.user_id"

// Manager mints and persists an anonymous server-issued session identifier.
type Manager struct {
	CookieName string
	TTL        time.Duration
	Now        func() time.Time
}

func NewManager() *Manager {
	return &Manager{
		CookieName: defaultCookieName,
		TTL:        defaultTTL,
		Now:        time.Now,
	}
}

func (m *Manager) Ensure(w http.ResponseWriter, r *http.Request) string {
	if existing := strings.TrimSpace(m.read(r)); existing != "" {
		return existing
	}

	userID := uuid.NewString()
	http.SetCookie(w, &http.Cookie{
		Name:     m.cookieName(),
		Value:    userID,
		Path:     "/",
		HttpOnly: true,
		SameSite: m.sameSite(r),
		Secure:   requestIsHTTPS(r),
		MaxAge:   int(m.ttl().Seconds()),
		Expires:  m.now().Add(m.ttl()),
	})
	return userID
}

func ContextWithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDContextKey, userID)
}

func UserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDContextKey).(string)
	if !ok || strings.TrimSpace(userID) == "" {
		return "", false
	}
	return userID, true
}

func (m *Manager) read(r *http.Request) string {
	cookie, err := r.Cookie(m.cookieName())
	if err != nil {
		return ""
	}
	return cookie.Value
}

func (m *Manager) cookieName() string {
	if strings.TrimSpace(m.CookieName) == "" {
		return defaultCookieName
	}
	return m.CookieName
}

func (m *Manager) ttl() time.Duration {
	if m == nil || m.TTL <= 0 {
		return defaultTTL
	}
	return m.TTL
}

func (m *Manager) now() time.Time {
	if m == nil || m.Now == nil {
		return time.Now().UTC()
	}
	return m.Now().UTC()
}

func (m *Manager) sameSite(r *http.Request) http.SameSite {
	if requestIsHTTPS(r) && isCrossOriginRequest(r) {
		return http.SameSiteNoneMode
	}
	return http.SameSiteLaxMode
}

func requestIsHTTPS(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}
	return strings.EqualFold(strings.TrimSpace(r.Header.Get("X-Forwarded-Proto")), "https")
}

func isCrossOriginRequest(r *http.Request) bool {
	origin := strings.TrimSpace(r.Header.Get("Origin"))
	if origin == "" {
		return false
	}
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	return !strings.EqualFold(u.Host, strings.TrimSpace(r.Host))
}
