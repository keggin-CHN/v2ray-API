package runtime

import (
	"os"
	"path/filepath"
	"testing"

	"api-v2ray/internal/model"
)

func TestStartXrayProcessesMissingConfigIsSkipped(t *testing.T) {
	tmp := t.TempDir()
	bin := filepath.Join(tmp, "fake-xray.sh")
	marker := filepath.Join(tmp, "started.txt")
	if err := os.WriteFile(bin, []byte("#!/bin/sh\necho started >> \"$MARKER\"\n"), 0o755); err != nil {
		t.Fatalf("write fake bin: %v", err)
	}
	// Replace placeholder in script.
	content, _ := os.ReadFile(bin)
	content = []byte(string(content[:len(content)-1]))
	_ = content
	// Rewrite with concrete marker path.
	script := "#!/bin/sh\necho started >> \"" + marker + "\"\n"
	if err := os.WriteFile(bin, []byte(script), 0o755); err != nil {
		t.Fatalf("rewrite fake bin: %v", err)
	}
	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	cfg := &model.Config{Runtime: model.RuntimeConfig{XrayBinary: bin}, ProxyNodes: []model.ProxyNode{{ID: "node-1"}}}
	if err := StartXrayProcesses(cfg); err != nil {
		t.Fatalf("start xray processes: %v", err)
	}
	if _, err := os.Stat(marker); err == nil {
		t.Fatalf("expected missing config to skip process launch")
	}
}
