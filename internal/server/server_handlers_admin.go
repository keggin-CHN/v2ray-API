package server

import (
	"context"
	"net/http"
	"strings"

	"api-v2ray/internal/app"
	"api-v2ray/internal/openai"
	appruntime "api-v2ray/internal/runtime"
)

func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		cfg, err := s.configStore.Load()
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		cfg.Server.AdminToken = ""
		writeJSON(w, http.StatusOK, ConfigResponse{Path: s.configStore.Path, Config: *cfg})
	case http.MethodPost:
		var req ConfigUpdateRequest
		if err := decodeJSONBody(r, maxRequestBodyBytes, &req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid json body"})
			return
		}
		oldCfg, err := s.configStore.Load()
		if err == nil && strings.TrimSpace(req.Config.Server.AdminToken) == "" {
			req.Config.Server.AdminToken = oldCfg.Server.AdminToken
		}
		if err := s.configStore.Save(&req.Config); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		masked := req.Config
		masked.Server.AdminToken = ""
		writeJSON(w, http.StatusOK, ConfigResponse{Path: s.configStore.Path, Config: masked})
	default:
		openai.WriteError(w, http.StatusMethodNotAllowed, "invalid_request_error", "method_not_allowed", "method not allowed")
	}
}

func (s *Server) handleAdminToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		openai.WriteError(w, http.StatusMethodNotAllowed, "invalid_request_error", "method_not_allowed", "method not allowed")
		return
	}
	var req struct {
		Token string `json:"token"`
	}
	if err := decodeJSONBody(r, maxAuthBodyBytes, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid json body"})
		return
	}
	if strings.TrimSpace(req.Token) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "token must not be empty"})
		return
	}
	cfg, err := s.configStore.Load()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	cfg.Server.AdminToken = strings.TrimSpace(req.Token)
	if err := s.configStore.Save(cfg); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	setSessionCookie(w, r, tokenHash(cfg.Server.AdminToken))
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handleConfigApply(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		openai.WriteError(w, http.StatusMethodNotAllowed, "invalid_request_error", "method_not_allowed", "method not allowed")
		return
	}
	var req ConfigUpdateRequest
	if err := decodeJSONBody(r, maxRequestBodyBytes, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid json body"})
		return
	}

	oldCfg, err := s.configStore.Load()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	if strings.TrimSpace(req.Config.Server.AdminToken) == "" {
		req.Config.Server.AdminToken = oldCfg.Server.AdminToken
	}

	masked := req.Config
	masked.Server.AdminToken = ""

	boot, err := app.Bootstrap(r.Context(), &req.Config)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, ApplyConfigResponse{
			Path:   s.configStore.Path,
			Config: masked,
			Result: boot,
			Error:  "dry-run bootstrap failed: " + err.Error(),
		})
		return
	}

	liveCfg := req.Config
	if boot != nil {
		liveCfg.ProxyNodes = boot.FlatResult.Nodes
	}
	if startErr := appruntime.StartXrayProcesses(&liveCfg); startErr != nil {
		writeJSON(w, http.StatusBadGateway, ApplyConfigResponse{
			Path:   s.configStore.Path,
			Config: masked,
			Result: boot,
			Error:  "runtime launch failed: " + startErr.Error(),
		})
		return
	}

	if err := s.configStore.Save(&req.Config); err != nil {
		_ = appruntime.StartXrayProcesses(oldCfg)
		s.applyLiveConfig(oldCfg, s.bootstrap)
		writeJSON(w, http.StatusInternalServerError, ApplyConfigResponse{
			Path:   s.configStore.Path,
			Config: masked,
			Result: boot,
			Error:  "save config failed and rollback attempted: " + err.Error(),
		})
		return
	}

	s.applyLiveConfig(&liveCfg, boot)
	writeJSON(w, http.StatusOK, ApplyConfigResponse{Path: s.configStore.Path, Config: masked, Result: boot})
}

func (s *Server) handleBootstrap(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		openai.WriteError(w, http.StatusMethodNotAllowed, "invalid_request_error", "method_not_allowed", "method not allowed")
		return
	}
	_, _, _, boot := s.snapshotState()
	writeJSON(w, http.StatusOK, BootstrapResponse{Result: boot})
}

func (s *Server) handleBootstrapRun(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		openai.WriteError(w, http.StatusMethodNotAllowed, "invalid_request_error", "method_not_allowed", "method not allowed")
		return
	}
	cfg, err := s.configStore.Load()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, BootstrapResponse{Error: err.Error()})
		return
	}
	boot, err := app.Bootstrap(context.Background(), cfg)
	if boot != nil {
		liveCfg := *cfg
		liveCfg.ProxyNodes = boot.FlatResult.Nodes
		s.applyLiveConfig(&liveCfg, boot)
	}
	if err != nil {
		writeJSON(w, http.StatusBadGateway, BootstrapResponse{Result: boot, Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, BootstrapResponse{Result: boot})
}