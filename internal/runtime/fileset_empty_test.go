package runtime

import "testing"

func TestBuildFileSetEmpty(t *testing.T) {
	fs := BuildFileSet("/tmp/runtime", nil)
	if len(fs.ConfigPaths) != 0 || len(fs.LogPaths) != 0 {
		t.Fatalf("expected empty fileset, got %#v", fs)
	}
}
