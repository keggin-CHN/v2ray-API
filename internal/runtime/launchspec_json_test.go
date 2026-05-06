package runtime

import (
	"encoding/json"
	"testing"
)

func TestLaunchSpecJSONShape(t *testing.T) {
	spec := LaunchSpec{NodeID: "n1", ConfigPath: "/tmp/n1.json", Command: "xray", Args: []string{"run", "-c", "/tmp/n1.json"}}
	b, err := json.Marshal(spec)
	if err != nil { t.Fatalf("marshal err: %v", err) }
	var got map[string]any
	if err := json.Unmarshal(b, &got); err != nil { t.Fatalf("unmarshal err: %v", err) }
	if got["node_id"] != "n1" || got["command"] != "xray" {
		t.Fatalf("unexpected json shape: %+v", got)
	}
}
