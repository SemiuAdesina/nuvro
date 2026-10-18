package postgres

import (
	"context"
	"time"

	queries "github.com/SemiuAdesina/nuvro/backend/internal/adapter/postgres/sqlc"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/identity"
)

func (s *Store) LinkIdentityToken(ctx context.Context, userID, provider string, sealed []byte, linkedAt time.Time) error {
	identifier, err := identifier(userID)
	if err != nil {
		return err
	}
	return databaseError(s.queries(ctx).LinkIdentityToken(ctx, queries.LinkIdentityTokenParams{
		UserID: identifier, Provider: provider, SealedRefreshToken: sealed, LinkedAt: &linkedAt,
	}))
}

func (s *Store) RevocableIdentities(ctx context.Context, limit int) ([]identity.Sealed, error) {
	rows, err := s.queries(ctx).RevocableIdentities(ctx, bounded(limit))
	if err != nil {
		return nil, databaseError(err)
	}
	return toSealed(rows), nil
}

func (s *Store) IdentitiesForUser(ctx context.Context, userID string) ([]identity.Sealed, error) {
	identifier, err := identifier(userID)
	if err != nil {
		return nil, err
	}
	rows, err := s.queries(ctx).IdentitiesForUser(ctx, identifier)
	if err != nil {
		return nil, databaseError(err)
	}
	sealed := make([]identity.Sealed, 0, len(rows))
	for _, row := range rows {
		sealed = append(sealed, identity.Sealed{
			ID: uuidText(row.ID), UserID: uuidText(row.UserID),
			Provider: row.Provider, Token: row.SealedRefreshToken,
		})
	}
	return sealed, nil
}

func (s *Store) MarkIdentityRevoked(ctx context.Context, identityID string, revokedAt time.Time) error {
	identifier, err := identifier(identityID)
	if err != nil {
		return err
	}
	return databaseError(s.queries(ctx).MarkIdentityRevoked(ctx, queries.MarkIdentityRevokedParams{
		ID: identifier, RevokedAt: &revokedAt,
	}))
}

func toSealed(rows []queries.RevocableIdentitiesRow) []identity.Sealed {
	sealed := make([]identity.Sealed, 0, len(rows))
	for _, row := range rows {
		sealed = append(sealed, identity.Sealed{
			ID: uuidText(row.ID), UserID: uuidText(row.UserID),
			Provider: row.Provider, Token: row.SealedRefreshToken,
		})
	}
	return sealed
}
