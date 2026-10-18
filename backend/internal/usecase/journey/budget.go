package journey

import (
	"context"

	"github.com/SemiuAdesina/nuvro/backend/internal/apperror"
	domain "github.com/SemiuAdesina/nuvro/backend/internal/domain/budget"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/money"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	budgetengine "github.com/SemiuAdesina/nuvro/backend/internal/engine/budget"
)

func (s *Service) Budget(ctx context.Context, principal user.Principal, tripID string) (domain.Trip, error) {
	if s.accountant == nil {
		return domain.Trip{}, apperror.New(apperror.CodeUnavailable, "Budget evidence is unavailable.")
	}
	record, err := s.trips.Read(ctx, principal, tripID)
	if err != nil {
		return domain.Trip{}, err
	}
	currency, err := money.ParseCurrency(record.Currency)
	if err != nil {
		return domain.Trip{}, apperror.New(apperror.CodeInvalidState, "This trip has no valid currency.")
	}
	nights := 0
	if record.StartDate != nil && record.EndDate != nil {
		nights = int(record.EndDate.Sub(*record.StartDate).Hours() / 24)
	}
	country := ""
	if record.DestinationID != nil {
		if place, lookupErr := s.catalog.Destination(ctx, *record.DestinationID); lookupErr == nil {
			country = place.CountryCode
		}
	}
	return s.accountant.Calculate(ctx, budgetengine.Input{
		Query: budgetengine.Query{
			TripID: tripID, Nights: max(nights, 1), Travelers: int(record.Adults + record.Children),
			Currency: currency, Country: country,
		},
		Activities: s.activityCosts(ctx, tripID, currency),
		On:         s.timer.Now(),
	})
}

func (s *Service) activityCosts(ctx context.Context, tripID string, currency money.Currency) []money.Amount {
	plan, err := s.store.Itinerary(ctx, tripID)
	if err != nil {
		return nil
	}
	amounts := make([]money.Amount, 0, len(plan.Items))
	for _, item := range plan.Items {
		if item.EstimatedCostMinor <= 0 || item.CostCurrency != currency.String() {
			continue
		}
		amount, amountErr := money.New(item.EstimatedCostMinor, currency)
		if amountErr != nil {
			continue
		}
		amounts = append(amounts, amount)
	}
	return amounts
}
