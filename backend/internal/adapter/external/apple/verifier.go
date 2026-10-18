package apple

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
)

const ProviderName = "apple"
const DefaultKeyEndpoint = "https://appleid.apple.com/auth/keys"
const DefaultIssuer = "https://appleid.apple.com"

type Settings struct {
	Audience     string
	Issuer       string
	KeyEndpoint  string
	KeyTTL       time.Duration
	RefreshFloor time.Duration
	Leeway       time.Duration
	Timeout      time.Duration
}

type Verifier struct {
	settings Settings
	keys     *keyCache
	timer    clock.Clock
}

func NewVerifier(settings Settings, timer clock.Clock) (*Verifier, error) {
	if settings.Audience == "" {
		return nil, errors.New("apple: audience is required")
	}
	settings = withDefaults(settings)
	client := &http.Client{Timeout: settings.Timeout}
	return &Verifier{settings: settings, keys: newKeyCache(settings.KeyEndpoint, client, settings.KeyTTL, settings.RefreshFloor), timer: timer}, nil
}

func (v *Verifier) Verify(ctx context.Context, token string) (user.Identity, error) {
	parsed, err := parse(token)
	if err != nil {
		return user.Identity{}, err
	}
	now := v.timer.Now()
	key, err := v.keys.key(ctx, parsed.header.KeyID, now)
	if err != nil {
		return user.Identity{}, ErrInvalidToken
	}
	if err := parsed.verify(key, v.settings.Issuer, v.settings.Audience, now, v.settings.Leeway); err != nil {
		return user.Identity{}, ErrInvalidToken
	}
	return user.Identity{Provider: ProviderName, Subject: parsed.claims.Subject}, nil
}

func withDefaults(settings Settings) Settings {
	if settings.Issuer == "" {
		settings.Issuer = DefaultIssuer
	}
	if settings.KeyEndpoint == "" {
		settings.KeyEndpoint = DefaultKeyEndpoint
	}
	if settings.KeyTTL <= 0 {
		settings.KeyTTL = time.Hour
	}
	if settings.RefreshFloor <= 0 {
		settings.RefreshFloor = time.Minute
	}
	if settings.Leeway <= 0 {
		settings.Leeway = 30 * time.Second
	}
	if settings.Timeout <= 0 {
		settings.Timeout = 5 * time.Second
	}
	return settings
}
