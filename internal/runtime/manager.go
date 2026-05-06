package runtime

import (
	"encoding/json"
	"os"
	"path/filepath"

	"api-v2ray/internal/xrayruntime"
)

type State struct {
	Instances []xrayruntime.GeneratedInstance `json:"instances"`
}

func NewState(instances []xrayruntime.GeneratedInstance) State {
	copied := append([]xrayruntime.GeneratedInstance{}, instances...)
	return State{Instances: copied}
}

func SaveState(path string, instances []xrayruntime.GeneratedInstance) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(NewState(instances), "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}
