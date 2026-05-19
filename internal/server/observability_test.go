package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWithRequestID_GeneratesWhenMissing(t *testing.T) {
	h := withRequestID(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("unexpected status: %d", rec.Code)
	}
	if strings.TrimSpace(rec.Header().Get("X-Request-ID")) == "" {
		t.Fatalf("expected X-Request-ID header")
	}
}

func TestWithRequestID_PreservesIncomingHeader(t *testing.T) {
	h := withRequestID(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set("X-Request-ID", "req-123")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if got := rec.Header().Get("X-Request-ID"); got != "req-123" {
		t.Fatalf("expected request id preserved, got: %s", got)
	}
}

func TestStatusRecorderPreservesFirstStatus(t *testing.T) {
	rec := httptest.NewRecorder()
	wrapped := &statusRecorder{ResponseWriter: rec, status: http.StatusOK}

	wrapped.WriteHeader(http.StatusAccepted)
	wrapped.WriteHeader(http.StatusInternalServerError)

	if wrapped.status != http.StatusAccepted {
		t.Fatalf("expected first status preserved, got %d", wrapped.status)
	}
	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected response status 202, got %d", rec.Code)
	}
}

func TestWithRecover_Returns500OnPanic(t *testing.T) {
	h := withRecover(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		panic("boom")
	}))
	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body["error"] != "internal server error" {
		t.Fatalf("unexpected error body: %+v", body)
	}
}

func TestDecodeAndNormalizeRequestBodyRejectsMultipleJSONValues(t *testing.T) {
	_, _, err := decodeAndNormalizeRequestBody(strings.NewReader(`{"model":"gpt-5.5"} {"model":"gpt-4"}`))
	if err == nil || !strings.Contains(err.Error(), "multiple json values") {
		t.Fatalf("expected multiple json values error, got: %v", err)
	}
}

func TestFetchIP_UsesFallbackURL(t *testing.T) {
	bad := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("failed"))
	}))
	defer bad.Close()

	good := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("1.2.3.4"))
	}))
	defer good.Close()

	ip, err := fetchIP(nil, bad.URL, good.URL)
	if err != nil {
		t.Fatalf("expected fallback success, got err: %v", err)
	}
	if ip != "1.2.3.4" {
		t.Fatalf("unexpected ip: %s", ip)
	}
}