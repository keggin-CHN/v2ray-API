package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"api-v2ray/internal/app"
	"api-v2ray/internal/config"
	"api-v2ray/internal/proxyruntime"
	appruntime "api-v2ray/internal/runtime"
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
	if boot != nil {
		cfg.ProxyNodes = boot.FlatResult.Nodes
	}

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

	httpServer := &http.Server{
		Addr:              cfg.Server.Listen,
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("api-v2ray listening on %s", cfg.Server.Listen)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	<-quit
	log.Println("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("server shutdown: %v", err)
	}
	log.Println("server stopped")
}
