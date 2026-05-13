package upstream

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"api-v2ray/internal/model"
	"api-v2ray/internal/proxyruntime"
)

type Client struct {
	mu      sync.RWMutex
	clients map[string]*http.Client
	stats   ClientStats
}

type ClientStats struct {
	RequestsTotal int64 `json:"requests_total"`
	FailuresTotal int64 `json:"failures_total"`
	Status2xx     int64 `json:"status_2xx"`
	Status4xx     int64 `json:"status_4xx"`
	Status5xx     int64 `json:"status_5xx"`
}

func New() *Client { return &Client{clients: make(map[string]*http.Client)} }

func (c *Client) ChatCompletionsRaw(ctx context.Context, incomingHeader http.Header, upstream model.Upstream, endpoint proxyruntime.Endpoint, body []byte) (*http.Response, error) {
	return c.postRaw(ctx, incomingHeader, strings.TrimRight(upstream.BaseURL, "/")+"/chat/completions", upstream, endpoint, body)
}

func (c *Client) EmbeddingsRaw(ctx context.Context, incomingHeader http.Header, upstream model.Upstream, endpoint proxyruntime.Endpoint, body []byte) (*http.Response, error) {
	return c.postRaw(ctx, incomingHeader, strings.TrimRight(upstream.BaseURL, "/")+"/embeddings", upstream, endpoint, body)
}

func (c *Client) postRaw(ctx context.Context, incomingHeader http.Header, target string, upstream model.Upstream, endpoint proxyruntime.Endpoint, body []byte) (*http.Response, error) {
	start := time.Now()
	atomic.AddInt64(&c.stats.RequestsTotal, 1)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, target, bytes.NewReader(body))
	if err != nil {
		atomic.AddInt64(&c.stats.FailuresTotal, 1)
		return nil, fmt.Errorf("build request: %w", err)
	}
	copyAllowedHeaders(httpReq.Header, incomingHeader)
	httpReq.Header.Set("Content-Type", "application/json")
	if upstream.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+upstream.APIKey)
	}
	client := c.getOrCreateClient(upstream, endpoint)
	resp, err := client.Do(httpReq)
	elapsed := time.Since(start)

	if err != nil {
		atomic.AddInt64(&c.stats.FailuresTotal, 1)
		log.Printf("upstream_request_fail upstream=%s endpoint=%s:%d cost_ms=%d err=%v", upstream.ID, endpoint.Host, endpoint.Port, elapsed.Milliseconds(), err)
		return nil, fmt.Errorf("request upstream via proxy: %w", err)
	}
	switch {
	case resp.StatusCode >= 500:
		atomic.AddInt64(&c.stats.Status5xx, 1)
	case resp.StatusCode >= 400:
		atomic.AddInt64(&c.stats.Status4xx, 1)
	case resp.StatusCode >= 200:
		atomic.AddInt64(&c.stats.Status2xx, 1)
	}
	log.Printf("upstream_request_done upstream=%s endpoint=%s:%d status=%d cost_ms=%d", upstream.ID, endpoint.Host, endpoint.Port, resp.StatusCode, elapsed.Milliseconds())
	return resp, nil
}

func (c *Client) getOrCreateClient(upstream model.Upstream, endpoint proxyruntime.Endpoint) *http.Client {
	key := fmt.Sprintf("%s:%d|%d", endpoint.Host, endpoint.Port, upstream.TimeoutSeconds)
	c.mu.RLock()
	if client, ok := c.clients[key]; ok {
		c.mu.RUnlock()
		return client
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()
	if client, ok := c.clients[key]; ok {
		return client
	}
	client := buildHTTPClient(upstream, endpoint)
	c.clients[key] = client
	return client
}

func (c *Client) InvalidateAll() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for k, client := range c.clients {
		if t, ok := client.Transport.(*http.Transport); ok {
			t.CloseIdleConnections()
		}
		delete(c.clients, k)
	}
}

var allowedHeaders = map[string]bool{
	"Accept":                       true,
	"Accept-Encoding":              true,
	"Accept-Language":              true,
	"User-Agent":                   true,
	"Openai-Beta":                  true,
	"X-Request-Id":                 true,
	"X-Stainless-Lang":             true,
	"X-Stainless-Package-Version":  true,
	"X-Stainless-Os":               true,
	"X-Stainless-Arch":             true,
	"X-Stainless-Runtime":          true,
	"X-Stainless-Runtime-Version":  true,
}

func copyAllowedHeaders(dst, src http.Header) {
	for k, values := range src {
		ck := http.CanonicalHeaderKey(k)
		if !allowedHeaders[ck] {
			continue
		}
		for _, v := range values {
			dst.Add(ck, v)
		}
	}
}

func buildHTTPClient(upstream model.Upstream, endpoint proxyruntime.Endpoint) *http.Client {
	proxyURL := &url.URL{
		Scheme: endpoint.Scheme,
		Host:   net.JoinHostPort(endpoint.Host, fmt.Sprintf("%d", endpoint.Port)),
	}
	timeout := time.Duration(upstream.TimeoutSeconds) * time.Second
	if timeout == 0 {
		timeout = 120 * time.Second
	}
	transport := &http.Transport{
		Proxy:                  http.ProxyURL(proxyURL),
		DialContext:            (&net.Dialer{Timeout: 15 * time.Second, KeepAlive: 30 * time.Second}).DialContext,
		MaxIdleConns:           64,
		MaxIdleConnsPerHost:    16,
		IdleConnTimeout:        90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout:  timeout,
		ExpectContinueTimeout:  1 * time.Second,
		ForceAttemptHTTP2:      true,
	}
	return &http.Client{Transport: transport, Timeout: timeout}
}

func (c *Client) StatsSnapshot() ClientStats {
	return ClientStats{
		RequestsTotal: atomic.LoadInt64(&c.stats.RequestsTotal),
		FailuresTotal: atomic.LoadInt64(&c.stats.FailuresTotal),
		Status2xx:     atomic.LoadInt64(&c.stats.Status2xx),
		Status4xx:     atomic.LoadInt64(&c.stats.Status4xx),
		Status5xx:     atomic.LoadInt64(&c.stats.Status5xx),
	}
}

func CopyResponse(w http.ResponseWriter, resp *http.Response) error {
	defer resp.Body.Close()
	for k, values := range resp.Header {
		for _, v := range values {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	buf := make([]byte, 32*1024)
	_, err := io.CopyBuffer(w, resp.Body, buf)
	return err
}
