package subscription

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"api-v2ray/internal/model"
)

func ParseAny(subscriptionID, text string) ([]model.ProxyNode, string, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, "empty", nil
	}
	if strings.HasPrefix(text, "{") {
		nodes, err := parseSingboxLike(subscriptionID, text)
		if err == nil && len(nodes) > 0 {
			return nodes, "sing-box", nil
		}
	}
	if looksLikeClash(text) {
		nodes, err := parseClashLike(subscriptionID, text)
		if err == nil && len(nodes) > 0 {
			return nodes, "clash", nil
		}
	}
	nodes, err := ParseSubscriptionText(subscriptionID, text)
	if err != nil {
		return nil, "uri-lines", err
	}
	return nodes, "uri-lines", nil
}

func looksLikeClash(text string) bool {
	lower := strings.ToLower(text)
	return strings.Contains(lower, "proxies:") || strings.Contains(lower, "proxy-providers:")
}

func parseSingboxLike(subscriptionID, text string) ([]model.ProxyNode, error) {
	var root map[string]any
	if err := json.Unmarshal([]byte(text), &root); err != nil {
		return nil, err
	}
	outbounds, _ := root["outbounds"].([]any)
	var nodes []model.ProxyNode
	for idx, item := range outbounds {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		typ := asString(m["type"])
		server := asString(m["server"])
		if typ == "" || server == "" {
			continue
		}
		port := asInt(m["server_port"])
		tag := asString(m["tag"])
		raw := buildURIFromMap(typ, m)
		nodes = append(nodes, model.ProxyNode{
			ID:             fmt.Sprintf("%s-%s-%d", subscriptionID, normalizeScheme(typ), idx+1),
			Name:           defaultName(tag, typ, idx),
			Scheme:         normalizeScheme(typ),
			Host:           server,
			Port:           port,
			SubscriptionID: subscriptionID,
			Tags:           inferTags(tag),
			RawURI:         raw,
		})
	}
	return nodes, nil
}

func parseClashLike(subscriptionID, text string) ([]model.ProxyNode, error) {
	lines := strings.Split(text, "\n")
	var nodes []model.ProxyNode
	current := map[string]string{}
	idx := 0
	emit := func() {
		if current["type"] == "" || current["server"] == "" {
			current = map[string]string{}
			return
		}
		idx++
		scheme := normalizeScheme(current["type"])
		name := current["name"]
		port := 0
		fmt.Sscanf(current["port"], "%d", &port)
		raw := buildURIFromKV(scheme, current)
		nodes = append(nodes, model.ProxyNode{
			ID:             fmt.Sprintf("%s-%s-%d", subscriptionID, scheme, idx),
			Name:           defaultName(name, scheme, idx-1),
			Scheme:         scheme,
			Host:           current["server"],
			Port:           port,
			SubscriptionID: subscriptionID,
			Tags:           inferTags(name),
			RawURI:         raw,
		})
		current = map[string]string{}
	}
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "- ") {
			if len(current) > 0 {
				emit()
			}
			trimmed = strings.TrimPrefix(trimmed, "- ")
		}
		if !strings.Contains(trimmed, ":") {
			continue
		}
		parts := strings.SplitN(trimmed, ":", 2)
		key := strings.TrimSpace(parts[0])
		val := strings.Trim(strings.TrimSpace(parts[1]), `"'`)
		if key == "name" || key == "type" || key == "server" || key == "port" || key == "uuid" || key == "password" || key == "cipher" || key == "sni" {
			current[key] = val
		}
	}
	if len(current) > 0 {
		emit()
	}
	return nodes, nil
}

func buildURIFromMap(typ string, m map[string]any) string {
	scheme := normalizeScheme(typ)
	server := asString(m["server"])
	port := asInt(m["server_port"])
	tag := url.QueryEscape(asString(m["tag"]))
	switch scheme {
	case "vless", "vmess", "tuic":
		uuid := asString(m["uuid"])
		return fmt.Sprintf("%s://%s@%s:%d#%s", scheme, uuid, server, port, tag)
	case "trojan", "hysteria2", "naive", "http", "socks", "ss":
		pass := asString(m["password"])
		if pass == "" {
			pass = asString(m["password"])
		}
		return fmt.Sprintf("%s://%s@%s:%d#%s", scheme, pass, server, port, tag)
	default:
		return fmt.Sprintf("%s://%s:%d#%s", scheme, server, port, tag)
	}
}

func buildURIFromKV(scheme string, kv map[string]string) string {
	server := kv["server"]
	port := kv["port"]
	name := url.QueryEscape(kv["name"])
	switch scheme {
	case "vless", "vmess", "tuic":
		return fmt.Sprintf("%s://%s@%s:%s#%s", scheme, kv["uuid"], server, port, name)
	case "trojan", "hysteria2", "naive", "http", "socks", "ss":
		secret := kv["password"]
		if secret == "" {
			secret = kv["cipher"]
		}
		return fmt.Sprintf("%s://%s@%s:%s#%s", scheme, secret, server, port, name)
	default:
		return fmt.Sprintf("%s://%s:%s#%s", scheme, server, port, name)
	}
}

func asString(v any) string {
	s, _ := v.(string)
	return strings.TrimSpace(s)
}

func asInt(v any) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case string:
		var out int
		fmt.Sscanf(n, "%d", &out)
		return out
	default:
		return 0
	}
}

func defaultName(name, scheme string, idx int) string {
	if strings.TrimSpace(name) != "" {
		return name
	}
	return fmt.Sprintf("%s-%d", scheme, idx+1)
}
