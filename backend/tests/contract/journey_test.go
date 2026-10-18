package contract_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

func createTrip(t *testing.T, environment *environment) string {
	t.Helper()
	start := contractNow.AddDate(0, 0, 3).Format(time.DateOnly)
	end := contractNow.AddDate(0, 0, 9).Format(time.DateOnly)
	body := fmt.Sprintf(`{"destinationId":"madeira-pt","originAirport":"LIS","adults":2,`+
		`"preferredCurrency":"EUR","startDate":%q,"endDate":%q}`, start, end)
	response, decoded := environment.call(t, http.MethodPost, "/v1/trips", body, true)
	if response.Code != http.StatusCreated {
		t.Fatalf("trip creation returned %d %v", response.Code, decoded)
	}
	return text(t, decoded, "id")
}

func TestDestinationDetailReturnsEverySectionWithFreshness(t *testing.T) {
	environment := newEnvironment(t)
	response, decoded := environment.call(t, http.MethodGet, "/v1/destinations/madeira-pt", "", true)
	if response.Code != http.StatusOK {
		t.Fatalf("returned %d %v", response.Code, decoded)
	}
	if member(t, decoded["destination"], "id") != "madeira-pt" {
		t.Fatalf("destination = %v", decoded["destination"])
	}
	for _, name := range []string{"weatherSection", "placesSection", "mediaSection"} {
		section := object(t, decoded, name)
		if section["status"] != "available" {
			t.Fatalf("%s = %v", name, section)
		}
		freshness := object(t, section, "freshness")
		if freshness["observedAt"] == nil || freshness["status"] != "live" {
			t.Fatalf("%s freshness = %v", name, freshness)
		}
	}
	assertDestinationContent(t, decoded)
}

func assertDestinationContent(t *testing.T, decoded map[string]any) {
	t.Helper()
	weather := object(t, decoded, "weather")
	if weather["availability"] != "available" || weather["current"] == nil {
		t.Fatalf("weather = %v", weather)
	}
	places := list(t, decoded, "thingsToDo")
	if len(places) == 0 {
		t.Fatal("no places returned")
	}
	first, ok := places[0].(map[string]any)
	if !ok || first["openingStatus"] != "open" || first["category"] == nil {
		t.Fatalf("place = %v", places[0])
	}
	assertCreditedMedia(t, list(t, decoded, "media"))
}

func assertCreditedMedia(t *testing.T, media []any) {
	t.Helper()
	if len(media) == 0 {
		t.Fatal("no media returned")
	}
	asset, ok := media[0].(map[string]any)
	if !ok {
		t.Fatal("media is not an object")
	}
	attribution := list(t, asset, "attribution")
	if len(attribution) == 0 {
		t.Fatal("media attribution was stripped")
	}
	credit, ok := attribution[0].(map[string]any)
	if !ok || credit["text"] == "" || credit["provider"] == "" {
		t.Fatalf("attribution = %v", attribution[0])
	}
}

func TestDestinationMediaEndpointReturnsTheSameAssets(t *testing.T) {
	environment := newEnvironment(t)
	response, decoded := environment.call(t, http.MethodGet, "/v1/destinations/madeira-pt/media", "", true)
	if response.Code != http.StatusOK {
		t.Fatalf("returned %d %v", response.Code, decoded)
	}
	if text(t, decoded, "destinationId") != "madeira-pt" {
		t.Fatalf("destination id = %v", decoded)
	}
	if len(list(t, decoded, "media")) == 0 {
		t.Fatal("no media returned")
	}
}

func TestUnknownDestinationIsNotFound(t *testing.T) {
	environment := newEnvironment(t)
	response, problem := environment.call(t, http.MethodGet, "/v1/destinations/atlantis", "", true)
	if response.Code != http.StatusNotFound || problem["code"] != "not_found" {
		t.Fatalf("returned %d %v", response.Code, problem)
	}
}

