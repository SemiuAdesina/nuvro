package contract_test

import (
	"net/http"
	"testing"
)

func TestATravellerGoesFromPreferencesToAClosedAccount(t *testing.T) {
	environment := newEnvironment(t)

	setPreferences(t, environment)
	tripID := createTrip(t, environment)
	planTheTrip(t, environment, tripID)
	leaveFeedback(t, environment, tripID)
	assertSomethingWasLearned(t, environment)
	closeTheAccount(t, environment)
}

func setPreferences(t *testing.T, environment *environment) {
	t.Helper()
	body := `{"interests":["food","nature"],"tripPace":"relaxed","weatherPreference":"warm"}`
	response, decoded := environment.call(t, http.MethodPatch, "/v1/me/preferences", body, true)
	if response.Code != http.StatusOK {
		t.Fatalf("preferences returned %d %v", response.Code, decoded)
	}
	if list(t, decoded, "interests") == nil {
		t.Fatalf("preferences did not come back: %v", decoded)
	}
}

func planTheTrip(t *testing.T, environment *environment, tripID string) {
	t.Helper()
	response, decoded := environment.call(t, http.MethodGet, "/v1/trips/"+tripID, "", true)
	if response.Code != http.StatusOK {
		t.Fatalf("reading the trip returned %d %v", response.Code, decoded)
	}
	if decoded["id"] != tripID {
		t.Fatalf("trip = %v, want the one just created", decoded)
	}
	listed, body := environment.call(t, http.MethodGet, "/v1/trips", "", true)
	if listed.Code != http.StatusOK || len(list(t, body, "items")) == 0 {
		t.Fatalf("the trip is missing from the library: %d %v", listed.Code, body)
	}
}

func leaveFeedback(t *testing.T, environment *environment, tripID string) {
	t.Helper()
	learnCoastalTaste(t, environment)
	body := `{"positives":["local_food"],"pace":"just_right","returnIntent":"would_return"}`
	response, decoded := environment.call(t, http.MethodPut, "/v1/trips/"+tripID+"/feedback", body, true)
	if response.Code != http.StatusOK {
		t.Fatalf("feedback returned %d %v", response.Code, decoded)
	}
}

func assertSomethingWasLearned(t *testing.T, environment *environment) {
	t.Helper()
	response, decoded := environment.call(t, http.MethodGet, "/v1/me/personalization", "", true)
	if response.Code != http.StatusOK {
		t.Fatalf("personalization returned %d %v", response.Code, decoded)
	}
	if decoded["historyLearningEnabled"] != true {
		t.Fatalf("history learning = %v, want it on by default", decoded["historyLearningEnabled"])
	}
	insight := firstInsight(t, environment)
	if insight["key"] == "" {
		t.Fatalf("nothing was learned across the journey: %v", insight)
	}
}

func closeTheAccount(t *testing.T, environment *environment) {
	t.Helper()
	response, decoded := environment.call(t, http.MethodDelete, "/v1/me", `{"reason":"no_longer_travelling"}`, true)
	if response.Code != http.StatusAccepted {
		t.Fatalf("closing the account returned %d %v", response.Code, decoded)
	}
	if decoded["status"] != "pending" || decoded["purgeAfter"] == nil {
		t.Fatalf("deletion = %v, want a pending record with a schedule", decoded)
	}
	if len(environment.store.closed) != 1 {
		t.Fatalf("closed = %v, want exactly the traveller who asked", environment.store.closed)
	}
}

func TestTurningOffLearningStopsInferenceForTheWholeJourney(t *testing.T) {
	environment := newEnvironment(t)
	off := `{"historyLearningEnabled":false}`
	response, decoded := environment.call(t, http.MethodPatch, "/v1/me/personalization", off, true)
	if response.Code != http.StatusOK {
		t.Fatalf("disabling learning returned %d %v", response.Code, decoded)
	}
	learnCoastalTaste(t, environment)

	listed, body := environment.call(t, http.MethodGet, "/v1/me/personalization/insights", "", true)
	if listed.Code != http.StatusOK {
		t.Fatalf("insights returned %d %v", listed.Code, body)
	}
	if insights := list(t, body, "insights"); len(insights) != 0 {
		t.Fatalf("insights = %v, want nothing learned while learning is off", insights)
	}
}
