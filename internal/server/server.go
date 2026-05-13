package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"

	"api-v2ray/internal/app"
	"api-v2ray/internal/openai"
	"api-v2ray/internal/proxyruntime"
	"api-v2ray/internal/router"
	"api-v2ray/internal/upstream"
)

const maxRequestBodyBytes = 32 << 20
const maxAuthBodyBytes = 4096
const maxImportURIBodyBytes = 65536

type Server struct {
	mu             sync.RWMutex
	routerSvc      *router.Service
	proxyRegistry  *proxyruntime.Registry
	upstreamClient *upstream.Client
	configStore    ConfigStore
	bootstrap      *app.BootstrapResult
}

func New(routerSvc *router.Service, proxyRegistry *proxyruntime.Registry, upstreamClient *upstream.Client, extra ...any) *Server {
	s := &Server{routerSvc: routerSvc, proxyRegistry: proxyRegistry, upstreamClient: upstreamClient}
	if len(extra) > 0 {
		if cs, ok := extra[0].(ConfigStore); ok {
			s.configStore = cs
		}
	}
	if len(extra) > 1 {
		if boot, ok := extra[1].(*app.BootstrapResult); ok {
			s.bootstrap = boot
		}
	}
	return s
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/login", s.handleLoginPage)
	mux.HandleFunc("/api/login", s.handleLogin)
	mux.HandleFunc("/api/logout", s.handleLogout)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		s.requireAuth(serveHTML(indexHTML))(w, r)
	})
	mux.HandleFunc("/config", s.requireAuth(serveHTML(configHTML)))
	mux.HandleFunc("/bootstrap", s.requireAuth(serveHTML(bootstrapHTML)))
	mux.HandleFunc("/ui/app.js", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		w.Header().Set("Cache-Control", "public, max-age=3600")
		_, _ = w.Write([]byte(appJS))
	})
	mux.HandleFunc("/ui/styles.css", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
		w.Header().Set("Cache-Control", "public, max-age=3600")
		_, _ = w.Write([]byte(stylesCSS))
	})
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	})
	mux.HandleFunc("/api/config", s.requireAuth(s.handleConfig))
	mux.HandleFunc("/api/config/apply", s.requireAuth(s.handleConfigApply))
	mux.HandleFunc("/api/import/uri", s.requireAuth(s.handleImportURI))
	mux.HandleFunc("/api/import/subscription", s.requireAuth(s.handleImportSubscription))
	mux.HandleFunc("/api/bootstrap", s.requireAuth(s.handleBootstrap))
	mux.HandleFunc("/api/bootstrap/run", s.requireAuth(s.handleBootstrapRun))
	mux.HandleFunc("/api/health/routes", s.requireAuth(s.handleRouteHealth))
	mux.HandleFunc("/api/diagnostics/exit-ip", s.requireAuth(s.handleExitIPProbe))
	mux.HandleFunc("/api/metrics/upstream", s.requireAuth(s.handleUpstreamMetrics))
	mux.HandleFunc("/api/metrics/runtime", s.requireAuth(s.handleRuntimeMetrics))
	mux.HandleFunc("/api/admin/token", s.requireAuth(s.handleAdminToken))
	mux.HandleFunc("/api/restart", s.requireAuth(s.handleRestart))
	mux.HandleFunc("/debug/bootstrap", s.requireAuth(s.handleBootstrap))
	mux.HandleFunc("/v1/models", s.handleModels)
	mux.HandleFunc("/v1/chat/completions", s.handleChatCompletions)
	mux.HandleFunc("/v1/embeddings", s.handleEmbeddings)
	return chain(mux, withRecover, withRequestID, withAccessLog)
}


func (s *Server) handleModels(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		openai.WriteError(w, http.StatusMethodNotAllowed, "invalid_request_error", "method_not_allowed", "method not allowed")
		return
	}
	routerSvc, _, _, _ := s.snapshotState()
	models := routerSvc.Models()
	resp := openai.ModelsResponse{Object: "list"}
	for _, m := range models {
		resp.Data = append(resp.Data, openai.ModelInfo{ID: m, Object: "model", OwnedBy: "api-v2ray"})
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleRouteHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		openai.WriteError(w, http.StatusMethodNotAllowed, "invalid_request_error", "method_not_allowed", "method not allowed")
		return
	}
	routerSvc, _, _, _ := s.snapshotState()
	writeJSON(w, http.StatusOK, map[string]any{"routes": routerSvc.HealthSnapshot()})
}

func (s *Server) handleChatCompletions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		openai.WriteError(w, http.StatusMethodNotAllowed, "invalid_request_error", "method_not_allowed", "method not allowed")
		return
	}
	body, modelName, err := decodeAndNormalizeRequestBody(io.LimitReader(r.Body, maxRequestBodyBytes))
	if err != nil {
		openai.WriteError(w, http.StatusBadRequest, "invalid_request_error", "invalid_json", "invalid json body")
		return
	}
	resp, err := s.tryCandidates(r, modelName, body, true)
	if err != nil {
		openai.WriteError(w, http.StatusBadGateway, "upstream_error", "request_failed", err.Error())
		return
	}
	if err := upstream.CopyResponse(w, resp); err != nil {
		log.Printf("copy upstream response: %v", err)
	}
}

