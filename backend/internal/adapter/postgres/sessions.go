package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	queries "github.com/SemiuAdesina/nuvro/backend/internal/adapter/postgres/sqlc"
	"github.com/SemiuAdesina/nuvro/backend/internal/apperror"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/session"
)

func (s *Store) InsertSession(ctx context.Context, value session.Session, refreshHash []byte) error {
	sessionID, err := identifier(value.ID)
	if err != nil {
		return err
	}
	userID, err := identifier(value.UserID)
	if err != nil {
		return err
	}
	familyID, err := identifier(value.FamilyID)
	if err != nil {
		return err
	}
	return databaseError(s.queries(ctx).InsertSession(ctx, queries.InsertSessionParams{
		ID: sessionID, UserID: userID, FamilyID: familyID, RefreshHash: refreshHash,
		IssuedAt: value.IssuedAt.UTC(), ExpiresAt: value.ExpiresAt.UTC(),
	}))
}

func (s *Store) SessionByRefreshHash(ctx context.Context, refreshHash []byte) (session.Session, error) {
	row, err := s.queries(ctx).GetSessionByRefreshHash(ctx, refreshHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return session.Session{}, apperror.New(apperror.CodeUnauthenticated, "Authentication is required.")
	}
	if err != nil {
		return session.Session{}, databaseError(err)
	}
	return session.Session{
		ID: uuidText(row.ID), UserID: uuidText(row.UserID), FamilyID: uuidText(row.FamilyID),
		IssuedAt: row.IssuedAt.UTC(), ExpiresAt: row.ExpiresAt.UTC(), RotatedAt: row.RotatedAt, RevokedAt: row.RevokedAt,
	}, nil
}

func (s *Store) SessionAccess(ctx context.Context, sessionID string) (bool, error) {
	identity, err := identifier(sessionID)
	if err != nil {
		return false, err
	}
	row, err := s.queries(ctx).GetSessionAccess(ctx, identity)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, databaseError(err)
	}
	return row.RevokedAt == nil && row.Status == "active", nil
}

func (s *Store) MarkSessionRotated(ctx context.Context, sessionID string, at time.Time) error {
	identity, err := identifier(sessionID)
	if err != nil {
		return err
	}
	at = at.UTC()
	return databaseError(s.queries(ctx).MarkSessionRotated(ctx, queries.MarkSessionRotatedParams{ID: identity, RotatedAt: &at}))
}

func (s *Store) RevokeSession(ctx context.Context, sessionID string, at time.Time) error {
	identity, err := identifier(sessionID)
	if err != nil {
		return err
	}
	at = at.UTC()
	return databaseError(s.queries(ctx).RevokeSession(ctx, queries.RevokeSessionParams{ID: identity, RevokedAt: &at}))
}

func (s *Store) RevokeSessionFamily(ctx context.Context, familyID string, at time.Time) error {
	identity, err := identifier(familyID)
	if err != nil {
		return err
	}
	at = at.UTC()
	return databaseError(s.queries(ctx).RevokeSessionFamily(ctx, queries.RevokeSessionFamilyParams{FamilyID: identity, RevokedAt: &at}))
}

func (s *Store) RevokeUserSessions(ctx context.Context, userID string, at time.Time) error {
	identity, err := identifier(userID)
	if err != nil {
		return err
	}
	at = at.UTC()
	return databaseError(s.queries(ctx).RevokeUserSessions(ctx, queries.RevokeUserSessionsParams{UserID: identity, RevokedAt: &at}))
}
