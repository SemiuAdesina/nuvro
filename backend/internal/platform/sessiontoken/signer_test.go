package sessiontoken_test

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/platform/sessiontoken"
)

const traveller = "3f1c9c8e-0e3a-4f0a-9b1f-1a2b3c4d5e6f"
const sessionID = "9a8b7c6d-5e4f-4a3b-8c2d-1e0f9a8b7c6d"

func issued() time.Time { return time.Date(2026, 9, 9, 12, 0, 0, 0, time.UTC) }

func signer(t *testing.T, fill byte) *sessiontoken.Signer {
	t.Helper()
	value, err := sessiontoken.NewSigner(bytes.Repeat([]byte{fill}, sessiontoken.MinimumKeyLength))
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func TestASignedTokenVerifiesWithTheSameKey(t *testing.T) {
	token, err := signer(t, 'k').Sign(traveller, sessionID, issued(), 15*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	claims, err := signer(t, 'k').Verify(token, issued().Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if claims.Subject != traveller || claims.SessionID != sessionID || claims.Issuer != sessiontoken.Issuer {
		t.Fatalf("claims = %+v", claims)
	}
	if !sessiontoken.Looks(token) {
		t.Fatal("a session token was not recognised as one")
	}
}

func TestATokenIsRefusedOnceItExpires(t *testing.T) {
	token, err := signer(t, 'k').Sign(traveller, sessionID, issued(), 15*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := signer(t, 'k').Verify(token, issued().Add(15*time.Minute)); !errors.Is(err, sessiontoken.ErrInvalid) {
		t.Fatalf("err = %v, want an expired token refused at the boundary", err)
	}
}

func TestAnotherKeyCannotVerifyTheToken(t *testing.T) {
	token, err := signer(t, 'k').Sign(traveller, sessionID, issued(), time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := signer(t, 'j').Verify(token, issued()); !errors.Is(err, sessiontoken.ErrInvalid) {
		t.Fatalf("err = %v, want a refusal", err)
	}
}

func TestTamperedClaimsAreRefused(t *testing.T) {
	token, err := signer(t, 'k').Sign(traveller, sessionID, issued(), time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	parts := strings.Split(token, ".")
	forged := base64.RawURLEncoding.EncodeToString([]byte(`{"iss":"nuvro","sub":"someone-else","jti":"x","iat":1,"exp":9999999999}`))
	if _, err := signer(t, 'k').Verify(parts[0]+"."+forged+"."+parts[2], issued()); !errors.Is(err, sessiontoken.ErrInvalid) {
		t.Fatalf("err = %v, want forged claims refused", err)
	}
}

func TestUnsignedAndForeignTokensAreRefused(t *testing.T) {
	none := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none"}`))
	claims := base64.RawURLEncoding.EncodeToString([]byte(`{"iss":"nuvro","sub":"a","jti":"b","iat":1,"exp":9999999999}`))
	for name, token := range map[string]string{
		"alg none":     none + "." + claims + ".",
		"apple shaped": base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256","kid":"k"}`)) + "." + claims + ".sig",
		"two segments": "a.b",
		"empty":        "",
		"oversized":    strings.Repeat("a", 5000),
		"wrong issuer": signedWithIssuer(t, "someone"),
	} {
		if _, err := signer(t, 'k').Verify(token, issued()); !errors.Is(err, sessiontoken.ErrInvalid) {
			t.Fatalf("%s was accepted: %v", name, err)
		}
	}
	if sessiontoken.Looks("not.a.token") || sessiontoken.Looks("") {
		t.Fatal("garbage looks like a session token")
	}
}

func signedWithIssuer(t *testing.T, issuer string) string {
	t.Helper()
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	claims := base64.RawURLEncoding.EncodeToString([]byte(
		`{"iss":"` + issuer + `","sub":"` + traveller + `","jti":"` + sessionID + `","iat":1,"exp":9999999999}`))
	mac := hmac.New(sha256.New, bytes.Repeat([]byte{'k'}, sessiontoken.MinimumKeyLength))
	mac.Write([]byte(header + "." + claims))
	return header + "." + claims + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func TestAShortKeyAndANilSignerAreRefused(t *testing.T) {
	if _, err := sessiontoken.NewSigner(bytes.Repeat([]byte{'k'}, 16)); !errors.Is(err, sessiontoken.ErrKey) {
		t.Fatal("a 16 byte key was accepted")
	}
	var absent *sessiontoken.Signer
	if _, err := absent.Sign(traveller, sessionID, issued(), time.Hour); !errors.Is(err, sessiontoken.ErrKey) {
		t.Fatal("a nil signer signed a token")
	}
	if _, err := absent.Verify("a.b.c", issued()); !errors.Is(err, sessiontoken.ErrInvalid) {
		t.Fatal("a nil signer verified a token")
	}
}
