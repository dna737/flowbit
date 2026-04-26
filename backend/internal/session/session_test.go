package session

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestEnsure_reusesExistingCookie(t *testing.T) {
	m := NewManager()
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/jobs", nil)
	req.AddCookie(&http.Cookie{Name: defaultCookieName, Value: "user-123"})

	got := m.Ensure(rr, req)

	if got != "user-123" {
		t.Fatalf("want existing cookie value, got %q", got)
	}
	if cookies := rr.Result().Cookies(); len(cookies) != 0 {
		t.Fatalf("did not expect a new cookie, got %d", len(cookies))
	}
}

func TestEnsure_setsCookieWhenMissing(t *testing.T) {
	now := time.Date(2026, 4, 23, 12, 0, 0, 0, time.UTC)
	m := &Manager{Now: func() time.Time { return now }}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/jobs", nil)

	got := m.Ensure(rr, req)

	cookies := rr.Result().Cookies()
	if got == "" {
		t.Fatal("expected generated session id")
	}
	if len(cookies) != 1 {
		t.Fatalf("want 1 cookie got %d", len(cookies))
	}
	if cookies[0].Name != defaultCookieName {
		t.Fatalf("unexpected cookie name %q", cookies[0].Name)
	}
	if cookies[0].Value != got {
		t.Fatalf("cookie value %q does not match user id %q", cookies[0].Value, got)
	}
	if !cookies[0].HttpOnly {
		t.Fatal("expected HttpOnly cookie")
	}
	if cookies[0].SameSite != http.SameSiteLaxMode {
		t.Fatalf("unexpected SameSite %v", cookies[0].SameSite)
	}
}

func TestEnsure_usesSameSiteNoneForSecureCrossOriginRequests(t *testing.T) {
	m := NewManager()
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "https://api.example.com/jobs", nil)
	req.Host = "api.example.com"
	req.Header.Set("Origin", "https://app.example.net")
	req.Header.Set("X-Forwarded-Proto", "https")

	_ = m.Ensure(rr, req)

	cookies := rr.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("want 1 cookie got %d", len(cookies))
	}
	if cookies[0].SameSite != http.SameSiteNoneMode {
		t.Fatalf("unexpected SameSite %v", cookies[0].SameSite)
	}
	if !cookies[0].Secure {
		t.Fatal("expected Secure cookie")
	}
	if !strings.EqualFold(cookies[0].Path, "/") {
		t.Fatalf("unexpected cookie path %q", cookies[0].Path)
	}
}
