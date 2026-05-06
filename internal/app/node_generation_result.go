package app

import "api-v2ray/internal/model"

type NodeGenerationResult struct {
	Nodes []model.ProxyNode `json:"nodes"`
}
