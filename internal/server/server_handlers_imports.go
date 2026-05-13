package server

import (
	"net/http"
	"strings"

	"api-v2ray/internal/model"
	"api-v2ray/internal/openai"
)

func (s *Server) handleImportURI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		openai.WriteError(w, http.StatusMethodNotAllowed, "invalid_request_error", "method_not_allowed", "method not allowed")
		return
	}
	var req ImportURIRequest
	if err := decodeJSONBody(r, maxImportURIBodyBytes, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid json body"})
		return
	}
	if strings.TrimSpace(req.RawURI) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "raw_uri must not be empty"})
		return
	}
	node, err := buildNodeFromURI(req.RawURI)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, ImportPreviewResponse{Nodes: []model.ProxyNode{node}})
}

func (s *Server) handleImportSubscription(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		openai.WriteError(w, http.StatusMethodNotAllowed, "invalid_request_error", "method_not_allowed", "method not allowed")
		return
	}
	var req ImportSubscriptionRequest
	if err := decodeJSONBody(r, maxRequestBodyBytes, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid json body"})
		return
	}
	if strings.TrimSpace(req.URL) == "" && strings.TrimSpace(req.Text) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "url or text must not be empty"})
		return
	}
	nodes, format, err := s.previewSubscription(r.Context(), req)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, ImportPreviewResponse{Format: format, Nodes: nodes})
}