package runtime

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteManifestContent(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "manifest.json")
	files := []string{"/tmp/runtime/xray/n1.json", "/tmp/runtime/logs/n1.log"}
	if err := WriteManifest(path, files); err != nil {
		t.Fatalf("write manifest err: %v", err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read err: %v", err)
	}
	var got Manifest
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("json err: %v", err)
	}
	if got.Version != "v1" || len(got.Files) != 2 {
		t.Fatalf("unexpected manifest: %+v", got)
	}
}
