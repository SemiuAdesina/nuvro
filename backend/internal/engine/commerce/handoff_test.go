package commerce_test

import (
	"context"
	"errors"
	"testing"
	"time"

	domain "github.com/SemiuAdesina/nuvro/backend/internal/domain/commerce"
	engine "github.com/SemiuAdesina/nuvro/backend/internal/engine/commerce"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/id"
)

type registry struct{ partners map[string]domain.Partner }

func (r registry) Partner(code string) (domain.Partner, bool) {
	partner, found := r.partners[code]
	return partner, found
}

type offerStore struct{ offers map[string]engine.StoredOffer }

func (o *offerStore) SaveOffers(_ context.Context, offers []engine.StoredOffer, _ time.Duration) error {
	if o.offers == nil {
		o.offers = map[string]engine.StoredOffer{}
	}
	for _, offer := range offers {
		o.offers[offer.Offer.ID] = offer
	}
	return nil
}

func (o *offerStore) Offer(_ context.Context, offerID string) (engine.StoredOffer, bool) {
	offer, found := o.offers[offerID]
	return offer, found
}

type publisher struct {
	seen []domain.Commercials
	err  error
}

func (p *publisher) Publish(_ context.Context, _ domain.Handoff, commercials domain.Commercials) error {
	p.seen = append(p.seen, commercials)
	return p.err
}

func approvedPartner() domain.Partner {
	return domain.Partner{
		Code: "fixture", DisplayName: "Fixture Travel", Status: domain.Enabled,
		Products:       []domain.ProductType{domain.Flight, domain.Stay},
		AllowedDomains: []string{"partner.invalid"},
	}
}

func storedOffer(t *testing.T, destination string) engine.StoredOffer {
	t.Helper()
	offer := base(t, "offer-1", domain.Flight, 20_000)
	return engine.StoredOffer{
		Offer: offer, Destination: destination,
		Commercials: domain.Commercials{OfferID: offer.ID, Partner: "fixture", CommissionMinor: 900, CommissionCurrency: "EUR"},
	}
}

func newHandoffEngine(t *testing.T, store *offerStore, attribution engine.AttributionPublisher) *engine.Engine {
	t.Helper()
	return engine.NewEngine(registry{partners: map[string]domain.Partner{"fixture": approvedPartner()}},
		store, attribution, id.NewUUID(), clock.NewFixed(now))
}

func TestHandoffCarriesDisclosureAndAnOpaqueToken(t *testing.T) {
	store := &offerStore{}
	if err := store.SaveOffers(t.Context(), []engine.StoredOffer{
		storedOffer(t, "https://partner.invalid/book/offer-1"),
	}, time.Hour); err != nil {
		t.Fatal(err)
	}
	attribution := &publisher{}
	handoff, err := newHandoffEngine(t, store, attribution).CreateHandoff(t.Context(), engine.HandoffRequest{
		OfferID: "offer-1", UserID: "3f1c9c8e-0e3a-4f0a-9b1f-1a2b3c4d5e6f",
	})
	if err != nil {
		t.Fatal(err)
	}
	if handoff.Disclosure.Code != domain.EligibleBookingDisclosure || handoff.Disclosure.Partner != "Fixture Travel" {
		t.Fatalf("disclosure = %+v", handoff.Disclosure)
	}
	if handoff.AttributionToken == "" || handoff.AttributionToken == handoff.UserID {
		t.Fatalf("attribution token is not opaque: %q", handoff.AttributionToken)
	}
	if handoff.DestinationURL != "https://partner.invalid/book/offer-1" {
		t.Fatalf("destination = %q", handoff.DestinationURL)
	}
	if !handoff.ExpiresAt.After(handoff.CreatedAt) {
		t.Fatal("the handoff has no expiry")
	}
	if len(attribution.seen) != 1 || attribution.seen[0].CommissionMinor != 900 {
		t.Fatalf("commission was not published for accounting: %+v", attribution.seen)
	}
}

func TestHandoffRefusesDestinationsOutsideTheAllowlist(t *testing.T) {
	for name, destination := range map[string]string{
		"other host":     "https://evil.invalid/book",
		"lookalike host": "https://partner.invalid.evil.invalid/book",
		"plain http":     "http://partner.invalid/book",
		"credentials":    "https://user:pass@partner.invalid/book",
		"not a url":      "javascript:alert(1)",
		"empty":          "",
	} {
		store := &offerStore{}
		if err := store.SaveOffers(t.Context(), []engine.StoredOffer{storedOffer(t, destination)}, time.Hour); err != nil {
			t.Fatal(err)
		}
		_, err := newHandoffEngine(t, store, nil).CreateHandoff(t.Context(), engine.HandoffRequest{
			OfferID: "offer-1", UserID: "3f1c9c8e-0e3a-4f0a-9b1f-1a2b3c4d5e6f",
		})
		if !errors.Is(err, domain.ErrUntrustedDestination) {
			t.Fatalf("%s produced %v", name, err)
		}
	}
}

func TestSubdomainsOfAnApprovedPartnerAreAccepted(t *testing.T) {
	store := &offerStore{}
	if err := store.SaveOffers(t.Context(), []engine.StoredOffer{
		storedOffer(t, "https://book.partner.invalid/offer-1"),
	}, time.Hour); err != nil {
		t.Fatal(err)
	}
	if _, err := newHandoffEngine(t, store, nil).CreateHandoff(t.Context(), engine.HandoffRequest{
		OfferID: "offer-1", UserID: "3f1c9c8e-0e3a-4f0a-9b1f-1a2b3c4d5e6f",
	}); err != nil {
		t.Fatal(err)
	}
}

