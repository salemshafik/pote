// Package main is the entry point for the Pote user-service.
// This service handles user profiles, contacts management,
// and email invite functionality.
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
	svcCfg := config.LoadServiceConfig(loader, 8082)
	_ = config.LoadDatabaseConfig(loader, "USER_DB_NAME")
	loader.MustValidate()

	// Initialize logger
	log := logger.New(logger.Config{
		ServiceName: "user-service",
		Level:       svcCfg.LogLevel,
	})

	// TODO: Initialize database connection
	// TODO: Run migrations
	// TODO: Initialize router with handlers

	log.Info("starting user-service", "port", svcCfg.Port)

	addr := fmt.Sprintf(":%d", svcCfg.Port)
	server := &http.Server{
		Addr: addr,
	}

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Info("shutting down user-service")
		server.Close()
	}()

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Error("server error", "error", err)
		os.Exit(1)
	}
}
