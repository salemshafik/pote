// Package main is the entry point for the Pote auth-service.
// This service handles user registration, login (email/password),
// Google OAuth, JWT issuance, and token refresh.
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
	// Load configuration
	loader := config.NewLoader()
	svcCfg := config.LoadServiceConfig(loader, 8081)
	_ = config.LoadDatabaseConfig(loader, "AUTH_DB_NAME")
	loader.MustValidate()

	// Initialize logger
	log := logger.New(logger.Config{
		ServiceName: "auth-service",
		Level:       svcCfg.LogLevel,
	})

	// TODO: Initialize database connection
	// TODO: Run migrations
	// TODO: Initialize router with handlers
	// TODO: Set up Google OAuth provider

	log.Info("starting auth-service", "port", svcCfg.Port)

	// Start HTTP server
	addr := fmt.Sprintf(":%d", svcCfg.Port)
	server := &http.Server{
		Addr: addr,
		// Handler: router, // TODO: wire up chi router
	}

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Info("shutting down auth-service")
		server.Close()
	}()

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Error("server error", "error", err)
		os.Exit(1)
	}
}
