package app

type NodeCollectionResult struct {
	GenerationResult NodeGenerationResult `json:"generation_result"`
	FetchResult      NodeFetchResult      `json:"fetch_result"`
	ArtifactResult   NodeArtifactResult   `json:"artifact_result"`
}
