package runtime

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"api-v2ray/internal/xrayruntime"
)

func TestSaveStateContent(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "runtime-state.json")
	instances := []xrayruntime.GeneratedInstance{{NodeID: "n1", ConfigPath: "/tmp/runtime/xray/n1.json", ListenHost: "127.0.0.1", ListenPort: 21001}}
	if err := SaveState(path, instances); err != nil { t.Fatalf("save state err: %v", err) }
	b, err := os.ReadFile(path)
	if err != nil { t.Fatalf("read err: %v", err) }
	var got State
	if err := json.Unmarshal(b, &got); err != nil { t.Fatalf("json err: %v", err) }
	if len(got.Instances) != 1 || got.Instances[0].NodeID != "n1" { t.Fatalf("unexpected state: %+v", got) }
}
