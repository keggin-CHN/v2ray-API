package app

import "api-v2ray/internal/model"

func mergeFetchedNodes(base []model.ProxyNode, fetched []model.ProxyNode) []model.ProxyNode {
	out := append([]model.ProxyNode{}, base...)
	out = append(out, fetched...)
	return out
}
