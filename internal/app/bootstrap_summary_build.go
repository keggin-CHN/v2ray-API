package app

func BuildBootstrapSummary(res *BootstrapResult) BootstrapSummary {
	if res == nil {
		return BootstrapSummary{}
	}
	summary := BootstrapSummary{}
	summary.NodeCount = len(res.FlatResult.Nodes)
	summary.GeneratedCount = len(res.FlatResult.GeneratedXRAY)
	if res.NodeStage != nil {
		summary.FetchErrorCount = len(res.NodeStage.FetchResult.Errors)
	}
	if res.RuntimeStage != nil {
		summary.LaunchSpecCount = len(res.RuntimeStage.LaunchResult.LaunchSpecs)
	}
	return summary
}
