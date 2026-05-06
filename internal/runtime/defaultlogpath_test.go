package runtime

import "testing"

func TestDefaultLogPathContainsNodeID(t *testing.T) {
	p := DefaultLogPath("/tmp/runtime", "node-123")
	if p == "" || p[len(p)-len("node-123.log"):] != "node-123.log" {
		t.Fatalf("unexpected log path: %s", p)
	}
}
