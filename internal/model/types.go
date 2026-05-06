package model

type ServerConfig struct {
	Listen     string `json:"listen"`
	AdminToken string `json:"admin_token"`
}

type RuntimeConfig struct {
	Dir                   string `json:"dir"`
	XrayBinary            string `json:"xray_binary"`
	BasePort              int    `json:"base_port"`
	HealthcheckURL        string `json:"healthcheck_url"`
	SubscriptionCacheFile string `json:"subscription_cache_file"`
}

type CooldownStep struct {
	AfterFailures   int `json:"after_failures"`
	DurationSeconds int `json:"duration_seconds"`
}

type FailoverConfig struct {
	CooldownSteps []CooldownStep `json:"cooldown_steps"`
}

type Upstream struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	BaseURL        string   `json:"base_url"`
	APIKey         string   `json:"api_key"`
	Models         []string `json:"models"`
	BindingID      string   `json:"binding_id"`
	Priority       int      `json:"priority"`
	Enabled        bool     `json:"enabled"`
	TimeoutSeconds int      `json:"timeout_seconds"`
}

type Binding struct {
	ID         string `json:"id"`
	UpstreamID string `json:"upstream_id"`
	NodeID     string `json:"node_id"`
	Mode       string `json:"mode"`
}

type ProxyNode struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Scheme         string   `json:"scheme"`
	Host           string   `json:"host"`
	Port           int      `json:"port"`
	SubscriptionID string   `json:"subscription_id"`
	Tags           []string `json:"tags"`
	RawURI         string   `json:"raw_uri"`
}

type Subscription struct {
	ID                     string `json:"id"`
	Name                   string `json:"name"`
	URL                    string `json:"url"`
	RefreshIntervalSeconds int    `json:"refresh_interval_seconds"`
}

type Config struct {
	Server        ServerConfig   `json:"server"`
	Runtime       RuntimeConfig  `json:"runtime"`
	Failover      FailoverConfig `json:"failover"`
	Upstreams     []Upstream     `json:"upstreams"`
	Bindings      []Binding      `json:"bindings"`
	ProxyNodes    []ProxyNode    `json:"proxy_nodes"`
	Subscriptions []Subscription `json:"subscriptions"`
}
