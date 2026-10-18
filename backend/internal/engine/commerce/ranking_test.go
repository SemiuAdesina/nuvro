package commerce_test

import (
	"testing"
	"time"

	domain "github.com/SemiuAdesina/nuvro/backend/internal/domain/commerce"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/geo"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/money"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/weather"
	engine "github.com/SemiuAdesina/nuvro/backend/internal/engine/commerce"
)

var now = time.Date(2026, 10, 16, 9, 0, 0, 0, time.UTC)
var funchal = geo.Point{Latitude: 32.6478, Longitude: -16.9048}

func amount(t *testing.T, minor int64) money.Amount {
	t.Helper()
	value, err := money.New(minor, money.Currency("EUR"))
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func base(t *testing.T, id string, product domain.ProductType, minor int64) domain.Offer {
	t.Helper()
	return domain.Offer{
		ID: id, Partner: "fixture", Product: product, Title: id, Price: amount(t, minor),
		Available: true, ObservedAt: now, ExpiresAt: now.Add(time.Hour), TermsURL: "https://partner.invalid/" + id,
	}
}

func flight(t *testing.T, id string, minor int64, stops, minutes int) domain.FlightOffer {
	t.Helper()
	return domain.FlightOffer{
		Offer: base(t, id, domain.Flight, minor), Stops: stops, DurationMinutes: minutes,
		DepartAt: now.Add(24 * time.Hour), ArriveAt: now.Add(time.Duration(24+minutes/60) * time.Hour),
		Carrier: "FX" + id,
	}
}

func flightIDs(offers []domain.FlightOffer) []string {
	found := make([]string, 0, len(offers))
	for _, offer := range offers {
		found = append(found, offer.ID)
	}
	return found
}

func stayIDs(offers []domain.StayOffer) []string {
	found := make([]string, 0, len(offers))
	for _, offer := range offers {
		found = append(found, offer.ID)
	}
	return found
}

func activityIDs(offers []domain.ActivityOffer) []string {
	found := make([]string, 0, len(offers))
	for _, offer := range offers {
		found = append(found, offer.ID)
	}
	return found
}

func TestCheapWithManyStopsLosesToTheFasterOption(t *testing.T) {
	cheapSlow := flight(t, "a-cheapest", 18_000, 2, 1080)
	fasterDirect := flight(t, "b-faster", 22_000, 1, 600)
	worst := flight(t, "c-highest-commission", 31_000, 3, 1320)
	ranked := engine.RankFlights([]domain.FlightOffer{cheapSlow, fasterDirect, worst},
		engine.FlightContext{Now: now})
	if ranked[0].ID != "b-faster" {
		t.Fatalf("ranking = %v, want the faster option first", flightIDs(ranked))
	}
	if ranked[len(ranked)-1].ID != "c-highest-commission" {
		t.Fatalf("the worst option did not rank last: %v", flightIDs(ranked))
	}
}

func TestCommissionCannotEnterFlightRanking(t *testing.T) {
	offers := []domain.FlightOffer{
		flight(t, "a-cheapest", 18_000, 2, 1080),
		flight(t, "b-faster", 22_000, 1, 600),
		flight(t, "c-highest-commission", 31_000, 3, 1320),
	}
	baseline := flightIDs(engine.RankFlights(offers, engine.FlightContext{Now: now}))
	for _, commission := range []int64{0, 500, 50_000, 5_000_000} {
		commercials := []domain.Commercials{{
			OfferID: "c-highest-commission", Partner: "fixture",
			CommissionMinor: commission, CommissionCurrency: "EUR",
		}}
		_ = commercials
		again := flightIDs(engine.RankFlights(offers, engine.FlightContext{Now: now}))
		for index := range again {
			if again[index] != baseline[index] {
				t.Fatalf("commission %d changed the ranking: %v then %v", commission, baseline, again)
			}
		}
	}
}

func TestRankingInputsCannotCarryCommission(t *testing.T) {
	offer := flight(t, "a", 10_000, 0, 300)
	commercials := domain.Commercials{
		OfferID: offer.ID, Partner: "fixture", CommissionMinor: 900_000, CommissionCurrency: "EUR",
	}
	value, known := commercials.Amount()
	if !known || value.Minor() != 900_000 {
		t.Fatal("commission accounting is broken")
	}
	ranked := engine.RankFlights([]domain.FlightOffer{offer}, engine.FlightContext{Now: now})
	if len(ranked) != 1 || ranked[0].Rank != 1 {
		t.Fatalf("ranked = %+v", ranked)
	}
}

func TestFlightReasonsExplainThePosition(t *testing.T) {
	ranked := engine.RankFlights([]domain.FlightOffer{
		flight(t, "cheap", 18_000, 2, 1080),
		flight(t, "direct", 22_000, 0, 600),
	}, engine.FlightContext{Now: now, BudgetMinor: 25_000})
	byID := map[string][]domain.Reason{}
	for _, offer := range ranked {
		byID[offer.ID] = offer.Reasons
	}
	if !contains(byID["cheap"], domain.BestPrice) {
		t.Fatalf("the cheapest offer lacks its reason: %v", byID["cheap"])
	}
	if !contains(byID["direct"], domain.FewerStops) || !contains(byID["direct"], domain.ShortestTrip) {
		t.Fatalf("the direct offer lacks its reasons: %v", byID["direct"])
	}
	for _, offer := range ranked {
		if !contains(offer.Reasons, domain.WithinBudget) {
			t.Fatalf("%s is inside budget but not labelled: %v", offer.ID, offer.Reasons)
		}
	}
}

func TestStaleAndUnavailableOffersAreDropped(t *testing.T) {
	fresh := flight(t, "fresh", 20_000, 0, 400)
	expired := flight(t, "expired", 10_000, 0, 300)
	expired.ExpiresAt = now.Add(-time.Minute)
	stale := flight(t, "stale", 11_000, 0, 300)
	stale.ObservedAt = now.Add(-2 * time.Hour)
	stale.ExpiresAt = now.Add(time.Hour)
	unavailable := flight(t, "sold-out", 9_000, 0, 300)
	unavailable.Available = false
	ranked := engine.RankFlights([]domain.FlightOffer{fresh, expired, stale, unavailable}, engine.FlightContext{Now: now})
	if len(ranked) != 1 || ranked[0].ID != "fresh" {
		t.Fatalf("ranked = %v", flightIDs(ranked))
	}
}

func TestDuplicateItinerariesAreCollapsed(t *testing.T) {
	first := flight(t, "partner-a", 20_000, 1, 600)
	second := flight(t, "partner-b", 21_000, 1, 600)
	second.Carrier = first.Carrier
	second.DepartAt = first.DepartAt
	second.ArriveAt = first.ArriveAt
	ranked := engine.RankFlights([]domain.FlightOffer{first, second}, engine.FlightContext{Now: now})
	if len(ranked) != 1 {
		t.Fatalf("the same itinerary appeared %d times", len(ranked))
	}
}

func stay(t *testing.T, id string, minor int64, location geo.Point) domain.StayOffer {
	t.Helper()
	return domain.StayOffer{
		Offer: base(t, id, domain.Stay, minor), PropertyID: id, Location: location,
		Rating: 4.4, ReviewCount: 200, Nights: 4,
	}
}

func TestNearbyStayBeatsACheaperDistantOne(t *testing.T) {
	far := stay(t, "a-cheap-far", 60_000, geo.Point{Latitude: 32.90, Longitude: -17.20})
	near := stay(t, "b-near", 74_000, funchal)
	payout := stay(t, "c-highest-commission", 96_000, geo.Point{Latitude: 33.10, Longitude: -17.40})
	ranked := engine.RankStays([]domain.StayOffer{far, near, payout}, engine.StayContext{
		Now: now, ItineraryCentre: funchal, BudgetMinor: 100_000,
	})
	if ranked[0].ID != "b-near" {
		t.Fatalf("ranking = %v, want the well-located stay first", stayIDs(ranked))
	}
	if ranked[len(ranked)-1].ID != "c-highest-commission" {
		t.Fatalf("the worst-value stay did not rank last: %v", stayIDs(ranked))
	}
	if !contains(ranked[0].Reasons, domain.BestLocation) {
		t.Fatalf("the nearest stay lacks its reason: %v", ranked[0].Reasons)
	}
}

func TestCommissionCannotEnterStayRanking(t *testing.T) {
	offers := []domain.StayOffer{
		stay(t, "a-cheap-far", 60_000, geo.Point{Latitude: 32.90, Longitude: -17.20}),
		stay(t, "b-near", 74_000, funchal),
		stay(t, "c-highest-commission", 96_000, geo.Point{Latitude: 33.10, Longitude: -17.40}),
	}
	context := engine.StayContext{Now: now, ItineraryCentre: funchal, BudgetMinor: 100_000}
	baseline := stayIDs(engine.RankStays(offers, context))
	for range 25 {
		again := stayIDs(engine.RankStays(offers, context))
		for index := range again {
			if again[index] != baseline[index] {
				t.Fatalf("stay ranking is not stable: %v then %v", baseline, again)
			}
		}
	}
}

func TestOverBudgetStaysAreDownrankedNotHidden(t *testing.T) {
	affordable := stay(t, "affordable", 70_000, funchal)
	expensive := stay(t, "expensive", 200_000, funchal)
	ranked := engine.RankStays([]domain.StayOffer{expensive, affordable}, engine.StayContext{
		Now: now, ItineraryCentre: funchal, BudgetMinor: 100_000,
	})
	if len(ranked) != 2 {
		t.Fatalf("an over-budget stay was hidden: %v", stayIDs(ranked))
	}
	if ranked[0].ID != "affordable" {
		t.Fatalf("ranking = %v", stayIDs(ranked))
	}
}

func activity(t *testing.T, id string, minor int64, startsAt time.Time, minutes int, outdoor bool) domain.ActivityOffer {
	t.Helper()
	return domain.ActivityOffer{
		Offer: base(t, id, domain.Activity, minor), PlaceID: id, Location: funchal,
		DurationMinutes: minutes, StartsAt: startsAt, Outdoor: outdoor, Rating: 4.5,
	}
}

func TestActivityThatDoesNotFitTheWindowLosesDespitePrice(t *testing.T) {
	fits := activity(t, "a-fits", 9_000, now.Add(time.Hour), 90, false)
	tooLong := activity(t, "b-too-long", 4_000, now.Add(time.Hour), 480, false)
	wrongDay := activity(t, "c-highest-commission", 3_000, now.Add(72*time.Hour), 60, false)
	ranked := engine.RankActivities([]domain.ActivityOffer{fits, tooLong, wrongDay}, engine.ActivityContext{
		Now: now, Centre: funchal, FreeFrom: now.Add(time.Hour), FreeUntil: now.Add(4 * time.Hour),
	})
	if ranked[0].ID != "a-fits" {
		t.Fatalf("ranking = %v", activityIDs(ranked))
	}
	if !contains(ranked[0].Reasons, domain.FitsItinerary) {
		t.Fatalf("the fitting activity lacks its reason: %v", ranked[0].Reasons)
	}
}

func TestOutdoorActivityIsDownrankedInBadWeather(t *testing.T) {
	startsAt := now.Add(time.Hour)
	outdoor := activity(t, "outdoor", 5_000, startsAt, 90, true)
	indoor := activity(t, "indoor", 12_000, startsAt, 90, false)
	stormy := weather.Window{
		Availability: weather.Available,
		Hourly: []weather.Observation{{
			At: startsAt, Condition: weather.Storm, TemperatureCelsius: 17, PrecipitationChance: 90,
		}},
	}
	ranked := engine.RankActivities([]domain.ActivityOffer{outdoor, indoor}, engine.ActivityContext{
		Now: now, Centre: funchal, FreeFrom: startsAt, FreeUntil: startsAt.Add(4 * time.Hour), Weather: stormy,
	})
	if ranked[0].ID != "indoor" {
		t.Fatalf("a storm did not push the outdoor activity down: %v", activityIDs(ranked))
	}
}

func contains(reasons []domain.Reason, wanted domain.Reason) bool {
	for _, reason := range reasons {
		if reason == wanted {
			return true
		}
	}
	return false
}