func (s *Server) handleEmbeddings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		openai.WriteError(w, http.StatusMethodNotAllowed, "invalid_request_error", "method_not_allowed", "method not allowed")
		return
	}
	body, modelName, err := decodeAndNormalizeRequestBody(io.LimitReader(r.Body, maxRequestBodyBytes))
	if err != nil {
		openai.WriteError(w, http.StatusBadRequest, "invalid_request_error", "invalid_json", "invalid json body")
		return
	}
	resp, err := s.tryCandidates(r, modelName, body, false)
	if err != nil {
		openai.WriteError(w, http.StatusBadGateway, "upstream_error", "request_failed", err.Error())
		return
	}
	if err := upstream.CopyResponse(w, resp); err != nil {
		log.Printf("copy upstream response: %v", err)
	}
}

func (s *Server) tryCandidates(r *http.Request, modelName string, body []byte, chat bool) (*http.Response, error) {
	routerSvc, proxyRegistry, upstreamClient, _ := s.snapshotState()
	candidates, err := routerSvc.ResolveCandidatesByModel(modelName)
	if err != nil {
		return nil, err
	}
	var failures []string
	for _, c := range candidates {
		endpoint, err := proxyRegistry.Get(c.Binding.NodeID)
		if err != nil {
			errText := "proxy not found: " + err.Error()
			routerSvc.MarkFailure(c, router.FailureProxyRegistryError, errText)
			failures = append(failures, fmt.Sprintf("%s via %s: %s", c.Upstream.ID, c.Node.ID, errText))
			continue
		}

		var resp *http.Response
		if chat {
			resp, err = upstreamClient.ChatCompletionsRaw(r.Context(), r.Header, c.Upstream, *endpoint, body)
		} else {
			resp, err = upstreamClient.EmbeddingsRaw(r.Context(), r.Header, c.Upstream, *endpoint, body)
		}
		if err != nil {
			kind := classifyRequestFailure(err)
			routerSvc.MarkFailure(c, kind, err.Error())
			failures = append(failures, fmt.Sprintf("%s via %s: %v", c.Upstream.ID, c.Node.ID, err))
			continue
		}
		if isRetryableStatus(resp.StatusCode) {
			b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
			_ = resp.Body.Close()
			errText := fmt.Sprintf("upstream status %d: %s", resp.StatusCode, strings.TrimSpace(string(b)))
			routerSvc.MarkFailure(c, classifyStatusFailure(resp.StatusCode), errText)
			failures = append(failures, fmt.Sprintf("%s via %s: %s", c.Upstream.ID, c.Node.ID, errText))
			continue
		}
		if resp.StatusCode >= 400 {
			routerSvc.MarkFailure(c, classifyStatusFailure(resp.StatusCode), fmt.Sprintf("upstream status %d", resp.StatusCode))
			return resp, nil
		}
		routerSvc.MarkSuccess(c)
		return resp, nil
	}
	if len(failures) == 0 {
		return nil, fmt.Errorf("all upstream candidates failed for model %s", modelName)
	}
	return nil, fmt.Errorf("all upstream candidates failed for model %s: %s", modelName, strings.Join(failures, " | "))
}

func isRetryableStatus(code int) bool {
	return code == http.StatusTooManyRequests || code >= 500
}

func classifyStatusFailure(code int) router.FailureKind {
	switch {
	case code == http.StatusTooManyRequests:
		return router.FailureUpstream429
	case code == http.StatusUnauthorized || code == http.StatusForbidden:
		return router.FailureUpstreamAuthError
	case code == http.StatusNotFound:
		return router.FailureModelNotFound
	case code >= 500:
		return router.FailureUpstream5xx
	case code >= 400:
		return router.FailureUpstream4xx
	default:
		return router.FailureUpstreamUnknown
	}
}

func classifyRequestFailure(err error) router.FailureKind {
	if err == nil {
		return router.FailureUpstreamUnknown
	}
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "proxy") || strings.Contains(msg, "socks") {
		return router.FailureProxyConnectError
	}
	if strings.Contains(msg, "timeout") || strings.Contains(msg, "deadline exceeded") {
		return router.FailureUpstreamTimeout
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return router.FailureUpstreamTimeout
	}
	return router.FailureUpstreamUnknown
}

func decodeAndNormalizeRequestBody(body io.Reader) ([]byte, string, error) {
	raw, err := io.ReadAll(body)
	if err != nil {
		return nil, "", err
	}
	if len(raw) == 0 {
		return nil, "", fmt.Errorf("empty request body")
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, "", err
	}
	modelName, _ := payload["model"].(string)
	modelName = strings.TrimSpace(openai.CanonicalModelName(modelName))
	if modelName == "" {
		return nil, "", fmt.Errorf("model is required")
	}
	payload["model"] = modelName
	normalized, err := json.Marshal(payload)
	if err != nil {
		return nil, "", err
	}
	return normalized, modelName, nil
}
