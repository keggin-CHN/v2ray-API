package app

type BootstrapSummary struct {
	NodeCount       int `json:"node_count"`
	GeneratedCount  int `json:"generated_count"`
	FetchErrorCount int `json:"fetch_error_count"`
	LaunchSpecCount int `json:"launch_spec_count"`
}
