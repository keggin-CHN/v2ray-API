package server

import (
    "net/http"
    appruntime "api-v2ray/internal/runtime"
    "api-v2ray/internal/openai"
)


func (s *Server) handleRestart(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        openai.WriteError(w, http.StatusMethodNotAllowed, "invalid_request_error", "method_not_allowed", "method not allowed")
        return
    }
    cfg, err := s.configStore.Load()
    if err != nil {
        writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
        return
    }
    if err := appruntime.StartXrayProcesses(cfg); err != nil {
        writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
        return
    }
    s.applyLiveConfig(cfg, s.bootstrap)
    writeJSON(w, http.StatusOK, map[string]any{"ok": true, "msg": "xray processes started"})
}
