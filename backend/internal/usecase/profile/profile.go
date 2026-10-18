package profile

import (
	"context"

	"github.com/SemiuAdesina/nuvro/backend/internal/apperror"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
)

type Repository interface {
	Profile(context.Context, string) (user.Profile, error)
	SaveProfile(context.Context, user.Profile) error
	Preferences(context.Context, string) (user.Preferences, error)
	SavePreferences(context.Context, string, user.Preferences) error
}

type Transaction interface {
	Within(context.Context, func(context.Context) error) error
}

type Service struct {
	repository  Repository
	transaction Transaction
}

func NewService(repository Repository, transaction Transaction) *Service {
	return &Service{repository: repository, transaction: transaction}
}

func (s *Service) Profile(ctx context.Context, principal user.Principal) (user.Profile, error) {
	if err := requirePrincipal(principal); err != nil {
		return user.Profile{}, err
	}
	return s.repository.Profile(ctx, principal.UserID)
}

func (s *Service) UpdateProfile(ctx context.Context, principal user.Principal, patch user.ProfilePatch) (user.Profile, error) {
	if err := requirePrincipal(principal); err != nil {
		return user.Profile{}, err
	}
	var updated user.Profile
	err := s.transaction.Within(ctx, func(txctx context.Context) error {
		current, err := s.repository.Profile(txctx, principal.UserID)
		if err != nil {
			return err
		}
		updated, err = current.Apply(patch)
		if err != nil {
			return apperror.New(apperror.CodeInvalidInput, "Profile values are invalid.")
		}
		return s.repository.SaveProfile(txctx, updated)
	})
	if err != nil {
		return user.Profile{}, err
	}
	return updated, nil
}

func (s *Service) Preferences(ctx context.Context, principal user.Principal) (user.Preferences, error) {
	if err := requirePrincipal(principal); err != nil {
		return user.Preferences{}, err
	}
	return s.repository.Preferences(ctx, principal.UserID)
}

func (s *Service) UpdatePreferences(ctx context.Context, principal user.Principal, patch user.PreferencesPatch) (user.Preferences, error) {
	if err := requirePrincipal(principal); err != nil {
		return user.Preferences{}, err
	}
	var updated user.Preferences
	err := s.transaction.Within(ctx, func(txctx context.Context) error {
		current, err := s.repository.Preferences(txctx, principal.UserID)
		if err != nil {
			return err
		}
		updated, err = current.Apply(patch)
		if err != nil {
			return apperror.New(apperror.CodeInvalidInput, "Preference values are invalid.")
		}
		return s.repository.SavePreferences(txctx, principal.UserID, updated)
	})
	if err != nil {
		return user.Preferences{}, err
	}
	return updated, nil
}

func requirePrincipal(principal user.Principal) error {
	if principal.UserID == "" {
		return apperror.New(apperror.CodeUnauthenticated, "Authentication is required.")
	}
	return nil
}
