package app

import "api-v2ray/internal/model"

func CopyNodeCollectionResult(res *NodeCollectionResult) *NodeCollectionResult {
	if res == nil {
		return nil
	}
	copyRes := *res
	copyRes.GenerationResult.Nodes = append([]model.ProxyNode{}, res.GenerationResult.Nodes...)
	copyRes.FetchResult.Errors = append([]string{}, res.FetchResult.Errors...)
	copyRes.ArtifactResult.SubscriptionCache = res.ArtifactResult.SubscriptionCache
	return &copyRes
}
