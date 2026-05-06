package app

import (
	"api-v2ray/internal/model"
	"api-v2ray/internal/runtime"
)

func BuildRuntimeLaunchResult(cfg *model.Config, launchSpecs []runtime.LaunchSpec) RuntimeLaunchResult {
	binary := ""
	if cfg != nil {
		binary = cfg.Runtime.XrayBinary
	}
	enabled := binary != ""
	copiedSpecs := append([]runtime.LaunchSpec{}, launchSpecs...)
	return RuntimeLaunchResult{
		Binary:      binary,
		LaunchSpecs: copiedSpecs,
		Enabled:     enabled,
		Ready:       enabled && len(copiedSpecs) > 0,
		Started:     false,
		PIDs:        []int{},
		Errors:      []string{},
	}
}
