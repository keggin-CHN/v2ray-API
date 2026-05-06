package runtime

import (
	"fmt"
	"os"
	"os/exec"

	"api-v2ray/internal/model"
)

func KillExistingXray() {
	_ = exec.Command("pkill", "-f", "xray").Run()
}

func StartXrayProcesses(cfg *model.Config) error {
	KillExistingXray()
	for _, node := range cfg.ProxyNodes {
		cfgPath := fmt.Sprintf("runtime/xray/%s.json", node.ID)
		if _, err := os.Stat(cfgPath); err != nil {
			continue
		}
		cmd := exec.Command(cfg.Runtime.XrayBinary, "run", "-c", cfgPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("start xray for %s: %w", node.ID, err)
		}
		if err := cmd.Process.Release(); err != nil {
			return fmt.Errorf("release xray process for %s: %w", node.ID, err)
		}
	}
	return nil
}
