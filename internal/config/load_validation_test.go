package config

import (
	"strings"
	"testing"

	"api-v2ray/internal/model"
)

func TestValidateRejectsDuplicateUpstreamID(t *testing.T) {
	cfg := &model.Config{
		Upstreams: []model.Upstream{
			{ID: "u1", BaseURL: "https://api.openai.com/v1"},
			{ID: "u1", BaseURL: "https://example.com/v1"},
		},
	}
	err := validate(cfg)
	if err == nil || !strings.Contains(err.Error(), "duplicate upstream id") {
		t.Fatalf("expected duplicate upstream id error, got: %v", err)
	}
}

func TestValidateRejectsInvalidUpstreamBaseURL(t *testing.T) {
	cfg := &model.Config{
		Upstreams: []model.Upstream{
			{ID: "u1", BaseURL: "not-a-url"},
		},
	}
	err := validate(cfg)
	if err == nil || !strings.Contains(err.Error(), "base_url is invalid") {
		t.Fatalf("expected invalid base_url error, got: %v", err)
	}
}

func TestValidateRejectsBindingUnknownNode(t *testing.T) {
	cfg := &model.Config{
		Upstreams: []model.Upstream{
			{ID: "u1", BaseURL: "https://api.openai.com/v1"},
		},
		ProxyNodes: []model.ProxyNode{
			{ID: "n1"},
		},
		Bindings: []model.Binding{
			{ID: "b1", UpstreamID: "u1", NodeID: "n404"},
		},
	}
	err := validate(cfg)
	if err == nil || !strings.Contains(err.Error(), "unknown node_id") {
		t.Fatalf("expected unknown node_id error, got: %v", err)
	}
}

func TestValidateAcceptsConsistentReferences(t *testing.T) {
	cfg := &model.Config{
		Upstreams: []model.Upstream{
			{ID: "u1", BaseURL: "https://api.openai.com/v1", BindingID: "b1"},
		},
		ProxyNodes: []model.ProxyNode{
			{ID: "n1"},
		},
		Bindings: []model.Binding{
			{ID: "b1", UpstreamID: "u1", NodeID: "n1"},
		},
		Failover: model.FailoverConfig{
			CooldownSteps: []model.CooldownStep{
				{AfterFailures: 3, DurationSeconds: 30},
				{AfterFailures: 1, DurationSeconds: 10},
			},
		},
	}
	if err := validate(cfg); err != nil {
		t.Fatalf("expected valid config, got err: %v", err)
	}
	if got := cfg.Failover.CooldownSteps[0].AfterFailures; got != 1 {
		t.Fatalf("expected cooldown steps to be sorted, first after_failures=1, got: %d", got)
	}
}