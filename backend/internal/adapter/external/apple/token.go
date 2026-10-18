package apple

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

const maximumTokenLength = 8192

var ErrInvalidToken = errors.New("apple: identity token rejected")

type header struct {
	Algorithm string `json:"alg"`
	KeyID     string `json:"kid"`
}

type claims struct {
	Issuer    string `json:"iss"`
	Subject   string `json:"sub"`
	Audience  any    `json:"aud"`
	ExpiresAt int64  `json:"exp"`
	IssuedAt  int64  `json:"iat"`
}

type parsedToken struct {
	header  header
	claims  claims
	signed  []byte
	digest  []byte
	rawSign []byte
}

func parse(token string) (parsedToken, error) {
	if len(token) == 0 || len(token) > maximumTokenLength {
		return parsedToken{}, ErrInvalidToken
	}
	first := strings.IndexByte(token, '.')
	last := strings.LastIndexByte(token, '.')
	if first <= 0 || last <= first || last == len(token)-1 {
		return parsedToken{}, ErrInvalidToken
	}
	var result parsedToken
	if err := decodeSegment(token[:first], &result.header); err != nil {
		return parsedToken{}, err
	}
	if err := decodeSegment(token[first+1:last], &result.claims); err != nil {
		return parsedToken{}, err
	}
	signature, err := base64.RawURLEncoding.DecodeString(token[last+1:])
	if err != nil {
		return parsedToken{}, ErrInvalidToken
	}
	result.rawSign = signature
	result.signed = []byte(token[:last])
	digest := sha256.Sum256(result.signed)
	result.digest = digest[:]
	return result, nil
}

func decodeSegment(segment string, target any) error {
	raw, err := base64.RawURLEncoding.DecodeString(segment)
	if err != nil {
		return ErrInvalidToken
	}
	if err := json.Unmarshal(raw, target); err != nil {
		return ErrInvalidToken
	}
	return nil
}

func (t parsedToken) verify(key *rsa.PublicKey, issuer, audience string, now time.Time, leeway time.Duration) error {
	if t.header.Algorithm != "RS256" || t.claims.Issuer != issuer || t.claims.Subject == "" {
		return ErrInvalidToken
	}
	if !matchesAudience(t.claims.Audience, audience) {
		return ErrInvalidToken
	}
	if t.claims.ExpiresAt == 0 || now.After(time.Unix(t.claims.ExpiresAt, 0).Add(leeway)) {
		return ErrInvalidToken
	}
	if t.claims.IssuedAt != 0 && now.Add(leeway).Before(time.Unix(t.claims.IssuedAt, 0)) {
		return ErrInvalidToken
	}
	if err := rsa.VerifyPKCS1v15(key, crypto.SHA256, t.digest, t.rawSign); err != nil {
		return ErrInvalidToken
	}
	return nil
}

func matchesAudience(value any, expected string) bool {
	switch typed := value.(type) {
	case string:
		return typed == expected
	case []any:
		for _, candidate := range typed {
			if text, ok := candidate.(string); ok && text == expected {
				return true
			}
		}
	}
	return false
}
