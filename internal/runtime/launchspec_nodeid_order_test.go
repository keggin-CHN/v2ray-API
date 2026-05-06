package runtime

import (
	"testing"

	"api-v2ray/internal/xrayruntime"
)

func TestBuildLaunchSpecsPreservesNodeOrder(t *testing.T) {
	instances := []xrayruntime.GeneratedInstance{
		{NodeID: "z-node", ConfigPath: "/tmp/runtime/xray/z.json"},
		{NodeID: "a-node", ConfigPath: "/tmp/runtime/xray/a.json"},
	}
	specs := BuildLaunchSpecs("xray", instances)
	if len(specs) != 2 { t.Fatalf("expected 2 specs, got %d", len(specs)) }
	if specs[0].NodeID != "z-node" || specs[1].NodeID != "a-node" {
		t.Fatalf("unexpected spec order: %#v", specs)
	}
}
