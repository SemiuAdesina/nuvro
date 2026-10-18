package config

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"
)

type Environment string

const (
	EnvironmentDevelopment Environment = "development"
	EnvironmentStaging     Environment = "staging"
	EnvironmentProduction  Environment = "production"
)

var ErrInvalid = errors.New("config: invalid configuration")

type Lookup func(key string) (string, bool)

type HTTP struct {
	Port              int
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ShutdownTimeout   time.Duration
	MaxHeaderBytes    int
}

type Database struct{ URL string }
type Redis struct{ URL string }

type Config struct {
	Environment    Environment
	LogLevel       string
	HTTP           HTTP
	Database       Database
	Redis          Redis
	Authentication Authentication
	Providers      Providers
	Runtime        Runtime
	Admission      Admission
	Proxy          Proxy
}

func (c Config) RequiresBackingServices() bool {
	return c.Environment == EnvironmentStaging || c.Environment == EnvironmentProduction
}

func Load(lookup Lookup) (Config, error) {
	values := reader{lookup: lookup}

	cfg := Config{
		Environment: Environment(values.text("NUVRO_ENV", string(EnvironmentDevelopment))),
		LogLevel:    values.text("NUVRO_LOG_LEVEL", "info"),
		HTTP: HTTP{
			Port:              values.number("NUVRO_HTTP_PORT", 8080),
			ReadHeaderTimeout: values.duration("NUVRO_HTTP_READ_HEADER_TIMEOUT", 5*time.Second),
			ReadTimeout:       values.duration("NUVRO_HTTP_READ_TIMEOUT", 15*time.Second),
			WriteTimeout:      values.duration("NUVRO_HTTP_WRITE_TIMEOUT", 30*time.Second),
			IdleTimeout:       values.duration("NUVRO_HTTP_IDLE_TIMEOUT", 120*time.Second),
			ShutdownTimeout:   values.duration("NUVRO_HTTP_SHUTDOWN_TIMEOUT", 20*time.Second),
			MaxHeaderBytes:    values.number("NUVRO_HTTP_MAX_HEADER_BYTES", 1<<20),
		},
		Database: Database{URL: values.text("NUVRO_DATABASE_URL", "")},
		Redis:    Redis{URL: values.text("NUVRO_REDIS_URL", "")},
	}
	cfg.Authentication = values.authentication()
	cfg.Providers = values.providers()
	cfg.Runtime = values.runtime()
	cfg.Admission = values.admission()
	cfg.Proxy = values.proxy()

	if err := values.err(); err != nil {
		return Config{}, err
	}
	if err := cfg.validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) validate() error {
	var problems []string

	switch c.Environment {
	case EnvironmentDevelopment, EnvironmentStaging, EnvironmentProduction:
	default:
		problems = append(problems, fmt.Sprintf("NUVRO_ENV %q must be development, staging or production", c.Environment))
	}

	switch c.LogLevel {
	case "debug", "info", "warn", "error":
	default:
		problems = append(problems, fmt.Sprintf("NUVRO_LOG_LEVEL %q must be debug, info, warn or error", c.LogLevel))
	}

	if c.HTTP.Port < 1 || c.HTTP.Port > 65535 {
		problems = append(problems, fmt.Sprintf("NUVRO_HTTP_PORT %d must be between 1 and 65535", c.HTTP.Port))
	}
	if c.HTTP.MaxHeaderBytes < 1 {
		problems = append(problems, "NUVRO_HTTP_MAX_HEADER_BYTES must be positive")
	}

	problems = append(problems, c.timeoutProblems()...)
	problems = append(problems, c.connectionProblems()...)
	problems = append(problems, c.providerProblems()...)
	problems = append(problems, c.runtimeProblems()...)
	problems = append(problems, c.admissionProblems()...)
	problems = append(problems, c.proxyProblems()...)
	problems = append(problems, c.authenticationProblems()...)

	if c.RequiresBackingServices() {
		if c.Database.URL == "" {
			problems = append(problems, "NUVRO_DATABASE_URL is required outside development")
		}
		if c.Redis.URL == "" {
			problems = append(problems, "NUVRO_REDIS_URL is required outside development")
		}
	}

	if len(problems) == 0 {
		return nil
	}
	return fmt.Errorf("%w: %s", ErrInvalid, strings.Join(problems, "; "))
}

func (c Config) timeoutProblems() []string {
	timeouts := map[string]time.Duration{
		"NUVRO_HTTP_READ_HEADER_TIMEOUT": c.HTTP.ReadHeaderTimeout,
		"NUVRO_HTTP_READ_TIMEOUT":        c.HTTP.ReadTimeout,
		"NUVRO_HTTP_WRITE_TIMEOUT":       c.HTTP.WriteTimeout,
		"NUVRO_HTTP_IDLE_TIMEOUT":        c.HTTP.IdleTimeout,
		"NUVRO_HTTP_SHUTDOWN_TIMEOUT":    c.HTTP.ShutdownTimeout,
	}

	names := make([]string, 0, len(timeouts))
	for name := range timeouts {
		names = append(names, name)
	}
	slices.Sort(names)

	var problems []string
	for _, name := range names {
		if timeouts[name] <= 0 {
			problems = append(problems, name+" must be a positive duration")
		}
	}
	return problems
}

type reader struct {
	lookup   Lookup
	problems []string
}

func (r *reader) text(key string, fallback string) string {
	value, ok := r.lookup(key)
	if !ok || strings.TrimSpace(value) == "" {
		return fallback
	}
	return strings.TrimSpace(value)
}

func (r *reader) number(key string, fallback int) int {
	raw, ok := r.lookup(key)
	if !ok || strings.TrimSpace(raw) == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		r.problems = append(r.problems, fmt.Sprintf("%s %q is not an integer", key, raw))
		return fallback
	}
	return parsed
}

func (r *reader) duration(key string, fallback time.Duration) time.Duration {
	raw, ok := r.lookup(key)
	if !ok || strings.TrimSpace(raw) == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(strings.TrimSpace(raw))
	if err != nil {
		r.problems = append(r.problems, fmt.Sprintf("%s %q is not a duration", key, raw))
		return fallback
	}
	return parsed
}

func (r *reader) err() error {
	if len(r.problems) == 0 {
		return nil
	}
	return fmt.Errorf("%w: %s", ErrInvalid, strings.Join(r.problems, "; "))
}
