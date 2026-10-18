package feasibility_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/destination"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/discovery"
	domain "github.com/SemiuAdesina/nuvro/backend/internal/domain/feasibility"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/money"
	engine "github.com/SemiuAdesina/nuvro/backend/internal/engine/feasibility"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
)

var evaluatedAt = time.Date(2026, 9, 9, 12, 0, 0, 0, time.UTC)

type entryProvider struct {
	byPassport map[string]engine.EntryEvidence
	err        error
	queries    []engine.EntryQuery
}

func (e *entryProvider) Entry(_ context.Context, query engine.EntryQuery) (engine.EntryEvidence, error) {
	e.queries = append(e.queries, query)
	if e.err != nil {
		return engine.EntryEvidence{}, e.err
	}
	evidence, found := e.byPassport[query.PassportCountry]
	if !found {
		return engine.EntryEvidence{}, errors.New("no fixture for this passport")
	}
	return evidence, nil
}

type flightProvider struct {
	evidence engine.FlightEvidence
	err      error
}

func (f flightProvider) Flights(context.Context, engine.FlightQuery) (engine.FlightEvidence, error) {
	return f.evidence, f.err
}

type costProvider struct {
	evidence engine.CostEvidence
	err      error
}

func (c costProvider) Estimate(context.Context, engine.CostQuery) (engine.CostEvidence, error) {
	return c.evidence, c.err
}

type rateProvider struct {
	micros int64
	err    error
	calls  int
}

func (r *rateProvider) Rate(_ context.Context, _, _ money.Currency, _ time.Time) (engine.Rate, error) {
	r.calls++
	if r.err != nil {
		return engine.Rate{}, r.err
	}
	rate, err := money.NewRate(r.micros)
	if err != nil {
		return engine.Rate{}, err
	}
	return engine.Rate{Rate: rate, Source: "fixture", ObservedAt: evaluatedAt}, nil
}

