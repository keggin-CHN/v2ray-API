package runtime

import (
	"testing"

	"api-v2ray/internal/xrayruntime"
)

func TestPlanLaunchSpecConsistency(t *testing.T) {
	layout := NewLayout("/tmp/runtime", "/tmp/runtime/subscriptions/cache.json")
	instances := []xrayruntime.GeneratedInstance{
		{NodeID: "n1", ConfigPath: "/tmp/runtime/xray/n1.json"},
		{NodeID: "n2", ConfigPath: "/tmp/runtime/xray/n2.json"},
	}
	plan := BuildPlan(layout, instances, "xray")
	if len(plan.LaunchSpecs) != len(plan.Instances) {
		t.Fatalf("launch specs / instances mismatch: %d vs %d", len(plan.LaunchSpecs), len(plan.Instances))
	}
}
