package config_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"strings"
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/platform/config"
)

var productionCursorSecret = base64.StdEncoding.EncodeToString([]byte(strings.Repeat("s", 32)))

var productionSigningKey = base64.StdEncoding.EncodeToString([]byte(strings.Repeat("g", 32)))

var productionSealingKey = base64.StdEncoding.EncodeToString([]byte(strings.Repeat("t", 32)))

func signingKeyPEM() string {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic(err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}))
}

func productionSettings() map[string]string {
	return map[string]string{
		"NUVRO_ENV":                 "production",
		"NUVRO_DATABASE_URL":        "postgres://localhost:5432/nuvro",
		"NUVRO_REDIS_URL":           "redis://localhost:6379/0",
		"NUVRO_APPLE_AUDIENCE":      "com.vectixlogic.nuvro",
		"NUVRO_CURSOR_SECRET":       productionCursorSecret,
		"NUVRO_APPLE_TEAM_ID":       "ABCDE12345",
		"NUVRO_APPLE_KEY_ID":        "KEY1234567",
		"NUVRO_APPLE_PRIVATE_KEY":   signingKeyPEM(),
		"NUVRO_TOKEN_SEALING_KEY":   productionSealingKey,
		"NUVRO_SESSION_SIGNING_KEY": productionSigningKey,
	}
}

func TestAuthenticationDefaultsAndOverrides(t *testing.T) {
	cfg, err := config.Load(lookupFrom(map[string]string{}))
	if err != nil {
		t.Fatal(err)
	}
	settings := cfg.Authentication
	if settings.AppleIssuer != "https://appleid.apple.com" || settings.AppleKeyEndpoint != "https://appleid.apple.com/auth/keys" {
		t.Fatalf("unexpected Apple defaults: %+v", settings)
	}
	if settings.AppleKeyTTL != time.Hour || settings.UsesApple() || settings.UsesDevelopmentCredential() {
		t.Fatalf("development defaults are not inert: %+v", settings)
	}
	configured := productionSettings()
	configured["NUVRO_APPLE_KEY_TTL"] = "15m"
	cfg, loadErr := config.Load(lookupFrom(configured))
	if loadErr != nil {
		t.Fatal(loadErr)
	}
	if !cfg.Authentication.UsesApple() || cfg.Authentication.AppleKeyTTL != 15*time.Minute {
		t.Fatalf("overrides ignored: %+v", cfg.Authentication)
	}
	if len(cfg.Authentication.CursorSecret) != 32 {
		t.Fatal("cursor secret was not decoded")
	}
}

func TestDeployedEnvironmentsRequireVerifiedIdentity(t *testing.T) {
	for _, missing := range []string{
		"NUVRO_APPLE_AUDIENCE", "NUVRO_CURSOR_SECRET", "NUVRO_APPLE_TEAM_ID",
		"NUVRO_APPLE_KEY_ID", "NUVRO_APPLE_PRIVATE_KEY", "NUVRO_TOKEN_SEALING_KEY", "NUVRO_SESSION_SIGNING_KEY",
	} {
		settings := productionSettings()
		delete(settings, missing)
		if _, err := config.Load(lookupFrom(settings)); err == nil {
			t.Fatalf("production started without %s", missing)
		}
	}
}

func TestDevelopmentCredentialIsRefusedOutsideDevelopment(t *testing.T) {
	settings := productionSettings()
	settings["NUVRO_DEV_AUTH_CREDENTIAL"] = "synthetic-development-credential"
	_, err := config.Load(lookupFrom(settings))
	if err == nil {
		t.Fatal("production accepted a development credential")
	}
	if !strings.Contains(err.Error(), "NUVRO_DEV_AUTH_CREDENTIAL") {
		t.Fatalf("unclear failure: %v", err)
	}
	weak := map[string]string{"NUVRO_DEV_AUTH_CREDENTIAL": "short"}
	if _, weakErr := config.Load(lookupFrom(weak)); weakErr == nil {
		t.Fatal("weak development credential accepted")
	}
	accepted := map[string]string{"NUVRO_DEV_AUTH_CREDENTIAL": "synthetic-development-credential"}
	cfg, acceptedErr := config.Load(lookupFrom(accepted))
	if acceptedErr != nil || !cfg.Authentication.UsesDevelopmentCredential() {
		t.Fatalf("development credential rejected in development: %v", acceptedErr)
	}
}

func TestAuthenticationRejectsMalformedSettings(t *testing.T) {
	for name, override := range map[string]map[string]string{
		"plaintext secret":   {"NUVRO_CURSOR_SECRET": "not base64!!"},
		"short secret":       {"NUVRO_CURSOR_SECRET": base64.StdEncoding.EncodeToString([]byte("tooshort"))},
		"insecure endpoint":  {"NUVRO_APPLE_KEY_ENDPOINT": "http://appleid.apple.com/auth/keys"},
		"negative key cache": {"NUVRO_APPLE_KEY_TTL": "-1m"},
	} {
		if _, err := config.Load(lookupFrom(override)); err == nil {
			t.Fatalf("%s was accepted", name)
		}
	}
}

func TestASealingKeyOfTheWrongLengthIsRefused(t *testing.T) {
	settings := productionSettings()
	settings["NUVRO_TOKEN_SEALING_KEY"] = base64.StdEncoding.EncodeToString([]byte(strings.Repeat("t", 16)))
	if _, err := config.Load(lookupFrom(settings)); err == nil {
		t.Fatal("a 16 byte sealing key was accepted")
	}
}

func TestASignInKeyThatIsNotAnECPrivateKeyIsRefused(t *testing.T) {
	settings := productionSettings()
	settings["NUVRO_APPLE_PRIVATE_KEY"] = base64.StdEncoding.EncodeToString([]byte("not a pem block"))
	if _, err := config.Load(lookupFrom(settings)); err == nil {
		t.Fatal("a key that is not a PKCS#8 EC private key was accepted")
	}
}

func TestAProductionConfigurationCanRevokeSignInTokens(t *testing.T) {
	cfg, err := config.Load(lookupFrom(productionSettings()))
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Authentication.RevokesAppleTokens() {
		t.Fatal("a complete production configuration cannot revoke Sign in with Apple tokens")
	}
}

func TestAShortSessionSigningKeyIsRefused(t *testing.T) {
	settings := productionSettings()
	settings["NUVRO_SESSION_SIGNING_KEY"] = base64.StdEncoding.EncodeToString([]byte(strings.Repeat("g", 16)))
	if _, err := config.Load(lookupFrom(settings)); err == nil {
		t.Fatal("a 16 byte session signing key was accepted")
	}
}

func TestDevelopmentIssuesSessionsOnlyWhenAKeyIsPresent(t *testing.T) {
	without, err := config.Load(lookupFrom(map[string]string{}))
	if err != nil {
		t.Fatal(err)
	}
	if without.Authentication.IssuesSessions() {
		t.Fatal("development claimed to issue sessions with no signing key")
	}
	with, err := config.Load(lookupFrom(map[string]string{"NUVRO_SESSION_SIGNING_KEY": productionSigningKey}))
	if err != nil {
		t.Fatal(err)
	}
	if !with.Authentication.IssuesSessions() {
		t.Fatal("development cannot issue sessions with a signing key")
	}
}
