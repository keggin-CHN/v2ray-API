package app

import (
	"context"
	"fmt"
	"strings"

	"api-v2ray/internal/model"
	"api-v2ray/internal/subscription"
)

func CollectNodes(ctx context.Context, cfg *model.Config, svc *subscription.Service) ([]model.ProxyNode, []string, error) {
	nodes := append([]model.ProxyNode{}, cfg.ProxyNodes...)
	var errs []string
	for _, sub := range cfg.Subscriptions {
		fetched, err := fetchSubscriptionNodes(ctx, svc, sub)
		if err != nil {
			errs = append(errs, err.Error())
			continue
		}
		nodes = mergeFetchedNodes(nodes, fetched)
	}
	if len(errs) > 0 {
		return nodes, append([]string{}, errs...), fmt.Errorf("subscription fetch failures: %s", strings.Join(errs, "; "))
	}
	return nodes, nil, nil
}
