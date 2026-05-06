package runtime

import (
	"testing"

	"api-v2ray/internal/xrayruntime"
)

func TestBuildLaunchSpecs(t *testing.T) {
	instances := []xrayruntime.GeneratedInstance{{NodeID: "n1", ConfigPath: "/tmp/n1.json"}}
	specs := BuildLaunchSpecs("xray", instances)
	if len(specs) != 1 { t.Fatalf("want 1 spec, got %d", len(specs)) }
	if specs[0].Command != "xray" { t.Fatalf("unexpected command: %s", specs[0].Command) }
}

func TestDefaultLogPath(t *testing.T) {
	p := DefaultLogPath("/tmp/runtime", "n1")
	if p == "" { t.Fatalf("empty log path") }
}
