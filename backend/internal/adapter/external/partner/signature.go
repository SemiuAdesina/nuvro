package partner

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
)

var ErrUnsigned = errors.New("partner: request carries no signature")
var ErrInvalidSignature = errors.New("partner: signature does not match")
var ErrNotConfigured = errors.New("partner: no webhook secret configured")

const MaximumPayloadBytes = 256 << 10

type Verifier interface {
	Partner() string
	Verify(payload []byte, headers map[string]string) error
}

type HMACVerifier struct {
	partner string
	secret  []byte
	header  string
}

func NewHMACVerifier(partnerCode, header, secret string) *HMACVerifier {
	return &HMACVerifier{partner: partnerCode, header: header, secret: []byte(secret)}
}

func (v *HMACVerifier) Partner() string { return v.partner }

func (v *HMACVerifier) Verify(payload []byte, headers map[string]string) error {
	if len(v.secret) == 0 {
		return ErrNotConfigured
	}
	provided := strings.TrimSpace(headerValue(headers, v.header))
	if provided == "" {
		return ErrUnsigned
	}
	decoded, err := hex.DecodeString(strings.TrimPrefix(provided, "sha256="))
	if err != nil {
		return ErrInvalidSignature
	}
	mac := hmac.New(sha256.New, v.secret)
	mac.Write(payload)
	if !hmac.Equal(decoded, mac.Sum(nil)) {
		return ErrInvalidSignature
	}
	return nil
}

func headerValue(headers map[string]string, name string) string {
	for key, value := range headers {
		if strings.EqualFold(key, name) {
			return value
		}
	}
	return ""
}
