// Package main is the entry point for the Pote message-service.
// This service handles message CRUD, history retrieval, and
// AI command parsing (@chatgpt, @claude, @gemini + leave).
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
	svcCfg := config.LoadServiceConfig(loader, 8084)
	_ = config.LoadDatabaseConfig(loader, "MESSAGE_DB_NAME")
	loader.MustValidate()

	log := logger.New(logger.Config{
		ServiceName: "message-service",
		Level:       svcCfg.LogLevel,
	})

	// TODO: Initialize database connection
	// TODO: Run migrations
	// TODO: Initialize router with handlers
	// TODO: Set up Pub/Sub publisher for ai-requests topic
	// TODO: Set up Pub/Sub subscriber for ai-responses topic
	// TODO: Set up Redis publisher for realtime message delivery

	log.Info("starting message-service", "port", svcCfg.Port)

	addr := fmt.Sprintf(":%d", svcCfg.Port)
	server := &http.Server{Addr: addr}

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Info("shutting down message-service")
		server.Close()
	}()

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Error("server error", "error", err)
		os.Exit(1)
	}
}
