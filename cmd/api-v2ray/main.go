package main

import (
	"context"
	"flag"
	"log"
	"net/http"

	"api-v2ray/internal/app"
	"api-v2ray/internal/config"
	appruntime "api-v2ray/internal/runtime"
	"api-v2ray/internal/proxyruntime"
	"api-v2ray/internal/router"
	"api-v2ray/internal/server"
	"api-v2ray/internal/upstream"
)

func main() {
	configPath := flag.String("config", "./configs/config.example.json", "path to config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	boot, err := app.Bootstrap(context.Background(), cfg)
	if err != nil {
		log.Fatalf("bootstrap: %v", err)
	}
	cfg.ProxyNodes = boot.FlatResult.Nodes

	routerSvc := router.New(cfg)
	proxyRegistry := proxyruntime.New(cfg.ProxyNodes)
	if boot != nil && len(boot.FlatResult.GeneratedXRAY) > 0 {
		proxyRegistry = proxyruntime.NewFromGenerated(boot.FlatResult.GeneratedXRAY)
		if err := appruntime.StartXrayProcesses(cfg); err != nil {
			log.Printf("start xray processes: %v", err)
		}
	}
	upstreamClient := upstream.New()

	srv := server.New(routerSvc, proxyRegistry, upstreamClient, server.ConfigStore{Path: *configPath}, boot)

	log.Printf("api-v2ray listening on %s", cfg.Server.Listen)
	if err := http.ListenAndServe(cfg.Server.Listen, srv.Handler()); err != nil {
		log.Fatal(err)
	}
}
