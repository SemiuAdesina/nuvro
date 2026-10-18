package session

import (
	"context"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/apperror"
	domain "github.com/SemiuAdesina/nuvro/backend/internal/domain/session"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/sessiontoken"
)

const maximumCredential = 8192

type IdentityVerifier interface {
	Verify(context.Context, string) (user.Identity, error)
}

type Accounts interface {
	Resolve(context.Context, user.Identity) (user.Principal, error)
}

type Linker interface {
	Link(context.Context, user.Principal, string) error
}

type Store interface {
	InsertSession(context.Context, domain.Session, []byte) error
	SessionByRefreshHash(context.Context, []byte) (domain.Session, error)
	SessionAccess(context.Context, string) (bool, error)
	MarkSessionRotated(context.Context, string, time.Time) error
	RevokeSession(context.Context, string, time.Time) error
	RevokeSessionFamily(context.Context, string, time.Time) error
	RevokeUserSessions(context.Context, string, time.Time) error
}

type Identifiers interface{ Generate() (string, error) }

type Dependencies struct {
	Identity    IdentityVerifier
	Accounts    Accounts
	Linker      Linker
	Store       Store
	Signer      *sessiontoken.Signer
	Identifiers Identifiers
	Secrets     Identifiers
}

type Service struct {
	identity    IdentityVerifier
	accounts    Accounts
	linker      Linker
	store       Store
	signer      *sessiontoken.Signer
	identifiers Identifiers
	secrets     Identifiers
	timer       clock.Clock
}

func NewService(dependencies Dependencies, timer clock.Clock) *Service {
	return &Service{
		identity: dependencies.Identity, accounts: dependencies.Accounts, linker: dependencies.Linker,
		store: dependencies.Store, signer: dependencies.Signer, identifiers: dependencies.Identifiers,
		secrets: dependencies.Secrets, timer: timer,
	}
}

func (s *Service) Issue(ctx context.Context, credential, authorizationCode string) (domain.Grant, error) {
	if credential == "" || len(credential) > maximumCredential {
		return domain.Grant{}, unauthenticated()
	}
	identity, err := s.identity.Verify(ctx, credential)
	if err != nil {
		return domain.Grant{}, unauthenticated()
	}
	principal, err := s.accounts.Resolve(ctx, identity)
	if err != nil {
		return domain.Grant{}, err
	}
	if authorizationCode != "" && s.linker != nil {
		_ = s.linker.Link(ctx, principal, authorizationCode)
	}
	familyID, err := s.identifiers.Generate()
	if err != nil {
		return domain.Grant{}, internal(err)
	}
	return s.grant(ctx, principal.UserID, familyID)
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (domain.Grant, error) {
	if refreshToken == "" || len(refreshToken) > maximumCredential {
		return domain.Grant{}, unauthenticated()
	}
	now := s.timer.Now().UTC()
	current, err := s.store.SessionByRefreshHash(ctx, domain.HashRefreshToken(refreshToken))
	if err != nil {
		return domain.Grant{}, err
	}
	if current.Reused() {
		_ = s.store.RevokeSessionFamily(ctx, current.FamilyID, now)
		return domain.Grant{}, unauthenticated()
	}
	if !current.Usable(now) {
		return domain.Grant{}, unauthenticated()
	}
	if err := s.store.MarkSessionRotated(ctx, current.ID, now); err != nil {
		return domain.Grant{}, err
	}
	return s.grant(ctx, current.UserID, current.FamilyID)
}

func (s *Service) Revoke(ctx context.Context, principal user.Principal) error {
	if principal.Provider != domain.Provider || principal.Subject == "" {
		return nil
	}
	return s.store.RevokeSession(ctx, principal.Subject, s.timer.Now().UTC())
}

func (s *Service) RevokeAll(ctx context.Context, userID string) error {
	return s.store.RevokeUserSessions(ctx, userID, s.timer.Now().UTC())
}

func (s *Service) Authenticate(ctx context.Context, accessToken string) (user.Principal, bool, error) {
	if s == nil || !sessiontoken.Looks(accessToken) {
		return user.Principal{}, false, nil
	}
	claims, err := s.signer.Verify(accessToken, s.timer.Now())
	if err != nil {
		return user.Principal{}, true, unauthenticated()
	}
	usable, err := s.store.SessionAccess(ctx, claims.SessionID)
	if err != nil {
		return user.Principal{}, true, err
	}
	if !usable {
		return user.Principal{}, true, unauthenticated()
	}
	return user.Principal{UserID: claims.Subject, Provider: domain.Provider, Subject: claims.SessionID}, true, nil
}

func (s *Service) grant(ctx context.Context, userID, familyID string) (domain.Grant, error) {
	now := s.timer.Now().UTC()
	sessionID, err := s.identifiers.Generate()
	if err != nil {
		return domain.Grant{}, internal(err)
	}
	refreshToken, err := s.secrets.Generate()
	if err != nil {
		return domain.Grant{}, internal(err)
	}
	started := domain.Start(sessionID, familyID, userID, now)
	if insertErr := s.store.InsertSession(ctx, started, domain.HashRefreshToken(refreshToken)); insertErr != nil {
		return domain.Grant{}, insertErr
	}
	accessToken, err := s.signer.Sign(userID, sessionID, now, domain.AccessLifetime)
	if err != nil {
		return domain.Grant{}, internal(err)
	}
	return domain.Grant{
		SessionID: sessionID, UserID: userID, AccessToken: accessToken,
		RefreshToken: refreshToken, ExpiresAt: now.Add(domain.AccessLifetime),
	}, nil
}

func unauthenticated() error {
	return apperror.New(apperror.CodeUnauthenticated, "Authentication is required.")
}

func internal(cause error) error {
	return apperror.Wrap(apperror.CodeInternal, "The session could not be created.", cause)
}
