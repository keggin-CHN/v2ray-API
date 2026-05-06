package runtime

import "testing"

func TestBuildFileSetUsesDefaultLogPath(t *testing.T) {
	specs := []LaunchSpec{{NodeID: "n1", ConfigPath: "/tmp/runtime/xray/n1.json"}}
	fs := BuildFileSet("/tmp/runtime", specs)
	if len(fs.LogPaths) != 1 { t.Fatalf("want 1 log path, got %d", len(fs.LogPaths)) }
	if fs.LogPaths[0] == "" { t.Fatalf("empty log path") }
}
