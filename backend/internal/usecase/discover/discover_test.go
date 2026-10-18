package discover_test

import (
	"testing"

	"github.com/SemiuAdesina/nuvro/backend/internal/apperror"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/destination"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/discovery"
	domain "github.com/SemiuAdesina/nuvro/backend/internal/domain/feasibility"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	engine "github.com/SemiuAdesina/nuvro/backend/internal/engine/feasibility"
	"github.com/SemiuAdesina/nuvro/backend/internal/engine/understanding"
	"github.com/SemiuAdesina/nuvro/backend/internal/usecase/discover"
)

const traveller = "3f1c9c8e-0e3a-4f0a-9b1f-1a2b3c4d5e6f"
const thailand = "phuket-th"

func destinations() []destination.Destination {
	return []destination.Destination{
		place("zanzibar-tz", "TZ", "ZNZ", []string{"warm"}, []string{"beach", "food", "relaxation"}, "quiet"),
		place("phuket-th", "TH", "HKT", []string{"warm"}, []string{"beach", "food", "relaxation"}, "lively"),
		place("bridgetown-bb", "BB", "BGI", []string{"warm"}, []string{"beach", "relaxation"}, "quiet"),
		place("reykjavik-is", "IS", "KEF", []string{"cold"}, []string{"nature", "adventure"}, "quiet"),
	}
}

func warmBeachRequest(t *testing.T, budgetMinor int64, currency string) discover.Request {
	t.Helper()
	budget := amount(t, budgetMinor, currency)
	return discover.Request{
		Overrides: understanding.Supplied{
			Budget: &budget, Interests: []string{"beach", "food", "relaxation"},
			Weather: "warm", Crowds: "quiet", Paperwork: "keep_it_simple",
		},
		Party:  discovery.Party{Adults: 2},
		Window: discovery.Window{Start: evaluatedAt.AddDate(0, 1, 0), End: evaluatedAt.AddDate(0, 1, 7)},
	}
}

func profileFor(country, airport, currency string) profiles {
	return profiles{
		profile: user.Profile{
			UserID: traveller, PreferredLanguage: "en", Locale: "en", PreferredCurrency: currency,
			HomeCountry: &country, DepartureAirport: &airport, Units: "metric", HourCycle: "24",
		},
		preferences: user.DefaultPreferences(),
	}
}

func rankOf(result discover.Response, id string) int {
	for _, recommendation := range result.Result.Recommendations {
		if recommendation.Destination.ID == id {
			return recommendation.Rank
		}
	}
	return -1
}

func nigeriaProviders() engine.Providers {
	return engine.Providers{
		Entry: entryProvider{rules: map[string]map[string]entryRule{
			"NG": {
				"TZ": {mode: domain.VisaFree, preparation: domain.NoPreparation},
				"BB": {mode: domain.VisaFree, preparation: domain.NoPreparation},
				"TH": {mode: domain.VisaRequired, preparation: domain.HighPreparation},
				"IS": {mode: domain.VisaRequired, preparation: domain.HighPreparation},
			},
			"US": {
				"TZ": {mode: domain.EVisa, preparation: domain.LowPreparation},
				"BB": {mode: domain.VisaFree, preparation: domain.NoPreparation},
				"TH": {mode: domain.VisaFree, preparation: domain.NoPreparation},
				"IS": {mode: domain.VisaFree, preparation: domain.NoPreparation},
			},
		}},
		Flight: flightProvider{routes: map[string]engine.FlightEvidence{
			"LOS-ZNZ": {Available: true, Stops: 1, DurationMinutes: 480},
			"LOS-HKT": {Available: true, Stops: 2, DurationMinutes: 1250},
			"LOS-BGI": {Available: true, Stops: 2, DurationMinutes: 1100},
			"LOS-KEF": {Available: true, Stops: 2, DurationMinutes: 900},
			"JFK-ZNZ": {Available: true, Stops: 2, DurationMinutes: 1150},
			"JFK-HKT": {Available: true, Stops: 1, DurationMinutes: 1200},
			"JFK-BGI": {Available: true, Stops: 0, DurationMinutes: 280},
			"JFK-KEF": {Available: true, Stops: 0, DurationMinutes: 330},
		}},
		Costs: costProvider{currency: "USD", perNight: map[string]int64{
			"zanzibar-tz": 12_000, "phuket-th": 11_000, "bridgetown-bb": 26_000, "reykjavik-is": 30_000,
		}},
		Rates: rateProvider{micros: map[string]int64{"USD-NGN": 1_500_000_000, "USD-USD": 1_000_000}},
	}
}

