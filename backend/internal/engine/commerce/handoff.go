package commerce

import (
	"context"
	"errors"
	"time"

	domain "github.com/SemiuAdesina/nuvro/backend/internal/domain/commerce"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
)

var ErrOfferExpired = errors.New("commerce: offer expired")
var ErrOfferUnavailable = errors.New("commerce: offer unavailable")
var ErrPartnerUnavailable = errors.New("commerce: partner unavailable")
var ErrPriceChanged = errors.New("commerce: price changed")

type Registry interface {
	Partner(code string) (domain.Partner, bool)
}

type Identifiers interface{ Generate() (string, error) }

type Engine struct {
	registry    Registry
	offers      OfferStore
	attribution AttributionPublisher
	identifiers Identifiers
	timer       clock.Clock
}

func NewEngine(registry Registry, offers OfferStore, attribution AttributionPublisher,
	identifiers Identifiers, timer clock.Clock) *Engine {
	return &Engine{
		registry: registry, offers: offers, attribution: attribution,
		identifiers: identifiers, timer: timer,
	}
}

type HandoffRequest struct {
	OfferID string
	TripID  string
	UserID  string
}

func (e *Engine) CreateHandoff(ctx context.Context, request HandoffRequest) (domain.Handoff, error) {
	stored, partner, err := e.bookable(ctx, request.OfferID)
	if err != nil {
		return domain.Handoff{}, err
	}
	now := e.timer.Now().UTC()
	handoffID, err := e.identifiers.Generate()
	if err != nil {
		return domain.Handoff{}, err
	}
	token, err := e.identifiers.Generate()
	if err != nil {
		return domain.Handoff{}, err
	}
	handoff := domain.Handoff{
		ID: handoffID, OfferID: stored.Offer.ID, TripID: request.TripID, UserID: request.UserID,
		Partner: partner.Code, Product: stored.Offer.Product, DestinationURL: stored.Destination,
		Price: stored.Offer.Price, Title: stored.Offer.Title,
		Disclosure:       domain.Disclosure{Code: domain.EligibleBookingDisclosure, Partner: partner.DisplayName},
		AttributionToken: token, CreatedAt: now, ExpiresAt: now.Add(domain.HandoffLifetime),
	}
	if err := handoff.Validate(); err != nil {
		return domain.Handoff{}, err
	}
	if e.attribution != nil {
		if err := e.attribution.Publish(ctx, handoff, stored.Commercials); err != nil {
			return domain.Handoff{}, err
		}
	}
	return handoff, nil
}

func (e *Engine) bookable(ctx context.Context, offerID string) (StoredOffer, domain.Partner, error) {
	stored, found := e.offers.Offer(ctx, offerID)
	if !found {
		return StoredOffer{}, domain.Partner{}, ErrOfferExpired
	}
	if !stored.Offer.Available {
		return StoredOffer{}, domain.Partner{}, ErrOfferUnavailable
	}
	if !stored.Offer.FreshnessAt(e.timer.Now().UTC()).Bookable() {
		return StoredOffer{}, domain.Partner{}, ErrOfferExpired
	}
	partner, known := e.registry.Partner(stored.Offer.Partner)
	if !known || !partner.Status.Usable() || !partner.Supports(stored.Offer.Product) {
		return StoredOffer{}, domain.Partner{}, ErrPartnerUnavailable
	}
	if err := partner.AllowsDestination(stored.Destination); err != nil {
		return StoredOffer{}, domain.Partner{}, err
	}
	return stored, partner, nil
}

func (e *Engine) Destination(handoff domain.Handoff) (string, error) {
	partner, known := e.registry.Partner(handoff.Partner)
	if !known || !partner.Status.Usable() {
		return "", ErrPartnerUnavailable
	}
	if handoff.Expired(e.timer.Now()) {
		return "", ErrOfferExpired
	}
	if err := partner.AllowsDestination(handoff.DestinationURL); err != nil {
		return "", err
	}
	return handoff.DestinationURL, nil
}

func (e *Engine) Store(ctx context.Context, offers []StoredOffer, ttl time.Duration) error {
	if e.offers == nil || len(offers) == 0 {
		return nil
	}
	return e.offers.SaveOffers(ctx, offers, ttl)
}