func amount(t *testing.T, minor int64, currency string) money.Amount {
	t.Helper()
	value, err := money.New(minor, money.Currency(currency))
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func traveller(t *testing.T, passports []string) discovery.Intent {
	t.Helper()
	return discovery.Intent{
		Traveler: discovery.Traveler{ResidenceCountry: "NG", PassportCountries: passports},
		Origin:   discovery.Origin{DepartureAirport: "LOS"},
		Party:    discovery.Party{Adults: 2},
		Window:   discovery.Window{Start: evaluatedAt.AddDate(0, 1, 0), End: evaluatedAt.AddDate(0, 1, 7)},
		Budget:   amount(t, 500_000_000, "NGN"),
	}
}

func target() destination.Destination {
	return destination.Destination{ID: "zanzibar-tz", Name: "Zanzibar", CountryCode: "TZ", PrimaryAirport: "ZNZ", Timezone: "Africa/Dar_es_Salaam"}
}

func costs(t *testing.T, minimum, expected, maximum int64, currency string) costProvider {
	t.Helper()
	return costProvider{evidence: engine.CostEvidence{
		Minimum: amount(t, minimum, currency), Expected: amount(t, expected, currency), Maximum: amount(t, maximum, currency),
		Confidence: domain.Medium, Components: []string{"flight", "stay", "food"}, ObservedAt: evaluatedAt,
	}}
}

func evaluate(t *testing.T, providers engine.Providers, intent discovery.Intent) domain.Result {
	t.Helper()
	result := engine.New(providers, clock.NewFixed(evaluatedAt)).Evaluate(t.Context(), intent, target())
	if err := result.Validate(); err != nil {
		t.Fatalf("invalid result: %v", err)
	}
	return result
}

func TestNoProvidersProducesHonestUnknownEvidence(t *testing.T) {
	result := evaluate(t, engine.Providers{}, traveller(t, []string{"NG"}))
	if result.Access.Status != domain.UnknownAccess || result.Access.EntryMode != domain.UnknownEntry {
		t.Fatalf("unconfigured entry provider claimed knowledge: %+v", result.Access)
	}
	if result.Affordability.Status != domain.InsufficientEvidence || result.Affordability.HasEstimate {
		t.Fatalf("unconfigured cost provider invented an estimate: %+v", result.Affordability)
	}
	if result.Flight.Status != domain.FlightUnknown || result.Overall != domain.UnknownOverall {
		t.Fatalf("unknown evidence was not carried through: %+v", result)
	}
}

func TestMissingPassportLeavesAccessUnknown(t *testing.T) {
	providers := engine.Providers{
		Entry:  &entryProvider{byPassport: map[string]engine.EntryEvidence{"NG": {EntryMode: domain.VisaFree, Preparation: domain.NoPreparation, Confidence: domain.High}}},
		Flight: flightProvider{evidence: engine.FlightEvidence{Available: true, Stops: 0, DurationMinutes: 320, Confidence: domain.High, ObservedAt: evaluatedAt}},
		Costs:  costs(t, 300_000_000, 350_000_000, 400_000_000, "NGN"),
	}
	intent := traveller(t, nil)
	result := evaluate(t, providers, intent)
	if result.Access.Status != domain.UnknownAccess {
		t.Fatalf("access was asserted without a passport: %+v", result.Access)
	}
	if len(result.Access.Reasons) == 0 || result.Access.Reasons[0] != "passport_context_required" {
		t.Fatalf("the reason was not explained: %+v", result.Access.Reasons)
	}
	if result.Affordability.Status != domain.StrongFit {
		t.Fatalf("affordability should still be assessed: %+v", result.Affordability)
	}
}

func TestBestPassportIsChosenWithoutPreferringAnyCountry(t *testing.T) {
	source := &entryProvider{byPassport: map[string]engine.EntryEvidence{
		"NG": {EntryMode: domain.VisaRequired, Preparation: domain.HighPreparation, Confidence: domain.High},
		"PT": {EntryMode: domain.VisaFree, Preparation: domain.NoPreparation, Confidence: domain.High},
	}}
	providers := engine.Providers{Entry: source, Costs: costs(t, 1, 2, 3, "NGN")}
	forward := evaluate(t, providers, traveller(t, []string{"NG", "PT"}))
	reversed := evaluate(t, providers, traveller(t, []string{"PT", "NG"}))
	if forward.Access.Status != domain.Easy || reversed.Access.Status != domain.Easy {
		t.Fatalf("the stronger document was not selected: %+v %+v", forward.Access, reversed.Access)
	}
	if forward.Access.PassportUsed != "PT" || reversed.Access.PassportUsed != "PT" {
		t.Fatal("the chosen document was not recorded, or ordering changed the outcome")
	}
	if len(source.queries) != 4 {
		t.Fatalf("every held document should be evaluated, saw %d queries", len(source.queries))
	}
}

func TestAffordabilityConvertsIntoTheTravellerCurrencyWithProvenance(t *testing.T) {
	rates := &rateProvider{micros: 1_500_000_000}
	providers := engine.Providers{Costs: costs(t, 2_000_00, 2_400_00, 2_800_00, "USD"), Rates: rates}
	intent := traveller(t, []string{"NG"})
	result := evaluate(t, providers, intent)
	if result.Affordability.Budget.Currency().String() != "NGN" {
		t.Fatal("the traveller's budget currency was replaced")
	}
	if result.Affordability.Estimate.Expected.Currency().String() != "NGN" {
		t.Fatalf("the estimate was not converted: %+v", result.Affordability.Estimate.Expected)
	}
	if result.Affordability.Conversion == nil || result.Affordability.Conversion.Provider != "fixture" {
		t.Fatal("conversion provenance was not preserved")
	}
	if result.Affordability.Conversion.Source.Currency().String() != "USD" {
		t.Fatal("the original currency was not preserved")
	}
	if rates.calls == 0 {
		t.Fatal("the exchange rate port was never consulted")
	}
}

func TestMissingExchangeRateNeverInventsAnEstimate(t *testing.T) {
	providers := engine.Providers{Costs: costs(t, 2_000_00, 2_400_00, 2_800_00, "USD")}
	result := evaluate(t, providers, traveller(t, []string{"NG"}))
	if result.Affordability.Status != domain.InsufficientEvidence || result.Affordability.HasEstimate {
		t.Fatalf("a cross-currency estimate was invented: %+v", result.Affordability)
	}
	failing := engine.Providers{Costs: costs(t, 1, 2, 3, "USD"), Rates: &rateProvider{err: errors.New("rate provider down")}}
	degraded := evaluate(t, failing, traveller(t, []string{"NG"}))
	if degraded.Affordability.Status != domain.InsufficientEvidence {
		t.Fatal("a failing rate provider produced an estimate")
	}
}

func TestAffordabilityStatusFollowsTheBudget(t *testing.T) {
	for name, test := range map[string]struct {
		minimum, expected, maximum int64
		want                       domain.AffordabilityStatus
	}{
		"comfortable": {200_000_000, 300_000_000, 400_000_000, domain.StrongFit},
		"good":        {400_000_000, 450_000_000, 490_000_000, domain.GoodFit},
		"tight":       {400_000_000, 480_000_000, 600_000_000, domain.Tight},
		"over":        {600_000_000, 700_000_000, 800_000_000, domain.OverBudget},
	} {
		providers := engine.Providers{Costs: costs(t, test.minimum, test.expected, test.maximum, "NGN")}
		result := evaluate(t, providers, traveller(t, []string{"NG"}))
		if result.Affordability.Status != test.want {
			t.Fatalf("%s: status = %q, want %q", name, result.Affordability.Status, test.want)
		}
	}
}

func TestPartialProviderFailureDegradesInsteadOfFailing(t *testing.T) {
	providers := engine.Providers{
		Entry:  &entryProvider{byPassport: map[string]engine.EntryEvidence{"NG": {EntryMode: domain.EVisa, Preparation: domain.MediumPreparation, Confidence: domain.High}}},
		Flight: flightProvider{err: errors.New("flight provider unavailable")},
		Costs:  costs(t, 300_000_000, 350_000_000, 400_000_000, "NGN"),
	}
	result := evaluate(t, providers, traveller(t, []string{"NG"}))
	if result.Access.Status != domain.Manageable {
		t.Fatalf("entry evidence was lost: %+v", result.Access)
	}
	if result.Affordability.Status != domain.StrongFit {
		t.Fatalf("cost evidence was lost: %+v", result.Affordability)
	}
	if result.Flight.Status != domain.FlightUnknown {
		t.Fatalf("a failed flight provider was not reported as unknown: %+v", result.Flight)
	}
	if !hasCaution(result, "flight_evidence_unknown") {
		t.Fatalf("no caution was raised: %v", result.Cautions)
	}
}

func TestNoRouteMakesTheTripInfeasible(t *testing.T) {
	providers := engine.Providers{
		Entry:  &entryProvider{byPassport: map[string]engine.EntryEvidence{"NG": {EntryMode: domain.VisaFree, Preparation: domain.NoPreparation, Confidence: domain.High}}},
		Flight: flightProvider{evidence: engine.FlightEvidence{Available: false, Confidence: domain.High, ObservedAt: evaluatedAt}},
		Costs:  costs(t, 1, 2, 3, "NGN"),
	}
	result := evaluate(t, providers, traveller(t, []string{"NG"}))
	if result.Overall != domain.Infeasible {
		t.Fatalf("overall = %q, want not_feasible", result.Overall)
	}
}

func TestLongFlightAndTransitReviewRaiseCautions(t *testing.T) {
	providers := engine.Providers{
		Entry: &entryProvider{byPassport: map[string]engine.EntryEvidence{"NG": {
			EntryMode: domain.EVisa, Preparation: domain.MediumPreparation, Confidence: domain.High,
			TransitRisk: domain.TransitReviewNeeded,
		}}},
		Flight: flightProvider{evidence: engine.FlightEvidence{Available: true, Stops: 2, DurationMinutes: 1400, Confidence: domain.Medium, ObservedAt: evaluatedAt}},
		Costs:  costs(t, 300_000_000, 350_000_000, 400_000_000, "NGN"),
	}
	intent := traveller(t, []string{"NG"})
	intent.Preferences.MaxFlightMinutes = 600
	result := evaluate(t, providers, intent)
	if !hasCaution(result, "long_flight") || !hasCaution(result, "transit_review_required") {
		t.Fatalf("cautions = %v", result.Cautions)
	}
}

func hasCaution(result domain.Result, want string) bool {
	for _, caution := range result.Cautions {
		if caution == want {
			return true
		}
	}
	return false
}
