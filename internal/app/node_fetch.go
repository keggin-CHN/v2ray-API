package app

import (
	"context"
	"fmt"

	"api-v2ray/internal/model"
	"api-v2ray/internal/subscription"
)

func fetchSubscriptionNodes(ctx context.Context, svc *subscription.Service, sub model.Subscription) ([]model.ProxyNode, error) {
	nodes, err := svc.Fetch(ctx, sub)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", sub.ID, err)
	}
	return nodes, nil
}
