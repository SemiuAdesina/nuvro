package commerce_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/apperror"
	domain "github.com/SemiuAdesina/nuvro/backend/internal/domain/commerce"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/geo"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/itinerary"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	engine "github.com/SemiuAdesina/nuvro/backend/internal/engine/commerce"
	usecase "github.com/SemiuAdesina/nuvro/backend/internal/usecase/commerce"
)

type flightPartner struct {
	partner domain.Partner
	offers  []domain.FlightOffer
	priced  []domain.Commercials
	err     error
}

func (f flightPartner) Partner() domain.Partner { return f.partner }

func (f flightPartner) Flights(context.Context, engine.FlightQuery) ([]domain.FlightOffer, []domain.Commercials, error) {
	return f.offers, f.priced, f.err
}

type stayPartner struct {
	partner domain.Partner
	offers  []domain.StayOffer
	err     error
}

func (s stayPartner) Partner() domain.Partner { return s.partner }

func (s stayPartner) Stays(context.Context, engine.StayQuery) ([]domain.StayOffer, []domain.Commercials, error) {
	return s.offers, nil, s.err
}

func flightOffer(t *testing.T, id string, minor int64, stops, minutes int) domain.FlightOffer {
	t.Helper()
	return domain.FlightOffer{
		Offer: domain.Offer{
			ID: id, Partner: "fixture", Product: domain.Flight, Title: id, Price: amount(t, minor),
			Available: true, ObservedAt: now, ExpiresAt: now.Add(time.Hour),
			TermsURL: "https://partner.invalid/book/" + id,
		},
		Stops: stops, DurationMinutes: minutes, Carrier: "FX" + id,
		DepartAt: now.Add(24 * time.Hour), ArriveAt: now.Add(30 * time.Hour),
	}
}

func stayOffer(t *testing.T, id string, minor int64, location geo.Point) domain.StayOffer {
	t.Helper()
	return domain.StayOffer{
		Offer: domain.Offer{
			ID: id, Partner: "fixture", Product: domain.Stay, Title: id, Price: amount(t, minor),
			Available: true, ObservedAt: now, ExpiresAt: now.Add(time.Hour),
			TermsURL: "https://partner.invalid/book/" + id,
		},
		PropertyID: id, Location: location, Rating: 4.4, ReviewCount: 300, Nights: 4,
	}
}

