package contract_test

import (
	"net/http"
	"testing"
)

func learnCoastalTaste(t *testing.T, environment *environment) {
	t.Helper()
	for range 2 {
		tripID := createTrip(t, environment)
		body := `{"positives":["coastal_places"],"pace":"just_right"}`
		response, decoded := environment.call(t, http.MethodPut, "/v1/trips/"+tripID+"/feedback", body, true)
		if response.Code != http.StatusOK {
			t.Fatalf("feedback returned %d %v", response.Code, decoded)
		}
	}
}

func firstInsight(t *testing.T, environment *environment) map[string]any {
	t.Helper()
	response, decoded := environment.call(t, http.MethodGet, "/v1/me/personalization/insights", "", true)
	if response.Code != http.StatusOK {
		t.Fatalf("insights returned %d %v", response.Code, decoded)
	}
	insights := list(t, decoded, "insights")
	if len(insights) == 0 {
		t.Fatalf("no insight was learned: %v", decoded)
	}
	value, ok := insights[0].(map[string]any)
	if !ok {
		t.Fatalf("insight is not an object: %v", insights[0])
	}
	return value
}

func TestFeedbackAcrossTripsBecomesAnEditableInsight(t *testing.T) {
	environment := newEnvironment(t)
	learnCoastalTaste(t, environment)
	insight := firstInsight(t, environment)
	if insight["key"] != "interest.coastal" || insight["status"] != "active" {
		t.Fatalf("insight = %v", insight)
	}
	if insight["editable"] != true || insight["evidenceCount"] != float64(2) {
		t.Fatalf("insight = %v", insight)
	}
	if insight["summaryCode"] != "insight.interest.coastal" {
		t.Fatalf("summary code = %v, want a code the client renders", insight["summaryCode"])
	}
}

func TestConfirmingAnInsightPromotesItsOrigin(t *testing.T) {
	environment := newEnvironment(t)
	learnCoastalTaste(t, environment)
	insightID := text(t, firstInsight(t, environment), "id")
	response, updated := environment.call(t, http.MethodPatch,
		"/v1/me/personalization/insights/"+insightID, `{"status":"confirmed"}`, true)
	if response.Code != http.StatusOK || updated["status"] != "confirmed" {
		t.Fatalf("returned %d %v", response.Code, updated)
	}
	_, settings := environment.call(t, http.MethodGet, "/v1/me/personalization", "", true)
	for _, entry := range list(t, settings, "learned") {
		if member(t, entry, "key") == "interest.coastal" {
			if member(t, entry, "origin") != "confirmed_learned" {
				t.Fatalf("origin = %v, want the confirmed origin", entry)
			}
			return
		}
	}
	t.Fatalf("the confirmed preference is missing from %v", settings["learned"])
}

func TestDismissingAnInsightRemovesItsInfluence(t *testing.T) {
	environment := newEnvironment(t)
	learnCoastalTaste(t, environment)
	insightID := text(t, firstInsight(t, environment), "id")
	response, updated := environment.call(t, http.MethodPatch,
		"/v1/me/personalization/insights/"+insightID, `{"status":"dismissed"}`, true)
	if response.Code != http.StatusOK || updated["status"] != "dismissed" {
		t.Fatalf("returned %d %v", response.Code, updated)
	}
	_, settings := environment.call(t, http.MethodGet, "/v1/me/personalization", "", true)
	for _, entry := range list(t, settings, "learned") {
		if member(t, entry, "key") == "interest.coastal" {
			t.Fatalf("a dismissed preference still influences: %v", settings["learned"])
		}
	}
	suppressed := list(t, settings, "suppressed")
	if len(suppressed) == 0 || suppressed[0] != "interest.coastal" {
		t.Fatalf("suppressed = %v", suppressed)
	}
}

func TestTurningOffHistoryLearningSilencesEveryInference(t *testing.T) {
	environment := newEnvironment(t)
	learnCoastalTaste(t, environment)
	response, settings := environment.call(t, http.MethodPatch, "/v1/me/personalization",
		`{"historyLearningEnabled":false}`, true)
	if response.Code != http.StatusOK || settings["historyLearningEnabled"] != false {
		t.Fatalf("returned %d %v", response.Code, settings)
	}
	if len(list(t, settings, "learned")) != 0 {
		t.Fatalf("learning is off but inferences remain: %v", settings["learned"])
	}
	if settings["rulesVersion"] != "personalization-v1" {
		t.Fatalf("rules version = %v", settings["rulesVersion"])
	}
}

func TestClearingLearningLeavesTheTripAndItsFeedbackReadable(t *testing.T) {
	environment := newEnvironment(t)
	tripID := createTrip(t, environment)
	submitted, stored := environment.call(t, http.MethodPut, "/v1/trips/"+tripID+"/feedback",
		`{"positives":["local_food"],"negatives":["long_transfers"],"returnIntent":"would_return"}`, true)
	if submitted.Code != http.StatusOK || stored["returnIntent"] != "would_return" {
		t.Fatalf("returned %d %v", submitted.Code, stored)
	}
	cleared, _ := environment.call(t, http.MethodDelete, "/v1/me/personalization", "", true)
	if cleared.Code != http.StatusNoContent {
		t.Fatalf("clear returned %d", cleared.Code)
	}
	response, decoded := environment.call(t, http.MethodGet, "/v1/me/personalization/insights", "", true)
	if response.Code != http.StatusOK || len(list(t, decoded, "insights")) != 0 {
		t.Fatalf("insights survived the clear: %d %v", response.Code, decoded)
	}
	readback, feedback := environment.call(t, http.MethodGet, "/v1/trips/"+tripID+"/feedback", "", true)
	if readback.Code != http.StatusOK || feedback["tripId"] != tripID {
		t.Fatalf("returned %d %v", readback.Code, feedback)
	}
}

func TestFeedbackRejectsMalformedSubmissions(t *testing.T) {
	environment := newEnvironment(t)
	tripID := createTrip(t, environment)
	for name, body := range map[string]string{
		"nothing said":    `{}`,
		"invented aspect": `{"positives":["shopping"]}`,
		"both ways":       `{"positives":["local_food"],"negatives":["local_food"]}`,
		"unknown field":   `{"positives":["local_food"],"mood":"good"}`,
		"invented intent": `{"returnIntent":"maybe"}`,
	} {
		response, problem := environment.call(t, http.MethodPut, "/v1/trips/"+tripID+"/feedback", body, true)
		if response.Code != http.StatusBadRequest || problem["code"] != "validation_error" {
			t.Fatalf("%s returned %d %v", name, response.Code, problem)
		}
	}
}

func TestPersonalizationRequiresAuthentication(t *testing.T) {
	environment := newEnvironment(t)
	for _, path := range []string{"/v1/me/personalization", "/v1/me/personalization/insights"} {
		response, problem := environment.call(t, http.MethodGet, path, "", false)
		if response.Code != http.StatusUnauthorized || problem["code"] != "authentication_required" {
			t.Fatalf("%s returned %d %v", path, response.Code, problem)
		}
	}
}