func TestExpiredUnavailableAndUnknownOffersAreRefused(t *testing.T) {
	base := storedOffer(t, "https://partner.invalid/book")
	expired := base
	expired.Offer.ExpiresAt = now.Add(-time.Minute)
	unavailable := base
	unavailable.Offer.Available = false
	for name, test := range map[string]struct {
		stored engine.StoredOffer
		want   error
	}{
		"expired":     {expired, engine.ErrOfferExpired},
		"unavailable": {unavailable, engine.ErrOfferUnavailable},
	} {
		store := &offerStore{}
		if err := store.SaveOffers(t.Context(), []engine.StoredOffer{test.stored}, time.Hour); err != nil {
			t.Fatal(err)
		}
		_, err := newHandoffEngine(t, store, nil).CreateHandoff(t.Context(), engine.HandoffRequest{
			OfferID: "offer-1", UserID: "3f1c9c8e-0e3a-4f0a-9b1f-1a2b3c4d5e6f",
		})
		if !errors.Is(err, test.want) {
			t.Fatalf("%s produced %v, want %v", name, err, test.want)
		}
	}
	_, err := newHandoffEngine(t, &offerStore{}, nil).CreateHandoff(t.Context(), engine.HandoffRequest{
		OfferID: "never-seen", UserID: "3f1c9c8e-0e3a-4f0a-9b1f-1a2b3c4d5e6f",
	})
	if !errors.Is(err, engine.ErrOfferExpired) {
		t.Fatalf("an unknown offer produced %v", err)
	}
}

func TestDisabledPartnersCannotReceiveTraffic(t *testing.T) {
	disabled := approvedPartner()
	disabled.Status = domain.Disabled
	store := &offerStore{}
	if err := store.SaveOffers(t.Context(), []engine.StoredOffer{
		storedOffer(t, "https://partner.invalid/book"),
	}, time.Hour); err != nil {
		t.Fatal(err)
	}
	subject := engine.NewEngine(registry{partners: map[string]domain.Partner{"fixture": disabled}},
		store, nil, id.NewUUID(), clock.NewFixed(now))
	if _, err := subject.CreateHandoff(t.Context(), engine.HandoffRequest{
		OfferID: "offer-1", UserID: "3f1c9c8e-0e3a-4f0a-9b1f-1a2b3c4d5e6f",
	}); !errors.Is(err, engine.ErrPartnerUnavailable) {
		t.Fatalf("error = %v", err)
	}
}

func TestFollowingAnExpiredHandoffIsRefused(t *testing.T) {
	subject := newHandoffEngine(t, &offerStore{}, nil)
	handoff := domain.Handoff{
		ID: "handoff", Partner: "fixture", DestinationURL: "https://partner.invalid/book",
		CreatedAt: now.Add(-2 * time.Hour), ExpiresAt: now.Add(-time.Hour),
	}
	if _, err := subject.Destination(handoff); !errors.Is(err, engine.ErrOfferExpired) {
		t.Fatalf("error = %v", err)
	}
	fresh := handoff
	fresh.ExpiresAt = now.Add(time.Hour)
	destination, err := subject.Destination(fresh)
	if err != nil || destination != "https://partner.invalid/book" {
		t.Fatalf("%q %v", destination, err)
	}
	tampered := fresh
	tampered.DestinationURL = "https://evil.invalid/book"
	if _, err := subject.Destination(tampered); !errors.Is(err, domain.ErrUntrustedDestination) {
		t.Fatal("a tampered destination was followed")
	}
}

func TestAttributionFailureBlocksTheHandoff(t *testing.T) {
	store := &offerStore{}
	if err := store.SaveOffers(t.Context(), []engine.StoredOffer{
		storedOffer(t, "https://partner.invalid/book"),
	}, time.Hour); err != nil {
		t.Fatal(err)
	}
	failing := &publisher{err: errors.New("attribution unavailable")}
	if _, err := newHandoffEngine(t, store, failing).CreateHandoff(t.Context(), engine.HandoffRequest{
		OfferID: "offer-1", UserID: "3f1c9c8e-0e3a-4f0a-9b1f-1a2b3c4d5e6f",
	}); err == nil {
		t.Fatal("a handoff was issued without recording attribution")
	}
}

func TestConversionStatusNormalisation(t *testing.T) {
	for raw, want := range map[string]domain.ConversionStatus{
		"booking.created":     domain.BookingCreated,
		"BOOKING_COMPLETED":   domain.BookingCompleted,
		" confirmed ":         domain.BookingCompleted,
		"cancelled":           domain.BookingCancelled,
		"refund":              domain.Refunded,
		"commission_approved": domain.CommissionApproved,
		"rejected":            domain.CommissionRejected,
	} {
		status, err := engine.NormalizeStatus(raw)
		if err != nil || status != want {
			t.Fatalf("%q became %q %v, want %q", raw, status, err, want)
		}
	}
	if _, err := engine.NormalizeStatus("something_new"); !errors.Is(err, engine.ErrUnknownEvent) {
		t.Fatal("an unknown provider status was accepted")
	}
}

func TestLaterStatusesSupersedeEarlierOnes(t *testing.T) {
	if !engine.Supersedes(domain.BookingCreated, domain.BookingCompleted) {
		t.Fatal("completion did not supersede creation")
	}
	if engine.Supersedes(domain.Refunded, domain.BookingCreated) {
		t.Fatal("a late duplicate reopened a refunded booking")
	}
	if !engine.Supersedes(domain.CommissionPending, domain.CommissionApproved) {
		t.Fatal("approval did not supersede pending")
	}
}
