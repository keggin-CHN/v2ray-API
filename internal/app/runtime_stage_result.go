package app

type RuntimeStageResult struct {
	GenerationResult RuntimeGenerationResult `json:"generation_result"`
	PlanResult       RuntimePlanResult       `json:"plan_result"`
	ArtifactResult   RuntimeArtifactResult   `json:"artifact_result"`
	LaunchResult     RuntimeLaunchResult     `json:"launch_result"`
}
