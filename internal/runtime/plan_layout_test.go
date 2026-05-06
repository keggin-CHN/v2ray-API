package runtime

import "testing"

func TestBuildPlanKeepsLayout(t *testing.T) {
	layout := NewLayout("/tmp/runtime", "/tmp/runtime/subscriptions/cache.json")
	plan := BuildPlan(layout, nil, "xray")
	if plan.Layout.RootDir != "/tmp/runtime" { t.Fatalf("unexpected root dir: %s", plan.Layout.RootDir) }
}
