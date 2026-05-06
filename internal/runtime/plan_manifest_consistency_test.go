package runtime

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"api-v2ray/internal/xrayruntime"
)

func TestPlanManifestConsistency(t *testing.T) {
	tmp := t.TempDir()
	layout := NewLayout(filepath.Join(tmp, "runtime"), filepath.Join(tmp, "runtime", "subscriptions", "cache.json"))
	instances := []xrayruntime.GeneratedInstance{{NodeID: "n1", ConfigPath: filepath.Join(tmp, "runtime", "xray", "n1.json")}}
	plan := BuildPlan(layout, instances, "xray")
	files := append([]string{}, plan.FileSet.ConfigPaths...)
	files = append(files, plan.FileSet.LogPaths...)
	if err := os.MkdirAll(filepath.Dir(layout.ManifestPath), 0o755); err != nil { t.Fatalf("mkdir err: %v", err) }
	if err := WriteManifest(layout.ManifestPath, files); err != nil { t.Fatalf("write manifest err: %v", err) }
	b, err := os.ReadFile(layout.ManifestPath)
	if err != nil { t.Fatalf("read err: %v", err) }
	var got Manifest
	if err := json.Unmarshal(b, &got); err != nil { t.Fatalf("json err: %v", err) }
	if len(got.Files) != len(files) { t.Fatalf("want %d files, got %d", len(files), len(got.Files)) }
}
