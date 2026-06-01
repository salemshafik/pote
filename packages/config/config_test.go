package config_test

import (
	"os"
	"testing"
	"time"

	"github.com/salemshafik/pote/packages/config"
)

func TestStringRequired(t *testing.T) {
	os.Setenv("TEST_KEY", "hello")
	defer os.Unsetenv("TEST_KEY")

	l := config.NewLoader()
	val := l.String("TEST_KEY")
	if val != "hello" {
		t.Errorf("expected hello, got %s", val)
	}
	if err := l.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestStringFallback(t *testing.T) {
	os.Unsetenv("TEST_MISSING")

	l := config.NewLoader()
	val := l.String("TEST_MISSING", "default")
	if val != "default" {
		t.Errorf("expected default, got %s", val)
	}
	if err := l.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestStringMissing(t *testing.T) {
	os.Unsetenv("TEST_REQUIRED")

	l := config.NewLoader()
	_ = l.String("TEST_REQUIRED")
	if err := l.Validate(); err == nil {
		t.Error("expected validation error for missing required var")
	}
}

func TestInt(t *testing.T) {
	os.Setenv("TEST_INT", "42")
	defer os.Unsetenv("TEST_INT")

	l := config.NewLoader()
	val := l.Int("TEST_INT")
	if val != 42 {
		t.Errorf("expected 42, got %d", val)
	}
}

func TestIntInvalid(t *testing.T) {
	os.Setenv("TEST_INT_BAD", "not-a-number")
	defer os.Unsetenv("TEST_INT_BAD")

	l := config.NewLoader()
	_ = l.Int("TEST_INT_BAD")
	if err := l.Validate(); err == nil {
		t.Error("expected validation error for invalid integer")
	}
}

func TestBool(t *testing.T) {
	tests := []struct {
		value    string
		expected bool
	}{
		{"true", true},
		{"1", true},
		{"yes", true},
		{"false", false},
		{"0", false},
		{"no", false},
	}

	for _, tt := range tests {
		os.Setenv("TEST_BOOL", tt.value)
		l := config.NewLoader()
		val := l.Bool("TEST_BOOL")
		if val != tt.expected {
			t.Errorf("Bool(%q) = %v, want %v", tt.value, val, tt.expected)
		}
	}
	os.Unsetenv("TEST_BOOL")
}

func TestDuration(t *testing.T) {
	os.Setenv("TEST_DUR", "15m")
	defer os.Unsetenv("TEST_DUR")

	l := config.NewLoader()
	val := l.Duration("TEST_DUR")
	if val != 15*time.Minute {
		t.Errorf("expected 15m, got %v", val)
	}
}

func TestDatabaseDSN(t *testing.T) {
	os.Setenv("POSTGRES_HOST", "localhost")
	os.Setenv("POSTGRES_PORT", "5432")
	os.Setenv("POSTGRES_USER", "pote")
	os.Setenv("POSTGRES_PASSWORD", "secret")
	os.Setenv("AUTH_DB_NAME", "pote_auth")
	defer func() {
		os.Unsetenv("POSTGRES_HOST")
		os.Unsetenv("POSTGRES_PORT")
		os.Unsetenv("POSTGRES_USER")
		os.Unsetenv("POSTGRES_PASSWORD")
		os.Unsetenv("AUTH_DB_NAME")
	}()

	l := config.NewLoader()
	db := config.LoadDatabaseConfig(l, "AUTH_DB_NAME")
	expected := "postgres://pote:secret@localhost:5432/pote_auth?sslmode=disable"
	if db.DSN() != expected {
		t.Errorf("expected %s, got %s", expected, db.DSN())
	}
}

func TestMultipleErrors(t *testing.T) {
	os.Unsetenv("MISSING_1")
	os.Unsetenv("MISSING_2")

	l := config.NewLoader()
	_ = l.String("MISSING_1")
	_ = l.String("MISSING_2")
	err := l.Validate()
	if err == nil {
		t.Error("expected validation error")
	}
}
