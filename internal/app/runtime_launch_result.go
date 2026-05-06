package app

import "api-v2ray/internal/runtime"

type RuntimeLaunchResult struct {
	Binary      string               `json:"binary"`
	LaunchSpecs []runtime.LaunchSpec `json:"launch_specs"`
	Enabled     bool                 `json:"enabled"`
	Ready       bool                 `json:"ready"`
	Started     bool                 `json:"started"`
	PIDs        []int                `json:"pids"`
	Errors      []string             `json:"errors"`
}
