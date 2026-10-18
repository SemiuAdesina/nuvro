package commerce

import (
	"context"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/apperror"
	commercedomain "github.com/SemiuAdesina/nuvro/backend/internal/domain/commerce"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/money"
	domain "github.com/SemiuAdesina/nuvro/backend/internal/domain/pricewatch"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
)

type WatchRequest struct {
	TripID        string
	Product       string
	Target        domain.Target
	BaselineMinor int64
	Currency      string
	Threshold     domain.Threshold
	ExpiresAt     time.Time
}

func (s *Service) CreateWatch(ctx context.Context, principal user.Principal, request WatchRequest) (domain.Watch, error) {
	if principal.UserID == "" {
		return domain.Watch{}, apperror.New(apperror.CodeUnauthenticated, "Authentication is required.")
	}
	if err := s.assertEntitled(ctx, principal); err != nil {
		return domain.Watch{}, err
	}
	count, err := s.watches.ActivePriceWatchCount(ctx, principal.UserID)
	if err != nil {
		return domain.Watch{}, err
	}
	if count >= domain.MaximumPerUser {
		return domain.Watch{}, apperror.New(apperror.CodeConflict, "You are already watching the maximum number of prices.")
	}
	if request.TripID != "" {
		if _, tripErr := s.trips.Read(ctx, principal, request.TripID); tripErr != nil {
			return domain.Watch{}, tripErr
		}
	}
	baseline, err := money.New(request.BaselineMinor, money.Currency(request.Currency))
	if err != nil {
		return domain.Watch{}, apperror.New(apperror.CodeInvalidInput, "A valid baseline price is required.")
	}
	identifier, err := s.identifiers.Generate()
	if err != nil {
		return domain.Watch{}, apperror.Wrap(apperror.CodeInternal, "", err)
	}
	now := s.timer.Now().UTC()
	threshold := request.Threshold
	if threshold.PercentageDrop == 0 {
		threshold = domain.DefaultThreshold()
	}
	watch := domain.Watch{
		ID: identifier, UserID: principal.UserID, TripID: request.TripID,
		Product: commercedomain.ProductType(request.Product), Target: request.Target,
		Baseline: baseline, Threshold: threshold, Status: domain.Active,
		CreatedAt: now, NextCheckAt: NextCheck(now, request.Target.StartDate),
		ExpiresAt: expiryFor(request, now),
	}
	if err := watch.Validate(); err != nil {
		return domain.Watch{}, apperror.Wrap(apperror.CodeInvalidInput, "That price watch is not valid.", err)
	}
	if err := s.watches.SavePriceWatch(ctx, watch); err != nil {
		return domain.Watch{}, err
	}
	return watch, nil
}

func (s *Service) Watches(ctx context.Context, principal user.Principal) ([]domain.Watch, error) {
	if principal.UserID == "" {
		return nil, apperror.New(apperror.CodeUnauthenticated, "Authentication is required.")
	}
	return s.watches.PriceWatches(ctx, principal.UserID)
}

func (s *Service) DeleteWatch(ctx context.Context, principal user.Principal, watchID string) error {
	if principal.UserID == "" {
		return apperror.New(apperror.CodeUnauthenticated, "Authentication is required.")
	}
	watch, err := s.watches.PriceWatch(ctx, watchID)
	if err != nil || watch.UserID != principal.UserID {
		return apperror.New(apperror.CodeNotFound, "Price watch not found.")
	}
	return s.watches.EndPriceWatch(ctx, watchID, domain.Ended)
}

func (s *Service) assertEntitled(ctx context.Context, principal user.Principal) error {
	if s.entitlements == nil {
		return apperror.New(apperror.CodeUnavailable, "Price monitoring is not available yet.")
	}
	allowed, err := s.entitlements.AllowsPriceMonitoring(ctx, principal)
	if err != nil {
		return err
	}
	if !allowed {
		return apperror.New(apperror.CodeForbidden, "Price monitoring needs an active subscription.")
	}
	return nil
}

func expiryFor(request WatchRequest, now time.Time) time.Time {
	if !request.ExpiresAt.IsZero() {
		return request.ExpiresAt.UTC()
	}
	if !request.Target.StartDate.IsZero() && request.Target.StartDate.After(now) {
		return request.Target.StartDate.UTC()
	}
	return now.Add(24 * time.Hour)
}

func NextCheck(now, departure time.Time) time.Time {
	if departure.IsZero() {
		return now.Add(24 * time.Hour)
	}
	remaining := departure.UTC().Sub(now.UTC())
	switch {
	case remaining <= 0:
		return now.Add(24 * time.Hour)
	case remaining <= 7*24*time.Hour:
		return now.Add(6 * time.Hour)
	case remaining <= 30*24*time.Hour:
		return now.Add(12 * time.Hour)
	default:
		return now.Add(48 * time.Hour)
	}
}
