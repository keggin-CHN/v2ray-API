package runtime

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"api-v2ray/internal/model"
)

type ProcessState struct {
	NodeID    string `json:"node_id"`
	PID       int    `json:"pid"`
	Config    string `json:"config"`
	StartedAt string `json:"started_at"`
}

type ProcessStateFile struct {
	Processes []ProcessState `json:"processes"`
}

func runtimeDir(cfg *model.Config) string {
	if cfg != nil && cfg.Runtime.Dir != "" {
		return cfg.Runtime.Dir
	}
	return "runtime"
}

func xrayConfigPath(cfg *model.Config, nodeID string) string {
	return filepath.Join(runtimeDir(cfg), "xray", nodeID+".json")
}

func processStatePath(cfg *model.Config) string {
	return filepath.Join(runtimeDir(cfg), "xray", "processes.json")
}

func ProcessStatePath(cfg *model.Config) string {
	return processStatePath(cfg)
}

func loadProcessState(path string) (ProcessStateFile, error) {
	var st ProcessStateFile
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return st, nil
		}
		return st, err
	}
	if len(b) == 0 {
		return st, nil
	}
	if err := json.Unmarshal(b, &st); err != nil {
		return st, err
	}
	return st, nil
}

func LoadProcessState(path string) (ProcessStateFile, error) {
	return loadProcessState(path)
}

func saveProcessState(path string, st ProcessStateFile) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

func killPID(pid int) {
	if pid <= 0 {
		return
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return
	}
	if runtime.GOOS == "windows" {
		_ = proc.Kill()
		return
	}
	_ = proc.Signal(syscall.SIGTERM)
	time.Sleep(300 * time.Millisecond)
	_ = proc.Kill()
}

func stopTrackedXray(cfg *model.Config) error {
	statePath := processStatePath(cfg)
	st, err := loadProcessState(statePath)
	if err != nil {
		return err
	}
	for _, p := range st.Processes {
		killPID(p.PID)
	}
	if err := os.Remove(statePath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func StartXrayProcesses(cfg *model.Config) error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}
	bin := strings.TrimSpace(cfg.Runtime.XrayBinary)
	if bin == "" {
		return fmt.Errorf("runtime.xray_binary is empty")
	}
	if err := stopTrackedXray(cfg); err != nil {
		return fmt.Errorf("stop tracked xray processes: %w", err)
	}

	var started []ProcessState
	for _, node := range cfg.ProxyNodes {
		nodeID := strings.TrimSpace(node.ID)
		if nodeID == "" {
			continue
		}
		cfgPath := xrayConfigPath(cfg, nodeID)
		if _, err := os.Stat(cfgPath); err != nil {
			continue
		}
		cmd := exec.Command(bin, "run", "-c", cfgPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Start(); err != nil {
			stopStartedProcesses(started)
			return fmt.Errorf("start xray for %s: %w", nodeID, err)
		}
		if cmd.Process == nil {
			stopStartedProcesses(started)
			return fmt.Errorf("start xray for %s: empty process", nodeID)
		}
		started = append(started, ProcessState{
			NodeID:    nodeID,
			PID:       cmd.Process.Pid,
			Config:    cfgPath,
			StartedAt: time.Now().UTC().Format(time.RFC3339),
		})
		if err := cmd.Process.Release(); err != nil {
			stopStartedProcesses(started)
			return fmt.Errorf("release xray process for %s: %w", nodeID, err)
		}
	}
	return saveProcessState(processStatePath(cfg), ProcessStateFile{Processes: started})
}

func stopStartedProcesses(processes []ProcessState) {
	for _, p := range processes {
		killPID(p.PID)
	}
}
