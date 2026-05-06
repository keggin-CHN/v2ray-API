package app

import (
	"api-v2ray/internal/model"
	"api-v2ray/internal/runtime"
	"api-v2ray/internal/xrayruntime"
)

func buildRuntimePlan(cfg *model.Config, layout runtime.Layout, generated []xrayruntime.GeneratedInstance) (runtime.Plan, error) {
	plan := runtime.BuildPlan(layout, generated, cfg.Runtime.XrayBinary)
	return plan, nil
}

func writeRuntimePlanArtifacts(layout runtime.Layout, plan runtime.Plan) error {
	if err := runtime.WriteArtifacts(layout, plan.Artifacts()); err != nil {
		return wrapStep(stepWriteRuntimeArtifacts, err)
	}
	return nil
}
