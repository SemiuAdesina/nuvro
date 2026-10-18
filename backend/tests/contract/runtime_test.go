package contract_test

import (
	"fmt"
	"net/http"
	"testing"
)

func activeTrip(t *testing.T, environment *environment) string {
	t.Helper()
	tripID := createTrip(t, environment)
	response, decoded := environment.call(t, http.MethodPost, "/v1/trips/"+tripID+"/itinerary/generate", "{}", true)
	if response.Code != http.StatusOK {
		t.Fatalf("generation returned %d %v", response.Code, decoded)
	}
	return tripID
}

func TestActiveTripSnapshotIsInTheDestinationTimezone(t *testing.T) {
	environment := newEnvironment(t)
	tripID := activeTrip(t, environment)
	path := fmt.Sprintf("/v1/trips/%s/active?latitude=32.6478&longitude=-16.9048", tripID)
	response, decoded := environment.call(t, http.MethodGet, path, "", true)
	if response.Code != http.StatusOK {
		t.Fatalf("returned %d %v", response.Code, decoded)
	}
	if text(t, decoded, "timezone") != "Atlantic/Madeira" {
		t.Fatalf("timezone = %v", decoded["timezone"])
	}
	if decoded["phase"] == nil || decoded["currentDay"] == nil || decoded["generatedAt"] == nil {
		t.Fatalf("snapshot = %v", decoded)
	}
	departure := object(t, decoded, "departure")
	if departure["status"] == nil {
		t.Fatalf("departure = %v", departure)
	}
	assertQuickActions(t, list(t, decoded, "actions"))
	if _, present := decoded["alerts"]; !present {
		t.Fatal("alerts were omitted")
	}
}

func assertQuickActions(t *testing.T, actions []any) {
	t.Helper()
	if len(actions) == 0 {
		t.Fatal("no quick actions returned")
	}
	for _, entry := range actions {
		action, ok := entry.(map[string]any)
		if !ok || action["type"] == nil || action["enabled"] == nil {
			t.Fatalf("action = %v", entry)
		}
	}
}

func TestActiveTripRequiresAuthentication(t *testing.T) {
	environment := newEnvironment(t)
	tripID := activeTrip(t, environment)
	response, problem := environment.call(t, http.MethodGet, "/v1/trips/"+tripID+"/active", "", false)
	if response.Code != http.StatusUnauthorized || problem["code"] != "authentication_required" {
		t.Fatalf("returned %d %v", response.Code, problem)
	}
}

func TestActiveTripRejectsImpossibleCoordinates(t *testing.T) {
	environment := newEnvironment(t)
	tripID := activeTrip(t, environment)
	path := fmt.Sprintf("/v1/trips/%s/active?latitude=200&longitude=0", tripID)
	response, problem := environment.call(t, http.MethodGet, path, "", true)
	if response.Code != http.StatusBadRequest || problem["code"] != "validation_error" {
		t.Fatalf("returned %d %v", response.Code, problem)
	}
}

func TestReplanProposesThenAppliesOnlyWhenAsked(t *testing.T) {
	environment := newEnvironment(t)
	tripID := activeTrip(t, environment)
	response, proposal := environment.call(t, http.MethodPost, "/v1/trips/"+tripID+"/replan",
		`{"trigger":"weather"}`, true)
	if response.Code != http.StatusOK {
		t.Fatalf("propose returned %d %v", response.Code, proposal)
	}
	if proposal["itineraryVersion"] == nil || proposal["expiresAt"] == nil {
		t.Fatalf("proposal = %v", proposal)
	}
	assertProposalShape(t, proposal)
	before, stored := environment.call(t, http.MethodGet, "/v1/trips/"+tripID+"/itinerary", "", true)
	if before.Code != http.StatusOK {
		t.Fatalf("read returned %d %v", before.Code, stored)
	}
	version, ok := stored["version"].(float64)
	if !ok {
		t.Fatalf("version = %v", stored["version"])
	}
	proposalID := text(t, proposal, "id")
	applied, updated := environment.call(t, http.MethodPost,
		"/v1/trips/"+tripID+"/replan/"+proposalID+"/apply", "", true)
	if applied.Code != http.StatusOK {
		t.Fatalf("apply returned %d %v", applied.Code, updated)
	}
	newVersion, ok := updated["version"].(float64)
	if !ok || newVersion <= version {
		t.Fatalf("version did not advance: %v to %v", version, updated["version"])
	}
	repeat, problem := environment.call(t, http.MethodPost,
		"/v1/trips/"+tripID+"/replan/"+proposalID+"/apply", "", true)
	if repeat.Code != http.StatusConflict || problem["code"] != "conflict" {
		t.Fatalf("reapplying returned %d %v", repeat.Code, problem)
	}
}

