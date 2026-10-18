package apple

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
)

const DefaultExchangeEndpoint = "https://appleid.apple.com/auth/token"
const DefaultRevokeEndpoint = "https://appleid.apple.com/auth/revoke"
const maximumTokenPayload = 1 << 18

var ErrExchangeRejected = errors.New("apple: the authorization code was rejected")
var ErrRevokeFailed = errors.New("apple: the token could not be revoked")

type Identities struct {
	credentials    Credentials
	tokenEndpoint  string
	revokeEndpoint string
	client         *http.Client
	timer          clock.Clock
}

type IdentitySettings struct {
	Credentials    Credentials
	TokenEndpoint  string
	RevokeEndpoint string
	Timeout        time.Duration
}

func NewIdentities(settings IdentitySettings, timer clock.Clock) (*Identities, error) {
	if !settings.Credentials.Valid() {
		return nil, errors.New("apple: sign in credentials are incomplete")
	}
	if settings.TokenEndpoint == "" {
		settings.TokenEndpoint = DefaultExchangeEndpoint
	}
	if settings.RevokeEndpoint == "" {
		settings.RevokeEndpoint = DefaultRevokeEndpoint
	}
	if settings.Timeout <= 0 {
		settings.Timeout = 10 * time.Second
	}
	return &Identities{
		credentials: settings.Credentials, tokenEndpoint: settings.TokenEndpoint,
		revokeEndpoint: settings.RevokeEndpoint,
		client:         &http.Client{Timeout: settings.Timeout}, timer: timer,
	}, nil
}

func (i *Identities) Exchange(ctx context.Context, authorizationCode string) (string, error) {
	secret, err := i.credentials.clientSecret(i.timer.Now(), MaximumSecretLifetime)
	if err != nil {
		return "", err
	}
	form := url.Values{
		"client_id": {i.credentials.ClientID}, "client_secret": {secret},
		"code": {authorizationCode}, "grant_type": {"authorization_code"},
	}
	var payload struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := i.post(ctx, i.tokenEndpoint, form, &payload); err != nil {
		return "", ErrExchangeRejected
	}
	if payload.RefreshToken == "" {
		return "", ErrExchangeRejected
	}
	return payload.RefreshToken, nil
}

func (i *Identities) Revoke(ctx context.Context, refreshToken string) error {
	secret, err := i.credentials.clientSecret(i.timer.Now(), MaximumSecretLifetime)
	if err != nil {
		return err
	}
	form := url.Values{
		"client_id": {i.credentials.ClientID}, "client_secret": {secret},
		"token": {refreshToken}, "token_type_hint": {"refresh_token"},
	}
	if err := i.post(ctx, i.revokeEndpoint, form, nil); err != nil {
		return ErrRevokeFailed
	}
	return nil
}

func (i *Identities) post(ctx context.Context, endpoint string, form url.Values, into any) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := i.client.Do(request)
	if err != nil {
		return err
	}
	defer func() { _ = response.Body.Close() }()
	body, err := io.ReadAll(io.LimitReader(response.Body, maximumTokenPayload))
	if err != nil {
		return err
	}
	if response.StatusCode < 200 || response.StatusCode > 299 {
		return errors.New("apple: endpoint returned " + response.Status)
	}
	if into == nil {
		return nil
	}
	return json.Unmarshal(body, into)
}
