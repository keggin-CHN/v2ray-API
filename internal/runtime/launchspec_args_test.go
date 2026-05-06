package runtime

import (
	"testing"

	"api-v2ray/internal/xrayruntime"
)

func TestBuildLaunchSpecsArgs(t *testing.T) {
	instances := []xrayruntime.GeneratedInstance{{NodeID: "n1", ConfigPath: "/tmp/runtime/xray/n1.json"}}
	specs := BuildLaunchSpecs("xray", instances)
	if len(specs) != 1 { t.Fatalf("want 1 spec, got %d", len(specs)) }
	if len(specs[0].Args) != 3 { t.Fatalf("unexpected args len: %d", len(specs[0].Args)) }
	if specs[0].Args[0] != "run" || specs[0].Args[1] != "-c" || specs[0].Args[2] != "/tmp/runtime/xray/n1.json" {
		t.Fatalf("unexpected args: %#v", specs[0].Args)
	}
}
