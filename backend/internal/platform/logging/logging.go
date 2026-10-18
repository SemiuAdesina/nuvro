package logging

import (
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/SemiuAdesina/nuvro/backend/internal/platform/config"
)

func New(destination io.Writer, cfg config.Config) (*slog.Logger, error) {
	level, err := ParseLevel(cfg.LogLevel)
	if err != nil {
		return nil, err
	}

	options := &slog.HandlerOptions{Level: level}

	var handler slog.Handler
	if cfg.Environment == config.EnvironmentDevelopment {
		handler = slog.NewTextHandler(destination, options)
	} else {
		handler = slog.NewJSONHandler(destination, options)
	}

	return slog.New(handler).With(
		slog.String("service", "nuvro"),
		slog.String("environment", string(cfg.Environment)),
	), nil
}

func ParseLevel(raw string) (slog.Level, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return 0, fmt.Errorf("logging: unknown level %q", raw)
	}
}
