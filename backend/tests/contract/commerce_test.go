package contract_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

func flightSearchBody(t *testing.T) string {
	t.Helper()
	return fmt.Sprintf(`{"origin":"LIS","destination":"FNC","departDate":%q,"budget":{"amountMinor":30000,"currency":"EUR"}}`,
		contractNow.AddDate(0, 1, 0).Format(time.DateOnly))
}

func TestFlightSearchReturnsRankedOffersWithDisclosure(t *testing.T) {
	environment := newEnvironment(t)
	tripID := createTrip(t, environment)
	response, decoded := environment.call(t, http.MethodPost,
		"/v1/trips/"+tripID+"/flights/search", flightSearchBody(t), true)
	if response.Code != http.StatusOK {
		t.Fatalf("returned %d %v", response.Code, decoded)
	}
	offers := list(t, decoded, "offers")
	if len(offers) != 2 {
		t.Fatalf("offers = %v", offers)
	}
	top, ok := offers[0].(map[string]any)
	if !ok || top["id"] != "fast" {
		t.Fatalf("the cheap slow flight with high commission won: %v", offers[0])
	}
	if top["rank"] != float64(1) || top["freshness"] != "live" {
		t.Fatalf("top offer = %v", top)
	}
	if len(list(t, top, "reasons")) == 0 {
		t.Fatal("the top offer carries no reasons")
	}
	assertNoCommissionLeak(t, top)
	disclosure := object(t, decoded, "disclosure")
	if disclosure["code"] != "affiliate.eligible_booking" || disclosure["partner"] == "" {
		t.Fatalf("disclosure = %v", disclosure)
	}
	status := object(t, decoded, "status")
	if status["rankingVersion"] != "flight-rank-v1" {
		t.Fatalf("ranking version = %v", status["rankingVersion"])
	}
	if len(list(t, status, "partners")) == 0 {
		t.Fatal("partner status was omitted")
	}
}

func assertNoCommissionLeak(t *testing.T, offer map[string]any) {
	t.Helper()
	for _, forbidden := range []string{"commission", "commissionMinor", "payout", "tracking", "commercials"} {
		if _, present := offer[forbidden]; present {
			t.Fatalf("the offer leaked %q to the client: %v", forbidden, offer)
		}
	}
}

func TestStaySearchPrefersLocationOverPrice(t *testing.T) {
	environment := newEnvironment(t)
	tripID := createTrip(t, environment)
	body := fmt.Sprintf(`{"checkIn":%q,"checkOut":%q,"budget":{"amountMinor":100000,"currency":"EUR"}}`,
		contractNow.AddDate(0, 1, 0).Format(time.DateOnly), contractNow.AddDate(0, 1, 4).Format(time.DateOnly))
	response, decoded := environment.call(t, http.MethodPost, "/v1/trips/"+tripID+"/stays/search", body, true)
	if response.Code != http.StatusOK {
		t.Fatalf("returned %d %v", response.Code, decoded)
	}
	offers := list(t, decoded, "offers")
	top, ok := offers[0].(map[string]any)
	if !ok || top["id"] != "near" {
		t.Fatalf("the cheaper distant stay won: %v", offers)
	}
	assertNoCommissionLeak(t, top)
}

func TestActivitySearchReturnsAnEmptyResultRatherThanFailing(t *testing.T) {
	environment := newEnvironment(t)
	tripID := createTrip(t, environment)
	response, decoded := environment.call(t, http.MethodPost, "/v1/trips/"+tripID+"/activities/search", "{}", true)
	if response.Code != http.StatusOK {
		t.Fatalf("returned %d %v", response.Code, decoded)
	}
	if len(list(t, decoded, "offers")) != 0 {
		t.Fatal("offers appeared with no configured activity partner")
	}
	if object(t, decoded, "status")["rankingVersion"] == "" {
		t.Fatal("the ranking version was omitted")
	}
}

func TestCommerceSearchesRefuseAnonymousCallersAndMalformedBodies(t *testing.T) {
	environment := newEnvironment(t)
	tripID := createTrip(t, environment)
	response, problem := environment.call(t, http.MethodPost,
		"/v1/trips/"+tripID+"/flights/search", flightSearchBody(t), false)
	if response.Code != http.StatusUnauthorized || problem["code"] != "authentication_required" {
		t.Fatalf("returned %d %v", response.Code, problem)
	}
	for name, body := range map[string]string{
		"no route":       `{"departDate":"2026-11-16"}`,
		"lowercase code": `{"origin":"lis","destination":"FNC","departDate":"2026-11-16"}`,
		"unknown field":  `{"origin":"LIS","destination":"FNC","departDate":"2026-11-16","surprise":true}`,
		"invented cabin": `{"origin":"LIS","destination":"FNC","departDate":"2026-11-16","cabin":"spacious"}`,
	} {
		bad, badProblem := environment.call(t, http.MethodPost, "/v1/trips/"+tripID+"/flights/search", body, true)
		if bad.Code != http.StatusBadRequest || badProblem["code"] != "validation_error" {
			t.Fatalf("%s returned %d %v", name, bad.Code, badProblem)
		}
	}
}

