package server

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSetSessionCookieSecureFlagFollowsTLS(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "https://example.com", nil)

	setSessionCookie(w, r, "hashed-token")

	res := w.Result()
	defer res.Body.Close()
	cookies := res.Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie, got %d", len(cookies))
	}
	if !cookies[0].Secure {
		t.Fatalf("expected secure cookie when request is tls")
	}
	if cookies[0].HttpOnly != true {
		t.Fatalf("expected httponly cookie")
	}
}

func TestClearSessionCookieExpiresImmediately(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "http://example.com", nil)

	clearSessionCookie(w, r)

	header := w.Header().Get("Set-Cookie")
	if !strings.Contains(header, "Max-Age=0") && !strings.Contains(header, "Max-Age=-1") {
		t.Fatalf("expected cookie max-age to clear session, got: %s", header)
	}
}