package testutil

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
)

func NewMockOpenAIServer() *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/chat/completions", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id": "chatcmpl-mock",
			"object": "chat.completion",
			"choices": []any{map[string]any{
				"index": 0,
				"message": map[string]any{"role": "assistant", "content": "ok"},
			}},
		})
	})
	mux.HandleFunc("/v1/embeddings", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"object": "list",
			"data": []any{map[string]any{"embedding": []float64{0.1, 0.2}, "index": 0}},
		})
	})
	return httptest.NewServer(mux)
}
