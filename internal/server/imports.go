package server

import (
	"context"
	"fmt"
	"strings"

	"api-v2ray/internal/model"
	"api-v2ray/internal/subscription"
	"api-v2ray/internal/xrayruntime"
)

type ImportURIRequest struct {
	RawURI string `json:"raw_uri"`
}

type ImportSubscriptionRequest struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	URL    string `json:"url"`
	Format string `json:"format"`
	Text   string `json:"text"`
}

type ImportPreviewResponse struct {
	Format string            `json:"format"`
	Nodes  []model.ProxyNode `json:"nodes"`
}

func buildNodeFromURI(raw string) (model.ProxyNode, error) {
	raw = strings.TrimSpace(raw)
	parsed, ok := xrayruntime.ParseNode(raw)
	if !ok {
		return model.ProxyNode{}, fmt.Errorf("unsupported or invalid uri")
	}
	id := sanitizeID(parsed.Name)
	if id == "" {
		id = sanitizeID(parsed.Scheme + "-" + parsed.Host)
	}
	if id == "" {
		id = parsed.Scheme + "-node"
	}
	return model.ProxyNode{
		ID:             id,
		Name:           defaultNodeString(parsed.Name, id),
		Scheme:         parsed.Scheme,
		Host:           parsed.Host,
		Port:           parsed.Port,
		SubscriptionID: "manual",
		Tags:           inferNodeTags(parsed),
		RawURI:         raw,
	}, nil
}

func defaultNodeString(v, fallback string) string {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	return v
}

func inferNodeTags(parsed xrayruntime.ParsedNode) []string {
	var tags []string
	for _, v := range []string{parsed.Scheme, parsed.Security, parsed.Network, parsed.Flow} {
		v = strings.TrimSpace(strings.ToLower(v))
		if v != "" {
			tags = append(tags, strings.ReplaceAll(v, " ", "-"))
		}
	}
	return tags
}

func sanitizeID(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	s = strings.ReplaceAll(s, " ", "-")
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			b.WriteRune(r)
		}
	}
	return strings.Trim(b.String(), "-_")
}

func (s *Server) previewSubscription(ctx context.Context, req ImportSubscriptionRequest) ([]model.ProxyNode, string, error) {
	subID := strings.TrimSpace(req.ID)
	if subID == "" {
		subID = "imported-subscription"
	}
	if strings.TrimSpace(req.Text) != "" {
		nodes, kind, err := subscription.ParseAny(subID, req.Text)
		return nodes, kind, err
	}
	svc := subscription.New()
	nodes, err := svc.Fetch(ctx, model.Subscription{ID: subID, Name: strings.TrimSpace(req.Name), URL: strings.TrimSpace(req.URL)})
	return nodes, defaultNodeString(strings.TrimSpace(req.Format), "remote"), err
}
