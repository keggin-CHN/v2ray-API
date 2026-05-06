package runtime

import (
	"path/filepath"
	"testing"
)

func TestLayoutPaths(t *testing.T) {
	root := "/tmp/runtime"
	cache := filepath.Join(root, "subscriptions", "cache.json")
	l := NewLayout(root, cache)
	if l.ManifestPath != filepath.Join(root, "manifest.json") { t.Fatalf("unexpected manifest path: %s", l.ManifestPath) }
	if l.RuntimeStatePath != filepath.Join(root, "runtime-state.json") { t.Fatalf("unexpected runtime state path: %s", l.RuntimeStatePath) }
}
