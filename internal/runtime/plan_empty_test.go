package runtime

import "testing"

func TestBuildPlanEmpty(t *testing.T) {
	layout := NewLayout("/tmp/runtime", "/tmp/runtime/subscriptions/cache.json")
	plan := BuildPlan(layout, nil, "xray")
	if len(plan.Instances) != 0 || len(plan.LaunchSpecs) != 0 {
		t.Fatalf("expected empty plan, got %#v", plan)
	}
}
