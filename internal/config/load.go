package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"api-v2ray/internal/model"
)

func Load(path string) (*model.Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg model.Config
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".json":
		if err := json.Unmarshal(b, &cfg); err != nil {
			return nil, fmt.Errorf("parse json config: %w", err)
		}
	case ".yaml", ".yml":
		cfg, err = parseYAMLCompat(string(b))
		if err != nil {
			return nil, fmt.Errorf("parse yaml config: %w", err)
		}
	default:
		if err := json.Unmarshal(b, &cfg); err != nil {
			return nil, fmt.Errorf("unsupported config extension %s, and json fallback failed: %w", ext, err)
		}
	}

	applyDefaults(&cfg)
	if err := validate(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func applyDefaults(cfg *model.Config) {
	if cfg.Server.Listen == "" {
		cfg.Server.Listen = ":8080"
	}
	if cfg.Runtime.Dir == "" {
		cfg.Runtime.Dir = "./runtime"
	}
	if cfg.Runtime.BasePort == 0 {
		cfg.Runtime.BasePort = 21000
	}
	if cfg.Runtime.HealthcheckURL == "" {
		cfg.Runtime.HealthcheckURL = "https://api.openai.com"
	}
	if cfg.Runtime.SubscriptionCacheFile == "" {
		cfg.Runtime.SubscriptionCacheFile = "./runtime/subscriptions/cache.json"
	}
	if len(cfg.Failover.CooldownSteps) == 0 {
		cfg.Failover.CooldownSteps = []model.CooldownStep{
			{AfterFailures: 1, DurationSeconds: 10},
			{AfterFailures: 2, DurationSeconds: 30},
			{AfterFailures: 3, DurationSeconds: 90},
			{AfterFailures: 5, DurationSeconds: 300},
		}
	}
	for i := range cfg.Upstreams {
		if cfg.Upstreams[i].TimeoutSeconds == 0 {
			cfg.Upstreams[i].TimeoutSeconds = 120
		}
	}
}

func validate(cfg *model.Config) error {
	if len(cfg.Upstreams) == 0 {
		return errors.New("no upstreams configured")
	}

	upstreamIDs := make(map[string]struct{}, len(cfg.Upstreams))
	for i, upstream := range cfg.Upstreams {
		id := strings.TrimSpace(upstream.ID)
		if id == "" {
			return fmt.Errorf("upstreams[%d].id must not be empty", i)
		}
		if _, exists := upstreamIDs[id]; exists {
			return fmt.Errorf("duplicate upstream id: %s", id)
		}
		upstreamIDs[id] = struct{}{}

		if err := validateHTTPBaseURL(upstream.BaseURL); err != nil {
			return fmt.Errorf("upstreams[%d].base_url is invalid: %w", i, err)
		}
	}

	nodeIDs := make(map[string]struct{}, len(cfg.ProxyNodes))
	for i, node := range cfg.ProxyNodes {
		id := strings.TrimSpace(node.ID)
		if id == "" {
			return fmt.Errorf("proxy_nodes[%d].id must not be empty", i)
		}
		if _, exists := nodeIDs[id]; exists {
			return fmt.Errorf("duplicate proxy node id: %s", id)
		}
		nodeIDs[id] = struct{}{}
	}

	bindingIDs := make(map[string]struct{}, len(cfg.Bindings))
	for i, binding := range cfg.Bindings {
		id := strings.TrimSpace(binding.ID)
		if id == "" {
			return fmt.Errorf("bindings[%d].id must not be empty", i)
		}
		if _, exists := bindingIDs[id]; exists {
			return fmt.Errorf("duplicate binding id: %s", id)
		}
		bindingIDs[id] = struct{}{}

		if _, ok := upstreamIDs[strings.TrimSpace(binding.UpstreamID)]; !ok {
			return fmt.Errorf("bindings[%d].unknown upstream_id: %s", i, binding.UpstreamID)
		}
		if _, ok := nodeIDs[strings.TrimSpace(binding.NodeID)]; !ok {
			return fmt.Errorf("bindings[%d].unknown node_id: %s", i, binding.NodeID)
		}
	}

	for i, upstream := range cfg.Upstreams {
		bindingID := strings.TrimSpace(upstream.BindingID)
		if bindingID == "" {
			continue
		}
		if _, ok := bindingIDs[bindingID]; !ok {
			return fmt.Errorf("upstreams[%d].unknown binding_id: %s", i, upstream.BindingID)
		}
	}

	for _, step := range cfg.Failover.CooldownSteps {
		if step.AfterFailures <= 0 {
			return errors.New("failover.cooldown_steps.after_failures must be > 0")
		}
		if step.DurationSeconds < 0 {
			return errors.New("failover.cooldown_steps.duration_seconds must be >= 0")
		}
	}
	sort.Slice(cfg.Failover.CooldownSteps, func(i, j int) bool {
		return cfg.Failover.CooldownSteps[i].AfterFailures < cfg.Failover.CooldownSteps[j].AfterFailures
	})
	return nil
}

func validateHTTPBaseURL(raw string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return errors.New("must not be empty")
	}
	u, err := url.ParseRequestURI(raw)
	if err != nil {
		return err
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("unsupported scheme %q", u.Scheme)
	}
	if u.Host == "" {
		return errors.New("host is required")
	}
	return nil
}
