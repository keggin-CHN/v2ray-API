package app

import (
	"context"

	"api-v2ray/internal/model"
)

type BootstrapResult struct {
	Summary      BootstrapSummary      `json:"summary"`
	FlatResult   BootstrapFlatResult   `json:"flat_result"`
	NodeStage    *NodeCollectionResult `json:"node_stage,omitempty"`
	RuntimeStage *RuntimeStageResult   `json:"runtime_stage,omitempty"`
}

func Bootstrap(ctx context.Context, cfg *model.Config) (*BootstrapResult, error) {
	layout, err := bootstrapLayout(cfg)
	if err != nil {
		return nil, err
	}

	nodeResult, err := bootstrapNodes(ctx, cfg)
	if err != nil {
		return NewBootstrapResult(nodeResult, nil), err
	}

	runtimeResult, err := bootstrapRuntimeOutputs(cfg, layout, nodeResult.GenerationResult.Nodes)
	if err != nil {
		return NewBootstrapResult(nodeResult, runtimeResult), err
	}

	return NewBootstrapResult(nodeResult, runtimeResult), nil
}
