// Package main is the entry point for the Pote auth-service.
// This service handles user registration, login (email/password),
// Google OAuth, JWT issuance, and token refresh.
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	authutils "github.com/salemshafik/pote/packages/auth-utils"
	"github.com/salemshafik/pote/packages/config"
	"github.com/salemshafik/pote/packages/logger"
	"github.com/salemshafik/pote/services/auth-service/internal/database"
	"github.com/salemshafik/pote/services/auth-service/internal/handler"
	"github.com/salemshafik/pote/services/auth-service/internal/repository"
	"github.com/salemshafik/pote/services/auth-service/internal/service"
)

func main() {
	// ---- Configuration ----
	loader := config.NewLoader()
	svcCfg := config.LoadServiceConfig(loader, 8081)
	dbCfg := config.LoadDatabaseConfig(loader, "AUTH_DB_NAME")

	// JWT configuration
	jwtSecret := loader.String("JWT_SECRET")
	jwtIssuer := loader.String("JWT_ISSUER", "pote")
	accessExpiry := loader.Duration("JWT_ACCESS_TOKEN_EXPIRY", 15*time.Minute)
	refreshExpiry := loader.Duration("JWT_REFRESH_TOKEN_EXPIRY", 7*24*time.Hour)

	// Google OAuth configuration
	googleClientID := loader.String("GOOGLE_CLIENT_ID", "")
	googleClientSecret := loader.String("GOOGLE_CLIENT_SECRET", "")
	googleRedirectURL := loader.String("GOOGLE_REDIRECT_URL", "http://localhost:8081/api/v1/auth/google/callback")
	frontendURL := loader.String("FRONTEND_URL", "http://localhost:3000")

	loader.MustValidate()

	// ---- Logger ----
	log := logger.New(logger.Config{
		ServiceName: "auth-service",
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

	// ---- Token config ----
	tokenCfg := authutils.TokenConfig{
		Secret:             jwtSecret,
		Issuer:             jwtIssuer,
		AccessTokenExpiry:  accessExpiry,
		RefreshTokenExpiry: refreshExpiry,
	}

	// ---- Google OAuth config ----
	oauthCfg := &oauth2.Config{
		ClientID:     googleClientID,
		ClientSecret: googleClientSecret,
		RedirectURL:  googleRedirectURL,
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}

	// ---- Dependency injection ----
	userRepo := repository.NewUserRepository(pool)
	tokenRepo := repository.NewTokenRepository(pool)
	authService := service.NewAuthService(userRepo, tokenRepo, tokenCfg, log)
	authHandler := handler.NewAuthHandler(authService)
	oauthHandler := handler.NewOAuthHandler(authService, oauthCfg, frontendURL)

	// ---- Router ----
	router := handler.NewRouter(authHandler, oauthHandler)

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

	log.Info("starting auth-service", "port", svcCfg.Port, "env", svcCfg.Env)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Error("server error", "error", err)
		os.Exit(1)
	}

	log.Info("auth-service stopped")
}
