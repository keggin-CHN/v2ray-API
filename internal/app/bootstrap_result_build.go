package app

func NewBootstrapResult(nodes *NodeCollectionResult, runtimeResult *RuntimeStageResult) *BootstrapResult {
	res := &BootstrapResult{}
	res.FlatResult = BuildBootstrapFlatResult(nodes, runtimeResult)
	res.NodeStage = CopyNodeCollectionResult(nodes)
	res.RuntimeStage = CopyRuntimeStageResult(runtimeResult)
	res.Summary = BuildBootstrapSummary(res)
	return res
}
