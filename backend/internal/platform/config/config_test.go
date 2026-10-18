package config_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/platform/config"
)

func lookupFrom(values map[string]string) config.Lookup {
	return func(key string) (string, bool) {
		value, ok := values[key]
		return value, ok
	}
}

func TestLoadAppliesDevelopmentDefaults(t *testing.T) {
	cfg, err := config.Load(lookupFrom(map[string]string{}))
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.Environment != config.EnvironmentDevelopment {
		t.Errorf("Environment = %q, want development", cfg.Environment)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("LogLevel = %q, want info", cfg.LogLevel)
	}
	if cfg.HTTP.Port != 8080 {
		t.Errorf("Port = %d, want 8080", cfg.HTTP.Port)
	}
	if cfg.HTTP.ReadHeaderTimeout != 5*time.Second {
		t.Errorf("ReadHeaderTimeout = %v, want 5s", cfg.HTTP.ReadHeaderTimeout)
	}
	if cfg.HTTP.MaxHeaderBytes != 1<<20 {
		t.Errorf("MaxHeaderBytes = %d, want 1048576", cfg.HTTP.MaxHeaderBytes)
	}
	if cfg.RequiresBackingServices() {
		t.Error("RequiresBackingServices() = true, want false in development")
	}
}

func TestLoadReadsOverrides(t *testing.T) {
	cfg, err := config.Load(lookupFrom(map[string]string{
		"NUVRO_ENV":                   "production",
		"NUVRO_LOG_LEVEL":             "warn",
		"NUVRO_HTTP_PORT":             " 9090 ",
		"NUVRO_HTTP_WRITE_TIMEOUT":    "45s",
		"NUVRO_HTTP_SHUTDOWN_TIMEOUT": "1m",
		"NUVRO_DATABASE_URL":          "postgres://localhost:5432/nuvro",
		"NUVRO_REDIS_URL":             "redis://localhost:6379/0",
		"NUVRO_APPLE_AUDIENCE":        "com.vectixlogic.nuvro",
		"NUVRO_APPLE_TEAM_ID":         "ABCDE12345",
		"NUVRO_APPLE_KEY_ID":          "KEY1234567",
		"NUVRO_APPLE_PRIVATE_KEY":     signingKeyPEM(),
		"NUVRO_TOKEN_SEALING_KEY":     productionSealingKey,
		"NUVRO_SESSION_SIGNING_KEY":   productionSigningKey,
		"NUVRO_CURSOR_SECRET":         productionCursorSecret,
	}))
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.Environment != config.EnvironmentProduction {
		t.Errorf("Environment = %q, want production", cfg.Environment)
	}
	if cfg.HTTP.Port != 9090 {
		t.Errorf("Port = %d, want 9090", cfg.HTTP.Port)
	}
	if cfg.HTTP.WriteTimeout != 45*time.Second {
		t.Errorf("WriteTimeout = %v, want 45s", cfg.HTTP.WriteTimeout)
	}
	if cfg.HTTP.ShutdownTimeout != time.Minute {
		t.Errorf("ShutdownTimeout = %v, want 1m", cfg.HTTP.ShutdownTimeout)
	}
	if !cfg.RequiresBackingServices() {
		t.Error("RequiresBackingServices() = false, want true in production")
	}
}

func TestLoadTreatsBlankValueAsUnset(t *testing.T) {
	cfg, err := config.Load(lookupFrom(map[string]string{"NUVRO_HTTP_PORT": "   "}))
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if cfg.HTTP.Port != 8080 {
		t.Fatalf("Port = %d, want the 8080 default", cfg.HTTP.Port)
	}
}

func TestLoadRejectsUnparsableValues(t *testing.T) {
	_, err := config.Load(lookupFrom(map[string]string{
		"NUVRO_HTTP_PORT":         "eighty-eighty",
		"NUVRO_HTTP_READ_TIMEOUT": "soon",
	}))
	if !errors.Is(err, config.ErrInvalid) {
		t.Fatalf("error = %v, want ErrInvalid", err)
	}
	for _, fragment := range []string{"NUVRO_HTTP_PORT", "NUVRO_HTTP_READ_TIMEOUT"} {
		if !strings.Contains(err.Error(), fragment) {
			t.Errorf("error %q does not mention %s", err, fragment)
		}
	}
}

func TestLoadRejectsUnknownEnvironmentAndLogLevel(t *testing.T) {
	_, err := config.Load(lookupFrom(map[string]string{
		"NUVRO_ENV":       "prod",
		"NUVRO_LOG_LEVEL": "verbose",
	}))
	if !errors.Is(err, config.ErrInvalid) {
		t.Fatalf("error = %v, want ErrInvalid", err)
	}
	if !strings.Contains(err.Error(), "NUVRO_ENV") || !strings.Contains(err.Error(), "NUVRO_LOG_LEVEL") {
		t.Fatalf("error %q does not report both problems", err)
	}
}

