package runtime

import "testing"

func TestBuildFileSetPreservesConfigOrder(t *testing.T) {
	specs := []LaunchSpec{
		{NodeID: "n2", ConfigPath: "/tmp/runtime/xray/n2.json"},
		{NodeID: "n1", ConfigPath: "/tmp/runtime/xray/n1.json"},
	}
	fs := BuildFileSet("/tmp/runtime", specs)
	if len(fs.ConfigPaths) != 2 { t.Fatalf("expected 2 config paths, got %d", len(fs.ConfigPaths)) }
	if fs.ConfigPaths[0] != "/tmp/runtime/xray/n2.json" || fs.ConfigPaths[1] != "/tmp/runtime/xray/n1.json" {
		t.Fatalf("unexpected config order: %#v", fs.ConfigPaths)
	}
}
