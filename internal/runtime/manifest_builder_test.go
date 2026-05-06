package runtime

import "testing"

func TestNewManifestCopiesFiles(t *testing.T) {
	files := []string{"/tmp/runtime/xray/n1.json", "/tmp/runtime/logs/n1.log"}
	m := NewManifest(files)
	files[0] = "mutated"
	if m.Version != "v1" { t.Fatalf("unexpected version: %s", m.Version) }
	if len(m.Files) != 2 { t.Fatalf("unexpected files len: %d", len(m.Files)) }
	if m.Files[0] != "/tmp/runtime/xray/n1.json" {
		t.Fatalf("manifest should copy input files, got %#v", m.Files)
	}
}
