package xrayruntime

import (
	"fmt"
	"strings"

	"api-v2ray/internal/model"
)

func BuildOutbound(node model.ProxyNode) map[string]any {
	parsed, ok := ParseNode(node.RawURI)
	if ok {
		switch parsed.Scheme {
		case "vless":
			return buildVLESSOutboundParsed(parsed)
		case "trojan":
			return buildTrojanOutboundParsed(parsed)
		case "vmess":
			return buildVMessOutboundParsed(parsed)
		case "hysteria2":
			return buildHysteria2OutboundParsed(parsed)
		case "tuic":
			return buildTUICOutboundParsed(parsed)
		case "socks", "socks5":
			return buildSOCKSOutboundParsed(parsed)
		case "http":
			return buildHTTPOutboundParsed(parsed)
		case "naive":
			return buildHTTPOutboundParsed(parsed)
		case "ss":
			return buildShadowsocksFallback(node, parsed)
		}
	}
	return map[string]any{
		"tag":      "direct",
		"protocol": "freedom",
		"settings": map[string]any{},
	}
}

func buildVLESSOutboundParsed(node ParsedNode) map[string]any {
	user := map[string]any{
		"id":         node.UUID,
		"encryption": defaultString(queryOrNone(node), "none"),
	}
	if node.Flow != "" {
		user["flow"] = node.Flow
	}
	stream := map[string]any{
		"network":  node.Network,
		"security": node.Security,
	}
	out := map[string]any{
		"tag":      "proxy",
		"protocol": "vless",
		"settings": map[string]any{
			"vnext": []any{map[string]any{
				"address": node.Host,
				"port":    node.Port,
				"users":   []any{user},
			}},
		},
		"streamSettings": stream,
	}
	if node.Network == "ws" {
		stream["wsSettings"] = map[string]any{
			"path":    defaultString(node.Path, "/"),
			"headers": map[string]any{"Host": node.HostHdr},
		}
	}
	if node.Network == "grpc" {
		stream["grpcSettings"] = map[string]any{"serviceName": node.ServiceName}
	}
	if node.Security == "tls" {
		stream["tlsSettings"] = map[string]any{
			"serverName": node.SNI,
			"allowInsecure": node.Insecure,
		}
	}
	if node.Security == "reality" {
		stream["realitySettings"] = map[string]any{
			"serverName":    node.SNI,
			"fingerprint":   node.Fingerprint,
			"publicKey":     node.PublicKey,
			"shortId":       node.ShortID,
			"spiderX":       defaultString(node.Path, "/"),
			"show":          false,
		}
	}
	if node.HeaderType != "" && node.Network == "tcp" {
		stream["tcpSettings"] = map[string]any{"header": map[string]any{"type": node.HeaderType}}
	}
	return out
}

func buildTrojanOutboundParsed(node ParsedNode) map[string]any {
	stream := map[string]any{
		"network":  node.Network,
		"security": node.Security,
	}
	if node.Security == "tls" || node.Security == "reality" {
		stream["tlsSettings"] = map[string]any{
			"serverName":    node.SNI,
			"allowInsecure": node.Insecure,
		}
	}
	return map[string]any{
		"tag":      "proxy",
		"protocol": "trojan",
		"settings": map[string]any{
			"servers": []any{map[string]any{
				"address":  node.Host,
				"port":     node.Port,
				"password": node.Password,
			}},
		},
		"streamSettings": stream,
	}
}

func buildVMessOutboundParsed(node ParsedNode) map[string]any {
	out := map[string]any{
		"tag":      "proxy",
		"protocol": "vmess",
		"settings": map[string]any{
			"vnext": []any{map[string]any{
				"address": node.Host,
				"port":    node.Port,
				"users": []any{map[string]any{
					"id":       node.UUID,
					"alterId":  0,
					"security": "auto",
				}},
			}},
		},
		"streamSettings": map[string]any{
			"network":  node.Network,
			"security": node.Security,
		},
	}
	if node.Network == "ws" {
		out["streamSettings"].(map[string]any)["wsSettings"] = map[string]any{
			"path":    defaultString(node.Path, "/"),
			"headers": map[string]any{"Host": node.HostHdr},
		}
	}
	return out
}

