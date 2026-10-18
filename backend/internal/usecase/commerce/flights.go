package commerce

import (
	"context"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/apperror"
	domain "github.com/SemiuAdesina/nuvro/backend/internal/domain/commerce"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	engine "github.com/SemiuAdesina/nuvro/backend/internal/engine/commerce"
)

type FlightRequest struct {
	TripID           string
	Origin           string
	Destination      string
	DepartDate       time.Time
	ReturnDate       time.Time
	Cabin            string
	BudgetMinor      int64
	MaximumMinutes   int
	PreferFewerStops bool
}

type FlightResult struct {
	Offers []domain.FlightOffer
	Status SearchStatus
}

func (s *Service) SearchFlights(ctx context.Context, principal user.Principal, request FlightRequest) (FlightResult, error) {
	record, _, err := s.context(ctx, principal, request.TripID)
	if err != nil {
		return FlightResult{}, err
	}
	if request.Origin == "" || request.Destination == "" {
		return FlightResult{}, apperror.New(apperror.CodeInvalidInput, "An origin and destination are required.")
	}
	now := s.timer.Now().UTC()
	query := engine.FlightQuery{
		Origin: request.Origin, Destination: request.Destination,
		DepartDate: request.DepartDate, ReturnDate: request.ReturnDate,
		Adults: int(record.Adults), Children: int(record.Children),
		Cabin: request.Cabin, Currency: record.Currency,
	}
	collected := make([]domain.FlightOffer, 0)
	commercials := make([]domain.Commercials, 0)
	destinations := map[string]string{}
	status := SearchStatus{Ranking: engine.FlightRankingVersion, SearchedAt: now}
	for _, provider := range s.providers.Flights {
		partner := provider.Partner()
		if !partner.Status.Usable() || !partner.Supports(domain.Flight) {
			status.Partners = append(status.Partners, partnerUnavailable(partner, "partner_disabled"))
			continue
		}
		offers, priced, providerErr := provider.Flights(ctx, query)
		if providerErr != nil {
			status.Partners = append(status.Partners, partnerUnavailable(partner, "partner_unavailable"))
			continue
		}
		status.Partners = append(status.Partners, PartnerStatus{Partner: partner.Code, Available: true})
		for _, offer := range offers {
			collected = append(collected, offer)
			destinations[offer.ID] = offer.TermsURL
		}
		commercials = append(commercials, priced...)
	}
	ranked := engine.RankFlights(collected, engine.FlightContext{
		Now: now, PreferFewerStops: request.PreferFewerStops,
		MaximumMinutes: request.MaximumMinutes, BudgetMinor: request.BudgetMinor,
	})
	s.store(ctx, flightOffers(ranked), commercials, destinations)
	return FlightResult{Offers: ranked, Status: status}, nil
}

func flightOffers(offers []domain.FlightOffer) []domain.Offer {
	values := make([]domain.Offer, 0, len(offers))
	for _, offer := range offers {
		values = append(values, offer.Offer)
	}
	return values
}
