package runtime

import (
	"testing"

	"api-v2ray/internal/xrayruntime"
)

func TestPlanArtifacts(t *testing.T) {
	layout := NewLayout("/tmp/runtime", "/tmp/runtime/subscriptions/cache.json")
	instances := []xrayruntime.GeneratedInstance{{NodeID: "n1", ConfigPath: "/tmp/runtime/xray/n1.json"}}
	plan := BuildPlan(layout, instances, "xray")
	artifacts := plan.Artifacts()
	if len(artifacts.Manifest.Files) != 2 { t.Fatalf("unexpected manifest files: %#v", artifacts.Manifest.Files) }
	if len(artifacts.State.Instances) != 1 { t.Fatalf("unexpected state instances: %#v", artifacts.State.Instances) }
}
