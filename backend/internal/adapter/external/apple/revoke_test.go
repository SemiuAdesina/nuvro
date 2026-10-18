package apple_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/adapter/external/apple"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
)

const teamID = "ABCDE12345"
const keyID = "KEY1234567"
const clientID = "com.vectixlogic.nuvro"

func signingKey(t *testing.T) *ecdsa.PrivateKey {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	return key
}

func encodedKey(t *testing.T, key *ecdsa.PrivateKey) []byte {
	t.Helper()
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})
}

func credentials(t *testing.T) apple.Credentials {
	t.Helper()
	return apple.Credentials{TeamID: teamID, KeyID: keyID, ClientID: clientID, PrivateKey: signingKey(t)}
}

func identities(t *testing.T, server *httptest.Server) *apple.Identities {
	t.Helper()
	value, err := apple.NewIdentities(apple.IdentitySettings{
		Credentials:   credentials(t),
		TokenEndpoint: server.URL + "/auth/token", RevokeEndpoint: server.URL + "/auth/revoke",
	}, clock.NewFixed(time.Date(2026, 9, 9, 12, 0, 0, 0, time.UTC)))
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func TestAPrivateKeyRoundTripsThroughPEM(t *testing.T) {
	key := signingKey(t)
	parsed, err := apple.ParsePrivateKey(encodedKey(t, key))
	if err != nil {
		t.Fatal(err)
	}
	if !parsed.Equal(key) {
		t.Fatal("the parsed key differs from the one written")
	}
}

func TestANonECPrivateKeyIsRefused(t *testing.T) {
	for name, encoded := range map[string][]byte{
		"not pem":   []byte("nonsense"),
		"empty":     nil,
		"wrong der": pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: []byte("nonsense")}),
	} {
		if _, err := apple.ParsePrivateKey(encoded); !errors.Is(err, apple.ErrPrivateKey) {
			t.Fatalf("%s was accepted as a sign in key", name)
		}
	}
}

func TestTheClientSecretCarriesTheClaimsAppleRequires(t *testing.T) {
	var received url.Values
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Error(err)
		}
		received = r.PostForm
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"refresh_token":"apple-refresh-token"}`))
	}))
	defer server.Close()

	token, err := identities(t, server).Exchange(t.Context(), "code-1")
	if err != nil {
		t.Fatal(err)
	}
	if token != "apple-refresh-token" {
		t.Fatalf("token = %q", token)
	}
	claims := claimsOf(t, received.Get("client_secret"))
	if claims["iss"] != teamID || claims["sub"] != clientID || claims["aud"] != apple.AppleIDAudience {
		t.Fatalf("claims = %v", claims)
	}
	assertLifetime(t, claims)
	if received.Get("grant_type") != "authorization_code" || received.Get("code") != "code-1" {
		t.Fatalf("form = %v", received)
	}
}

func assertLifetime(t *testing.T, claims map[string]any) {
	t.Helper()
	issued, issuedOK := claims["iat"].(float64)
	expires, expiresOK := claims["exp"].(float64)
	if !issuedOK || !expiresOK {
		t.Fatalf("the secret carries no usable lifetime: %v", claims)
	}
	if expires <= issued || expires-issued > float64(apple.MaximumSecretLifetime/time.Second) {
		t.Fatalf("a secret living from %v to %v exceeds what Apple accepts", issued, expires)
	}
}

func TestTheClientSecretIsSignedWithES256AndTheKeyIdentifier(t *testing.T) {
	var received url.Values
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		received = r.PostForm
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	if err := identities(t, server).Revoke(t.Context(), "apple-refresh-token"); err != nil {
		t.Fatal(err)
	}
	header := headerOf(t, received.Get("client_secret"))
	if header["alg"] != "ES256" || header["kid"] != keyID {
		t.Fatalf("header = %v, want an ES256 secret naming the key", header)
	}
	if received.Get("token") != "apple-refresh-token" || received.Get("token_type_hint") != "refresh_token" {
		t.Fatalf("form = %v", received)
	}
	if received.Get("client_id") != clientID {
		t.Fatalf("client id = %q", received.Get("client_id"))
	}
}

func TestARejectedExchangeIsReportedNotSilentlyEmpty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid_grant"}`))
	}))
	defer server.Close()
	if _, err := identities(t, server).Exchange(t.Context(), "code-1"); !errors.Is(err, apple.ErrExchangeRejected) {
		t.Fatalf("err = %v, want a rejected exchange", err)
	}
}

func TestAnExchangeWithoutARefreshTokenIsRejected(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"access_token":"only-an-access-token"}`))
	}))
	defer server.Close()
	if _, err := identities(t, server).Exchange(t.Context(), "code-1"); !errors.Is(err, apple.ErrExchangeRejected) {
		t.Fatalf("err = %v, want a rejected exchange", err)
	}
}

func TestAFailedRevocationIsReported(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()
	if err := identities(t, server).Revoke(t.Context(), "token"); !errors.Is(err, apple.ErrRevokeFailed) {
		t.Fatalf("err = %v, want a failed revocation", err)
	}
}

func TestIncompleteCredentialsAreRefusedAtConstruction(t *testing.T) {
	for name, credentials := range map[string]apple.Credentials{
		"no team":   {KeyID: keyID, ClientID: clientID, PrivateKey: signingKey(t)},
		"no key id": {TeamID: teamID, ClientID: clientID, PrivateKey: signingKey(t)},
		"no client": {TeamID: teamID, KeyID: keyID, PrivateKey: signingKey(t)},
		"no key":    {TeamID: teamID, KeyID: keyID, ClientID: clientID},
	} {
		if credentials.Valid() {
			t.Fatalf("%s was treated as complete", name)
		}
		if _, err := apple.NewIdentities(apple.IdentitySettings{Credentials: credentials}, clock.NewSystem()); err == nil {
			t.Fatalf("%s was accepted", name)
		}
	}
}

func claimsOf(t *testing.T, token string) map[string]any {
	t.Helper()
	return segmentOf(t, token, 1)
}

func headerOf(t *testing.T, token string) map[string]any {
	t.Helper()
	return segmentOf(t, token, 0)
}

func segmentOf(t *testing.T, token string, index int) map[string]any {
	t.Helper()
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("the client secret is not a JWT: %q", token)
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[index])
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	return decoded
}
