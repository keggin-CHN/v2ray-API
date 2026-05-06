package runtime

import (
	"testing"

	"api-v2ray/internal/xrayruntime"
)

func TestBuildPlan(t *testing.T) {
	layout := NewLayout("/tmp/runtime", "/tmp/runtime/subscriptions/cache.json")
	instances := []xrayruntime.GeneratedInstance{{NodeID: "n1", ConfigPath: "/tmp/runtime/xray/n1.json"}}
	plan := BuildPlan(layout, instances, "xray")
	if len(plan.LaunchSpecs) != 1 { t.Fatalf("want 1 launch spec, got %d", len(plan.LaunchSpecs)) }
	if len(plan.FileSet.ConfigPaths) != 1 { t.Fatalf("want 1 config path, got %d", len(plan.FileSet.ConfigPaths)) }
}
