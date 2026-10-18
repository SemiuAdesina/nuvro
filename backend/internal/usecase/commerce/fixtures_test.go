package commerce_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/apperror"
	domain "github.com/SemiuAdesina/nuvro/backend/internal/domain/commerce"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/destination"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/geo"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/itinerary"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/money"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/pricewatch"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/trip"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	engine "github.com/SemiuAdesina/nuvro/backend/internal/engine/commerce"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/id"
	usecase "github.com/SemiuAdesina/nuvro/backend/internal/usecase/commerce"
)

const owner = "3f1c9c8e-0e3a-4f0a-9b1f-1a2b3c4d5e6f"
const tripID = "8c2f21b0-4a55-4b39-9f0b-2c7f9c1a1a11"

var now = time.Date(2026, 10, 16, 9, 0, 0, 0, time.UTC)
var funchal = geo.Point{Latitude: 32.6478, Longitude: -16.9048}

type trips struct{ err error }

func (t trips) Read(_ context.Context, principal user.Principal, _ string) (trip.Trip, error) {
	if t.err != nil {
		return trip.Trip{}, t.err
	}
	if principal.UserID == "" {
		return trip.Trip{}, errors.New("unauthenticated")
	}
	destinationID := "madeira-pt"
	start := now.AddDate(0, 1, 0)
	end := start.AddDate(0, 0, 6)
	return trip.Trip{
		ID: tripID, OwnerID: owner, DestinationID: &destinationID,
		StartDate: &start, EndDate: &end, Adults: 2, Currency: "EUR", Status: trip.Planning,
	}, nil
}

type catalog struct{}

func (catalog) Destination(context.Context, string) (destination.Destination, error) {
	return destination.Destination{
		ID: "madeira-pt", Name: "Madeira", CountryCode: "PT", Timezone: "Atlantic/Madeira",
		Latitude: 32.7607, Longitude: -16.9595, PrimaryAirport: "FNC",
	}, nil
}

type itineraries struct{ items []itinerary.Item }

func (i itineraries) Itinerary(context.Context, string) (itinerary.Itinerary, error) {
	if len(i.items) == 0 {
		return itinerary.Itinerary{}, errors.New("no itinerary")
	}
	return itinerary.Itinerary{TripID: tripID, Items: i.items}, nil
}

type handoffStore struct {
	saved  map[string]domain.Handoff
	clicks map[string]int
	err    error
}

func newHandoffStore() *handoffStore {
	return &handoffStore{saved: map[string]domain.Handoff{}, clicks: map[string]int{}}
}

func (h *handoffStore) SaveHandoff(_ context.Context, handoff domain.Handoff) error {
	if h.err != nil {
		return h.err
	}
	h.saved[handoff.ID] = handoff
	return nil
}

func (h *handoffStore) Handoff(_ context.Context, handoffID string) (domain.Handoff, error) {
	handoff, found := h.saved[handoffID]
	if !found {
		return domain.Handoff{}, errors.New("not found")
	}
	return handoff, nil
}

func (h *handoffStore) RecordClick(_ context.Context, click domain.Click) error {
	if _, seen := h.clicks[click.HandoffID]; seen {
		return nil
	}
	h.clicks[click.HandoffID] = 1
	return nil
}

type conversionStore struct {
	claimed     map[string]bool
	conversions map[string]domain.Conversion
	processed   int
}

func newConversionStore() *conversionStore {
	return &conversionStore{claimed: map[string]bool{}, conversions: map[string]domain.Conversion{}}
}

func (c *conversionStore) ClaimWebhookEvent(_ context.Context, _, partner, providerEventID, _, _ string) (bool, error) {
	key := partner + "|" + providerEventID
	if c.claimed[key] {
		return false, nil
	}
	c.claimed[key] = true
	return true, nil
}

func (c *conversionStore) MarkWebhookProcessed(context.Context, string, string, time.Time) error {
	c.processed++
	return nil
}

func (c *conversionStore) SaveConversion(_ context.Context, conversion domain.Conversion) (domain.Conversion, error) {
	c.conversions[conversion.Partner+"|"+conversion.ProviderConversionID] = conversion
	return conversion, nil
}

