package sessiontoken

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

const Issuer = "nuvro"
const MinimumKeyLength = 32
const maximumTokenLength = 4096

var ErrKey = errors.New("sessiontoken: a signing key of at least 32 bytes is required")
var ErrInvalid = errors.New("sessiontoken: the token is not a valid session token")

type Claims struct {
	Issuer    string `json:"iss"`
	Subject   string `json:"sub"`
	SessionID string `json:"jti"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp"`
}

type Signer struct {
	key []byte
}

func NewSigner(key []byte) (*Signer, error) {
	if len(key) < MinimumKeyLength {
		return nil, ErrKey
	}
	return &Signer{key: key}, nil
}

func (s *Signer) Sign(userID, sessionID string, issuedAt time.Time, lifetime time.Duration) (string, error) {
	if s == nil {
		return "", ErrKey
	}
	header, err := segment(map[string]string{"alg": "HS256", "typ": "JWT"})
	if err != nil {
		return "", err
	}
	claims, err := segment(Claims{
		Issuer: Issuer, Subject: userID, SessionID: sessionID,
		IssuedAt: issuedAt.Unix(), ExpiresAt: issuedAt.Add(lifetime).Unix(),
	})
	if err != nil {
		return "", err
	}
	signed := header + "." + claims
	return signed + "." + s.signature(signed), nil
}

func (s *Signer) Verify(token string, now time.Time) (Claims, error) {
	if s == nil || len(token) > maximumTokenLength {
		return Claims{}, ErrInvalid
	}
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return Claims{}, ErrInvalid
	}
	var header struct {
		Algorithm string `json:"alg"`
	}
	if err := decode(parts[0], &header); err != nil || header.Algorithm != "HS256" {
		return Claims{}, ErrInvalid
	}
	expected := s.signature(parts[0] + "." + parts[1])
	if !hmac.Equal([]byte(expected), []byte(parts[2])) {
		return Claims{}, ErrInvalid
	}
	var claims Claims
	if err := decode(parts[1], &claims); err != nil || !claims.usable(now) {
		return Claims{}, ErrInvalid
	}
	return claims, nil
}

func (c Claims) usable(now time.Time) bool {
	if c.Issuer != Issuer || c.Subject == "" || c.SessionID == "" {
		return false
	}
	return c.ExpiresAt != 0 && now.Before(time.Unix(c.ExpiresAt, 0))
}

func (s *Signer) signature(payload string) string {
	mac := hmac.New(sha256.New, s.key)
	mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func segment(value any) (string, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(encoded), nil
}

func decode(part string, into any) error {
	raw, err := base64.RawURLEncoding.DecodeString(part)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, into)
}

func Looks(token string) bool {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return false
	}
	var header struct {
		Algorithm string `json:"alg"`
	}
	return decode(parts[0], &header) == nil && header.Algorithm == "HS256"
}
