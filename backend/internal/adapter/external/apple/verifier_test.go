package apple_test

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/adapter/external/apple"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
)

const audience = "com.vectixlogic.nuvro"

var issuedAt = time.Date(2026, 9, 9, 12, 0, 0, 0, time.UTC)

type issuer struct {
	key      *rsa.PrivateKey
	keyID    string
	requests atomic.Int64
	server   *httptest.Server
}

func newIssuer(t *testing.T) *issuer {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	source := &issuer{key: key, keyID: "primary"}
	source.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		source.requests.Add(1)
		payload := map[string]any{"keys": []map[string]string{{
			"kid": source.keyID, "kty": "RSA", "alg": "RS256",
			"n": base64.RawURLEncoding.EncodeToString(key.N.Bytes()),
			"e": base64.RawURLEncoding.EncodeToString(big.NewInt(int64(key.E)).Bytes()),
		}}}
		_ = json.NewEncoder(w).Encode(payload)
	}))
	t.Cleanup(source.server.Close)
	return source
}

func (i *issuer) token(t *testing.T, header, claims map[string]any) string {
	t.Helper()
	encode := func(value map[string]any) string {
		raw, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		return base64.RawURLEncoding.EncodeToString(raw)
	}
	signed := encode(header) + "." + encode(claims)
	digest := sha256.Sum256([]byte(signed))
	signature, err := rsa.SignPKCS1v15(rand.Reader, i.key, crypto.SHA256, digest[:])
	if err != nil {
		t.Fatal(err)
	}
	return signed + "." + base64.RawURLEncoding.EncodeToString(signature)
}

func (i *issuer) verifier(t *testing.T) *apple.Verifier {
	t.Helper()
	verifier, err := apple.NewVerifier(apple.Settings{Audience: audience, KeyEndpoint: i.server.URL}, clock.NewFixed(issuedAt))
	if err != nil {
		t.Fatal(err)
	}
	return verifier
}

func validClaims() map[string]any {
	return map[string]any{
		"iss": apple.DefaultIssuer, "aud": audience, "sub": "001234.abcdef",
		"iat": issuedAt.Add(-time.Minute).Unix(), "exp": issuedAt.Add(time.Hour).Unix(),
	}
}

func TestVerifierAcceptsAWellFormedIdentityToken(t *testing.T) {
	source := newIssuer(t)
	identity, err := source.verifier(t).Verify(t.Context(), source.token(t, map[string]any{"alg": "RS256", "kid": source.keyID}, validClaims()))
	if err != nil || identity.Provider != apple.ProviderName || identity.Subject != "001234.abcdef" {
		t.Fatalf("%+v %v", identity, err)
	}
}

func TestVerifierRejectsEveryUntrustedToken(t *testing.T) {
	source := newIssuer(t)
	other, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	forged := &issuer{key: other, keyID: source.keyID, server: source.server}
	cases := map[string]string{
		"wrong signer":   forged.token(t, map[string]any{"alg": "RS256", "kid": source.keyID}, validClaims()),
		"unknown key":    source.token(t, map[string]any{"alg": "RS256", "kid": "rotated-away"}, validClaims()),
		"algorithm none": source.token(t, map[string]any{"alg": "none", "kid": source.keyID}, validClaims()),
		"malformed":      "not-a-token",
		"empty":          "",
	}
	expired := validClaims()
	expired["exp"] = issuedAt.Add(-time.Hour).Unix()
	cases["expired"] = source.token(t, map[string]any{"alg": "RS256", "kid": source.keyID}, expired)
	future := validClaims()
	future["iat"] = issuedAt.Add(time.Hour).Unix()
	cases["not yet issued"] = source.token(t, map[string]any{"alg": "RS256", "kid": source.keyID}, future)
	foreignAudience := validClaims()
	foreignAudience["aud"] = "com.someone.else"
	cases["another audience"] = source.token(t, map[string]any{"alg": "RS256", "kid": source.keyID}, foreignAudience)
	foreignIssuer := validClaims()
	foreignIssuer["iss"] = "https://accounts.example.com"
	cases["another issuer"] = source.token(t, map[string]any{"alg": "RS256", "kid": source.keyID}, foreignIssuer)
	noSubject := validClaims()
	delete(noSubject, "sub")
	cases["no subject"] = source.token(t, map[string]any{"alg": "RS256", "kid": source.keyID}, noSubject)

	verifier := source.verifier(t)
	for name, token := range cases {
		identity, err := verifier.Verify(t.Context(), token)
		if err == nil {
			t.Fatalf("%s was accepted", name)
		}
		if identity.Subject != "" || identity.Provider != "" {
			t.Fatalf("%s leaked an identity", name)
		}
	}
}

func TestVerifierAcceptsAudienceArraysAndBoundsKeyRetrieval(t *testing.T) {
	source := newIssuer(t)
	claims := validClaims()
	claims["aud"] = []any{"com.someone.else", audience}
	verifier := source.verifier(t)
	token := source.token(t, map[string]any{"alg": "RS256", "kid": source.keyID}, claims)
	for range 5 {
		if _, err := verifier.Verify(t.Context(), token); err != nil {
			t.Fatal(err)
		}
	}
	if requests := source.requests.Load(); requests != 1 {
		t.Fatalf("key endpoint called %d times; the cache is not holding", requests)
	}
}

func TestVerifierRequiresAnAudience(t *testing.T) {
	if _, err := apple.NewVerifier(apple.Settings{}, clock.NewSystem()); err == nil {
		t.Fatal("verifier built without a configured audience")
	}
}
