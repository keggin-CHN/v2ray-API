package runtime

import "testing"

func TestBuildFileSetPreservesLogOrder(t *testing.T) {
	specs := []LaunchSpec{{NodeID: "b", ConfigPath: "/tmp/runtime/xray/b.json"}, {NodeID: "a", ConfigPath: "/tmp/runtime/xray/a.json"}}
	fs := BuildFileSet("/tmp/runtime", specs)
	if len(fs.LogPaths) != 2 { t.Fatalf("expected 2 log paths, got %d", len(fs.LogPaths)) }
	if fs.LogPaths[0] != "/tmp/runtime/logs/b.log" || fs.LogPaths[1] != "/tmp/runtime/logs/a.log" {
		t.Fatalf("unexpected log order: %#v", fs.LogPaths)
	}
}
