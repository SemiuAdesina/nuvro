package contract_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/apperror"
	domain "github.com/SemiuAdesina/nuvro/backend/internal/domain/commerce"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/geo"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/money"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/pricewatch"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	engine "github.com/SemiuAdesina/nuvro/backend/internal/engine/commerce"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/id"
	commerceusecase "github.com/SemiuAdesina/nuvro/backend/internal/usecase/commerce"
)

func (s *store) SaveHandoff(_ context.Context, handoff domain.Handoff) error {
	if s.handoffs == nil {
		s.handoffs = map[string]domain.Handoff{}
	}
	s.handoffs[handoff.ID] = handoff
	return nil
}

func (s *store) Handoff(_ context.Context, handoffID string) (domain.Handoff, error) {
	handoff, found := s.handoffs[handoffID]
	if !found {
		return domain.Handoff{}, errors.New("handoff not found")
	}
	return handoff, nil
}

func (s *store) RecordClick(context.Context, domain.Click) error { return nil }

func (s *store) SavePriceWatch(_ context.Context, watch pricewatch.Watch) error {
	if s.watches == nil {
		s.watches = map[string]pricewatch.Watch{}
	}
	s.watches[watch.ID] = watch
	return nil
}

func (s *store) PriceWatches(_ context.Context, userID string) ([]pricewatch.Watch, error) {
	values := make([]pricewatch.Watch, 0, len(s.watches))
	for _, watch := range s.watches {
		if watch.UserID == userID {
			values = append(values, watch)
		}
	}
	return values, nil
}

func (s *store) PriceWatch(_ context.Context, watchID string) (pricewatch.Watch, error) {
	watch, found := s.watches[watchID]
	if !found {
		return pricewatch.Watch{}, errors.New("watch not found")
	}
	return watch, nil
}

func (s *store) ActivePriceWatchCount(_ context.Context, userID string) (int64, error) {
	count := int64(0)
	for _, watch := range s.watches {
		if watch.UserID == userID && watch.Status == pricewatch.Active {
			count++
		}
	}
	return count, nil
}

func (s *store) EndPriceWatch(_ context.Context, watchID string, status pricewatch.Status) error {
	watch, found := s.watches[watchID]
	if !found {
		return errors.New("watch not found")
	}
	watch.Status = status
	s.watches[watchID] = watch
	return nil
}

func contractPartner() domain.Partner {
	return domain.Partner{
		Code: "fixture", DisplayName: "Fixture Travel", Status: domain.Enabled,
		Products:       []domain.ProductType{domain.Flight, domain.Stay, domain.Activity},
		AllowedDomains: []string{"partner.invalid"},
	}
}

type contractRegistry struct{}

func (contractRegistry) Partner(code string) (domain.Partner, bool) {
	if code != "fixture" {
		return domain.Partner{}, false
	}
	return contractPartner(), true
}

type contractOffers struct{ offers map[string]engine.StoredOffer }

func (c *contractOffers) SaveOffers(_ context.Context, offers []engine.StoredOffer, _ time.Duration) error {
	if c.offers == nil {
		c.offers = map[string]engine.StoredOffer{}
	}
	for _, offer := range offers {
		c.offers[offer.Offer.ID] = offer
	}
	return nil
}

func (c *contractOffers) Offer(_ context.Context, offerID string) (engine.StoredOffer, bool) {
	offer, found := c.offers[offerID]
	return offer, found
}

type flightPartnerFixture struct{}

func (flightPartnerFixture) Partner() domain.Partner { return contractPartner() }

