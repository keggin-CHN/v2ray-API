package subscription

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"api-v2ray/internal/model"
)

type Service struct{}

func New() *Service { return &Service{} }

func (s *Service) Fetch(ctx context.Context, sub model.Subscription) ([]model.ProxyNode, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, sub.URL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch subscription %s: %w", sub.ID, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return nil, fmt.Errorf("fetch subscription %s: status %d: %s", sub.ID, resp.StatusCode, strings.TrimSpace(string(msg)))
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, err
	}
	return ParseSubscriptionText(sub.ID, string(body))
}

func ParseSubscriptionText(subscriptionID, text string) ([]model.ProxyNode, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, nil
	}

	decoded := tryBase64(text)
	if decoded != "" && strings.Contains(decoded, "://") {
		text = decoded
	}

	lines := strings.Split(text, "\n")
	var out []model.ProxyNode
	for idx, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		node, err := parseNodeLine(subscriptionID, idx, line)
		if err != nil {
			continue
		}
		out = append(out, node)
	}
	return out, nil
}

func parseNodeLine(subscriptionID string, idx int, line string) (model.ProxyNode, error) {
	if strings.HasPrefix(line, "vmess://") {
		return parseVMess(subscriptionID, idx, line)
	}
	if strings.HasPrefix(line, "vless://") || strings.HasPrefix(line, "trojan://") || strings.HasPrefix(line, "ss://") || strings.HasPrefix(line, "hy2://") || strings.HasPrefix(line, "hysteria2://") || strings.HasPrefix(line, "tuic://") || strings.HasPrefix(line, "naive://") || strings.HasPrefix(line, "socks://") || strings.HasPrefix(line, "socks5://") || strings.HasPrefix(line, "http://") || strings.HasPrefix(line, "https://") {
		u, err := url.Parse(line)
		if err != nil {
			return model.ProxyNode{}, err
		}
		name := u.Fragment
		if name == "" {
			name = fmt.Sprintf("%s-%d", u.Scheme, idx+1)
		}
		port := 0
		if u.Port() != "" {
			fmt.Sscanf(u.Port(), "%d", &port)
		}
		return model.ProxyNode{
			ID:             fmt.Sprintf("%s-%s-%d", subscriptionID, normalizeScheme(u.Scheme), idx+1),
			Name:           name,
			Scheme:         normalizeScheme(u.Scheme),
			Host:           u.Hostname(),
			Port:           port,
			SubscriptionID: subscriptionID,
			Tags:           inferTags(name),
			RawURI:         line,
		}, nil
	}
	return model.ProxyNode{}, fmt.Errorf("unsupported line")
}

func parseVMess(subscriptionID string, idx int, line string) (model.ProxyNode, error) {
	raw := strings.TrimPrefix(line, "vmess://")
	decoded := tryBase64(raw)
	if decoded == "" {
		return model.ProxyNode{}, fmt.Errorf("invalid vmess base64")
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(decoded), &m); err != nil {
		return model.ProxyNode{}, err
	}
	name, _ := m["ps"].(string)
	host, _ := m["add"].(string)
	port := 0
	switch v := m["port"].(type) {
	case string:
		fmt.Sscanf(v, "%d", &port)
	case float64:
		port = int(v)
	}
	return model.ProxyNode{
		ID:             fmt.Sprintf("%s-vmess-%d", subscriptionID, idx+1),
		Name:           name,
		Scheme:         "vmess",
		Host:           host,
		Port:           port,
		SubscriptionID: subscriptionID,
		Tags:           inferTags(name),
		RawURI:         line,
	}, nil
}

func tryBase64(s string) string {
	s = strings.TrimSpace(s)
	for _, enc := range []*base64.Encoding{base64.StdEncoding, base64.RawStdEncoding, base64.URLEncoding, base64.RawURLEncoding} {
		b, err := enc.DecodeString(s)
		if err == nil {
			return string(b)
		}
	}
	return ""
}

func inferTags(name string) []string {
	lower := strings.ToLower(name)
	var tags []string
	for _, token := range []string{"hk", "hong kong", "jp", "japan", "sg", "singapore", "us", "tw", "kr", "de", "uk", "my", "reality", "vision", "hy2", "hysteria2", "tuic"} {
		if strings.Contains(lower, token) {
			tags = append(tags, strings.ReplaceAll(token, " ", "-"))
		}
	}
	return tags
}

func normalizeScheme(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "hy2" {
		return "hysteria2"
	}
	if s == "socks5" {
		return "socks"
	}
	if s == "https" {
		return "http"
	}
	return s
}
