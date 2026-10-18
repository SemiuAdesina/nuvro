package commerce

import (
	"context"
	"time"

	domain "github.com/SemiuAdesina/nuvro/backend/internal/domain/commerce"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/weather"
	engine "github.com/SemiuAdesina/nuvro/backend/internal/engine/commerce"
)

type ActivityRequest struct {
	TripID                string
	Date                  time.Time
	FreeFrom              time.Time
	FreeUntil             time.Time
	BudgetMinor           int64
	Interests             []string
	MaximumDistanceMeters int
	Weather               weather.Window
}

type ActivityResult struct {
	Offers []domain.ActivityOffer
	Status SearchStatus
}

func (s *Service) SearchActivities(ctx context.Context, principal user.Principal, request ActivityRequest) (ActivityResult, error) {
	record, centre, err := s.context(ctx, principal, request.TripID)
	if err != nil {
		return ActivityResult{}, err
	}
	now := s.timer.Now().UTC()
	itineraryCentre := s.itineraryCentre(ctx, request.TripID, centre)
	query := engine.ActivityQuery{
		Centre: itineraryCentre, Date: request.Date, Adults: int(record.Adults),
		Children: int(record.Children), Currency: record.Currency, Interests: request.Interests,
	}
	collected := make([]domain.ActivityOffer, 0)
	commercials := make([]domain.Commercials, 0)
	destinations := map[string]string{}
	status := SearchStatus{Ranking: engine.ActivityRankingVersion, SearchedAt: now}
	for _, provider := range s.providers.Activities {
		partner := provider.Partner()
		if !partner.Status.Usable() || !partner.Supports(domain.Activity) {
			status.Partners = append(status.Partners, partnerUnavailable(partner, "partner_disabled"))
			continue
		}
		offers, priced, providerErr := provider.Activities(ctx, query)
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
	ranked := engine.RankActivities(collected, engine.ActivityContext{
		Now: now, Centre: itineraryCentre, BudgetMinor: request.BudgetMinor,
		FreeFrom: request.FreeFrom, FreeUntil: request.FreeUntil, Weather: request.Weather,
		MaximumDistanceMeters: request.MaximumDistanceMeters,
	})
	s.store(ctx, activityOffers(ranked), commercials, destinations)
	return ActivityResult{Offers: ranked, Status: status}, nil
}

func activityOffers(offers []domain.ActivityOffer) []domain.Offer {
	values := make([]domain.Offer, 0, len(offers))
	for _, offer := range offers {
		values = append(values, offer.Offer)
	}
	return values
}