func TestLoadRejectsOutOfRangePort(t *testing.T) {
	for _, port := range []string{"0", "65536", "-1"} {
		if _, err := config.Load(lookupFrom(map[string]string{"NUVRO_HTTP_PORT": port})); !errors.Is(err, config.ErrInvalid) {
			t.Errorf("port %s error = %v, want ErrInvalid", port, err)
		}
	}
}

func TestLoadRejectsNonPositiveTimeout(t *testing.T) {
	_, err := config.Load(lookupFrom(map[string]string{"NUVRO_HTTP_IDLE_TIMEOUT": "0s"}))
	if !errors.Is(err, config.ErrInvalid) {
		t.Fatalf("error = %v, want ErrInvalid", err)
	}
	if !strings.Contains(err.Error(), "NUVRO_HTTP_IDLE_TIMEOUT") {
		t.Fatalf("error %q does not name the offending key", err)
	}
}

func TestLoadFailsFastWhenBackingServicesMissingOutsideDevelopment(t *testing.T) {
	_, err := config.Load(lookupFrom(map[string]string{"NUVRO_ENV": "staging"}))
	if !errors.Is(err, config.ErrInvalid) {
		t.Fatalf("error = %v, want ErrInvalid", err)
	}
	for _, fragment := range []string{"NUVRO_DATABASE_URL", "NUVRO_REDIS_URL"} {
		if !strings.Contains(err.Error(), fragment) {
			t.Errorf("error %q does not require %s", err, fragment)
		}
	}
}

func TestTrustedProxyConfigurationDefaultsToDirectDeployment(t *testing.T) {
	cfg, err := config.Load(lookupFrom(map[string]string{}))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Proxy.TrustedHops != 0 || len(cfg.Proxy.TrustedCIDRs) != 0 {
		t.Fatalf("proxy = %+v, want a direct deployment", cfg.Proxy)
	}
	resolver, err := cfg.Proxy.Resolver()
	if err != nil {
		t.Fatal(err)
	}
	if got := resolver.Resolve("203.0.113.9:51000", []string{"198.51.100.7"}); got != "203.0.113.9" {
		t.Fatalf("the default deployment honoured forwarding: %q", got)
	}
}

func TestTrustedProxyConfigurationReadsAChain(t *testing.T) {
	cfg, err := config.Load(lookupFrom(map[string]string{
		"NUVRO_TRUSTED_PROXY_HOPS":  "2",
		"NUVRO_TRUSTED_PROXY_CIDRS": "10.0.0.0/8, 2001:db8::/32 ,",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Proxy.TrustedHops != 2 || len(cfg.Proxy.TrustedCIDRs) != 2 {
		t.Fatalf("proxy = %+v, want two hops and two ranges", cfg.Proxy)
	}
	resolver, err := cfg.Proxy.Resolver()
	if err != nil {
		t.Fatal(err)
	}
	if got := resolver.Resolve("[2001:db8::4]:443", []string{"203.0.113.9, 10.0.0.7"}); got != "203.0.113.9" {
		t.Fatalf("the configured chain resolved to %q", got)
	}
}

func TestTrustedHopsWithoutTrustedRangesIsRefused(t *testing.T) {
	_, err := config.Load(lookupFrom(map[string]string{"NUVRO_TRUSTED_PROXY_HOPS": "1"}))
	if err == nil || !strings.Contains(err.Error(), "NUVRO_TRUSTED_PROXY_HOPS") {
		t.Fatalf("err = %v, want a refusal to trust forwarding with no trusted ranges", err)
	}
}

func TestAnInvalidTrustedRangeIsRefused(t *testing.T) {
	_, err := config.Load(lookupFrom(map[string]string{
		"NUVRO_TRUSTED_PROXY_HOPS": "1", "NUVRO_TRUSTED_PROXY_CIDRS": "10.0.0.4",
	}))
	if err == nil || !strings.Contains(err.Error(), "NUVRO_TRUSTED_PROXY_CIDRS") {
		t.Fatalf("err = %v, want a refusal of a bare address", err)
	}
}

func TestAnImplausibleHopCountIsRefused(t *testing.T) {
	_, err := config.Load(lookupFrom(map[string]string{
		"NUVRO_TRUSTED_PROXY_HOPS": "9", "NUVRO_TRUSTED_PROXY_CIDRS": "10.0.0.0/8",
	}))
	if err == nil || !strings.Contains(err.Error(), "NUVRO_TRUSTED_PROXY_HOPS") {
		t.Fatalf("err = %v, want a refusal of an implausible chain depth", err)
	}
}
