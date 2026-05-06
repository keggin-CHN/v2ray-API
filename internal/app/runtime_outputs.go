package app

import (
	"api-v2ray/internal/model"
	"api-v2ray/internal/runtime"
)

func GenerateRuntimeOutputs(cfg *model.Config, layout runtime.Layout, nodes []model.ProxyNode) (*RuntimeStageResult, error) {
	generated, err := generateXrayInstances(cfg, nodes)
	if err != nil {
		return nil, err
	}
	plan, err := buildRuntimePlan(cfg, layout, generated)
	if err != nil {
		return nil, wrapStep(stepBuildRuntimePlan, err)
	}
	if err := writeRuntimePlanArtifacts(layout, plan); err != nil {
		return nil, err
	}
	launchResult := BuildRuntimeLaunchResult(cfg, plan.LaunchSpecs)
	return &RuntimeStageResult{GenerationResult: RuntimeGenerationResult{Generated: generated}, PlanResult: RuntimePlanResult{Plan: plan}, ArtifactResult: RuntimeArtifactResult{Artifacts: plan.Artifacts()}, LaunchResult: launchResult}, nil
}
