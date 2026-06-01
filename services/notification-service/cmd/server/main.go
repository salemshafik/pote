// Package main is the entry point for the Pote notification-service.
// This service handles push notifications, email invites, and
// listens to Pub/Sub events for notification triggers.
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
	svcCfg := config.LoadServiceConfig(loader, 8087)
	_ = config.LoadDatabaseConfig(loader, "NOTIFICATION_DB_NAME")
	loader.MustValidate()

	log := logger.New(logger.Config{
		ServiceName: "notification-service",
		Level:       svcCfg.LogLevel,
	})

	// TODO: Initialize database connection
	// TODO: Run migrations
	// TODO: Set up Pub/Sub subscriber for notifications topic
	// TODO: Initialize email sender (SMTP / SendGrid / GCP)
	// TODO: Initialize router with handlers

	log.Info("starting notification-service", "port", svcCfg.Port)

	addr := fmt.Sprintf(":%d", svcCfg.Port)
	server := &http.Server{Addr: addr}

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Info("shutting down notification-service")
		server.Close()
	}()

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Error("server error", "error", err)
		os.Exit(1)
	}
}