func TestItineraryGenerationReadAndEdit(t *testing.T) {
	environment := newEnvironment(t)
	tripID := createTrip(t, environment)
	response, decoded := environment.call(t, http.MethodPost, "/v1/trips/"+tripID+"/itinerary/generate", `{"pace":"balanced"}`, true)
	if response.Code != http.StatusOK {
		t.Fatalf("generation returned %d %v", response.Code, decoded)
	}
	plan := object(t, decoded, "itinerary")
	if plan["algorithmVersion"] == "" || plan["timezone"] != "Atlantic/Madeira" {
		t.Fatalf("itinerary = %v", plan)
	}
	if len(list(t, decoded, "violations")) != 0 {
		t.Fatalf("the generated plan is invalid: %v", decoded["violations"])
	}
	days := list(t, plan, "days")
	itemID := firstItemID(t, days)

	readResponse, stored := environment.call(t, http.MethodGet, "/v1/trips/"+tripID+"/itinerary", "", true)
	if readResponse.Code != http.StatusOK {
		t.Fatalf("read returned %d %v", readResponse.Code, stored)
	}
	if len(list(t, stored, "days")) != len(days) {
		t.Fatal("the stored itinerary differs from the generated one")
	}

	patchResponse, patched := environment.call(t, http.MethodPatch,
		"/v1/trips/"+tripID+"/itinerary/items/"+itemID, `{"status":"reserved","userLocked":true}`, true)
	if patchResponse.Code != http.StatusOK {
		t.Fatalf("patch returned %d %v", patchResponse.Code, patched)
	}
	if patched["status"] != "reserved" || patched["userLocked"] != true {
		t.Fatalf("patched item = %v", patched)
	}
}

func TestItineraryEndpointsRefuseStrangersAndAnonymousCallers(t *testing.T) {
	environment := newEnvironment(t)
	tripID := createTrip(t, environment)
	response, problem := environment.call(t, http.MethodPost, "/v1/trips/"+tripID+"/itinerary/generate", "{}", false)
	if response.Code != http.StatusUnauthorized || problem["code"] != "authentication_required" {
		t.Fatalf("anonymous generation returned %d %v", response.Code, problem)
	}
	readResponse, readProblem := environment.call(t, http.MethodGet, "/v1/trips/"+tripID+"/itinerary", "", false)
	if readResponse.Code != http.StatusUnauthorized || readProblem["code"] != "authentication_required" {
		t.Fatalf("anonymous read returned %d %v", readResponse.Code, readProblem)
	}
}

func TestItineraryReadBeforeGenerationIsNotFound(t *testing.T) {
	environment := newEnvironment(t)
	tripID := createTrip(t, environment)
	response, problem := environment.call(t, http.MethodGet, "/v1/trips/"+tripID+"/itinerary", "", true)
	if response.Code != http.StatusNotFound || problem["code"] != "not_found" {
		t.Fatalf("returned %d %v", response.Code, problem)
	}
}

func TestBudgetIsReturnedInTheTripCurrency(t *testing.T) {
	environment := newEnvironment(t)
	tripID := createTrip(t, environment)
	response, decoded := environment.call(t, http.MethodGet, "/v1/trips/"+tripID+"/budget", "", true)
	if response.Code != http.StatusOK {
		t.Fatalf("returned %d %v", response.Code, decoded)
	}
	if text(t, decoded, "currency") != "EUR" {
		t.Fatalf("currency = %v", decoded["currency"])
	}
	categories := list(t, decoded, "categories")
	if len(categories) != 2 {
		t.Fatalf("categories = %v", categories)
	}
	stay, ok := categories[0].(map[string]any)
	if !ok || stay["category"] != "stay" {
		t.Fatalf("first category = %v", categories[0])
	}
	estimated := object(t, stay, "estimated")
	expected := object(t, estimated, "expected")
	if expected["currency"] != "EUR" || expected["amountMinor"] == nil {
		t.Fatalf("estimate = %v", estimated)
	}
	total := object(t, decoded, "estimatedTotal")
	if object(t, total, "expected")["currency"] != "EUR" {
		t.Fatal("the total is not in the trip currency")
	}
	if decoded["confidence"] != "low" {
		t.Fatalf("confidence = %v, want the weakest line", decoded["confidence"])
	}
}

