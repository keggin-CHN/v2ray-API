package app

import (
	"api-v2ray/internal/runtime"
	"api-v2ray/internal/xrayruntime"
)

func CopyRuntimeStageResult(res *RuntimeStageResult) *RuntimeStageResult {
	if res == nil {
		return nil
	}
	copyRes := *res
	copyRes.GenerationResult.Generated = append(copyRes.GenerationResult.Generated[:0:0], res.GenerationResult.Generated...)
	copyRes.PlanResult.Plan.LaunchSpecs = append([]runtime.LaunchSpec{}, res.PlanResult.Plan.LaunchSpecs...)
	copyRes.PlanResult.Plan.FileSet.ConfigPaths = append([]string{}, res.PlanResult.Plan.FileSet.ConfigPaths...)
	copyRes.PlanResult.Plan.FileSet.LogPaths = append([]string{}, res.PlanResult.Plan.FileSet.LogPaths...)
	copyRes.ArtifactResult.Artifacts.Manifest.Files = append([]string{}, res.ArtifactResult.Artifacts.Manifest.Files...)
	copyRes.ArtifactResult.Artifacts.State.Instances = append([]xrayruntime.GeneratedInstance{}, res.ArtifactResult.Artifacts.State.Instances...)
	copyRes.LaunchResult.LaunchSpecs = append([]runtime.LaunchSpec{}, res.LaunchResult.LaunchSpecs...)
	copyRes.LaunchResult.PIDs = append([]int{}, res.LaunchResult.PIDs...)
	copyRes.LaunchResult.Errors = append([]string{}, res.LaunchResult.Errors...)
	return &copyRes
}
