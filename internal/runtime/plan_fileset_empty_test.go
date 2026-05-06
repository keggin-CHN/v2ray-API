package runtime

import "testing"

func TestBuildPlanEmptyFileSet(t *testing.T) {
	layout := NewLayout("/tmp/runtime", "/tmp/runtime/subscriptions/cache.json")
	plan := BuildPlan(layout, nil, "xray")
	if len(plan.FileSet.ConfigPaths) != 0 || len(plan.FileSet.LogPaths) != 0 {
		t.Fatalf("expected empty fileset, got %#v", plan.FileSet)
	}
}
