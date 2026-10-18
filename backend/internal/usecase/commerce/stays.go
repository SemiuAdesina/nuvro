package commerce

import (
	"context"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/apperror"
	domain "github.com/SemiuAdesina/nuvro/backend/internal/domain/commerce"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	engine "github.com/SemiuAdesina/nuvro/backend/internal/engine/commerce"
)

type StayRequest struct {
	TripID            string
	CheckIn           time.Time
	CheckOut          time.Time
	BudgetMinor       int64
	RadiusMeters      int
	MaximumWalkMeters int
}

type StayResult struct {
	Offers []domain.StayOffer
	Status SearchStatus
}

func (s *Service) SearchStays(ctx context.Context, principal user.Principal, request StayRequest) (StayResult, error) {
	record, centre, err := s.context(ctx, principal, request.TripID)
	if err != nil {
		return StayResult{}, err
	}
	if request.CheckIn.IsZero() || request.CheckOut.IsZero() || !request.CheckOut.After(request.CheckIn) {
		return StayResult{}, apperror.New(apperror.CodeInvalidInput, "Provide a check-in date before the check-out date.")
	}
	now := s.timer.Now().UTC()
	itineraryCentre := s.itineraryCentre(ctx, request.TripID, centre)
	query := engine.StayQuery{
		Centre: itineraryCentre, CheckIn: request.CheckIn, CheckOut: request.CheckOut,
		Adults: int(record.Adults), Children: int(record.Children),
		Currency: record.Currency, RadiusMeters: request.RadiusMeters,
	}
	collected := make([]domain.StayOffer, 0)
	commercials := make([]domain.Commercials, 0)
	destinations := map[string]string{}
	status := SearchStatus{Ranking: engine.StayRankingVersion, SearchedAt: now}
	for _, provider := range s.providers.Stays {
		partner := provider.Partner()
		if !partner.Status.Usable() || !partner.Supports(domain.Stay) {
			status.Partners = append(status.Partners, partnerUnavailable(partner, "partner_disabled"))
			continue
		}
		offers, priced, providerErr := provider.Stays(ctx, query)
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
	ranked := engine.RankStays(collected, engine.StayContext{
		Now: now, ItineraryCentre: itineraryCentre, BudgetMinor: request.BudgetMinor,
		MaximumWalkMeters: request.MaximumWalkMeters,
	})
	s.store(ctx, stayOffers(ranked), commercials, destinations)
	return StayResult{Offers: ranked, Status: status}, nil
}

func stayOffers(offers []domain.StayOffer) []domain.Offer {
	values := make([]domain.Offer, 0, len(offers))
	for _, offer := range offers {
		values = append(values, offer.Offer)
	}
	return values
}
