package logger_test

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/salemshafik/pote/packages/logger"
)

func TestNew(t *testing.T) {
	var buf bytes.Buffer
	log := logger.New(logger.Config{
		ServiceName: "test-service",
		Level:       "debug",
		Output:      &buf,
	})

	log.Info("hello world")

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to parse log entry: %v", err)
	}

	if entry["service"] != "test-service" {
		t.Errorf("expected service=test-service, got %v", entry["service"])
	}
	if entry["msg"] != "hello world" {
		t.Errorf("expected msg=hello world, got %v", entry["msg"])
	}
	if entry["level"] != "INFO" {
		t.Errorf("expected level=INFO, got %v", entry["level"])
	}
}

func TestWithRequestID(t *testing.T) {
	var buf bytes.Buffer
	log := logger.New(logger.Config{
		ServiceName: "test-service",
		Level:       "info",
		Output:      &buf,
	})

	log.WithRequestID("req-123").Info("with request")

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to parse log entry: %v", err)
	}

	if entry["request_id"] != "req-123" {
		t.Errorf("expected request_id=req-123, got %v", entry["request_id"])
	}
}

func TestFromContext(t *testing.T) {
	var buf bytes.Buffer
	log := logger.New(logger.Config{
		ServiceName: "test-service",
		Level:       "info",
		Output:      &buf,
	})

	ctx := context.Background()
	ctx = logger.NewContext(ctx, "req-456")
	ctx = logger.NewContextWithUser(ctx, "user-789")

	log.FromContext(ctx).Info("context log")

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to parse log entry: %v", err)
	}

	if entry["request_id"] != "req-456" {
		t.Errorf("expected request_id=req-456, got %v", entry["request_id"])
	}
	if entry["user_id"] != "user-789" {
		t.Errorf("expected user_id=user-789, got %v", entry["user_id"])
	}
}

func TestLevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	log := logger.New(logger.Config{
		ServiceName: "test-service",
		Level:       "warn",
		Output:      &buf,
	})

	log.Info("should be filtered")

	if buf.Len() != 0 {
		t.Error("expected info message to be filtered at warn level")
	}

	log.Warn("should appear")

	if buf.Len() == 0 {
		t.Error("expected warn message to appear at warn level")
	}
}
