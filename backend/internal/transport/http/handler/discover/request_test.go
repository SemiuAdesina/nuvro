package discover

import (
	"strings"
	"testing"
	"time"

	openapi "github.com/oapi-codegen/runtime/types"

	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/api"
)

func body() api.MatchRequest {
	return api.MatchRequest{
		Adults:    2,
		Budget:    api.Money{AmountMinor: 500_000_000, Currency: "NGN"},
		StartDate: openapi.Date{Time: time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC)},
		EndDate:   openapi.Date{Time: time.Date(2026, 10, 19, 0, 0, 0, 0, time.UTC)},
	}
}

func TestRequestMappingPreservesTheTravellerCurrencyAndDates(t *testing.T) {
	request, ok := discoveryRequest(body())
	if !ok {
		t.Fatal("a valid body was rejected")
	}
	if request.Overrides.Budget == nil || request.Overrides.Budget.Currency().String() != "NGN" {
		t.Fatalf("budget currency lost: %+v", request.Overrides.Budget)
	}
	if request.Overrides.Budget.Minor() != 500_000_000 {
		t.Fatal("budget was not carried as integer minor units")
	}
	if request.Window.Start.Location() != time.UTC || request.Window.Start.Hour() != 0 {
		t.Fatalf("dates were not normalised to calendar days: %v", request.Window.Start)
	}
	if request.Window.Nights() != 7 || request.Party.Adults != 2 || request.Party.Children != 0 {
		t.Fatalf("window or party mapped wrongly: %+v %+v", request.Window, request.Party)
	}
}

func TestAbsentOptionalFieldsStayUnset(t *testing.T) {
	request, ok := discoveryRequest(body())
	if !ok {
		t.Fatal("rejected")
	}
	if request.Overrides.Weather != "" || request.Overrides.DepartureAirport != "" || request.Overrides.Interests != nil {
		t.Fatalf("absent fields became values: %+v", request.Overrides)
	}
	if request.Overrides.MaxFlightMinutes != nil || request.Query != "" || request.Limit != 0 {
		t.Fatalf("absent optional values were invented: %+v", request)
	}
}

func TestSuppliedOverridesAreCarried(t *testing.T) {
	full := body()
	weather := api.MatchRequestWeather("warm")
	paperwork := api.MatchRequestPaperwork("keep_it_simple")
	airport := "LOS"
	query := "somewhere warm"
	minutes := 600
	children := 1
	limit := 5
	interests := []string{"beach"}
	passports := []string{"NG", "US"}
	full.Weather = &weather
	full.Paperwork = &paperwork
	full.DepartureAirport = &airport
	full.Query = &query
	full.MaxFlightMinutes = &minutes
	full.Children = &children
	full.Limit = &limit
	full.Interests = &interests
	full.PassportCountries = &passports
	mapped, ok := discoveryRequest(full)
	if !ok {
		t.Fatal("rejected")
	}
	if mapped.Overrides.Weather != "warm" || mapped.Overrides.Paperwork != "keep_it_simple" {
		t.Fatalf("enum overrides lost: %+v", mapped.Overrides)
	}
	if mapped.Overrides.DepartureAirport != "LOS" || mapped.Query != "somewhere warm" {
		t.Fatal("text overrides lost")
	}
	if len(mapped.Overrides.PassportCountries) != 2 || mapped.Party.Children != 1 || mapped.Limit != 5 {
		t.Fatalf("collections or counts lost: %+v", mapped)
	}
	if mapped.Overrides.MaxFlightMinutes == nil || *mapped.Overrides.MaxFlightMinutes != 600 {
		t.Fatal("flight tolerance lost")
	}
}

func TestMalformedCurrencyIsRejectedAtTheBoundary(t *testing.T) {
	for _, currency := range []string{"ngn", "NGNX", "", "N1N"} {
		malformed := body()
		malformed.Budget.Currency = currency
		if _, ok := discoveryRequest(malformed); ok {
			t.Fatalf("currency %q was accepted", currency)
		}
	}
}

func FuzzDiscoveryRequestCurrency(f *testing.F) {
	f.Add("NGN")
	f.Add("usd")
	f.Add("")
	f.Add(strings.Repeat("A", 40))
	f.Fuzz(func(t *testing.T, currency string) {
		malformed := body()
		malformed.Budget.Currency = currency
		request, ok := discoveryRequest(malformed)
		if !ok {
			return
		}
		if request.Overrides.Budget == nil || request.Overrides.Budget.Currency().String() != currency {
			t.Fatalf("accepted %q but stored something else", currency)
		}
		if len(currency) != 3 {
			t.Fatalf("accepted a currency of length %d", len(currency))
		}
	})
}
