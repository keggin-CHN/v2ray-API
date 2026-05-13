package subscription

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"api-v2ray/internal/model"
)

func TestFetchRejectsNon2xxStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("forbidden"))
	}))
	defer srv.Close()

	s := New()
	_, err := s.Fetch(context.Background(), model.Subscription{ID: "sub1", URL: srv.URL})
	if err == nil || !strings.Contains(err.Error(), "status 403") {
		t.Fatalf("expected status error, got: %v", err)
	}
}

func TestFetchParsesURIOn2xx(t *testing.T) {
	line := "socks5://127.0.0.1:1080#local-node"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(line))
	}))
	defer srv.Close()

	s := New()
	nodes, err := s.Fetch(context.Background(), model.Subscription{ID: "sub2", URL: srv.URL})
	if err != nil {
		t.Fatalf("unexpected fetch err: %v", err)
	}
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	if nodes[0].Scheme != "socks" {
		t.Fatalf("expected normalized scheme socks, got: %s", nodes[0].Scheme)
	}
}