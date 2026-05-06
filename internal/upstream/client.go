package upstream

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"api-v2ray/internal/model"
	"api-v2ray/internal/proxyruntime"
)

type Client struct{}

func New() *Client { return &Client{} }

func (c *Client) ChatCompletionsRaw(ctx context.Context, incomingHeader http.Header, upstream model.Upstream, endpoint proxyruntime.Endpoint, body []byte) (*http.Response, error) {
	return c.postRaw(ctx, incomingHeader, strings.TrimRight(upstream.BaseURL, "/")+"/chat/completions", upstream, endpoint, body)
}

func (c *Client) EmbeddingsRaw(ctx context.Context, incomingHeader http.Header, upstream model.Upstream, endpoint proxyruntime.Endpoint, body []byte) (*http.Response, error) {
	return c.postRaw(ctx, incomingHeader, strings.TrimRight(upstream.BaseURL, "/")+"/embeddings", upstream, endpoint, body)
}

func (c *Client) postRaw(ctx context.Context, incomingHeader http.Header, target string, upstream model.Upstream, endpoint proxyruntime.Endpoint, body []byte) (*http.Response, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, target, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	copyAllowedHeaders(httpReq.Header, incomingHeader)
	httpReq.Header.Set("Content-Type", "application/json")
	if upstream.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+upstream.APIKey)
	}
	client, err := buildHTTPClient(upstream, endpoint)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request upstream via proxy: %w", err)
	}
	return resp, nil
}

func copyAllowedHeaders(dst, src http.Header) {
	allowed := map[string]bool{
		"Accept":              true,
		"Accept-Encoding":     true,
		"Accept-Language":     true,
		"User-Agent":          true,
		"OpenAI-Beta":         true,
		"X-Request-Id":        true,
		"X-Stainless-Lang":    true,
		"X-Stainless-Package-Version": true,
		"X-Stainless-OS":      true,
		"X-Stainless-Arch":    true,
		"X-Stainless-Runtime": true,
		"X-Stainless-Runtime-Version": true,
	}
	for k, values := range src {
		ck := http.CanonicalHeaderKey(k)
		if !allowed[ck] {
			continue
		}
		for _, v := range values {
			dst.Add(ck, v)
		}
	}
}

func buildHTTPClient(upstream model.Upstream, endpoint proxyruntime.Endpoint) (*http.Client, error) {
	proxyURL, err := url.Parse(fmt.Sprintf("%s://%s:%d", endpoint.Scheme, endpoint.Host, endpoint.Port))
	if err != nil {
		return nil, fmt.Errorf("proxy url: %w", err)
	}
	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
		DialContext: (&net.Dialer{Timeout: 15 * time.Second}).DialContext,
		ResponseHeaderTimeout: time.Duration(upstream.TimeoutSeconds) * time.Second,
	}
	return &http.Client{Transport: transport, Timeout: time.Duration(upstream.TimeoutSeconds) * time.Second}, nil
}

func CopyResponse(w http.ResponseWriter, resp *http.Response) error {
	defer resp.Body.Close()
	for k, values := range resp.Header {
		for _, v := range values {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	_, err := io.Copy(w, resp.Body)
	return err
}
