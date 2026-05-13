package server

import (
	"net/http"

	appruntime "api-v2ray/internal/runtime"
	"api-v2ray/internal/upstream"
)

type RuntimeMetricsResponse struct {
	ProcessStatePath string                         `json:"process_state_path"`
	Processes        []appruntime.ProcessState      `json:"processes"`
	Routes           map[string]any                 `json:"routes,omitempty"`
	Upstream         upstream.ClientStats           `json:"upstream"`
	Error            string                         `json:"error,omitempty"`
}

func (s *Server) handleUpstreamMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "method not allowed"})
		return
	}
	_, _, upstreamClient, _ := s.snapshotState()
	writeJSON(w, http.StatusOK, upstreamClient.StatsSnapshot())
}

func (s *Server) handleRuntimeMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "method not allowed"})
		return
	}

	cfg, err := s.configStore.Load()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, RuntimeMetricsResponse{
			Error: "load config failed: " + err.Error(),
		})
		return
	}

	statePath := appruntime.ProcessStatePath(cfg)
	st, err := appruntime.LoadProcessState(statePath)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, RuntimeMetricsResponse{
			ProcessStatePath: statePath,
			Error:            "load process state failed: " + err.Error(),
		})
		return
	}

	routerSvc, _, upstreamClient, _ := s.snapshotState()
	writeJSON(w, http.StatusOK, RuntimeMetricsResponse{
		ProcessStatePath: statePath,
		Processes:        st.Processes,
		Routes:           routerSvc.HealthSnapshot(),
		Upstream:         upstreamClient.StatsSnapshot(),
	})
}