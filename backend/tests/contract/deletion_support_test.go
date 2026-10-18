package contract_test

import (
	"context"
	"testing"

	"github.com/SemiuAdesina/nuvro/backend/internal/apperror"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/retention"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/id"
	"github.com/SemiuAdesina/nuvro/backend/internal/usecase/account"
)

func (s *store) RequestDeletion(_ context.Context, deletion retention.Deletion) (retention.Deletion, error) {
	if s.deletions == nil {
		s.deletions = map[string]retention.Deletion{}
	}
	if existing, found := s.deletions[deletion.UserID]; found {
		return existing, nil
	}
	s.deletions[deletion.UserID] = deletion
	profile := s.profiles[deletion.UserID]
	s.profiles[deletion.UserID] = profile
	s.closed = append(s.closed, deletion.UserID)
	return deletion, nil
}

func (s *store) Deletion(_ context.Context, userID string) (retention.Deletion, error) {
	found, ok := s.deletions[userID]
	if !ok {
		return retention.Deletion{}, apperror.New(apperror.CodeNotFound, "No deletion is pending.")
	}
	return found, nil
}

func eraser(t *testing.T, repository *store, timer clock.Clock) *account.Eraser {
	t.Helper()
	return account.NewEraser(repository, id.NewUUID(), timer)
}