func TestNearbySearchRanksForTheTimeAvailable(t *testing.T) {
	environment := newEnvironment(t)
	tripID := createTrip(t, environment)
	body := `{"location":{"latitude":32.6478,"longitude":-16.9048},"availableMinutes":60,"radiusMeters":2000}`
	response, decoded := environment.call(t, http.MethodPost, "/v1/trips/"+tripID+"/nearby/search", body, true)
	if response.Code != http.StatusOK {
		t.Fatalf("returned %d %v", response.Code, decoded)
	}
	ranked := list(t, decoded, "places")
	if len(ranked) == 0 {
		t.Fatal("no places returned")
	}
	first, ok := ranked[0].(map[string]any)
	if !ok {
		t.Fatal("ranked entry is not an object")
	}
	if first["fitsAvailableTime"] != true {
		t.Fatalf("the top result does not fit the time available: %v", first)
	}
	if len(list(t, first, "reasons")) == 0 {
		t.Fatal("no ranking reasons returned")
	}
	if object(t, first, "place")["id"] == nil {
		t.Fatal("the ranked place is missing")
	}
	fitting := 0
	for _, entry := range ranked {
		value, ok := entry.(map[string]any)
		if ok && value["fitsAvailableTime"] == true {
			fitting++
		}
	}
	if fitting == len(ranked) {
		t.Fatal("with one hour free, every place was said to fit")
	}
}

func TestNearbySearchRejectsMalformedRequests(t *testing.T) {
	environment := newEnvironment(t)
	tripID := createTrip(t, environment)
	for name, body := range map[string]string{
		"no location":   `{"availableMinutes":60}`,
		"bad latitude":  `{"location":{"latitude":200,"longitude":0}}`,
		"unknown field": `{"location":{"latitude":32.6,"longitude":-16.9},"surprise":true}`,
		"invented type": `{"location":{"latitude":32.6,"longitude":-16.9},"categories":["spaceport"]}`,
		"absurd radius": `{"location":{"latitude":32.6,"longitude":-16.9},"radiusMeters":900000}`,
	} {
		response, problem := environment.call(t, http.MethodPost, "/v1/trips/"+tripID+"/nearby/search", body, true)
		if response.Code != http.StatusBadRequest || problem["code"] != "validation_error" {
			t.Fatalf("%s returned %d %v", name, response.Code, problem)
		}
	}
}

func TestItineraryPatchRejectsInventedStatuses(t *testing.T) {
	environment := newEnvironment(t)
	tripID := createTrip(t, environment)
	if response, decoded := environment.call(t, http.MethodPost,
		"/v1/trips/"+tripID+"/itinerary/generate", "{}", true); response.Code != http.StatusOK {
		t.Fatalf("generation returned %d %v", response.Code, decoded)
	}
	itemID := "1c2f21b0-4a55-4b39-9f0b-2c7f9c1a1a99"
	response, problem := environment.call(t, http.MethodPatch,
		"/v1/trips/"+tripID+"/itinerary/items/"+itemID, `{"status":"teleported"}`, true)
	if response.Code != http.StatusBadRequest || problem["code"] != "validation_error" {
		t.Fatalf("returned %d %v", response.Code, problem)
	}
}

func firstItemID(t *testing.T, days []any) string {
	t.Helper()
	if len(days) == 0 {
		t.Fatal("no days generated")
	}
	firstDay, ok := days[0].(map[string]any)
	if !ok {
		t.Fatal("day is not an object")
	}
	items := list(t, firstDay, "items")
	if len(items) == 0 {
		t.Fatal("no items generated")
	}
	item, ok := items[0].(map[string]any)
	if !ok {
		t.Fatal("item is not an object")
	}
	if item["placeId"] == nil || item["status"] != "planned" || item["source"] != "generated" {
		t.Fatalf("item = %v", item)
	}
	identifier, ok := item["id"].(string)
	if !ok || identifier == "" {
		t.Fatal("item has no identifier")
	}
	return identifier
}
