// Package main is the entry point for the Pote chat-service.
// This service manages chat metadata, group creation, membership,
// and role-based access control for conversations.
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/salemshafik/pote/packages/config"
	"github.com/salemshafik/pote/packages/logger"
	"github.com/salemshafik/pote/services/chat-service/internal/database"
	"github.com/salemshafik/pote/services/chat-service/internal/handler"
	"github.com/salemshafik/pote/services/chat-service/internal/repository"
	"github.com/salemshafik/pote/services/chat-service/internal/service"
)

func main() {
	// ---- Configuration ----
	loader := config.NewLoader()
	svcCfg := config.LoadServiceConfig(loader, 8083)
	dbCfg := config.LoadDatabaseConfig(loader, "CHAT_DB_NAME")

	// JWT secret is required to validate access tokens issued by auth-service.
	jwtSecret := loader.String("JWT_SECRET")
	frontendURL := loader.String("FRONTEND_URL", "http://localhost:3000")

	loader.MustValidate()

	// ---- Logger ----
	log := logger.New(logger.Config{
		ServiceName: "chat-service",
		Level:       svcCfg.LogLevel,
	})

	// ---- Database ----
	ctx := context.Background()
	pool, err := database.Connect(ctx, dbCfg.DSN())
	if err != nil {
		log.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()
	log.Info("connected to database")

	// ---- Dependency injection ----
	chatRepo := repository.NewChatRepository(pool)
	memberRepo := repository.NewMemberRepository(pool)
	chatService := service.NewChatService(chatRepo, memberRepo, log)
	chatHandler := handler.NewChatHandler(chatService, log)

	// ---- Router ----
	router := handler.NewRouter(chatHandler, jwtSecret, frontendURL)

	// ---- HTTP Server ----
	addr := fmt.Sprintf(":%d", svcCfg.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// ---- Graceful shutdown ----
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigCh

		log.Info("received shutdown signal", "signal", sig.String())

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Error("graceful shutdown failed", "error", err)
		}
	}()

	log.Info("starting chat-service", "port", svcCfg.Port, "env", svcCfg.Env)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Error("server error", "error", err)
		os.Exit(1)
	}

	log.Info("chat-service stopped")
}
