package runtime

import (
	"testing"

	"api-v2ray/internal/xrayruntime"
)

func TestNewStateCopiesInstances(t *testing.T) {
	instances := []xrayruntime.GeneratedInstance{{NodeID: "n1", ConfigPath: "/tmp/runtime/xray/n1.json"}}
	state := NewState(instances)
	instances[0].NodeID = "mutated"
	if len(state.Instances) != 1 { t.Fatalf("unexpected instances len: %d", len(state.Instances)) }
	if state.Instances[0].NodeID != "n1" {
		t.Fatalf("state should copy input instances, got %#v", state.Instances)
	}
}
