package xrayruntime

import (
	"encoding/base64"
	"encoding/json"
	"net/url"
	"strings"
)

type ParsedNode struct {
	Scheme        string
	Host          string
	Port          int
	UUID          string
	Password      string
	Network       string
	Security      string
	Flow          string
	Fingerprint   string
	PublicKey     string
	ShortID       string
	HeaderType    string
	ALPN          string
	HostHdr       string
	Path          string
	SNI           string
	Name          string
	RawURI        string
	Insecure      bool
	Obfs          string
	ObfsPassword  string
	ServiceName   string
	Authority     string
	Username      string
}

func ParseNode(raw string) (ParsedNode, bool) {
	if strings.HasPrefix(raw, "vmess://") {
		return parseVMessNode(raw)
	}
	u, err := url.Parse(raw)
	if err != nil {
		return NormalizeParsedNode(ParsedNode{}), false
	}
	pn := ParsedNode{
		Scheme:       u.Scheme,
		Host:         u.Hostname(),
		UUID:         u.User.Username(),
		Username:     u.User.Username(),
		Password:     u.User.Username(),
		Network:      queryValue(u, "type", "tcp"),
		Security:     queryValue(u, "security", ""),
		Flow:         queryValue(u, "flow", ""),
		Fingerprint:  queryValue(u, "fp", queryValue(u, "fingerprint", "")),
		PublicKey:    queryValue(u, "pbk", queryValue(u, "publicKey", "")),
		ShortID:      queryValue(u, "sid", queryValue(u, "shortId", "")),
		HeaderType:   queryValue(u, "headerType", ""),
		ALPN:         queryValue(u, "alpn", ""),
		HostHdr:      queryValue(u, "host", ""),
		Path:         queryValue(u, "path", ""),
		SNI:          queryValue(u, "sni", queryValue(u, "host", "")),
		Obfs:         queryValue(u, "obfs", ""),
		ObfsPassword: queryValue(u, "obfs-password", queryValue(u, "obfsPassword", "")),
		ServiceName:  queryValue(u, "serviceName", ""),
		Authority:    queryValue(u, "authority", ""),
		Name:         u.Fragment,
		RawURI:       raw,
	}
	if u.Port() != "" {
		pn.Port = atoiSafe(u.Port())
	}
	if pass, ok := userPassword(u); ok {
		pn.Password = pass
	}
	if truthy(queryValue(u, "allowInsecure", queryValue(u, "insecure", ""))) {
		pn.Insecure = true
	}
	if pn.Scheme == "trojan" && pn.Security == "" {
		pn.Security = "tls"
	}
	return NormalizeParsedNode(pn), true
}

func parseVMessNode(raw string) (ParsedNode, bool) {
	encoded := strings.TrimPrefix(raw, "vmess://")
	decoded := tryDecode(encoded)
	if decoded == "" {
		return ParsedNode{}, false
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(decoded), &m); err != nil {
		return ParsedNode{}, false
	}
	return NormalizeParsedNode(ParsedNode{
		Scheme:   "vmess",
		Host:     asString(m["add"], ""),
		Port:     parsePort(m["port"]),
		UUID:     asString(m["id"], ""),
		Network:  asString(m["net"], "tcp"),
		Security: asString(m["tls"], ""),
		HostHdr:  asString(m["host"], ""),
		Path:     asString(m["path"], ""),
		Name:     asString(m["ps"], ""),
		RawURI:   raw,
	}), true
}

func EncodeVMessJSON(payload string) string {
	return "vmess://" + base64.StdEncoding.EncodeToString([]byte(payload))
}

func queryValue(u *url.URL, key, fallback string) string {
	if u == nil {
		return fallback
	}
	v := u.Query().Get(key)
	if v == "" {
		return fallback
	}
	return v
}

func defaultString(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}

func asString(v any, fallback string) string {
	s, ok := v.(string)
	if !ok || s == "" {
		return fallback
	}
	return s
}

func parsePort(v any) int {
	s := asString(v, "")
	if s == "" {
		return 0
	}
	return atoiSafe(s)
}

func userPassword(u *url.URL) (string, bool) {
	if u == nil || u.User == nil {
		return "", false
	}
	return u.User.Password()
}

func truthy(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	return s == "1" || s == "true" || s == "yes" || s == "on"
}