func (c *conversionStore) Conversion(_ context.Context, partner, providerConversionID string) (domain.Conversion, error) {
	conversion, found := c.conversions[partner+"|"+providerConversionID]
	if !found {
		return domain.Conversion{}, notFound()
	}
	return conversion, nil
}

type watchStore struct {
	saved map[string]pricewatch.Watch
	err   error
}

func newWatchStore() *watchStore { return &watchStore{saved: map[string]pricewatch.Watch{}} }

func (w *watchStore) SavePriceWatch(_ context.Context, watch pricewatch.Watch) error {
	if w.err != nil {
		return w.err
	}
	w.saved[watch.ID] = watch
	return nil
}

func (w *watchStore) PriceWatches(_ context.Context, userID string) ([]pricewatch.Watch, error) {
	values := make([]pricewatch.Watch, 0, len(w.saved))
	for _, watch := range w.saved {
		if watch.UserID == userID {
			values = append(values, watch)
		}
	}
	return values, nil
}

func (w *watchStore) PriceWatch(_ context.Context, watchID string) (pricewatch.Watch, error) {
	watch, found := w.saved[watchID]
	if !found {
		return pricewatch.Watch{}, errors.New("not found")
	}
	return watch, nil
}

func (w *watchStore) ActivePriceWatchCount(_ context.Context, userID string) (int64, error) {
	count := int64(0)
	for _, watch := range w.saved {
		if watch.UserID == userID && watch.Status == pricewatch.Active {
			count++
		}
	}
	return count, nil
}

func (w *watchStore) EndPriceWatch(_ context.Context, watchID string, status pricewatch.Status) error {
	watch, found := w.saved[watchID]
	if !found {
		return errors.New("not found")
	}
	watch.Status = status
	w.saved[watchID] = watch
	return nil
}

type entitlements struct {
	allowed bool
	err     error
}

func (e entitlements) AllowsPriceMonitoring(context.Context, user.Principal) (bool, error) {
	return e.allowed, e.err
}

type registry struct{ partners map[string]domain.Partner }

func (r registry) Partner(code string) (domain.Partner, bool) {
	partner, found := r.partners[code]
	return partner, found
}

type offerStore struct{ offers map[string]engine.StoredOffer }

func newOfferStore() *offerStore { return &offerStore{offers: map[string]engine.StoredOffer{}} }

func (o *offerStore) SaveOffers(_ context.Context, offers []engine.StoredOffer, _ time.Duration) error {
	for _, offer := range offers {
		o.offers[offer.Offer.ID] = offer
	}
	return nil
}

func (o *offerStore) Offer(_ context.Context, offerID string) (engine.StoredOffer, bool) {
	offer, found := o.offers[offerID]
	return offer, found
}

func approvedPartner(products ...domain.ProductType) domain.Partner {
	return domain.Partner{
		Code: "fixture", DisplayName: "Fixture Travel", Status: domain.Enabled,
		Products: products, AllowedDomains: []string{"partner.invalid"},
	}
}

func amount(t *testing.T, minor int64) money.Amount {
	t.Helper()
	value, err := money.New(minor, money.Currency("EUR"))
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func principal() user.Principal { return user.Principal{UserID: owner} }

func notFound() error {
	return apperror.New(apperror.CodeNotFound, "Conversion not found.")
}

func newService(t *testing.T, dependencies usecase.Dependencies) *usecase.Service {
	t.Helper()
	if dependencies.Trips == nil {
		dependencies.Trips = trips{}
	}
	if dependencies.Catalog == nil {
		dependencies.Catalog = catalog{}
	}
	if dependencies.Identifiers == nil {
		dependencies.Identifiers = id.NewUUID()
	}
	return usecase.NewService(dependencies, clock.NewFixed(now))
}

func commerceEngine(t *testing.T, offers *offerStore, partner domain.Partner) *engine.Engine {
	t.Helper()
	return engine.NewEngine(registry{partners: map[string]domain.Partner{partner.Code: partner}},
		offers, nil, id.NewUUID(), clock.NewFixed(now))
}
