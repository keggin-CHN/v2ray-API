package runtime

import (
	"encoding/json"
	"os"
	"path/filepath"

	"api-v2ray/internal/model"
)

func WriteSubscriptionCache(path string, nodes []model.ProxyNode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(map[string]any{"nodes": nodes}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}
