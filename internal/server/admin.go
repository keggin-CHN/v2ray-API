package server

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"api-v2ray/internal/app"
	"api-v2ray/internal/config"
	"api-v2ray/internal/model"
)

type ConfigStore struct {
	Path string
}

type ConfigUpdateRequest struct {
	Config model.Config `json:"config"`
}

type ConfigResponse struct {
	Path   string       `json:"path"`
	Config model.Config `json:"config"`
}

type BootstrapResponse struct {
	Result *app.BootstrapResult `json:"result,omitempty"`
	Error  string               `json:"error,omitempty"`
}

type ApplyConfigResponse struct {
	Path   string               `json:"path"`
	Config model.Config         `json:"config"`
	Result *app.BootstrapResult `json:"result,omitempty"`
	Error  string               `json:"error,omitempty"`
}

func (cs ConfigStore) Load() (*model.Config, error) {
	return config.Load(cs.Path)
}

func (cs ConfigStore) Save(cfg *model.Config) error {
	if err := os.MkdirAll(filepath.Dir(cs.Path), 0o755); err != nil {
		return err
	}
	if old, err := os.ReadFile(cs.Path); err == nil {
		backupDir := filepath.Join(filepath.Dir(cs.Path), ".history")
		if err := os.MkdirAll(backupDir, 0o755); err != nil {
			return err
		}
		backupName := filepath.Base(cs.Path) + "." + time.Now().Format("20060102-150405") + ".bak"
		if err := os.WriteFile(filepath.Join(backupDir, backupName), old, 0o644); err != nil {
			return err
		}
	}
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	return os.WriteFile(cs.Path, b, 0o644)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
