// Package main is the entry point for the Pote secrets-service.
// This service manages BYOK API key lifecycle: encrypt, store,
// retrieve (decrypted), and delete user-provided AI provider keys.
// Keys are stored encrypted via GCP Secret Manager / KMS.
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
	svcCfg := config.LoadServiceConfig(loader, 8088)
	_ = config.LoadDatabaseConfig(loader, "SECRETS_DB_NAME")
	loader.MustValidate()

	log := logger.New(logger.Config{
		ServiceName: "secrets-service",
		Level:       svcCfg.LogLevel,
	})

	// TODO: Initialize database connection
	// TODO: Run migrations
	// TODO: Initialize GCP Secret Manager / KMS client
	// TODO: Initialize router with handlers

	log.Info("starting secrets-service", "port", svcCfg.Port)

	addr := fmt.Sprintf(":%d", svcCfg.Port)
	server := &http.Server{Addr: addr}

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Info("shutting down secrets-service")
		server.Close()
	}()

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Error("server error", "error", err)
		os.Exit(1)
	}
}
