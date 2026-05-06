package runtime

import (
	"testing"

	"api-v2ray/internal/xrayruntime"
)

func TestPlanFileSetConsistency(t *testing.T) {
	layout := NewLayout("/tmp/runtime", "/tmp/runtime/subscriptions/cache.json")
	instances := []xrayruntime.GeneratedInstance{{NodeID: "n1", ConfigPath: "/tmp/runtime/xray/n1.json"}}
	plan := BuildPlan(layout, instances, "xray")
	if len(plan.FileSet.ConfigPaths) != len(plan.LaunchSpecs) {
		t.Fatalf("config paths / launch specs mismatch: %d vs %d", len(plan.FileSet.ConfigPaths), len(plan.LaunchSpecs))
	}
}
