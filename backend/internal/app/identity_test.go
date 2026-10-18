package app_test

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/SemiuAdesina/nuvro/backend/internal/app"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/config"
)

func settings(overrides map[string]string) config.Lookup {
	return func(key string) (string, bool) {
		value, ok := overrides[key]
		return value, ok
	}
}

func TestAPIStartsWithoutIdentityInDevelopment(t *testing.T) {
	cfg, err := config.Load(settings(map[string]string{}))
	if err != nil {
		t.Fatal(err)
	}
	api, err := app.NewAPI(cfg, &strings.Builder{})
	if err != nil || api == nil {
		t.Fatalf("the health-only development server did not start: %v", err)
	}
}

func TestVerifiedIdentityRequiresAStore(t *testing.T) {
	cfg, err := config.Load(settings(map[string]string{"NUVRO_APPLE_AUDIENCE": "com.vectixlogic.nuvro"}))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := app.NewAPI(cfg, &strings.Builder{}); err == nil {
		t.Fatal("identity was configured without a database")
	}
}

func TestDevelopmentCredentialIsRefusedWhenTheEnvironmentIsNotDevelopment(t *testing.T) {
	overrides := map[string]string{
		"NUVRO_ENV":                 "staging",
		"NUVRO_DATABASE_URL":        "postgres://localhost:5432/nuvro",
		"NUVRO_REDIS_URL":           "redis://localhost:6379/0",
		"NUVRO_APPLE_AUDIENCE":      "com.vectixlogic.nuvro",
		"NUVRO_CURSOR_SECRET":       base64.StdEncoding.EncodeToString([]byte(strings.Repeat("s", 32))),
		"NUVRO_DEV_AUTH_CREDENTIAL": "synthetic-development-credential",
	}
	if _, err := config.Load(settings(overrides)); err == nil {
		t.Fatal("staging accepted a development credential")
	}
}