func TestFlightSearchRanksAndCachesOffersForHandoff(t *testing.T) {
	offers := newOfferStore()
	provider := flightPartner{
		partner: approvedPartner(domain.Flight),
		offers: []domain.FlightOffer{
			flightOffer(t, "cheap-slow", 18_000, 2, 1080),
			flightOffer(t, "fast", 22_000, 0, 600),
		},
		priced: []domain.Commercials{{
			OfferID: "cheap-slow", Partner: "fixture", CommissionMinor: 90_000, CommissionCurrency: "EUR",
		}},
	}
	service := newService(t, usecase.Dependencies{
		Engine:    commerceEngine(t, offers, approvedPartner(domain.Flight)),
		Providers: usecase.Providers{Flights: []engine.FlightOfferProvider{provider}},
	})
	result, err := service.SearchFlights(t.Context(), principal(), usecase.FlightRequest{
		TripID: tripID, Origin: "LIS", Destination: "FNC", DepartDate: now.AddDate(0, 1, 0),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Offers) != 2 || result.Offers[0].ID != "fast" {
		t.Fatalf("ranking = %+v", result.Offers)
	}
	if result.Status.Ranking != engine.FlightRankingVersion {
		t.Fatalf("ranking version = %q", result.Status.Ranking)
	}
	if len(result.Status.Partners) != 1 || !result.Status.Partners[0].Available {
		t.Fatalf("partner status = %+v", result.Status.Partners)
	}
	stored, found := offers.Offer(t.Context(), "fast")
	if !found || stored.Destination != "https://partner.invalid/book/fast" {
		t.Fatalf("offer was not cached for handoff: %+v", stored)
	}
	commissioned, found := offers.Offer(t.Context(), "cheap-slow")
	if !found || commissioned.Commercials.CommissionMinor != 90_000 {
		t.Fatal("commission was not retained for accounting")
	}
}

func TestOnePartnerFailingDoesNotFailTheSearch(t *testing.T) {
	working := flightPartner{
		partner: approvedPartner(domain.Flight),
		offers:  []domain.FlightOffer{flightOffer(t, "available", 20_000, 0, 500)},
	}
	broken := flightPartner{
		partner: domain.Partner{
			Code: "broken", DisplayName: "Broken", Status: domain.Enabled,
			Products: []domain.ProductType{domain.Flight}, AllowedDomains: []string{"broken.invalid"},
		},
		err: errors.New("partner unavailable"),
	}
	disabled := flightPartner{
		partner: domain.Partner{
			Code: "off", DisplayName: "Off", Status: domain.Disabled,
			Products: []domain.ProductType{domain.Flight}, AllowedDomains: []string{"off.invalid"},
		},
	}
	service := newService(t, usecase.Dependencies{
		Engine:    commerceEngine(t, newOfferStore(), approvedPartner(domain.Flight)),
		Providers: usecase.Providers{Flights: []engine.FlightOfferProvider{working, broken, disabled}},
	})
	result, err := service.SearchFlights(t.Context(), principal(), usecase.FlightRequest{
		TripID: tripID, Origin: "LIS", Destination: "FNC", DepartDate: now.AddDate(0, 1, 0),
	})
	if err != nil {
		t.Fatalf("one partner outage failed the whole search: %v", err)
	}
	if len(result.Offers) != 1 || result.Offers[0].ID != "available" {
		t.Fatalf("offers = %+v", result.Offers)
	}
	statuses := map[string]bool{}
	for _, status := range result.Status.Partners {
		statuses[status.Partner] = status.Available
	}
	if statuses["broken"] || statuses["off"] || !statuses["fixture"] {
		t.Fatalf("partner statuses = %+v", result.Status.Partners)
	}
}

func TestNoResultsIsNotAnError(t *testing.T) {
	service := newService(t, usecase.Dependencies{
		Engine:    commerceEngine(t, newOfferStore(), approvedPartner(domain.Flight)),
		Providers: usecase.Providers{},
	})
	result, err := service.SearchFlights(t.Context(), principal(), usecase.FlightRequest{
		TripID: tripID, Origin: "LIS", Destination: "FNC", DepartDate: now.AddDate(0, 1, 0),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Offers) != 0 || result.Status.Ranking == "" {
		t.Fatalf("result = %+v", result)
	}
}

func TestStaySearchUsesTheItineraryCentre(t *testing.T) {
	near := stayOffer(t, "near", 74_000, funchal)
	far := stayOffer(t, "far", 60_000, geo.Point{Latitude: 33.0, Longitude: -17.3})
	plan := itineraries{items: []itinerary.Item{
		{ID: "a", TripID: tripID, Day: 1, StartAt: now, EndAt: now.Add(time.Hour),
			Kind: itinerary.Activity, Title: "Visit", Status: itinerary.Planned,
			Source: itinerary.Generated, Location: funchal},
	}}
	service := newService(t, usecase.Dependencies{
		Itineraries: plan,
		Engine:      commerceEngine(t, newOfferStore(), approvedPartner(domain.Stay)),
		Providers: usecase.Providers{Stays: []engine.StayOfferProvider{
			stayPartner{partner: approvedPartner(domain.Stay), offers: []domain.StayOffer{far, near}},
		}},
	})
	result, err := service.SearchStays(t.Context(), principal(), usecase.StayRequest{
		TripID: tripID, CheckIn: now.AddDate(0, 1, 0), CheckOut: now.AddDate(0, 1, 4), BudgetMinor: 100_000,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Offers[0].ID != "near" {
		t.Fatalf("the cheaper distant stay won: %+v", result.Offers)
	}
}

func TestSearchesRefuseAnonymousCallersAndBadInput(t *testing.T) {
	service := newService(t, usecase.Dependencies{
		Engine: commerceEngine(t, newOfferStore(), approvedPartner(domain.Flight)),
	})
	if _, err := service.SearchFlights(t.Context(), user.Principal{}, usecase.FlightRequest{
		TripID: tripID, Origin: "LIS", Destination: "FNC",
	}); err == nil {
		t.Fatal("an anonymous caller searched flights")
	}
	if _, err := service.SearchFlights(t.Context(), principal(), usecase.FlightRequest{
		TripID: tripID,
	}); apperror.CodeOf(err) != apperror.CodeInvalidInput {
		t.Fatal("a search without a route was accepted")
	}
	if _, err := service.SearchStays(t.Context(), principal(), usecase.StayRequest{
		TripID: tripID, CheckIn: now.AddDate(0, 1, 4), CheckOut: now.AddDate(0, 1, 0),
	}); apperror.CodeOf(err) != apperror.CodeInvalidInput {
		t.Fatal("a reversed stay range was accepted")
	}
}

func TestHandoffRecordsClickOnceAndRefusesStrangers(t *testing.T) {
	offers := newOfferStore()
	store := newHandoffStore()
	partner := approvedPartner(domain.Flight)
	service := newService(t, usecase.Dependencies{
		Handoffs: store, Engine: commerceEngine(t, offers, partner),
		Providers: usecase.Providers{Flights: []engine.FlightOfferProvider{
			flightPartner{partner: partner, offers: []domain.FlightOffer{flightOffer(t, "fast", 20_000, 0, 500)}},
		}},
	})
	if _, err := service.SearchFlights(t.Context(), principal(), usecase.FlightRequest{
		TripID: tripID, Origin: "LIS", Destination: "FNC", DepartDate: now.AddDate(0, 1, 0),
	}); err != nil {
		t.Fatal(err)
	}
	handoff, err := service.CreateHandoff(t.Context(), principal(), usecase.HandoffRequest{
		OfferID: "fast", TripID: tripID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if handoff.UserID != owner || handoff.Disclosure.Partner == "" {
		t.Fatalf("handoff = %+v", handoff)
	}
	destination, err := service.FollowHandoff(t.Context(), principal(), handoff.ID)
	if err != nil || destination != "https://partner.invalid/book/fast" {
		t.Fatalf("%q %v", destination, err)
	}
	if _, err := service.FollowHandoff(t.Context(), principal(), handoff.ID); err != nil {
		t.Fatal(err)
	}
	if store.clicks[handoff.ID] != 1 {
		t.Fatalf("duplicate clicks recorded: %d", store.clicks[handoff.ID])
	}
	stranger := user.Principal{UserID: "0c2f21b0-4a55-4b39-9f0b-2c7f9c1a1a99"}
	if _, err := service.FollowHandoff(t.Context(), stranger, handoff.ID); apperror.CodeOf(err) != apperror.CodeNotFound {
		t.Fatal("another traveller followed the handoff")
	}
}

func TestHandoffForAnExpiredOfferIsRefused(t *testing.T) {
	offers := newOfferStore()
	service := newService(t, usecase.Dependencies{
		Handoffs: newHandoffStore(), Engine: commerceEngine(t, offers, approvedPartner(domain.Flight)),
	})
	if _, err := service.CreateHandoff(t.Context(), principal(), usecase.HandoffRequest{
		OfferID: "never-searched",
	}); apperror.CodeOf(err) != apperror.CodeInvalidState {
		t.Fatalf("error = %v", err)
	}
}
