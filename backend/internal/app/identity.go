package app

import (
	"errors"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/sessiontoken"
	sessionusecase "github.com/SemiuAdesina/nuvro/backend/internal/usecase/session"

	"github.com/SemiuAdesina/nuvro/backend/internal/adapter/external/apple"
	"github.com/SemiuAdesina/nuvro/backend/internal/adapter/postgres"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/config"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/cursor"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/id"
	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/authentication"
	"github.com/SemiuAdesina/nuvro/backend/internal/usecase/account"
)

var errIdentityWithoutStore = errors.New("app: verified identity requires NUVRO_DATABASE_URL")

type identity struct {
	authenticator authentication.Authenticator
	sessions      *sessionusecase.Service
}

func newIdentity(cfg config.Config, store *postgres.Store, linker *account.Linker, timer clock.Clock) (identity, error) {
	verifier, err := newVerifier(cfg, timer)
	if err != nil || verifier == nil {
		return identity{}, err
	}
	if store == nil {
		return identity{}, errIdentityWithoutStore
	}
	resolver := account.NewResolver(store, store, id.NewUUID())
	sessions, err := newSessionService(cfg, store, verifier, resolver, linker, timer)
	if err != nil {
		return identity{}, err
	}
	authenticator := authentication.NewSessions(verifier, resolver).WithTokens(sessions)
	return identity{authenticator: authenticator, sessions: sessions}, nil
}

func newSessionService(cfg config.Config, store *postgres.Store, verifier authentication.Verifier,
	resolver *account.Resolver, linker *account.Linker, timer clock.Clock) (*sessionusecase.Service, error) {
	if !cfg.Authentication.IssuesSessions() {
		return nil, nil
	}
	signer, err := sessiontoken.NewSigner(cfg.Authentication.SessionSigningKey)
	if err != nil {
		return nil, err
	}
	dependencies := sessionusecase.Dependencies{
		Identity: verifier, Accounts: resolver, Store: store, Signer: signer,
		Identifiers: id.NewUUID(), Secrets: id.NewRandom(),
	}
	if linker != nil {
		dependencies.Linker = linker
	}
	return sessionusecase.NewService(dependencies, timer), nil
}

func newVerifier(cfg config.Config, timer clock.Clock) (authentication.Verifier, error) {
	settings := cfg.Authentication
	if settings.UsesDevelopmentCredential() {
		return authentication.NewDevelopmentVerifier(
			cfg.Environment == config.EnvironmentDevelopment,
			settings.DevelopmentCredential,
			settings.DevelopmentSubject,
		)
	}
	if !settings.UsesApple() {
		return nil, nil
	}
	return apple.NewVerifier(apple.Settings{
		Audience:    settings.AppleAudience,
		Issuer:      settings.AppleIssuer,
		KeyEndpoint: settings.AppleKeyEndpoint,
		KeyTTL:      settings.AppleKeyTTL,
	}, timer)
}

func newCursors(cfg config.Config) (cursor.Codec, error) {
	if len(cfg.Authentication.CursorSecret) > 0 {
		return cursor.New(cfg.Authentication.CursorSecret)
	}
	return cursor.NewEphemeral()
}
