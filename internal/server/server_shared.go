package server

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"api-v2ray/internal/app"
	"api-v2ray/internal/model"
	"api-v2ray/internal/proxyruntime"
	"api-v2ray/internal/router"
	"api-v2ray/internal/upstream"
)

func (s *Server) snapshotState() (*router.Service, *proxyruntime.Registry, *upstream.Client, *app.BootstrapResult) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.routerSvc, s.proxyRegistry, s.upstreamClient, s.bootstrap
}

func (s *Server) applyLiveConfig(cfg *model.Config, boot *app.BootstrapResult) {
	if cfg == nil {
		return
	}
	newRouter := router.New(cfg)
	newRegistry := buildProxyRegistry(cfg, boot)
	if s.upstreamClient != nil {
		s.upstreamClient.InvalidateAll()
	}
	s.mu.Lock()
	s.routerSvc = newRouter
	s.proxyRegistry = newRegistry
	s.bootstrap = boot
	s.mu.Unlock()
}

func decodeJSONBody(r *http.Request, maxBytes int64, dst any) error {
	if r == nil || r.Body == nil {
		return errors.New("empty request")
	}
	dec := json.NewDecoder(io.LimitReader(r.Body, maxBytes))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return err
	}
	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		if err == nil {
			return fmt.Errorf("multiple json values")
		}
		return err
	}
	return nil
}

func buildProxyRegistry(cfg *model.Config, boot *app.BootstrapResult) *proxyruntime.Registry {
	if boot != nil && len(boot.FlatResult.GeneratedXRAY) > 0 {
		return proxyruntime.NewFromGenerated(boot.FlatResult.GeneratedXRAY)
	}
	if cfg == nil {
		return proxyruntime.New(nil)
	}
	return proxyruntime.New(cfg.ProxyNodes)
}

func subtleCompare(got, expected string) bool {
	return subtle.ConstantTimeCompare([]byte(got), []byte(expected)) == 1
}
