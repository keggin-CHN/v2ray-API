package runtime

import "testing"

func TestBuildCommand(t *testing.T) {
	cmd := BuildCommand("xray", "/tmp/a.json")
	if cmd.Path == "" {
		t.Fatalf("empty command path")
	}
}
