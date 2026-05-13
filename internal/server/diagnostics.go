package server

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"api-v2ray/internal/openai"
)

type ExitIPProbeResponse struct {
	DirectIP     string `json:"direct_ip,omitempty"`
	ProxyIP      string `json:"proxy_ip,omitempty"`
	ProxyAddress string `json:"proxy_address,omitempty"`
	ProxyActive  bool   `json:"proxy_active"`
	SameExit     bool   `json:"same_exit"`
	Error        string `json:"error,omitempty"`
}

func (s *Server) handleExitIPProbe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		openai.WriteError(w, http.StatusMethodNotAllowed, "invalid_request_error", "method_not_allowed", "method not allowed")
		return
	}
	routerSvc, proxyRegistry, _, _ := s.snapshotState()
	models := routerSvc.Models()
	if len(models) == 0 {
		writeJSON(w, http.StatusBadRequest, ExitIPProbeResponse{Error: "no enabled models configured"})
		return
	}
	candidates, err := routerSvc.ResolveCandidatesByModel(models[0])
	if err != nil || len(candidates) == 0 {
		writeJSON(w, http.StatusBadGateway, ExitIPProbeResponse{Error: fmt.Sprintf("resolve candidates: %v", err)})
		return
	}
	endpoint, err := proxyRegistry.Get(candidates[0].Binding.NodeID)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, ExitIPProbeResponse{Error: fmt.Sprintf("proxy registry: %v", err)})
		return
	}
	proxyAddr := net.JoinHostPort(endpoint.Host, fmt.Sprintf("%d", endpoint.Port))
	resp := ExitIPProbeResponse{ProxyAddress: proxyAddr}

	probeURL := strings.TrimSpace(r.URL.Query().Get("url"))
	fallbackURL := strings.TrimSpace(r.URL.Query().Get("fallback_url"))
	directIP, directErr := fetchIP(nil, probeURL, fallbackURL)
	if directErr == nil {
		resp.DirectIP = directIP
	}
	proxyIP, proxyErr := fetchIP(newSocks5Client(endpoint.Host, endpoint.Port), probeURL, fallbackURL)
	if proxyErr == nil {
		resp.ProxyIP = proxyIP
	}
	resp.ProxyActive = resp.ProxyIP != ""
	resp.SameExit = resp.DirectIP != "" && resp.ProxyIP != "" && resp.DirectIP == resp.ProxyIP
	if directErr != nil || proxyErr != nil {
		resp.Error = fmt.Sprintf("direct=%v; proxy=%v", directErr, proxyErr)
	}
	writeJSON(w, http.StatusOK, resp)
}

func fetchIP(client *http.Client, probeURL, fallbackURL string) (string, error) {
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	targets := []string{}
	if strings.TrimSpace(probeURL) != "" {
		targets = append(targets, strings.TrimSpace(probeURL))
	}
	if strings.TrimSpace(fallbackURL) != "" {
		targets = append(targets, strings.TrimSpace(fallbackURL))
	}
	if len(targets) == 0 {
		targets = append(targets, "https://api.ipify.org")
	}

	var lastErr error
	for _, target := range targets {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
		if err != nil {
			cancel()
			lastErr = err
			continue
		}
		res, err := client.Do(req)
		if err != nil {
			cancel()
			lastErr = err
			continue
		}
		b, readErr := io.ReadAll(io.LimitReader(res.Body, 128))
		_ = res.Body.Close()
		cancel()
		if readErr != nil {
			lastErr = readErr
			continue
		}
		if res.StatusCode < 200 || res.StatusCode >= 300 {
			lastErr = fmt.Errorf("probe status %d", res.StatusCode)
			continue
		}
		ip := strings.TrimSpace(string(b))
		if ip == "" {
			lastErr = fmt.Errorf("empty probe response")
			continue
		}
		return ip, nil
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("all probes failed")
	}
	return "", lastErr
}

func newSocks5Client(host string, port int) *http.Client {
	dialer := func(ctx context.Context, network, addr string) (net.Conn, error) {
		d := &net.Dialer{Timeout: 15 * time.Second}
		conn, err := d.DialContext(ctx, "tcp", net.JoinHostPort(host, fmt.Sprintf("%d", port)))
		if err != nil {
			return nil, err
		}
		if _, err := conn.Write([]byte{0x05, 0x01, 0x00}); err != nil {
			_ = conn.Close()
			return nil, err
		}
		buf := make([]byte, 2)
		if _, err := io.ReadFull(conn, buf); err != nil {
			_ = conn.Close()
			return nil, err
		}
		if buf[0] != 0x05 || buf[1] != 0x00 {
			_ = conn.Close()
			return nil, fmt.Errorf("socks auth negotiation failed")
		}
		hostPart, portPart, err := net.SplitHostPort(addr)
		if err != nil {
			_ = conn.Close()
			return nil, err
		}
		portNum, err := net.LookupPort(network, portPart)
		if err != nil {
			_ = conn.Close()
			return nil, err
		}
		ip := net.ParseIP(hostPart)
		var req []byte
		if ip4 := ip.To4(); ip4 != nil {
			req = append([]byte{0x05, 0x01, 0x00, 0x01}, ip4...)
		} else if ip6 := ip.To16(); ip6 != nil && ip.To4() == nil {
			req = append([]byte{0x05, 0x01, 0x00, 0x04}, ip6...)
		} else {
			req = append([]byte{0x05, 0x01, 0x00, 0x03, byte(len(hostPart))}, []byte(hostPart)...)
		}
		req = append(req, byte(portNum>>8), byte(portNum))
		if _, err := conn.Write(req); err != nil {
			_ = conn.Close()
			return nil, err
		}
		head := make([]byte, 4)
		if _, err := io.ReadFull(conn, head); err != nil {
			_ = conn.Close()
			return nil, err
		}
		if head[1] != 0x00 {
			_ = conn.Close()
			return nil, fmt.Errorf("socks connect failed with code %d", head[1])
		}
		var skip int
		switch head[3] {
		case 0x01:
			skip = 4
		case 0x04:
			skip = 16
		case 0x03:
			lenBuf := make([]byte, 1)
			if _, err := io.ReadFull(conn, lenBuf); err != nil {
				_ = conn.Close()
				return nil, err
			}
			skip = int(lenBuf[0])
		default:
			_ = conn.Close()
			return nil, fmt.Errorf("unsupported socks atyp %d", head[3])
		}
		if skip > 0 {
			if _, err := io.CopyN(io.Discard, conn, int64(skip+2)); err != nil {
				_ = conn.Close()
				return nil, err
			}
		} else {
			if _, err := io.CopyN(io.Discard, conn, 2); err != nil {
				_ = conn.Close()
				return nil, err
			}
		}
		return conn, nil
	}
	tr := &http.Transport{DialContext: dialer}
	return &http.Client{Timeout: 20 * time.Second, Transport: tr}
}
