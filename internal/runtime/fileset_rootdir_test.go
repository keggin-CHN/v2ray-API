package runtime

import "testing"

func TestBuildFileSetUsesProvidedRootDir(t *testing.T) {
	specs := []LaunchSpec{{NodeID: "n9", ConfigPath: "/tmp/other/n9.json"}}
	fs := BuildFileSet("/srv/api-v2ray/runtime", specs)
	if len(fs.LogPaths) != 1 { t.Fatalf("expected 1 log path, got %d", len(fs.LogPaths)) }
	if fs.LogPaths[0] != "/srv/api-v2ray/runtime/logs/n9.log" {
		t.Fatalf("unexpected log path: %s", fs.LogPaths[0])
	}
}
