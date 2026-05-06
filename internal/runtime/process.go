package runtime

import "os/exec"

type ManagedProcess struct {
	Command string `json:"command"`
	Args    []string `json:"args"`
}

func BuildCommand(bin string, configPath string) *exec.Cmd {
	return exec.Command(bin, "run", "-c", configPath)
}
