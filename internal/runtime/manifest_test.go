package runtime

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteManifest(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "manifest.json")
	if err := WriteManifest(path, []string{"a.json", "b.json"}); err != nil {
		t.Fatalf("write manifest err: %v", err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read manifest err: %v", err)
	}
	if len(b) == 0 {
		t.Fatalf("manifest empty")
	}
}