func TestNigeriaTravellerPrefersLowerFrictionDestinations(t *testing.T) {
	service := newService(t, profileFor("NG", "LOS", "NGN"), &catalog{destinations: destinations()}, nigeriaProviders())
	result, err := service.Matches(t.Context(), user.Principal{UserID: traveller},
		withPassports(warmBeachRequest(t, 5_000_000_00, "NGN"), "NG"))
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Result.Recommendations) == 0 {
		t.Fatal("no recommendations")
	}
	top := result.Result.Recommendations[0]
	if top.Destination.ID != "zanzibar-tz" {
		t.Fatalf("top destination = %q, want the low-friction warm beach option", top.Destination.ID)
	}
	if top.Feasibility.Access.Status != domain.Easy || top.Feasibility.Affordability.Budget.Currency().String() != "NGN" {
		t.Fatalf("top result lost evidence or currency: %+v", top.Feasibility.Access)
	}
	if rankOf(result, "phuket-th") <= rankOf(result, "zanzibar-tz") {
		t.Fatal("a high-preparation destination outranked an equally enjoyable easy one")
	}
	if rankOf(result, "reykjavik-is") == 1 {
		t.Fatal("a cold destination topped a warm beach request")
	}
}

func TestUnitedStatesTravellerRanksDifferentlyOnTheSameCatalog(t *testing.T) {
	providers := nigeriaProviders()
	providers.Costs = costProvider{currency: "USD", perNight: map[string]int64{
		"zanzibar-tz": 12_000, "phuket-th": 11_000, "bridgetown-bb": 26_000, "reykjavik-is": 30_000,
	}}
	nigeria := newService(t, profileFor("NG", "LOS", "NGN"), &catalog{destinations: destinations()}, providers)
	unitedStates := newService(t, profileFor("US", "JFK", "USD"), &catalog{destinations: destinations()}, providers)
	fromLagos, err := nigeria.Matches(t.Context(), user.Principal{UserID: traveller},
		withPassports(warmBeachRequest(t, 5_000_000_00, "NGN"), "NG"))
	if err != nil {
		t.Fatal(err)
	}
	fromNewYork, err := unitedStates.Matches(t.Context(), user.Principal{UserID: traveller},
		withPassports(warmBeachRequest(t, 6_000_00, "USD"), "US"))
	if err != nil {
		t.Fatal(err)
	}
	if ordered(fromLagos) == ordered(fromNewYork) {
		t.Fatalf("both travellers received an identical ranking: %v", ordered(fromLagos))
	}
	if fromNewYork.Result.Recommendations[0].Feasibility.Affordability.Budget.Currency().String() != "USD" {
		t.Fatal("the traveller's currency was replaced")
	}
	if accessFor(fromLagos).Status == accessFor(fromNewYork).Status {
		t.Fatal("the same destination produced identical access for two different passports")
	}
	if rankOf(fromLagos, "bridgetown-bb") != -1 && rankOf(fromNewYork, "bridgetown-bb") == -1 {
		t.Fatal("affordability evidence did not differ between the two budgets")
	}
	for _, result := range []discover.Response{fromLagos, fromNewYork} {
		if result.Result.RankingVersion == "" || result.EvidenceVersion == "" {
			t.Fatal("a result is missing its algorithm provenance")
		}
	}
}

func ordered(result discover.Response) string {
	joined := ""
	for _, recommendation := range result.Result.Recommendations {
		joined += recommendation.Destination.ID + ";"
	}
	return joined
}

func TestDualPassportUnlocksTheStrongerAccessPath(t *testing.T) {
	providers := nigeriaProviders()
	catalogue := &catalog{destinations: destinations()}
	service := newService(t, profileFor("NG", "LOS", "NGN"), catalogue, providers)
	single, err := service.Matches(t.Context(), user.Principal{UserID: traveller},
		withPassports(warmBeachRequest(t, 5_000_000_00, "NGN"), "NG"))
	if err != nil {
		t.Fatal(err)
	}
	dual, err := service.Matches(t.Context(), user.Principal{UserID: traveller},
		withPassports(warmBeachRequest(t, 5_000_000_00, "NGN"), "NG", "US"))
	if err != nil {
		t.Fatal(err)
	}
	singleThailand := accessFor(single)
	dualThailand := accessFor(dual)
	if singleThailand.Status != domain.PreparationRequired {
		t.Fatalf("single passport access = %q", singleThailand.Status)
	}
	if dualThailand.Status != domain.Easy || dualThailand.PassportUsed != "US" {
		t.Fatalf("the second passport was not used: %+v", dualThailand)
	}
}

