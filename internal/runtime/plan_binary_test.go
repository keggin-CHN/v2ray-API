package runtime

import (
	"testing"

	"api-v2ray/internal/xrayruntime"
)

func TestBuildPlanCarriesBinaryIntoLaunchSpecs(t *testing.T) {
	layout := NewLayout("/tmp/runtime", "/tmp/runtime/subscriptions/cache.json")
	instances := []xrayruntime.GeneratedInstance{{NodeID: "n1", ConfigPath: "/tmp/runtime/xray/n1.json"}}
	plan := BuildPlan(layout, instances, "/usr/bin/xray")
	if len(plan.LaunchSpecs) != 1 { t.Fatalf("want 1 spec, got %d", len(plan.LaunchSpecs)) }
	if plan.LaunchSpecs[0].Command != "/usr/bin/xray" { t.Fatalf("unexpected command: %s", plan.LaunchSpecs[0].Command) }
}
