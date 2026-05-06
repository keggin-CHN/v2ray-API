package runtime

import (
	"path/filepath"

	"api-v2ray/internal/xrayruntime"
)

type LaunchSpec struct {
	NodeID     string `json:"node_id"`
	ConfigPath string `json:"config_path"`
	Command    string `json:"command"`
	Args       []string `json:"args"`
}

func BuildLaunchSpecs(bin string, instances []xrayruntime.GeneratedInstance) []LaunchSpec {
	out := make([]LaunchSpec, 0, len(instances))
	for _, inst := range instances {
		out = append(out, LaunchSpec{
			NodeID:     inst.NodeID,
			ConfigPath: inst.ConfigPath,
			Command:    bin,
			Args:       []string{"run", "-c", inst.ConfigPath},
		})
	}
	return out
}

func DefaultLogPath(rootDir string, nodeID string) string {
	return filepath.Join(rootDir, "logs", nodeID+".log")
}
