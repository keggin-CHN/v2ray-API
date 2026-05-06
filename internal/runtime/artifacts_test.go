package runtime

import (
	"os"
	"path/filepath"
	"testing"

	"api-v2ray/internal/xrayruntime"
)

func TestNewArtifactsFromPlan(t *testing.T) {
	layout := NewLayout("/tmp/runtime", "/tmp/runtime/subscriptions/cache.json")
	instances := []xrayruntime.GeneratedInstance{{NodeID: "n1", ConfigPath: "/tmp/runtime/xray/n1.json"}}
	plan := BuildPlan(layout, instances, "xray")
	artifacts := NewArtifacts(plan)
	if len(artifacts.Manifest.Files) != 2 { t.Fatalf("unexpected manifest files: %#v", artifacts.Manifest.Files) }
	if len(artifacts.State.Instances) != 1 { t.Fatalf("unexpected state instances: %#v", artifacts.State.Instances) }
}

func TestWriteArtifacts(t *testing.T) {
	tmp := t.TempDir()
	layout := NewLayout(filepath.Join(tmp, "runtime"), filepath.Join(tmp, "runtime", "subscriptions", "cache.json"))
	if err := os.MkdirAll(layout.RootDir, 0o755); err != nil { t.Fatalf("mkdir err: %v", err) }
	artifacts := Artifacts{Manifest: NewManifest([]string{"/tmp/runtime/xray/n1.json"}), State: NewState([]xrayruntime.GeneratedInstance{{NodeID: "n1", ConfigPath: "/tmp/runtime/xray/n1.json"}})}
	if err := WriteArtifacts(layout, artifacts); err != nil { t.Fatalf("write artifacts err: %v", err) }
	if _, err := os.Stat(layout.ManifestPath); err != nil { t.Fatalf("manifest stat err: %v", err) }
	if _, err := os.Stat(layout.RuntimeStatePath); err != nil { t.Fatalf("state stat err: %v", err) }
}
