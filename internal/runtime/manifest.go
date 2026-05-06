package runtime

import (
	"encoding/json"
	"os"
)

type Manifest struct {
	Version string   `json:"version"`
	Files   []string `json:"files"`
}

func NewManifest(files []string) Manifest {
	copied := append([]string{}, files...)
	return Manifest{Version: "v1", Files: copied}
}

func WriteManifest(path string, files []string) error {
	b, err := json.MarshalIndent(NewManifest(files), "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}
