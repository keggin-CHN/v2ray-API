package runtime

import (
	"path/filepath"
	"testing"
)

func TestDefaultLogPathUnderLogsDir(t *testing.T) {
	p := DefaultLogPath("/tmp/runtime", "n1")
	if filepath.Dir(p) != filepath.Join("/tmp/runtime", "logs") {
		t.Fatalf("unexpected dir: %s", filepath.Dir(p))
	}
}