func TestMissingPassportKeepsRecommendationsButNotAccessClaims(t *testing.T) {
	service := newService(t, profileFor("NG", "LOS", "NGN"), &catalog{destinations: destinations()}, nigeriaProviders())
	result, err := service.Matches(t.Context(), user.Principal{UserID: traveller}, warmBeachRequest(t, 5_000_000_00, "NGN"))
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Result.Recommendations) == 0 {
		t.Fatal("no recommendations without a passport")
	}
	for _, recommendation := range result.Result.Recommendations {
		if recommendation.Feasibility.Access.Status != domain.UnknownAccess {
			t.Fatalf("%s claimed access without passport context: %q", recommendation.Destination.ID, recommendation.Feasibility.Access.Status)
		}
		if !recommendation.Feasibility.Affordability.HasEstimate {
			t.Fatal("affordability should still be assessed without a passport")
		}
	}
	if len(result.MissingContext) == 0 || result.MissingContext[0] != "passport" {
		t.Fatalf("missing context = %v", result.MissingContext)
	}
}

func TestPartialProviderFailureStillReturnsRecommendations(t *testing.T) {
	providers := nigeriaProviders()
	providers.Flight = flightProvider{
		routes: map[string]engine.FlightEvidence{"LOS-ZNZ": {Available: true, Stops: 1, DurationMinutes: 480}},
		failed: map[string]bool{"LOS-HKT": true, "LOS-BGI": true, "LOS-KEF": true},
	}
	service := newService(t, profileFor("NG", "LOS", "NGN"), &catalog{destinations: destinations()}, providers)
	result, err := service.Matches(t.Context(), user.Principal{UserID: traveller},
		withPassports(warmBeachRequest(t, 5_000_000_00, "NGN"), "NG"))
	if err != nil {
		t.Fatalf("a failing flight provider broke the whole response: %v", err)
	}
	if len(result.Result.Recommendations) < 2 {
		t.Fatalf("recommendations collapsed to %d", len(result.Result.Recommendations))
	}
	degraded := 0
	for _, recommendation := range result.Result.Recommendations {
		if recommendation.Feasibility.Flight.Status == domain.FlightUnknown {
			degraded++
		}
	}
	if degraded == 0 {
		t.Fatal("no recommendation reported unknown flight evidence")
	}
}

func TestOverBudgetDestinationsAreNotRecommended(t *testing.T) {
	providers := nigeriaProviders()
	service := newService(t, profileFor("NG", "LOS", "NGN"), &catalog{destinations: destinations()}, providers)
	result, err := service.Matches(t.Context(), user.Principal{UserID: traveller},
		withPassports(warmBeachRequest(t, 60_000_00, "NGN"), "NG"))
	if err != nil {
		t.Fatal(err)
	}
	for _, recommendation := range result.Result.Recommendations {
		if recommendation.Feasibility.Affordability.Status == domain.OverBudget {
			t.Fatalf("%s is over budget but was recommended", recommendation.Destination.ID)
		}
	}
	if result.Result.Excluded == 0 {
		t.Fatal("nothing was excluded on a very small budget")
	}
}

func TestUnauthenticatedAndInvalidRequestsAreRefused(t *testing.T) {
	service := newService(t, profileFor("NG", "LOS", "NGN"), &catalog{destinations: destinations()}, nigeriaProviders())
	if _, err := service.Matches(t.Context(), user.Principal{}, warmBeachRequest(t, 5_000_000_00, "NGN")); apperror.CodeOf(err) != apperror.CodeUnauthenticated {
		t.Fatal("anonymous discovery")
	}
	noBudget := warmBeachRequest(t, 5_000_000_00, "NGN")
	noBudget.Overrides.Budget = nil
	if _, err := service.Matches(t.Context(), user.Principal{UserID: traveller}, noBudget); apperror.CodeOf(err) != apperror.CodeInvalidInput {
		t.Fatal("a request without a budget was accepted")
	}
	reversed := warmBeachRequest(t, 5_000_000_00, "NGN")
	reversed.Window.End = reversed.Window.Start.AddDate(0, 0, -3)
	if _, err := service.Matches(t.Context(), user.Principal{UserID: traveller}, reversed); apperror.CodeOf(err) != apperror.CodeInvalidInput {
		t.Fatal("a reversed date range was accepted")
	}
}

func TestCatalogFailurePropagates(t *testing.T) {
	failing := &catalog{err: apperror.New(apperror.CodeUnavailable, "The catalog is unavailable.")}
	service := newService(t, profileFor("NG", "LOS", "NGN"), failing, nigeriaProviders())
	if _, err := service.Matches(t.Context(), user.Principal{UserID: traveller}, warmBeachRequest(t, 5_000_000_00, "NGN")); err == nil {
		t.Fatal("a catalog failure was hidden")
	}
}

func withPassports(request discover.Request, passports ...string) discover.Request {
	request.Overrides.PassportCountries = passports
	return request
}

func accessFor(result discover.Response) domain.Access {
	for _, recommendation := range result.Result.Recommendations {
		if recommendation.Destination.ID == thailand {
			return recommendation.Feasibility.Access
		}
	}
	return domain.Access{}
}
