package config

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"api-v2ray/internal/model"
)

func parseYAMLCompat(src string) (model.Config, error) {

	var cfg model.Config
	lines := strings.Split(src, "\n")
	section := ""
	item := map[string]any{}

	flush := func() error {
		if len(item) == 0 || section == "" {
			item = map[string]any{}
			return nil
		}
		b, err := json.Marshal(item)
		if err != nil {
			return err
		}
		switch section {
		case "upstreams":
			decodeItem := map[string]any{}
			for k, v := range item {
				if k == "models" {
					continue
				}
				decodeItem[k] = v
			}
			b, err := json.Marshal(decodeItem)
			if err != nil {
				return err
			}
			var v model.Upstream
			if err := json.Unmarshal(b, &v); err != nil {
				return fmt.Errorf("upstream decode: %w", err)
			}
			if raw, ok := item["models"].(string); ok {
				v.Models = parseInlineStringArray(raw)
			}
			cfg.Upstreams = append(cfg.Upstreams, v)
		case "bindings":
			var v model.Binding
			if err := json.Unmarshal(b, &v); err != nil {
				return fmt.Errorf("binding decode: %w", err)
			}
			cfg.Bindings = append(cfg.Bindings, v)
		case "proxy_nodes":
			decodeItem := map[string]any{}
			for k, v := range item {
				if k == "tags" {
					continue
				}
				decodeItem[k] = v
			}
			b, err := json.Marshal(decodeItem)
			if err != nil {
				return err
			}
			var v model.ProxyNode
			if err := json.Unmarshal(b, &v); err != nil {
				return fmt.Errorf("proxy node decode: %w", err)
			}
			if raw, ok := item["tags"].(string); ok {
				v.Tags = parseInlineStringArray(raw)
			}
			cfg.ProxyNodes = append(cfg.ProxyNodes, v)
		case "subscriptions":
			var v model.Subscription
			if err := json.Unmarshal(b, &v); err != nil {
				return fmt.Errorf("subscription decode: %w", err)
			}
			cfg.Subscriptions = append(cfg.Subscriptions, v)
		}
		item = map[string]any{}
		return nil
	}

	for _, raw := range lines {
		line := strings.TrimRight(raw, " \t\r")
		trim := strings.TrimSpace(line)
		if trim == "" || strings.HasPrefix(trim, "#") {
			continue
		}
		if !strings.HasPrefix(line, " ") && strings.HasSuffix(trim, ":") {
			if err := flush(); err != nil {
				return cfg, err
			}
			section = strings.TrimSuffix(trim, ":")
			continue
		}
		if section == "server" || section == "runtime" {
			parts := strings.SplitN(trim, ":", 2)
			if len(parts) != 2 {
				continue
			}
			k := strings.TrimSpace(parts[0])
			v := cleanScalar(parts[1])
			switch section + "." + k {
			case "server.listen":
				cfg.Server.Listen = v
			case "server.admin_token":
				cfg.Server.AdminToken = v
			case "runtime.dir":
				cfg.Runtime.Dir = v
			case "runtime.xray_binary":
				cfg.Runtime.XrayBinary = v
			case "runtime.base_port":
				cfg.Runtime.BasePort, _ = strconv.Atoi(v)
			case "runtime.healthcheck_url":
				cfg.Runtime.HealthcheckURL = v
			case "runtime.subscription_cache_file":
				cfg.Runtime.SubscriptionCacheFile = v
			}
			continue
		}
		if strings.HasPrefix(trim, "- ") {
			if err := flush(); err != nil {
				return cfg, err
			}
			trim = strings.TrimPrefix(trim, "- ")
		}
		parts := strings.SplitN(trim, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		rawVal := cleanScalar(parts[1])
		switch rawVal {
		case "true":
			item[key] = true
		case "false":
			item[key] = false
		default:
			if n, err := strconv.Atoi(rawVal); err == nil {
				item[key] = n
			} else {
				item[key] = rawVal
			}
		}
	}
	if err := flush(); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func cleanScalar(v string) string {
	v = strings.TrimSpace(v)
	v = strings.Trim(v, `"`)
	v = strings.Trim(v, `'`)
	return v
}

func parseInlineStringArray(raw string) []string {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "[")
	raw = strings.TrimSuffix(raw, "]")
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = cleanScalar(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
