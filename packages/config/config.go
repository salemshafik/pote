// Package config provides utilities for loading and validating environment
// variables across all Pote microservices. It supports required variables,
// defaults, and type conversions with clear error reporting.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Loader reads environment variables with validation and type conversion.
type Loader struct {
	// errors collects all validation errors so they can be reported together.
	errors []string
}

// NewLoader creates a new config Loader.
func NewLoader() *Loader {
	return &Loader{}
}

// String returns the value of the environment variable identified by key.
// If the variable is not set and no fallback is provided, an error is recorded.
func (l *Loader) String(key string, fallback ...string) string {
	val := os.Getenv(key)
	if val != "" {
		return val
	}
	if len(fallback) > 0 {
		return fallback[0]
	}
	l.errors = append(l.errors, fmt.Sprintf("required env var %s is not set", key))
	return ""
}

// Int returns the integer value of the environment variable identified by key.
func (l *Loader) Int(key string, fallback ...int) int {
	val := os.Getenv(key)
	if val == "" {
		if len(fallback) > 0 {
			return fallback[0]
		}
		l.errors = append(l.errors, fmt.Sprintf("required env var %s is not set", key))
		return 0
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		l.errors = append(l.errors, fmt.Sprintf("env var %s=%q is not a valid integer", key, val))
		return 0
	}
	return n
}

// Bool returns the boolean value of the environment variable identified by key.
// Accepted truthy values: "true", "1", "yes". Everything else is false.
func (l *Loader) Bool(key string, fallback ...bool) bool {
	val := os.Getenv(key)
	if val == "" {
		if len(fallback) > 0 {
			return fallback[0]
		}
		return false
	}
	switch strings.ToLower(strings.TrimSpace(val)) {
	case "true", "1", "yes":
		return true
	default:
		return false
	}
}

// Duration returns a time.Duration parsed from the environment variable.
// Accepts Go duration strings (e.g., "15m", "1h", "30s").
func (l *Loader) Duration(key string, fallback ...time.Duration) time.Duration {
	val := os.Getenv(key)
	if val == "" {
		if len(fallback) > 0 {
			return fallback[0]
		}
		l.errors = append(l.errors, fmt.Sprintf("required env var %s is not set", key))
		return 0
	}
	d, err := time.ParseDuration(val)
	if err != nil {
		l.errors = append(l.errors, fmt.Sprintf("env var %s=%q is not a valid duration", key, val))
		return 0
	}
	return d
}

// StringSlice returns a slice of strings split by the given separator.
func (l *Loader) StringSlice(key, separator string, fallback ...[]string) []string {
	val := os.Getenv(key)
	if val == "" {
		if len(fallback) > 0 {
			return fallback[0]
		}
		return nil
	}
	parts := strings.Split(val, separator)
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// Validate checks if any errors were recorded during loading.
// Returns a combined error if any required variables are missing or invalid.
func (l *Loader) Validate() error {
	if len(l.errors) == 0 {
		return nil
	}
	return fmt.Errorf("config validation failed:\n  - %s", strings.Join(l.errors, "\n  - "))
}

// MustValidate calls Validate and panics if there are errors.
// Use this in main() to fail fast on misconfiguration.
func (l *Loader) MustValidate() {
	if err := l.Validate(); err != nil {
		panic(err)
	}
}

// ---------- Convenience: common service config ----------

// ServiceConfig holds common configuration shared by all Pote services.
type ServiceConfig struct {
	// Env is the deployment environment (development, staging, production).
	Env string
	// Port is the HTTP port the service listens on.
	Port int
	// LogLevel is the minimum log level (debug, info, warn, error).
	LogLevel string
}

// LoadServiceConfig loads the common service configuration from environment variables.
func LoadServiceConfig(loader *Loader, defaultPort int) ServiceConfig {
	return ServiceConfig{
		Env:      loader.String("ENV", "development"),
		Port:     loader.Int("PORT", defaultPort),
		LogLevel: loader.String("LOG_LEVEL", "info"),
	}
}

// DatabaseConfig holds PostgreSQL connection configuration.
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// LoadDatabaseConfig loads database configuration from environment variables.
func LoadDatabaseConfig(loader *Loader, dbNameEnvKey string) DatabaseConfig {
	return DatabaseConfig{
		Host:     loader.String("POSTGRES_HOST", "localhost"),
		Port:     loader.Int("POSTGRES_PORT", 5432),
		User:     loader.String("POSTGRES_USER", "pote"),
		Password: loader.String("POSTGRES_PASSWORD"),
		DBName:   loader.String(dbNameEnvKey),
		SSLMode:  loader.String("POSTGRES_SSL_MODE", "disable"),
	}
}

// DSN returns the PostgreSQL connection string.
func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.DBName, d.SSLMode,
	)
}

// RedisConfig holds Redis connection configuration.
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// LoadRedisConfig loads Redis configuration from environment variables.
func LoadRedisConfig(loader *Loader) RedisConfig {
	return RedisConfig{
		Host:     loader.String("REDIS_HOST", "localhost"),
		Port:     loader.Int("REDIS_PORT", 6379),
		Password: loader.String("REDIS_PASSWORD", ""),
		DB:       loader.Int("REDIS_DB", 0),
	}
}

// Addr returns the Redis address in host:port format.
func (r RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}
