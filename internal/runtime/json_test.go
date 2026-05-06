package runtime

import (
	"encoding/json"
	"testing"

	"api-v2ray/internal/xrayruntime"
)

func TestManifestJSON(t *testing.T) {
	b, err := NewManifest([]string{"/tmp/runtime/xray/n1.json"}).JSON()
	if err != nil { t.Fatalf("manifest json err: %v", err) }
	var got map[string]any
	if err := json.Unmarshal(b, &got); err != nil { t.Fatalf("unmarshal err: %v", err) }
	if got["version"] != "v1" { t.Fatalf("unexpected manifest json: %+v", got) }
}

func TestStateJSON(t *testing.T) {
	b, err := NewState([]xrayruntime.GeneratedInstance{{NodeID: "n1", ConfigPath: "/tmp/runtime/xray/n1.json"}}).JSON()
	if err != nil { t.Fatalf("state json err: %v", err) }
	var got map[string]any
	if err := json.Unmarshal(b, &got); err != nil { t.Fatalf("unmarshal err: %v", err) }
	if _, ok := got["instances"]; !ok { t.Fatalf("unexpected state json: %+v", got) }
}