func TestReplanRejectsInventedTriggers(t *testing.T) {
	environment := newEnvironment(t)
	tripID := activeTrip(t, environment)
	for name, body := range map[string]string{
		"invented trigger":  `{"trigger":"vibes"}`,
		"no trigger":        `{}`,
		"unknown field":     `{"trigger":"weather","surprise":true}`,
		"oversized message": `{"trigger":"weather","message":"` + longMessage() + `"}`,
	} {
		response, problem := environment.call(t, http.MethodPost, "/v1/trips/"+tripID+"/replan", body, true)
		if response.Code != http.StatusBadRequest || problem["code"] != "validation_error" {
			t.Fatalf("%s returned %d %v", name, response.Code, problem)
		}
	}
}

func TestAlertsListAndStatusLifecycle(t *testing.T) {
	environment := newEnvironment(t)
	tripID := activeTrip(t, environment)
	response, decoded := environment.call(t, http.MethodGet, "/v1/trips/"+tripID+"/alerts", "", true)
	if response.Code != http.StatusOK {
		t.Fatalf("returned %d %v", response.Code, decoded)
	}
	if _, present := decoded["alerts"]; !present {
		t.Fatalf("alerts key missing: %v", decoded)
	}
	missing := "0c2f21b0-4a55-4b39-9f0b-2c7f9c1a1a99"
	patch, problem := environment.call(t, http.MethodPatch,
		"/v1/trips/"+tripID+"/alerts/"+missing, `{"status":"read"}`, true)
	if patch.Code != http.StatusNotFound || problem["code"] != "not_found" {
		t.Fatalf("returned %d %v", patch.Code, problem)
	}
	invalid, invalidProblem := environment.call(t, http.MethodPatch,
		"/v1/trips/"+tripID+"/alerts/"+missing, `{"status":"expired"}`, true)
	if invalid.Code != http.StatusBadRequest || invalidProblem["code"] != "validation_error" {
		t.Fatalf("a client-set expiry returned %d %v", invalid.Code, invalidProblem)
	}
}

func TestTranslationReturnsTheLanguagePair(t *testing.T) {
	environment := newEnvironment(t)
	tripID := activeTrip(t, environment)
	response, decoded := environment.call(t, http.MethodPost, "/v1/trips/"+tripID+"/translate",
		`{"text":"Where is the market?","targetLanguage":"pt-PT"}`, true)
	if response.Code != http.StatusOK {
		t.Fatalf("returned %d %v", response.Code, decoded)
	}
	if text(t, decoded, "targetLanguage") != "pt-PT" {
		t.Fatalf("target = %v", decoded["targetLanguage"])
	}
	if text(t, decoded, "sourceText") != "Where is the market?" {
		t.Fatalf("source text = %v", decoded["sourceText"])
	}
	if decoded["translatedText"] == "" || decoded["confidence"] == nil {
		t.Fatalf("translation = %v", decoded)
	}
}

