// Package main is the entry point for the Pote realtime-service.
// This service manages WebSocket connections, Redis Pub/Sub subscriptions,
// typing indicators, read receipts, and user presence.
package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/salemshafik/pote/packages/config"
	"github.com/salemshafik/pote/packages/logger"
)

func main() {
	loader := config.NewLoader()
	svcCfg := config.LoadServiceConfig(loader, 8085)
	_ = config.LoadRedisConfig(loader)
	loader.MustValidate()

	log := logger.New(logger.Config{
		ServiceName: "realtime-service",
		Level:       svcCfg.LogLevel,
	})

	// TODO: Initialize Redis Pub/Sub subscriber
	// TODO: Initialize WebSocket hub (connection manager)
	// TODO: Set up WebSocket upgrade handler with JWT auth
	// TODO: Handle typing indicators, read receipts, presence

	log.Info("starting realtime-service", "port", svcCfg.Port)

	addr := fmt.Sprintf(":%d", svcCfg.Port)
	server := &http.Server{Addr: addr}

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Info("shutting down realtime-service")
		server.Close()
	}()

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Error("server error", "error", err)
		os.Exit(1)
	}
}
