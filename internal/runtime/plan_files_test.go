package runtime

import "testing"

func TestPlanManifestFiles(t *testing.T) {
	plan := Plan{FileSet: FileSet{ConfigPaths: []string{"/tmp/runtime/xray/n1.json"}, LogPaths: []string{"/tmp/runtime/logs/n1.log"}}}
	files := plan.ManifestFiles()
	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(files))
	}
	if files[0] != "/tmp/runtime/xray/n1.json" || files[1] != "/tmp/runtime/logs/n1.log" {
		t.Fatalf("unexpected files: %#v", files)
	}
}
