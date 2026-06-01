// Package main is the entry point for the Pote chat-service.
// This service manages chat metadata, group creation, membership,
// and role-based access control for conversations.
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
	svcCfg := config.LoadServiceConfig(loader, 8083)
	_ = config.LoadDatabaseConfig(loader, "CHAT_DB_NAME")
	loader.MustValidate()

	log := logger.New(logger.Config{
		ServiceName: "chat-service",
		Level:       svcCfg.LogLevel,
	})

	// TODO: Initialize database connection
	// TODO: Run migrations
	// TODO: Initialize router with handlers

	log.Info("starting chat-service", "port", svcCfg.Port)

	addr := fmt.Sprintf(":%d", svcCfg.Port)
	server := &http.Server{Addr: addr}

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Info("shutting down chat-service")
		server.Close()
	}()

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Error("server error", "error", err)
		os.Exit(1)
	}
}