func TestHandoffReturnsAnApprovedPartnerDestination(t *testing.T) {
	environment := newEnvironment(t)
	tripID := createTrip(t, environment)
	if response, decoded := environment.call(t, http.MethodPost,
		"/v1/trips/"+tripID+"/flights/search", flightSearchBody(t), true); response.Code != http.StatusOK {
		t.Fatalf("search returned %d %v", response.Code, decoded)
	}
	body := fmt.Sprintf(`{"offerId":"fast","tripId":%q}`, tripID)
	response, decoded := environment.call(t, http.MethodPost, "/v1/commerce/handoffs", body, true)
	if response.Code != http.StatusCreated {
		t.Fatalf("returned %d %v", response.Code, decoded)
	}
	if text(t, decoded, "destinationUrl") != "https://partner.invalid/book/fast" {
		t.Fatalf("destination = %v", decoded["destinationUrl"])
	}
	disclosure := object(t, decoded, "disclosure")
	if disclosure["code"] != "affiliate.eligible_booking" {
		t.Fatalf("disclosure = %v", disclosure)
	}
	if decoded["expiresAt"] == nil || decoded["price"] == nil {
		t.Fatalf("handoff = %v", decoded)
	}
	assertNoCommissionLeak(t, decoded)
	if _, present := decoded["attributionToken"]; present {
		t.Fatal("the attribution token was exposed to the client")
	}
}

func TestHandoffForAnUnknownOfferIsRefused(t *testing.T) {
	environment := newEnvironment(t)
	response, problem := environment.call(t, http.MethodPost, "/v1/commerce/handoffs",
		`{"offerId":"never-searched"}`, true)
	if response.Code != http.StatusConflict || problem["code"] != "invalid_state" {
		t.Fatalf("returned %d %v", response.Code, problem)
	}
	unauthenticated, unauthProblem := environment.call(t, http.MethodPost, "/v1/commerce/handoffs",
		`{"offerId":"fast"}`, false)
	if unauthenticated.Code != http.StatusUnauthorized || unauthProblem["code"] != "authentication_required" {
		t.Fatalf("returned %d %v", unauthenticated.Code, unauthProblem)
	}
}

func TestPriceWatchLifecycle(t *testing.T) {
	environment := newEnvironment(t)
	tripID := createTrip(t, environment)
	body := fmt.Sprintf(`{"tripId":%q,"product":"flight","baseline":{"amountMinor":40000,"currency":"EUR"},
	  "target":{"origin":"LIS","destination":"FNC","startDate":%q,"endDate":%q,"adults":2}}`,
		tripID, contractNow.AddDate(0, 1, 0).Format(time.DateOnly), contractNow.AddDate(0, 1, 6).Format(time.DateOnly))
	response, decoded := environment.call(t, http.MethodPost, "/v1/me/price-watches", body, true)
	if response.Code != http.StatusCreated {
		t.Fatalf("returned %d %v", response.Code, decoded)
	}
	if decoded["status"] != "active" || decoded["nextCheckAt"] == nil {
		t.Fatalf("watch = %v", decoded)
	}
	watchID := text(t, decoded, "id")

	listed, watches := environment.call(t, http.MethodGet, "/v1/me/price-watches", "", true)
	if listed.Code != http.StatusOK || len(list(t, watches, "watches")) != 1 {
		t.Fatalf("returned %d %v", listed.Code, watches)
	}
	deleted, _ := environment.call(t, http.MethodDelete, "/v1/me/price-watches/"+watchID, "", true)
	if deleted.Code != http.StatusNoContent {
		t.Fatalf("delete returned %d", deleted.Code)
	}
}

func TestPriceWatchRejectsMalformedRequests(t *testing.T) {
	environment := newEnvironment(t)
	for name, body := range map[string]string{
		"invented product": `{"product":"yacht","baseline":{"amountMinor":1,"currency":"EUR"},"target":{"startDate":"2026-11-16","endDate":"2026-11-20","adults":1}}`,
		"no baseline":      `{"product":"flight","target":{"startDate":"2026-11-16","endDate":"2026-11-20","adults":1}}`,
		"no travellers":    `{"product":"flight","baseline":{"amountMinor":1,"currency":"EUR"},"target":{"startDate":"2026-11-16","endDate":"2026-11-20","adults":0}}`,
		"unknown field":    `{"product":"flight","baseline":{"amountMinor":1,"currency":"EUR"},"target":{"startDate":"2026-11-16","endDate":"2026-11-20","adults":1},"surprise":true}`,
	} {
		response, problem := environment.call(t, http.MethodPost, "/v1/me/price-watches", body, true)
		if response.Code != http.StatusBadRequest || problem["code"] != "validation_error" {
			t.Fatalf("%s returned %d %v", name, response.Code, problem)
		}
	}
}

func TestPriceWatchesRequireAuthentication(t *testing.T) {
	environment := newEnvironment(t)
	response, problem := environment.call(t, http.MethodGet, "/v1/me/price-watches", "", false)
	if response.Code != http.StatusUnauthorized || problem["code"] != "authentication_required" {
		t.Fatalf("returned %d %v", response.Code, problem)
	}
}
