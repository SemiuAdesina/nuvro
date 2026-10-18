package apple

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"math/big"
	"time"
)

const AppleIDAudience = "https://appleid.apple.com"
const MaximumSecretLifetime = 15777000 * time.Second

var ErrPrivateKey = errors.New("apple: the sign in key must be a PKCS#8 EC private key")

type Credentials struct {
	TeamID     string
	KeyID      string
	ClientID   string
	PrivateKey *ecdsa.PrivateKey
}

func ParsePrivateKey(pemBytes []byte) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, ErrPrivateKey
	}
	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, ErrPrivateKey
	}
	key, ok := parsed.(*ecdsa.PrivateKey)
	if !ok {
		return nil, ErrPrivateKey
	}
	return key, nil
}

func (c Credentials) Valid() bool {
	return c.TeamID != "" && c.KeyID != "" && c.ClientID != "" && c.PrivateKey != nil
}

func (c Credentials) clientSecret(now time.Time, lifetime time.Duration) (string, error) {
	if !c.Valid() {
		return "", errors.New("apple: sign in credentials are incomplete")
	}
	if lifetime <= 0 || lifetime > MaximumSecretLifetime {
		lifetime = MaximumSecretLifetime
	}
	header, err := segment(map[string]string{"alg": "ES256", "kid": c.KeyID, "typ": "JWT"})
	if err != nil {
		return "", err
	}
	claims, err := segment(map[string]any{
		"iss": c.TeamID, "iat": now.Unix(), "exp": now.Add(lifetime).Unix(),
		"aud": AppleIDAudience, "sub": c.ClientID,
	})
	if err != nil {
		return "", err
	}
	signed := header + "." + claims
	signature, err := c.sign(signed)
	if err != nil {
		return "", err
	}
	return signed + "." + signature, nil
}

func (c Credentials) sign(payload string) (string, error) {
	digest := sha256.Sum256([]byte(payload))
	r, s, err := ecdsa.Sign(rand.Reader, c.PrivateKey, digest[:])
	if err != nil {
		return "", err
	}
	size := (c.PrivateKey.Curve.Params().BitSize + 7) / 8
	raw := make([]byte, 2*size)
	pad(raw[:size], r)
	pad(raw[size:], s)
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func pad(destination []byte, value *big.Int) {
	bytes := value.Bytes()
	copy(destination[len(destination)-len(bytes):], bytes)
}

func segment(value any) (string, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(encoded), nil
}
