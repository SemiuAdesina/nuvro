package account

import (
	"context"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/apperror"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/identity"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/secret"
)

const maximumAuthorizationCode = 512

type Provider interface {
	Exchange(context.Context, string) (string, error)
	Revoke(context.Context, string) error
}

type IdentityStore interface {
	LinkIdentityToken(ctx context.Context, userID, provider string, sealed []byte, linkedAt time.Time) error
	RevocableIdentities(context.Context, int) ([]identity.Sealed, error)
	IdentitiesForUser(context.Context, string) ([]identity.Sealed, error)
	MarkIdentityRevoked(ctx context.Context, identityID string, revokedAt time.Time) error
}

type Linker struct {
	provider Provider
	store    IdentityStore
	sealer   *secret.Sealer
	name     string
	timer    clock.Clock
}

func NewLinker(provider Provider, store IdentityStore, sealer *secret.Sealer, name string, timer clock.Clock) *Linker {
	return &Linker{provider: provider, store: store, sealer: sealer, name: name, timer: timer}
}

func (l *Linker) Link(ctx context.Context, principal user.Principal, authorizationCode string) error {
	if principal.UserID == "" {
		return apperror.New(apperror.CodeUnauthenticated, "Authentication is required.")
	}
	if authorizationCode == "" || len(authorizationCode) > maximumAuthorizationCode {
		return apperror.New(apperror.CodeInvalidInput, "That authorization code is not usable.")
	}
	if l == nil || l.provider == nil || l.sealer == nil {
		return apperror.New(apperror.CodeUnavailable, "Identity linking is unavailable.")
	}
	refreshToken, err := l.provider.Exchange(ctx, authorizationCode)
	if err != nil {
		return apperror.New(apperror.CodeInvalidInput, "That authorization code was rejected.")
	}
	sealed, err := l.sealer.Seal(refreshToken, []byte(principal.UserID))
	if err != nil {
		return apperror.New(apperror.CodeInternal, "The identity could not be stored.")
	}
	return l.store.LinkIdentityToken(ctx, principal.UserID, l.name, sealed, l.timer.Now().UTC())
}

func (l *Linker) RevokePending(ctx context.Context, limit int) (int, error) {
	if l == nil || l.provider == nil || l.sealer == nil {
		return 0, nil
	}
	pending, err := l.store.RevocableIdentities(ctx, limit)
	if err != nil {
		return 0, err
	}
	return l.revokeAll(ctx, pending)
}

func (l *Linker) RevokeFor(ctx context.Context, userID string) (int, error) {
	if l == nil || l.provider == nil || l.sealer == nil {
		return 0, nil
	}
	held, err := l.store.IdentitiesForUser(ctx, userID)
	if err != nil {
		return 0, err
	}
	return l.revokeAll(ctx, held)
}

func (l *Linker) revokeAll(ctx context.Context, held []identity.Sealed) (int, error) {
	revoked := 0
	for _, sealed := range held {
		if !sealed.Revocable() || sealed.Provider != l.name {
			continue
		}
		done, err := l.revoke(ctx, sealed)
		if err != nil {
			return revoked, err
		}
		if done {
			revoked++
		}
	}
	return revoked, nil
}

func (l *Linker) revoke(ctx context.Context, sealed identity.Sealed) (bool, error) {
	refreshToken, err := l.sealer.Open(sealed.Token, []byte(sealed.UserID))
	usable := err == nil
	if usable && !l.accepted(ctx, refreshToken) {
		return false, nil
	}
	return true, l.store.MarkIdentityRevoked(ctx, sealed.ID, l.timer.Now().UTC())
}

func (l *Linker) accepted(ctx context.Context, refreshToken string) bool {
	return l.provider.Revoke(ctx, refreshToken) == nil
}