func buildHysteria2OutboundParsed(node ParsedNode) map[string]any {
	settings := map[string]any{
		"version": 2,
		"address": node.Host,
		"port":    node.Port,
	}
	stream := map[string]any{
		"network":  "hysteria",
		"security": "tls",
		"tlsSettings": map[string]any{
			"serverName":    defaultString(node.SNI, node.Host),
			"allowInsecure": node.Insecure,
			"alpn":          splitCSV(node.ALPN),
		},
		"hysteriaSettings": map[string]any{
			"version": 2,
			"auth":    node.Password,
		},
	}
	if node.Obfs != "" {
		stream["hysteriaSettings"].(map[string]any)["obfs"] = map[string]any{
			"type":     node.Obfs,
			"password": node.ObfsPassword,
		}
	}
	return map[string]any{
		"tag":            "proxy",
		"protocol":       "hysteria",
		"settings":       settings,
		"streamSettings": stream,
	}
}

func buildTUICOutboundParsed(node ParsedNode) map[string]any {
	return map[string]any{
		"tag":      "proxy",
		"protocol": "tuic",
		"settings": map[string]any{
			"servers": []any{map[string]any{
				"address":  node.Host,
				"port":     node.Port,
				"uuid":     node.UUID,
				"password": node.Password,
				"alpn":     splitCSV(node.ALPN),
				"sni":      defaultString(node.SNI, node.Host),
			}},
		},
	}
}

func buildSOCKSOutboundParsed(node ParsedNode) map[string]any {
	server := map[string]any{"address": node.Host, "port": node.Port}
	if node.Username != "" || node.Password != "" {
		server["users"] = []any{map[string]any{"user": node.Username, "pass": node.Password}}
	}
	return map[string]any{
		"tag":      "proxy",
		"protocol": "socks",
		"settings": map[string]any{"servers": []any{server}},
	}
}

func buildHTTPOutboundParsed(node ParsedNode) map[string]any {
	server := map[string]any{"address": node.Host, "port": node.Port}
	if node.Username != "" || node.Password != "" {
		server["users"] = []any{map[string]any{"user": node.Username, "pass": node.Password}}
	}
	return map[string]any{
		"tag":      "proxy",
		"protocol": "http",
		"settings": map[string]any{"servers": []any{server}},
	}
}

func buildShadowsocksFallback(src model.ProxyNode, node ParsedNode) map[string]any {
	return map[string]any{
		"tag":      "proxy",
		"protocol": "shadowsocks",
		"settings": map[string]any{
			"servers": []any{map[string]any{
				"address":  defaultString(node.Host, src.Host),
				"port":     defaultInt(node.Port, src.Port),
				"password": node.Password,
				"method":   extractSSMethod(src.RawURI),
			}},
		},
	}
}

func extractSSMethod(raw string) string {
	if !strings.HasPrefix(raw, "ss://") {
		return "aes-256-gcm"
	}
	trimmed := strings.TrimPrefix(raw, "ss://")
	if i := strings.Index(trimmed, "#"); i >= 0 {
		trimmed = trimmed[:i]
	}
	if i := strings.Index(trimmed, "@"); i < 0 {
		decoded := tryDecode(trimmed)
		if decoded != "" {
			trimmed = decoded
		}
	}
	parts := strings.SplitN(trimmed, ":", 2)
	if len(parts) == 0 || parts[0] == "" {
		return "aes-256-gcm"
	}
	return parts[0]
}

func queryOrNone(node ParsedNode) string { return "none" }
func defaultInt(v, fallback int) int {
	if v == 0 { return fallback }
	return v
}
func splitCSV(v string) []string {
	v = strings.TrimSpace(v)
	if v == "" { return nil }
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" { out = append(out, p) }
	}
	return out
}
func atoiSafe(s string) int {
	var n int
	_, _ = fmt.Sscanf(s, "%d", &n)
	return n
}
func tryDecode(s string) string { return decodeBase64Any(s) }
