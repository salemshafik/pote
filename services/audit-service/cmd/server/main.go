// Package main is the entry point for the Pote audit-service.
// This service provides append-only security and event logging.
// It subscribes to Pub/Sub audit-events topic and persists all
// security-relevant actions for compliance and debugging.
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
	svcCfg := config.LoadServiceConfig(loader, 8089)
	_ = config.LoadDatabaseConfig(loader, "AUDIT_DB_NAME")
	loader.MustValidate()

	log := logger.New(logger.Config{
		ServiceName: "audit-service",
		Level:       svcCfg.LogLevel,
	})

	// TODO: Initialize database connection
	// TODO: Run migrations (append-only table)
	// TODO: Set up Pub/Sub subscriber for audit-events topic
	// TODO: Initialize router with read-only query handlers

	log.Info("starting audit-service", "port", svcCfg.Port)

	addr := fmt.Sprintf(":%d", svcCfg.Port)
	server := &http.Server{Addr: addr}

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Info("shutting down audit-service")
		server.Close()
	}()

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Error("server error", "error", err)
		os.Exit(1)
	}
}
