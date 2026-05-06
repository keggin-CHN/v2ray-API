package app

import (
	"api-v2ray/internal/model"
	"api-v2ray/internal/xrayruntime"
)

func BuildBootstrapFlatResult(nodes *NodeCollectionResult, runtimeResult *RuntimeStageResult) BootstrapFlatResult {
	res := BootstrapFlatResult{}
	if nodes != nil {
		res.Nodes = append([]model.ProxyNode{}, nodes.GenerationResult.Nodes...)
	}
	if runtimeResult != nil {
		res.GeneratedXRAY = append([]xrayruntime.GeneratedInstance{}, runtimeResult.GenerationResult.Generated...)
	}
	return res
}
