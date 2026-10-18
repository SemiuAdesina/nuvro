package commerce

import (
	"context"
	"errors"

	"github.com/SemiuAdesina/nuvro/backend/internal/apperror"
	domain "github.com/SemiuAdesina/nuvro/backend/internal/domain/commerce"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	engine "github.com/SemiuAdesina/nuvro/backend/internal/engine/commerce"
)

type HandoffRequest struct {
	OfferID string
	TripID  string
}

func (s *Service) CreateHandoff(ctx context.Context, principal user.Principal, request HandoffRequest) (domain.Handoff, error) {
	if principal.UserID == "" {
		return domain.Handoff{}, apperror.New(apperror.CodeUnauthenticated, "Authentication is required.")
	}
	if s.engine == nil {
		return domain.Handoff{}, apperror.New(apperror.CodeUnavailable, "Booking partners are unavailable.")
	}
	if request.TripID != "" {
		if _, err := s.trips.Read(ctx, principal, request.TripID); err != nil {
			return domain.Handoff{}, err
		}
	}
	handoff, err := s.engine.CreateHandoff(ctx, engine.HandoffRequest{
		OfferID: request.OfferID, TripID: request.TripID, UserID: principal.UserID,
	})
	if err != nil {
		return domain.Handoff{}, handoffError(err)
	}
	if err := s.handoffs.SaveHandoff(ctx, handoff); err != nil {
		return domain.Handoff{}, err
	}
	return handoff, nil
}

func (s *Service) FollowHandoff(ctx context.Context, principal user.Principal, handoffID string) (string, error) {
	if principal.UserID == "" {
		return "", apperror.New(apperror.CodeUnauthenticated, "Authentication is required.")
	}
	handoff, err := s.handoffs.Handoff(ctx, handoffID)
	if err != nil || handoff.UserID != principal.UserID {
		return "", apperror.New(apperror.CodeNotFound, "Handoff not found.")
	}
	destination, err := s.engine.Destination(handoff)
	if err != nil {
		return "", handoffError(err)
	}
	clickID, err := s.identifiers.Generate()
	if err != nil {
		return "", apperror.Wrap(apperror.CodeInternal, "", err)
	}
	click := domain.Click{
		ID: clickID, HandoffID: handoff.ID, Partner: handoff.Partner,
		Product: handoff.Product, OccurredAt: s.timer.Now().UTC(),
	}
	if err := s.handoffs.RecordClick(ctx, click); err != nil {
		return "", err
	}
	return destination, nil
}

func handoffError(err error) error {
	switch {
	case errors.Is(err, engine.ErrOfferExpired):
		return apperror.Wrap(apperror.CodeInvalidState, "That option is no longer available. Search again for current prices.", err)
	case errors.Is(err, engine.ErrOfferUnavailable):
		return apperror.Wrap(apperror.CodeInvalidState, "That option is no longer available. Search again for current prices.", err)
	case errors.Is(err, engine.ErrPartnerUnavailable):
		return apperror.Wrap(apperror.CodeUnavailable, "That booking partner is unavailable right now.", err)
	case errors.Is(err, domain.ErrUntrustedDestination):
		return apperror.Wrap(apperror.CodeInvalidState, "That option cannot be opened safely.", err)
	default:
		return apperror.Wrap(apperror.CodeInternal, "", err)
	}
}
