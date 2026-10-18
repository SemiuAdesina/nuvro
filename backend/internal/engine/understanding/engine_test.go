package understanding_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/discovery"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/money"
	"github.com/SemiuAdesina/nuvro/backend/internal/engine/understanding"
)

var now = time.Date(2026, 9, 9, 12, 0, 0, 0, time.UTC)

type interpreter struct {
	inferred understanding.Inferred
	err      error
	calls    int
	seen     string
}

func (i *interpreter) Interpret(_ context.Context, query string) (understanding.Inferred, error) {
	i.calls++
	i.seen = query
	return i.inferred, i.err
}

func budget(t *testing.T, minor int64, currency string) money.Amount {
	t.Helper()
	amount, err := money.New(minor, money.Currency(currency))
	if err != nil {
		t.Fatal(err)
	}
	return amount
}

func baseInput(t *testing.T) understanding.Input {
	t.Helper()
	amount := budget(t, 500_000_000, "NGN")
	return understanding.Input{
		Profile: discovery.Intent{
			Traveler:    discovery.Traveler{ResidenceCountry: "NG", PassportCountries: []string{"NG"}},
			Origin:      discovery.Origin{DepartureAirport: "LOS"},
			Preferences: discovery.Preferences{Weather: "mild", Crowds: "lively"},
		},
		Supplied: understanding.Supplied{Budget: &amount},
		Party:    discovery.Party{Adults: 2},
		Window:   discovery.Window{Start: now.AddDate(0, 1, 0), End: now.AddDate(0, 1, 7)},
		Now:      now,
	}
}

func TestProfileContextIsUsedWhenTheRequestOmitsIt(t *testing.T) {
	intent, err := understanding.New(nil).Understand(t.Context(), baseInput(t))
	if err != nil {
		t.Fatal(err)
	}
	if intent.Origin.DepartureAirport != "LOS" || intent.Traveler.ResidenceCountry != "NG" {
		t.Fatalf("profile context lost: %+v", intent)
	}
	if intent.Sources["departure_airport"] != discovery.SourceProfile {
		t.Fatal("profile provenance not recorded")
	}
	if intent.Budget.Currency().String() != "NGN" {
		t.Fatal("budget currency was replaced")
	}
}

func TestSuppliedValuesOverrideProfileAndInference(t *testing.T) {
	input := baseInput(t)
	input.Supplied.Weather = "warm"
	input.Supplied.Interests = []string{"Beach", "food", "beach"}
	input.Supplied.DepartureAirport = "ABV"
	input.Query = "somewhere cold and quiet"
	engine := understanding.New(&interpreter{inferred: understanding.Inferred{Weather: "cold", Crowds: "quiet"}})
	intent, err := engine.Understand(t.Context(), input)
	if err != nil {
		t.Fatal(err)
	}
	if intent.Preferences.Weather != "warm" {
		t.Fatalf("inferred text overrode an explicit field: %q", intent.Preferences.Weather)
	}
	if intent.Origin.DepartureAirport != "ABV" || intent.Sources["departure_airport"] != discovery.SourceRequest {
		t.Fatal("supplied origin not applied")
	}
	if len(intent.Preferences.Interests) != 2 {
		t.Fatalf("interests were not normalised: %v", intent.Preferences.Interests)
	}
	if intent.Preferences.Crowds != "lively" || intent.Sources["crowds"] != discovery.SourceProfile {
		t.Fatalf("profile value was replaced by inference: %+v", intent.Preferences)
	}
}

func TestInferenceOnlyFillsUnsetFields(t *testing.T) {
	input := baseInput(t)
	input.Profile.Preferences = discovery.Preferences{Weather: "any"}
	input.Query = "somewhere warm with beaches"
	engine := understanding.New(&interpreter{inferred: understanding.Inferred{Weather: "warm", Interests: []string{"beach"}}})
	intent, err := engine.Understand(t.Context(), input)
	if err != nil {
		t.Fatal(err)
	}
	if intent.Preferences.Weather != "warm" || intent.Sources["weather"] != discovery.SourceInferred {
		t.Fatalf("inference did not fill an unset field: %+v", intent.Preferences)
	}
	if len(intent.Preferences.Interests) != 1 || intent.Sources["interests"] != discovery.SourceInferred {
		t.Fatal("inferred interests were not applied")
	}
}

func TestInvalidOrFailedInterpretationIsDiscarded(t *testing.T) {
	for name, source := range map[string]*interpreter{
		"provider failure": {err: errors.New("model unavailable")},
		"invalid output":   {inferred: understanding.Inferred{Weather: "tropical"}},
		"invented enum":    {inferred: understanding.Inferred{Crowds: "packed to the brim"}},
	} {
		input := baseInput(t)
		input.Profile.Preferences = discovery.Preferences{}
		input.Query = "somewhere nice"
		intent, err := understanding.New(source).Understand(t.Context(), input)
		if err != nil {
			t.Fatalf("%s broke the request: %v", name, err)
		}
		if intent.Preferences.Weather != "" || intent.Preferences.Crowds != "" {
			t.Fatalf("%s leaked into the intent: %+v", name, intent.Preferences)
		}
	}
}

func TestQueryIsNotSentWithoutAnInterpreterOrWhenOversized(t *testing.T) {
	input := baseInput(t)
	input.Query = "somewhere warm"
	if _, err := understanding.New(nil).Understand(t.Context(), input); err != nil {
		t.Fatal(err)
	}
	source := &interpreter{}
	oversized := baseInput(t)
	oversized.Query = string(make([]byte, 501))
	if _, err := understanding.New(source).Understand(t.Context(), oversized); err != nil {
		t.Fatal(err)
	}
	if source.calls != 0 {
		t.Fatal("an oversized query reached the interpreter")
	}
}

func TestInvalidRequestsAreRejected(t *testing.T) {
	for name, mutate := range map[string]func(*understanding.Input){
		"no adults":      func(i *understanding.Input) { i.Party.Adults = 0 },
		"reversed dates": func(i *understanding.Input) { i.Window.End = i.Window.Start.AddDate(0, 0, -2) },
		"past departure": func(i *understanding.Input) { i.Window.Start = now.AddDate(0, 0, -10); i.Window.End = now },
		"too long":       func(i *understanding.Input) { i.Window.End = i.Window.Start.AddDate(1, 0, 0) },
		"bad passport":   func(i *understanding.Input) { i.Supplied.PassportCountries = []string{"NGA"} },
		"bad airport":    func(i *understanding.Input) { i.Supplied.DepartureAirport = "Lagos" },
	} {
		input := baseInput(t)
		mutate(&input)
		if _, err := understanding.New(nil).Understand(t.Context(), input); err == nil {
			t.Fatalf("%s was accepted", name)
		}
	}
}

func TestMissingContextIsReported(t *testing.T) {
	input := baseInput(t)
	input.Profile.Traveler = discovery.Traveler{}
	input.Profile.Origin = discovery.Origin{}
	intent, err := understanding.New(nil).Understand(t.Context(), input)
	if err != nil {
		t.Fatal(err)
	}
	missing := intent.MissingContext()
	if len(missing) != 3 {
		t.Fatalf("missing context = %v", missing)
	}
}
