package runtime

import (
	"os"
	"path/filepath"
	"strings"
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
	script := "#!/bin/sh\necho started >> \"" + marker + "\"\n"
	if err := os.WriteFile(bin, []byte(script), 0o755); err != nil {
		t.Fatalf("rewrite fake bin: %v", err)
	}
	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	cfg := &model.Config{
		Runtime: model.RuntimeConfig{XrayBinary: bin, Dir: "./runtime"},
		ProxyNodes: []model.ProxyNode{{ID: "node-1"}},
	}
	if err := StartXrayProcesses(cfg); err != nil {
		t.Fatalf("start xray processes: %v", err)
	}
	if _, err := os.Stat(marker); err == nil {
		t.Fatalf("expected missing config to skip process launch")
	}
}

func TestStartXrayProcessesUsesConfiguredRuntimeDir(t *testing.T) {
	tmp := t.TempDir()
	bin := filepath.Join(tmp, "fake-xray.sh")
	if err := os.WriteFile(bin, []byte("#!/bin/sh\nsleep 1\n"), 0o755); err != nil {
		t.Fatalf("write fake bin: %v", err)
	}

	runtimeDir := filepath.Join(tmp, "custom-runtime")
	cfgDir := filepath.Join(runtimeDir, "xray")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(cfgDir, "node-custom.json"), []byte("{}"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg := &model.Config{
		Runtime:    model.RuntimeConfig{XrayBinary: bin, Dir: runtimeDir},
		ProxyNodes: []model.ProxyNode{{ID: "node-custom"}},
	}
	if err := StartXrayProcesses(cfg); err != nil {
		t.Fatalf("start xray processes: %v", err)
	}

	statePath := filepath.Join(runtimeDir, "xray", "processes.json")
	b, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatalf("read state: %v", err)
	}
	if !strings.Contains(string(b), filepath.Join(runtimeDir, "xray", "node-custom.json")) {
		t.Fatalf("expected configured runtime config path in state, got: %s", string(b))
	}
}

func TestStartXrayProcessesWritesStateFile(t *testing.T) {
	tmp := t.TempDir()
	bin := filepath.Join(tmp, "fake-xray.sh")
	if err := os.WriteFile(bin, []byte("#!/bin/sh\nsleep 1\n"), 0o755); err != nil {
		t.Fatalf("write fake bin: %v", err)
	}

	cfgDir := filepath.Join(tmp, "runtime", "xray")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(cfgDir, "node-a.json"), []byte("{}"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	cfg := &model.Config{
		Runtime: model.RuntimeConfig{XrayBinary: bin, Dir: "./runtime"},
		ProxyNodes: []model.ProxyNode{{ID: "node-a"}},
	}
	if err := StartXrayProcesses(cfg); err != nil {
		t.Fatalf("start xray processes: %v", err)
	}

	statePath := filepath.Join("runtime", "xray", "processes.json")
	b, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatalf("read state: %v", err)
	}
	if !strings.Contains(string(b), "\"node_id\": \"node-a\"") {
		t.Fatalf("unexpected state content: %s", string(b))
	}
}
