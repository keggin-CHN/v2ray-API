package runtime

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"api-v2ray/internal/model"
)

func TestWriteSubscriptionCache(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "runtime", "subscriptions", "cache.json")
	nodes := []model.ProxyNode{{ID: "n1", Scheme: "socks5", Host: "127.0.0.1", Port: 21001}}
	if err := WriteSubscriptionCache(path, nodes); err != nil { t.Fatalf("write cache err: %v", err) }
	b, err := os.ReadFile(path)
	if err != nil { t.Fatalf("read cache err: %v", err) }
	var got map[string]any
	if err := json.Unmarshal(b, &got); err != nil { t.Fatalf("cache json err: %v", err) }
	if _, ok := got["nodes"]; !ok { t.Fatalf("unexpected cache json: %+v", got) }
}
