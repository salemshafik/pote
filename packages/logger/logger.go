// Package logger provides a standardized structured logging utility for all
// Pote microservices. It wraps Go's standard log/slog package to ensure
// consistent JSON-formatted log output across every service.
package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strings"
)

// contextKey is an unexported type used for context value keys to avoid collisions.
type contextKey string

const (
	// requestIDKey is the context key for storing the request ID.
	requestIDKey contextKey = "request_id"
	// userIDKey is the context key for storing the authenticated user ID.
	userIDKey contextKey = "user_id"
)

// Logger wraps slog.Logger with service-specific metadata.
type Logger struct {
	*slog.Logger
}

// Config holds the configuration for initializing a Logger.
type Config struct {
	// ServiceName is the name of the microservice (e.g., "auth-service").
	ServiceName string
	// Level is the minimum log level (debug, info, warn, error). Defaults to "info".
	Level string
	// Output is the writer for log output. Defaults to os.Stdout.
	Output io.Writer
}

// New creates a new Logger with the given configuration.
// All log entries automatically include the "service" field.
func New(cfg Config) *Logger {
	if cfg.Output == nil {
		cfg.Output = os.Stdout
	}

	level := parseLevel(cfg.Level)

	handler := slog.NewJSONHandler(cfg.Output, &slog.HandlerOptions{
		Level:     level,
		AddSource: level == slog.LevelDebug,
	})

	l := slog.New(handler).With(
		slog.String("service", cfg.ServiceName),
	)

	return &Logger{Logger: l}
}

// WithRequestID returns a new Logger with the request ID attached.
func (l *Logger) WithRequestID(requestID string) *Logger {
	return &Logger{
		Logger: l.Logger.With(slog.String("request_id", requestID)),
	}
}

// WithUserID returns a new Logger with the user ID attached.
func (l *Logger) WithUserID(userID string) *Logger {
	return &Logger{
		Logger: l.Logger.With(slog.String("user_id", userID)),
	}
}

// WithError returns a new Logger with the error attached.
func (l *Logger) WithError(err error) *Logger {
	return &Logger{
		Logger: l.Logger.With(slog.String("error", err.Error())),
	}
}

// WithField returns a new Logger with an additional key-value field.
func (l *Logger) WithField(key string, value any) *Logger {
	return &Logger{
		Logger: l.Logger.With(slog.Any(key, value)),
	}
}

// WithFields returns a new Logger with multiple key-value fields.
func (l *Logger) WithFields(fields map[string]any) *Logger {
	attrs := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		attrs = append(attrs, slog.Any(k, v))
	}
	return &Logger{
		Logger: l.Logger.With(attrs...),
	}
}

// ---- Context helpers ----

// NewContext returns a new context with the request ID stored.
func NewContext(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// NewContextWithUser returns a new context with the user ID stored.
func NewContextWithUser(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// FromContext creates a child Logger enriched with any request_id and user_id
// found in the context.
func (l *Logger) FromContext(ctx context.Context) *Logger {
	child := l
	if reqID, ok := ctx.Value(requestIDKey).(string); ok && reqID != "" {
		child = child.WithRequestID(reqID)
	}
	if uid, ok := ctx.Value(userIDKey).(string); ok && uid != "" {
		child = child.WithUserID(uid)
	}
	return child
}

// ---- Helpers ----

// parseLevel converts a string level name to slog.Level.
func parseLevel(level string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
