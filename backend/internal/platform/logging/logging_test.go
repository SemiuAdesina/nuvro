package logging_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"

	"github.com/SemiuAdesina/nuvro/backend/internal/platform/config"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/logging"
)

func TestParseLevelAcceptsSupportedNames(t *testing.T) {
	cases := map[string]slog.Level{
		"debug": slog.LevelDebug,
		"INFO":  slog.LevelInfo,
		" warn": slog.LevelWarn,
		"error": slog.LevelError,
	}

	for raw, want := range cases {
		got, err := logging.ParseLevel(raw)
		if err != nil {
			t.Errorf("ParseLevel(%q) returned error: %v", raw, err)
			continue
		}
		if got != want {
			t.Errorf("ParseLevel(%q) = %v, want %v", raw, got, want)
		}
	}
}

func TestParseLevelRejectsUnknownName(t *testing.T) {
	if _, err := logging.ParseLevel("verbose"); err == nil {
		t.Fatal("ParseLevel(\"verbose\") returned no error")
	}
}

func TestNewEmitsJSONOutsideDevelopment(t *testing.T) {
	var buffer bytes.Buffer
	cfg := config.Config{Environment: config.EnvironmentProduction, LogLevel: "info"}

	logger, err := logging.New(&buffer, cfg)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	logger.Info("api started", slog.Int("port", 8080))

	var record map[string]any
	if err := json.Unmarshal(buffer.Bytes(), &record); err != nil {
		t.Fatalf("output is not JSON: %v (%q)", err, buffer.String())
	}
	if record["service"] != "nuvro" {
		t.Errorf("service = %v, want nuvro", record["service"])
	}
	if record["environment"] != "production" {
		t.Errorf("environment = %v, want production", record["environment"])
	}
	if record["msg"] != "api started" {
		t.Errorf("msg = %v, want %q", record["msg"], "api started")
	}
}

func TestNewEmitsTextInDevelopment(t *testing.T) {
	var buffer bytes.Buffer
	cfg := config.Config{Environment: config.EnvironmentDevelopment, LogLevel: "info"}

	logger, err := logging.New(&buffer, cfg)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	logger.Info("api started")

	if json.Valid(bytes.TrimSpace(buffer.Bytes())) {
		t.Fatalf("development output should not be JSON, got %q", buffer.String())
	}
	if !strings.Contains(buffer.String(), "environment=development") {
		t.Fatalf("output %q does not carry the environment attribute", buffer.String())
	}
}

func TestNewHonoursConfiguredLevel(t *testing.T) {
	var buffer bytes.Buffer
	cfg := config.Config{Environment: config.EnvironmentProduction, LogLevel: "error"}

	logger, err := logging.New(&buffer, cfg)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	logger.Info("suppressed")
	if buffer.Len() != 0 {
		t.Fatalf("info record was emitted at error level: %q", buffer.String())
	}

	logger.Error("reported")
	if !strings.Contains(buffer.String(), "reported") {
		t.Fatalf("error record was not emitted: %q", buffer.String())
	}
}

func TestNewRejectsUnknownLevel(t *testing.T) {
	cfg := config.Config{Environment: config.EnvironmentProduction, LogLevel: "loud"}

	if _, err := logging.New(&bytes.Buffer{}, cfg); err == nil {
		t.Fatal("New returned no error for an unknown level")
	}
}
