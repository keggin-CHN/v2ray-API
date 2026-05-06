package runtime

import (
	"testing"

	"api-v2ray/internal/xrayruntime"
)

func TestBuildLaunchSpecsKeepsBinary(t *testing.T) {
	instances := []xrayruntime.GeneratedInstance{{NodeID: "n1", ConfigPath: "/tmp/runtime/xray/n1.json"}}
	specs := BuildLaunchSpecs("/usr/local/bin/xray", instances)
	if len(specs) != 1 { t.Fatalf("want 1 spec, got %d", len(specs)) }
	if specs[0].Command != "/usr/local/bin/xray" {
		t.Fatalf("unexpected command: %s", specs[0].Command)
	}
}
