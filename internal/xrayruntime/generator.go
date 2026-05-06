package xrayruntime

import (
	"encoding/json"
	"strings"
	"fmt"
	"os"
	"path/filepath"

	"api-v2ray/internal/model"
)

type GeneratedInstance struct {
	NodeID     string `json:"node_id"`
	ListenHost string `json:"listen_host"`
	ListenPort int    `json:"listen_port"`
	ConfigPath string `json:"config_path"`
	Enabled    bool   `json:"enabled"`
	Reason     string `json:"reason,omitempty"`
}

type Manager struct {
	runtimeDir string
	basePort   int
	xrayBinary string
}

func New(runtimeDir string, basePort int, xrayBinary string) *Manager {
	return &Manager{runtimeDir: runtimeDir, basePort: basePort, xrayBinary: xrayBinary}
}

func (m *Manager) Generate(nodes []model.ProxyNode) ([]GeneratedInstance, error) {
	if err := os.MkdirAll(filepath.Join(m.runtimeDir, "xray"), 0o755); err != nil {
		return nil, err
	}
	instances := make([]GeneratedInstance, 0, len(nodes))
	for i, node := range nodes {
		inst := GeneratedInstance{
			NodeID:     node.ID,
			ListenHost: "127.0.0.1",
			ListenPort: m.basePort + i + 1,
			ConfigPath: filepath.Join(m.runtimeDir, "xray", fmt.Sprintf("%s.json", sanitize(node.ID))),
		}
		cfg := buildStubConfig(node, inst.ListenPort)
		b, _ := json.MarshalIndent(cfg, "", "  ")
		_ = os.WriteFile(inst.ConfigPath, b, 0o644)
		inst.Enabled = false
		inst.Reason = "generated from parsed node; process launch pending"
		instances = append(instances, inst)
	}
	return instances, nil
}

func buildStubConfig(node model.ProxyNode, localPort int) map[string]any {
	return map[string]any{
		"log": map[string]any{"loglevel": "warning"},
		"inbounds": []any{
			map[string]any{
				"tag":      "socks-in",
				"port":     localPort,
				"listen":   "127.0.0.1",
				"protocol": "socks",
				"settings": map[string]any{"udp": false},
			},
		},
		"outbounds": []any{BuildOutbound(node)},
		"routing": map[string]any{
			"rules": []any{
				map[string]any{
					"type":        "field",
					"inboundTag":  []string{"socks-in"},
					"outboundTag": "proxy",
				},
			},
		},
		"api_v2ray_meta": map[string]any{
			"node_id": node.ID,
			"name":    node.Name,
			"scheme":  node.Scheme,
			"host":    node.Host,
			"port":    node.Port,
			"raw_uri": node.RawURI,
		},
	}
}

func sanitize(s string) string {
	replacer := func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_':
			return r
		default:
			return '-'
		}
	}
	return strings.Map(replacer, s)
}
