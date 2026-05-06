package app

import (
	"context"

	"api-v2ray/internal/model"
	"api-v2ray/internal/runtime"
	"api-v2ray/internal/subscription"
)

func bootstrapLayout(cfg *model.Config) (runtime.Layout, error) {
	layout := runtime.NewLayout(cfg.Runtime.Dir, cfg.Runtime.SubscriptionCacheFile)
	if err := runtime.EnsureLayout(layout); err != nil {
		return runtime.Layout{}, wrapStep(stepEnsureRuntimeLayout, err)
	}
	return layout, nil
}

func bootstrapNodes(ctx context.Context, cfg *model.Config) (*NodeCollectionResult, error) {
	nodes, fetchErrors, err := CollectNodes(ctx, cfg, subscription.New())
	if err != nil {
		return &NodeCollectionResult{GenerationResult: NodeGenerationResult{Nodes: nodes}, FetchResult: NodeFetchResult{Errors: fetchErrors}, ArtifactResult: NodeArtifactResult{SubscriptionCache: cfg.Runtime.SubscriptionCacheFile}}, wrapStep(stepCollectNodes, err)
	}
	if err := runtime.WriteSubscriptionCache(cfg.Runtime.SubscriptionCacheFile, nodes); err != nil {
		return &NodeCollectionResult{GenerationResult: NodeGenerationResult{Nodes: nodes}, FetchResult: NodeFetchResult{Errors: fetchErrors}, ArtifactResult: NodeArtifactResult{SubscriptionCache: cfg.Runtime.SubscriptionCacheFile}}, wrapStep(stepWriteSubscriptionCache, err)
	}
	return &NodeCollectionResult{GenerationResult: NodeGenerationResult{Nodes: nodes}, FetchResult: NodeFetchResult{Errors: fetchErrors}, ArtifactResult: NodeArtifactResult{SubscriptionCache: cfg.Runtime.SubscriptionCacheFile}}, nil
}

func bootstrapRuntimeOutputs(cfg *model.Config, layout runtime.Layout, nodes []model.ProxyNode) (*RuntimeStageResult, error) {
	result, err := GenerateRuntimeOutputs(cfg, layout, nodes)
	if err != nil {
		return nil, wrapStep(stepGenerateRuntimeOutputs, err)
	}
	return result, nil
}
