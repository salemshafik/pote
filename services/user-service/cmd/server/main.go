// Package main is the entry point for the Pote user-service.
// This service handles user profiles, contacts management,
// user search, and email invite functionality.
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
	"github.com/salemshafik/pote/services/user-service/internal/database"
	"github.com/salemshafik/pote/services/user-service/internal/handler"
	"github.com/salemshafik/pote/services/user-service/internal/repository"
	"github.com/salemshafik/pote/services/user-service/internal/service"
)

func main() {
	// ---- Configuration ----
	loader := config.NewLoader()
	svcCfg := config.LoadServiceConfig(loader, 8082)
	dbCfg := config.LoadDatabaseConfig(loader, "USER_DB_NAME")

	// JWT secret (for validating tokens from auth-service)
	jwtSecret := loader.String("JWT_SECRET")

	loader.MustValidate()

	// ---- Logger ----
	log := logger.New(logger.Config{
		ServiceName: "user-service",
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
	profileRepo := repository.NewProfileRepository(pool)
	contactRepo := repository.NewContactRepository(pool)
	inviteRepo := repository.NewInviteRepository(pool)
	userService := service.NewUserService(profileRepo, contactRepo, inviteRepo, log)
	userHandler := handler.NewUserHandler(userService)

	// ---- Router ----
	router := handler.NewRouter(userHandler, jwtSecret)

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

	log.Info("starting user-service", "port", svcCfg.Port, "env", svcCfg.Env)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Error("server error", "error", err)
		os.Exit(1)
	}

	log.Info("user-service stopped")
}
