package app

import (
	"api-v2ray/internal/model"
	"api-v2ray/internal/xrayruntime"
)

func generateXrayInstances(cfg *model.Config, nodes []model.ProxyNode) ([]xrayruntime.GeneratedInstance, error) {
	mgr := xrayruntime.New(cfg.Runtime.Dir, cfg.Runtime.BasePort, cfg.Runtime.XrayBinary)
	generated, err := mgr.Generate(nodes)
	if err != nil {
		return nil, wrapStep(stepGenerateXrayInstances, err)
	}
	return generated, nil
}
