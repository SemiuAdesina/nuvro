package partner_test

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"

	"github.com/SemiuAdesina/nuvro/backend/internal/adapter/external/partner"
)

const secret = "synthetic-test-secret"

func sign(payload []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

func TestValidSignaturesAreAccepted(t *testing.T) {
	payload := []byte(`{"event":"booking.created"}`)
	verifier := partner.NewHMACVerifier("fixture", "X-Partner-Signature", secret)
	if verifier.Partner() != "fixture" {
		t.Fatalf("partner = %q", verifier.Partner())
	}
	if err := verifier.Verify(payload, map[string]string{"X-Partner-Signature": sign(payload)}); err != nil {
		t.Fatal(err)
	}
	if err := verifier.Verify(payload, map[string]string{"x-partner-signature": "sha256=" + sign(payload)}); err != nil {
		t.Fatalf("a prefixed lowercase header was rejected: %v", err)
	}
}

func TestUnsignedAndForgedRequestsAreRefused(t *testing.T) {
	payload := []byte(`{"event":"booking.created"}`)
	verifier := partner.NewHMACVerifier("fixture", "X-Partner-Signature", secret)
	if err := verifier.Verify(payload, nil); !errors.Is(err, partner.ErrUnsigned) {
		t.Fatalf("an unsigned request produced %v", err)
	}
	if err := verifier.Verify(payload, map[string]string{"X-Partner-Signature": "  "}); !errors.Is(err, partner.ErrUnsigned) {
		t.Fatal("a blank signature was accepted")
	}
	if err := verifier.Verify(payload, map[string]string{"X-Partner-Signature": sign([]byte("other"))}); !errors.Is(err, partner.ErrInvalidSignature) {
		t.Fatal("a signature over different content was accepted")
	}
	if err := verifier.Verify(payload, map[string]string{"X-Partner-Signature": "not-hex"}); !errors.Is(err, partner.ErrInvalidSignature) {
		t.Fatal("a malformed signature was accepted")
	}
	tampered := []byte(`{"event":"booking.completed"}`)
	if err := verifier.Verify(tampered, map[string]string{"X-Partner-Signature": sign(payload)}); !errors.Is(err, partner.ErrInvalidSignature) {
		t.Fatal("a tampered payload was accepted")
	}
}

func TestAnUnconfiguredSecretNeverAcceptsAnything(t *testing.T) {
	payload := []byte(`{"event":"booking.created"}`)
	verifier := partner.NewHMACVerifier("fixture", "X-Partner-Signature", "")
	if err := verifier.Verify(payload, map[string]string{"X-Partner-Signature": sign(payload)}); !errors.Is(err, partner.ErrNotConfigured) {
		t.Fatalf("an unconfigured verifier produced %v", err)
	}
}
