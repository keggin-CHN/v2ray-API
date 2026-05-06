package runtime

import "testing"

func TestBuildFileSet(t *testing.T) {
	specs := []LaunchSpec{{NodeID: "n1", ConfigPath: "/tmp/n1.json"}, {NodeID: "n2", ConfigPath: "/tmp/n2.json"}}
	fs := BuildFileSet("/tmp/runtime", specs)
	if len(fs.ConfigPaths) != 2 || len(fs.LogPaths) != 2 {
		t.Fatalf("unexpected fileset sizes: %#v", fs)
	}
}
