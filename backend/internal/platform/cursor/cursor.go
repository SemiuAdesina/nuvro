package cursor

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strings"
	"time"
)

const minimumSecretLength = 32
const maximumTokenLength = 512

var ErrInvalid = errors.New("cursor: invalid cursor")
var ErrWeakSecret = errors.New("cursor: secret must be at least 32 bytes")

type Position struct {
	CreatedAt time.Time
	ID        string
}

type Codec struct{ secret []byte }

func New(secret []byte) (Codec, error) {
	if len(secret) < minimumSecretLength {
		return Codec{}, ErrWeakSecret
	}
	return Codec{secret: append([]byte(nil), secret...)}, nil
}

func NewEphemeral() (Codec, error) {
	secret := make([]byte, minimumSecretLength)
	if _, err := rand.Read(secret); err != nil {
		return Codec{}, err
	}
	return New(secret)
}

func (c Codec) Encode(scope, principal string, position Position) (string, error) {
	if scope == "" || principal == "" || position.ID == "" {
		return "", ErrInvalid
	}
	payload := position.CreatedAt.UTC().Format(time.RFC3339Nano) + "\x1f" + position.ID
	body := encode(payload)
	token := body + "." + encode(string(c.sign(scope, principal, body)))
	if len(token) > maximumTokenLength {
		return "", ErrInvalid
	}
	return token, nil
}

func (c Codec) Decode(scope, principal, token string) (Position, error) {
	if len(token) > maximumTokenLength {
		return Position{}, ErrInvalid
	}
	body, signature, found := strings.Cut(token, ".")
	if !found {
		return Position{}, ErrInvalid
	}
	presented, err := base64.RawURLEncoding.DecodeString(signature)
	if err != nil || !hmac.Equal(presented, c.sign(scope, principal, body)) {
		return Position{}, ErrInvalid
	}
	payload, err := base64.RawURLEncoding.DecodeString(body)
	if err != nil {
		return Position{}, ErrInvalid
	}
	instant, identifier, found := strings.Cut(string(payload), "\x1f")
	if !found || identifier == "" {
		return Position{}, ErrInvalid
	}
	createdAt, err := time.Parse(time.RFC3339Nano, instant)
	if err != nil {
		return Position{}, ErrInvalid
	}
	return Position{CreatedAt: createdAt.UTC(), ID: identifier}, nil
}

func (c Codec) sign(scope, principal, body string) []byte {
	mac := hmac.New(sha256.New, c.secret)
	mac.Write([]byte(scope))
	mac.Write([]byte{0x1e})
	mac.Write([]byte(principal))
	mac.Write([]byte{0x1e})
	mac.Write([]byte(body))
	return mac.Sum(nil)
}

func encode(value string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(value))
}
