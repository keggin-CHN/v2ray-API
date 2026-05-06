package runtime

import "testing"

func TestNewLayout(t *testing.T) {
	l := NewLayout("/tmp/api-v2ray-runtime", "/tmp/api-v2ray-runtime/subscriptions/cache.json")
	if l.XrayDir == "" || l.ManifestPath == "" || l.RuntimeStatePath == "" {
		t.Fatalf("layout fields should not be empty: %#v", l)
	}
}
