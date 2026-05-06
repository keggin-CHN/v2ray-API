package runtime

import "api-v2ray/internal/xrayruntime"

type Plan struct {
	Layout      Layout                          `json:"layout"`
	Instances   []xrayruntime.GeneratedInstance `json:"instances"`
	LaunchSpecs []LaunchSpec                    `json:"launch_specs"`
	FileSet     FileSet                         `json:"file_set"`
}

func BuildPlan(layout Layout, instances []xrayruntime.GeneratedInstance, bin string) Plan {
	launchSpecs := BuildLaunchSpecs(bin, instances)
	fileSet := BuildFileSet(layout.RootDir, launchSpecs)
	return Plan{
		Layout:      layout,
		Instances:   instances,
		LaunchSpecs: launchSpecs,
		FileSet:     fileSet,
	}
}