func TestTranslationRejectsEmptyAndOversizedText(t *testing.T) {
	environment := newEnvironment(t)
	tripID := activeTrip(t, environment)
	oversized := make([]byte, 5001)
	for index := range oversized {
		oversized[index] = 'a'
	}
	for name, body := range map[string]string{
		"empty":         `{"text":""}`,
		"missing text":  `{"targetLanguage":"pt-PT"}`,
		"oversized":     `{"text":"` + string(oversized) + `"}`,
		"unknown field": `{"text":"Hello","surprise":true}`,
	} {
		response, problem := environment.call(t, http.MethodPost, "/v1/trips/"+tripID+"/translate", body, true)
		if response.Code != http.StatusBadRequest || problem["code"] != "validation_error" {
			t.Fatalf("%s returned %d %v", name, response.Code, problem)
		}
	}
}

func TestPhrasesAreScopedToTheRequestedCategory(t *testing.T) {
	environment := newEnvironment(t)
	tripID := activeTrip(t, environment)
	response, decoded := environment.call(t, http.MethodGet,
		"/v1/trips/"+tripID+"/phrases?category=emergency", "", true)
	if response.Code != http.StatusOK {
		t.Fatalf("returned %d %v", response.Code, decoded)
	}
	if text(t, decoded, "targetLanguage") != "pt-PT" {
		t.Fatalf("target = %v", decoded["targetLanguage"])
	}
	phrases := list(t, decoded, "phrases")
	if len(phrases) == 0 {
		t.Fatal("no phrases returned")
	}
	for _, entry := range phrases {
		phrase, ok := entry.(map[string]any)
		if !ok || phrase["category"] != "emergency" {
			t.Fatalf("phrase = %v", entry)
		}
		if phrase["key"] == "" || phrase["sourceText"] == "" {
			t.Fatalf("phrase is incomplete: %v", phrase)
		}
	}
	invalid, problem := environment.call(t, http.MethodGet,
		"/v1/trips/"+tripID+"/phrases?category=smalltalk", "", true)
	if invalid.Code != http.StatusBadRequest || problem["code"] != "validation_error" {
		t.Fatalf("an invented category returned %d %v", invalid.Code, problem)
	}
}

func TestOfflinePackIsExplicitlyNotLive(t *testing.T) {
	environment := newEnvironment(t)
	tripID := activeTrip(t, environment)
	response, decoded := environment.call(t, http.MethodGet, "/v1/trips/"+tripID+"/offline-pack", "", true)
	if response.Code != http.StatusOK {
		t.Fatalf("returned %d %v", response.Code, decoded)
	}
	if text(t, decoded, "version") == "" || decoded["expiresAt"] == nil {
		t.Fatalf("pack = %v", decoded)
	}
	freshness := object(t, decoded, "freshness")
	if freshness["status"] != "cached" {
		t.Fatalf("an offline pack claimed live data: %v", freshness)
	}
	if len(list(t, decoded, "excluded")) == 0 {
		t.Fatal("no licensing exclusions were declared")
	}
	if len(list(t, decoded, "days")) == 0 {
		t.Fatal("the itinerary was not included")
	}
	if len(list(t, decoded, "phrases")) == 0 {
		t.Fatal("essential phrases were not included")
	}
}

func assertProposalShape(t *testing.T, proposal map[string]any) {
	t.Helper()
	changes := list(t, proposal, "changes")
	if len(changes) == 0 {
		t.Fatal("no changes proposed")
	}
	change, ok := changes[0].(map[string]any)
	if !ok || change["kind"] == nil {
		t.Fatalf("change = %v", changes[0])
	}
	if len(list(t, change, "reasons")) == 0 {
		t.Fatal("a change carried no reason")
	}
	if len(list(t, proposal, "preservedConstraints")) == 0 {
		t.Fatal("no preserved constraints were reported")
	}
}

func longMessage() string {
	value := make([]byte, 501)
	for index := range value {
		value[index] = 'x'
	}
	return string(value)
}