func (flightPartnerFixture) Flights(context.Context, engine.FlightQuery) ([]domain.FlightOffer, []domain.Commercials, error) {
	build := func(id string, minor int64, stops, minutes int) domain.FlightOffer {
		price, _ := money.New(minor, money.Currency("EUR"))
		return domain.FlightOffer{
			Offer: domain.Offer{
				ID: id, Partner: "fixture", Product: domain.Flight, Title: id, Price: price,
				Available: true, ObservedAt: contractNow, ExpiresAt: contractNow.Add(time.Hour),
				TermsURL: "https://partner.invalid/book/" + id,
			},
			Stops: stops, DurationMinutes: minutes, Carrier: "FX" + id,
			DepartAt: contractNow.Add(24 * time.Hour), ArriveAt: contractNow.Add(30 * time.Hour),
		}
	}
	return []domain.FlightOffer{
		build("cheap-slow", 18_000, 2, 1080),
		build("fast", 22_000, 0, 600),
	}, []domain.Commercials{{
		OfferID: "cheap-slow", Partner: "fixture", CommissionMinor: 90_000, CommissionCurrency: "EUR",
	}}, nil
}

type stayPartnerFixture struct{}

func (stayPartnerFixture) Partner() domain.Partner { return contractPartner() }

func (stayPartnerFixture) Stays(context.Context, engine.StayQuery) ([]domain.StayOffer, []domain.Commercials, error) {
	build := func(id string, minor int64, location geo.Point) domain.StayOffer {
		price, _ := money.New(minor, money.Currency("EUR"))
		return domain.StayOffer{
			Offer: domain.Offer{
				ID: id, Partner: "fixture", Product: domain.Stay, Title: id, Price: price,
				Available: true, ObservedAt: contractNow, ExpiresAt: contractNow.Add(time.Hour),
				TermsURL: "https://partner.invalid/book/" + id,
			},
			PropertyID: id, Location: location, Rating: 4.5, ReviewCount: 400, Nights: 4,
		}
	}
	return []domain.StayOffer{
		build("far", 60_000, geo.Point{Latitude: 33.0, Longitude: -17.3}),
		build("near", 74_000, geo.Point{Latitude: 32.6478, Longitude: -16.9048}),
	}, nil, nil
}

type contractEntitlements struct{}

func (contractEntitlements) AllowsPriceMonitoring(context.Context, user.Principal) (bool, error) {
	return true, nil
}

func commerceService(t *testing.T, repository *store, timer clock.Clock) *commerceusecase.Service {
	t.Helper()
	offers := &contractOffers{}
	commerceEngine := engine.NewEngine(contractRegistry{}, offers, nil, id.NewUUID(), timer)
	return commerceusecase.NewService(commerceusecase.Dependencies{
		Trips: tripAccess{store: repository}, Catalog: repository, Itineraries: repository,
		Handoffs: repository, Conversions: repository, Watches: repository,
		Entitlements: contractEntitlements{}, Engine: commerceEngine,
		Providers: commerceusecase.Providers{
			Flights: []engine.FlightOfferProvider{flightPartnerFixture{}},
			Stays:   []engine.StayOfferProvider{stayPartnerFixture{}},
		},
		Identifiers: id.NewUUID(),
	}, timer)
}

func (s *store) ClaimWebhookEvent(_ context.Context, _, partner, providerEventID, _, _ string) (bool, error) {
	if s.claimed == nil {
		s.claimed = map[string]bool{}
	}
	key := partner + "|" + providerEventID
	if s.claimed[key] {
		return false, nil
	}
	s.claimed[key] = true
	return true, nil
}

func (s *store) MarkWebhookProcessed(context.Context, string, string, time.Time) error { return nil }

func (s *store) SaveConversion(_ context.Context, conversion domain.Conversion) (domain.Conversion, error) {
	if s.conversions == nil {
		s.conversions = map[string]domain.Conversion{}
	}
	s.conversions[conversion.Partner+"|"+conversion.ProviderConversionID] = conversion
	return conversion, nil
}

func (s *store) Conversion(_ context.Context, partner, providerConversionID string) (domain.Conversion, error) {
	conversion, found := s.conversions[partner+"|"+providerConversionID]
	if !found {
		return domain.Conversion{}, apperror.New(apperror.CodeNotFound, "Conversion not found.")
	}
	return conversion, nil
}
