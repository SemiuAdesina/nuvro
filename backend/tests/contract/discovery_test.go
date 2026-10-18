package contract_test

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"
)

func discoveryBody(t *testing.T, overrides string) string {
	t.Helper()
	start := contractNow.AddDate(0, 1, 0).Format(time.DateOnly)
	end := contractNow.AddDate(0, 1, 7).Format(time.DateOnly)
	base := fmt.Sprintf(`"budget":{"amountMinor":500000000,"currency":"NGN"},"startDate":%q,"endDate":%q,"adults":2`, start, end)
	if overrides != "" {
		base += "," + overrides
	}
	return "{" + base + "}"
}

func TestDiscoverRequiresAVerifiedPrincipal(t *testing.T) {
	environment := newEnvironment(t)
	response, problem := environment.call(t, http.MethodPost, "/v1/discover/matches", discoveryBody(t, ""), false)
	if response.Code != http.StatusUnauthorized || problem["code"] != "authentication_required" {
		t.Fatalf("returned %d %v", response.Code, problem)
	}
}

func TestDiscoverReturnsRankedRecommendationsWithEvidence(t *testing.T) {
	environment := newEnvironment(t)
	body := discoveryBody(t, `"interests":["beach","food","relaxation"],"weather":"warm","crowds":"quiet","paperwork":"keep_it_simple","passportCountries":["NG"],"departureAirport":"LOS","residenceCountry":"NG"`)
	response, decoded := environment.call(t, http.MethodPost, "/v1/discover/matches", body, true)
	if response.Code != http.StatusOK {
		t.Fatalf("returned %d %v", response.Code, decoded)
	}
	if text(t, decoded, "algorithmVersion") != "match-v1" || text(t, decoded, "rankingVersion") != "rank-v1" {
		t.Fatalf("versions missing: %v", decoded)
	}
	recommendations := list(t, decoded, "recommendations")
	if len(recommendations) == 0 {
		t.Fatal("no recommendations returned")
	}
	top, ok := recommendations[0].(map[string]any)
	if !ok {
		t.Fatal("recommendation is not an object")
	}
	if member(t, top["destination"], "id") != "zanzibar-tz" {
		t.Fatalf("unexpected top destination: %v", top["destination"])
	}
	assertEvidence(t, top)
}

func assertEvidence(t *testing.T, top map[string]any) {
	t.Helper()
	match := object(t, top, "personalMatch")
	if match["score"] == nil || match["algorithmVersion"] != "match-v1" {
		t.Fatalf("personal match missing: %v", match)
	}
	if len(list(t, match, "components")) != 7 {
		t.Fatalf("score is not explainable: %v", match["components"])
	}
	access := object(t, top, "access")
	if access["status"] != "easy" || access["entryMode"] != "visa_free" || access["confidence"] != "high" {
		t.Fatalf("access evidence missing: %v", access)
	}
	affordability := object(t, top, "affordability")
	if object(t, affordability, "budget")["currency"] != "NGN" {
		t.Fatalf("the traveller's currency was not preserved: %v", affordability["budget"])
	}
	conversion := object(t, affordability, "conversion")
	if object(t, conversion, "original")["currency"] != "USD" || conversion["source"] == "" {
		t.Fatalf("conversion provenance incomplete: %v", conversion)
	}
	if object(t, top, "flight")["status"] != "available" {
		t.Fatalf("flight summary missing: %v", top["flight"])
	}
	if len(list(t, top, "reasons")) == 0 {
		t.Fatal("no structured reasons")
	}
}

func object(t *testing.T, source map[string]any, field string) map[string]any {
	t.Helper()
	value, ok := source[field].(map[string]any)
	if !ok {
		t.Fatalf("%s is not an object in %v", field, source)
	}
	return value
}

func TestDiscoverReportsMissingTravellerContextInsteadOfGuessing(t *testing.T) {
	environment := newEnvironment(t)
	response, decoded := environment.call(t, http.MethodPost, "/v1/discover/matches", discoveryBody(t, `"weather":"warm"`), true)
	if response.Code != http.StatusOK {
		t.Fatalf("returned %d %v", response.Code, decoded)
	}
	missing := list(t, decoded, "missingContext")
	if len(missing) == 0 {
		t.Fatal("an incomplete profile reported no missing context")
	}
	for _, recommendation := range list(t, decoded, "recommendations") {
		value, ok := recommendation.(map[string]any)
		if !ok {
			t.Fatal("recommendation is not an object")
		}
		access, ok := value["access"].(map[string]any)
		if !ok || access["status"] != "unknown" {
			t.Fatalf("access was claimed without traveller context: %v", value["access"])
		}
	}
}

func TestDiscoverRejectsMalformedRequestsBeforeAnyEngineRuns(t *testing.T) {
	environment := newEnvironment(t)
	start := contractNow.AddDate(0, 1, 0).Format(time.DateOnly)
	for name, body := range map[string]string{
		"no budget":       fmt.Sprintf(`{"startDate":%q,"endDate":%q,"adults":2}`, start, start),
		"unknown field":   discoveryBody(t, `"surprise":true`),
		"bad currency":    `{"budget":{"amountMinor":1,"currency":"ngn"},"startDate":"2026-10-12","endDate":"2026-10-19","adults":2}`,
		"zero adults":     discoveryBody(t, `"adults":0`),
		"invented enum":   discoveryBody(t, `"weather":"tropical"`),
		"bad airport":     discoveryBody(t, `"departureAirport":"Lagos"`),
		"oversized query": discoveryBody(t, `"query":"`+strings.Repeat("x", 501)+`"`),
	} {
		response, problem := environment.call(t, http.MethodPost, "/v1/discover/matches", body, true)
		if response.Code != http.StatusBadRequest || problem["code"] != "validation_error" {
			t.Fatalf("%s returned %d %v", name, response.Code, problem)
		}
	}
}

func TestDiscoverRejectsPastAndReversedDateRanges(t *testing.T) {
	environment := newEnvironment(t)
	past := contractNow.AddDate(0, -1, 0).Format(time.DateOnly)
	future := contractNow.AddDate(0, 1, 0).Format(time.DateOnly)
	for name, body := range map[string]string{
		"past departure": fmt.Sprintf(`{"budget":{"amountMinor":500000000,"currency":"NGN"},"startDate":%q,"endDate":%q,"adults":2}`, past, future),
		"reversed":       fmt.Sprintf(`{"budget":{"amountMinor":500000000,"currency":"NGN"},"startDate":%q,"endDate":%q,"adults":2}`, future, past),
	} {
		response, problem := environment.call(t, http.MethodPost, "/v1/discover/matches", body, true)
		if response.Code != http.StatusBadRequest || problem["code"] != "validation_error" {
			t.Fatalf("%s returned %d %v", name, response.Code, problem)
		}
	}
}
