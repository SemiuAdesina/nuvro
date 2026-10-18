package contract_test

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/apperror"
	sessiondomain "github.com/SemiuAdesina/nuvro/backend/internal/domain/session"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/id"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/sessiontoken"
	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/authentication"
	sessionusecase "github.com/SemiuAdesina/nuvro/backend/internal/usecase/session"
)

func (s *store) InsertSession(_ context.Context, value sessiondomain.Session, hash []byte) error {
	if s.sessions == nil {
		s.sessions = map[string]sessiondomain.Session{}
		s.sessionHashes = map[string]string{}
	}
	s.sessions[value.ID] = value
	s.sessionHashes[string(hash)] = value.ID
	return nil
}

func (s *store) SessionByRefreshHash(_ context.Context, hash []byte) (sessiondomain.Session, error) {
	sessionID, found := s.sessionHashes[string(hash)]
	if !found {
		return sessiondomain.Session{}, apperror.New(apperror.CodeUnauthenticated, "Authentication is required.")
	}
	return s.sessions[sessionID], nil
}

func (s *store) SessionAccess(_ context.Context, sessionID string) (bool, error) {
	found, ok := s.sessions[sessionID]
	if !ok || found.RevokedAt != nil {
		return false, nil
	}
	for _, closed := range s.closed {
		if closed == found.UserID {
			return false, nil
		}
	}
	return true, nil
}

func (s *store) MarkSessionRotated(_ context.Context, sessionID string, at time.Time) error {
	found := s.sessions[sessionID]
	found.RotatedAt = &at
	s.sessions[sessionID] = found
	return nil
}

func (s *store) RevokeSession(_ context.Context, sessionID string, at time.Time) error {
	found := s.sessions[sessionID]
	found.RevokedAt = &at
	s.sessions[sessionID] = found
	return nil
}

func (s *store) RevokeSessionFamily(_ context.Context, familyID string, at time.Time) error {
	for key, found := range s.sessions {
		if found.FamilyID == familyID {
			found.RevokedAt = &at
			s.sessions[key] = found
		}
	}
	return nil
}

func (s *store) RevokeUserSessions(_ context.Context, userID string, at time.Time) error {
	for key, found := range s.sessions {
		if found.UserID == userID {
			found.RevokedAt = &at
			s.sessions[key] = found
		}
	}
	return nil
}

func sessionService(t *testing.T, repository *store, verifier authentication.Verifier, timer clock.Clock) *sessionusecase.Service {
	t.Helper()
	signer, err := sessiontoken.NewSigner(bytes.Repeat([]byte{'s'}, sessiontoken.MinimumKeyLength))
	if err != nil {
		t.Fatal(err)
	}
	return sessionusecase.NewService(sessionusecase.Dependencies{
		Identity: verifier, Accounts: resolver{}, Store: repository, Signer: signer,
		Identifiers: id.NewUUID(), Secrets: id.NewRandom(),
	}, timer)
}
