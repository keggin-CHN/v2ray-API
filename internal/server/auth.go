package server

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
	"strings"
)

const sessionCookieName = "api_v2ray_session"

var buildInitialAdminToken = ""

func effectiveAdminToken(current string) string {
	current = strings.TrimSpace(current)
	if current != "" {
		return current
	}
	if strings.TrimSpace(buildInitialAdminToken) != "" {
		return strings.TrimSpace(buildInitialAdminToken)
	}
	return "change-me"
}

func tokenHash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func randomHex(n int) string {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "fallback-session-token"
	}
	return hex.EncodeToString(buf)
}

func (s *Server) sessionValue() string {
	cfg, err := s.configStore.Load()
	if err == nil {
		return tokenHash(effectiveAdminToken(cfg.Server.AdminToken))
	}
	return tokenHash(effectiveAdminToken(""))
}

func setSessionCookie(w http.ResponseWriter, r *http.Request, hashedToken string) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    hashedToken,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   r.TLS != nil,
		MaxAge:   86400 * 7,
	})
}

func clearSessionCookie(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   r.TLS != nil,
		MaxAge:   -1,
	})
}

func (s *Server) isAuthenticated(r *http.Request) bool {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil || cookie.Value == "" {
		return false
	}
	expected := s.sessionValue()
	return subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(expected)) == 1
}

func (s *Server) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.configStore.Path == "" {
			next(w, r)
			return
		}
		if s.isAuthenticated(r) {
			next(w, r)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/api/") || strings.HasPrefix(r.URL.Path, "/debug/") {
			writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "admin auth required"})
			return
		}
		http.Redirect(w, r, "/login", http.StatusFound)
	}
}
