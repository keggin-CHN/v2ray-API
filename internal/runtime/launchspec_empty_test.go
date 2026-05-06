package runtime

import (
	"testing"

	"api-v2ray/internal/xrayruntime"
)

func TestBuildLaunchSpecsEmpty(t *testing.T) {
	var instances []xrayruntime.GeneratedInstance
	specs := BuildLaunchSpecs("xray", instances)
	if len(specs) != 0 {
		t.Fatalf("expected empty specs, got %#v", specs)
	}
}
