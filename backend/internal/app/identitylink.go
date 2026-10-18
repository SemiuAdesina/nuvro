package app

import (
	"github.com/SemiuAdesina/nuvro/backend/internal/adapter/external/apple"
	"github.com/SemiuAdesina/nuvro/backend/internal/adapter/postgres"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/config"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/secret"
	"github.com/SemiuAdesina/nuvro/backend/internal/usecase/account"
)

func newLinker(cfg config.Config, store *postgres.Store, timer clock.Clock) (*account.Linker, error) {
	settings := cfg.Authentication
	if store == nil || !settings.RevokesAppleTokens() {
		return nil, nil
	}
	privateKey, err := apple.ParsePrivateKey(settings.ApplePrivateKey)
	if err != nil {
		return nil, err
	}
	identities, err := apple.NewIdentities(apple.IdentitySettings{
		Credentials: apple.Credentials{
			TeamID: settings.AppleTeamID, KeyID: settings.AppleKeyID,
			ClientID: settings.AppleAudience, PrivateKey: privateKey,
		},
		TokenEndpoint:  settings.AppleTokenEndpoint,
		RevokeEndpoint: settings.AppleRevokeEndpoint,
	}, timer)
	if err != nil {
		return nil, err
	}
	sealer, err := secret.NewSealer(settings.TokenSealingKey)
	if err != nil {
		return nil, err
	}
	return account.NewLinker(identities, store, sealer, apple.ProviderName, timer), nil
}

func revoker(cfg config.Config, store *postgres.Store, timer clock.Clock) *account.Linker {
	linker, err := newLinker(cfg, store, timer)
	if err != nil {
		return nil
	}
	return linker
}
