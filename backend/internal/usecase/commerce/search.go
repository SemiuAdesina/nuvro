package commerce

import (
	"context"
	"time"

	domain "github.com/SemiuAdesina/nuvro/backend/internal/domain/commerce"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/geo"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/trip"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	engine "github.com/SemiuAdesina/nuvro/backend/internal/engine/commerce"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
)

const OfferCacheTTL = 20 * time.Minute

type Service struct {
	trips        Trips
	catalog      Catalog
	itineraries  Itineraries
	handoffs     HandoffStore
	conversions  ConversionStore
	watches      WatchStore
	entitlements Entitlements
	alerts       Alerts
	engine       *engine.Engine
	providers    Providers
	identifiers  Identifiers
	timer        clock.Clock
}

type Dependencies struct {
	Trips        Trips
	Catalog      Catalog
	Itineraries  Itineraries
	Handoffs     HandoffStore
	Conversions  ConversionStore
	Watches      WatchStore
	Entitlements Entitlements
	Alerts       Alerts
	Engine       *engine.Engine
	Providers    Providers
	Identifiers  Identifiers
}

func NewService(dependencies Dependencies, timer clock.Clock) *Service {
	return &Service{
		trips: dependencies.Trips, catalog: dependencies.Catalog, itineraries: dependencies.Itineraries,
		handoffs: dependencies.Handoffs, conversions: dependencies.Conversions, watches: dependencies.Watches,
		entitlements: dependencies.Entitlements, alerts: dependencies.Alerts, engine: dependencies.Engine,
		providers: dependencies.Providers, identifiers: dependencies.Identifiers, timer: timer,
	}
}

type SearchStatus struct {
	Partners   []PartnerStatus
	Ranking    string
	SearchedAt time.Time
}

type PartnerStatus struct {
	Partner   string
	Available bool
	Reason    string
}

func (s *Service) context(ctx context.Context, principal user.Principal, tripID string) (trip.Trip, geo.Point, error) {
	record, err := s.trips.Read(ctx, principal, tripID)
	if err != nil {
		return trip.Trip{}, geo.Point{}, err
	}
	centre := geo.Point{}
	if record.DestinationID != nil && s.catalog != nil {
		if place, lookupErr := s.catalog.Destination(ctx, *record.DestinationID); lookupErr == nil {
			centre = geo.Point{Latitude: place.Latitude, Longitude: place.Longitude}
		}
	}
	return record, centre, nil
}

func (s *Service) itineraryCentre(ctx context.Context, tripID string, fallback geo.Point) geo.Point {
	if s.itineraries == nil {
		return fallback
	}
	plan, err := s.itineraries.Itinerary(ctx, tripID)
	if err != nil || len(plan.Items) == 0 {
		return fallback
	}
	latitude, longitude, count := 0.0, 0.0, 0
	for _, item := range plan.Items {
		if item.Location.IsZero() {
			continue
		}
		latitude += item.Location.Latitude
		longitude += item.Location.Longitude
		count++
	}
	if count == 0 {
		return fallback
	}
	return geo.Point{Latitude: latitude / float64(count), Longitude: longitude / float64(count)}
}

func (s *Service) store(ctx context.Context, offers []domain.Offer, commercials []domain.Commercials, destinations map[string]string) {
	if s.engine == nil {
		return
	}
	byOffer := make(map[string]domain.Commercials, len(commercials))
	for _, entry := range commercials {
		byOffer[entry.OfferID] = entry
	}
	stored := make([]engine.StoredOffer, 0, len(offers))
	for _, offer := range offers {
		stored = append(stored, engine.StoredOffer{
			Offer: offer, Commercials: byOffer[offer.ID], Destination: destinations[offer.ID],
		})
	}
	_ = s.engine.Store(ctx, stored, OfferCacheTTL)
}

func partnerUnavailable(partner domain.Partner, reason string) PartnerStatus {
	return PartnerStatus{Partner: partner.Code, Available: false, Reason: reason}
}
