package contract_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/SemiuAdesina/nuvro/backend/internal/platform/admission"
)

type recording struct {
	refuse    admission.Class
	refuseAll bool
	classes   []admission.Class
	subjects  []string
}

func (r *recording) Allow(_ context.Context, class admission.Class, subject string) (bool, error) {
	r.classes = append(r.classes, class)
	r.subjects = append(r.subjects, subject)
	return !r.refuseAll && class != r.refuse, nil
}

func TestTheRouterActuallyEnforcesAdmission(t *testing.T) {
	limiter := &recording{refuse: admission.StandardRead}
	environment := newEnvironment(t, withLimiter(limiter))
	response, problem := environment.call(t, http.MethodGet, "/v1/me", "", true)
	if response.Code != http.StatusTooManyRequests || problem["code"] != "rate_limited" {
		t.Fatalf("returned %d %v, want a throttled request", response.Code, problem)
	}
	if len(limiter.subjects) != 2 || limiter.subjects[1] != travellerID {
		t.Fatalf("subjects = %v, want the address guard then the traveller", limiter.subjects)
	}
	if limiter.classes[0] != admission.Authentication || limiter.classes[1] != admission.StandardRead {
		t.Fatalf("classes = %v", limiter.classes)
	}
}

func TestExpensiveSearchIsThrottledOnItsOwnBudget(t *testing.T) {
	limiter := &recording{}
	environment := newEnvironment(t, withLimiter(limiter))
	tripID := createTrip(t, environment)
	if len(limiter.classes) == 0 {
		t.Fatal("trip creation was never admitted")
	}
	limiter.classes = nil
	limiter.subjects = nil
	limiter.refuse = admission.ExpensiveSearch
	response, _ := environment.call(t, http.MethodPost, "/v1/trips/"+tripID+"/nearby/search",
		`{"location":{"latitude":32.6478,"longitude":-16.9048},"availableMinutes":60,"radiusMeters":2000}`, true)
	if response.Code != http.StatusTooManyRequests {
		t.Fatalf("returned %d, want a throttled search", response.Code)
	}
	if limiter.classes[1] != admission.ExpensiveSearch {
		t.Fatalf("classes = %v, want a search on the expensive budget", limiter.classes)
	}
}

func TestHealthProbesAreNeverThrottled(t *testing.T) {
	limiter := &recording{refuseAll: true}
	environment := newEnvironment(t, withLimiter(limiter))
	for _, path := range []string{"/health/live", "/health/ready"} {
		response, _ := environment.call(t, http.MethodGet, path, "", false)
		if response.Code != http.StatusOK {
			t.Fatalf("%s returned %d while the limiter refused everything", path, response.Code)
		}
	}
	if len(limiter.classes) != 0 {
		t.Fatalf("health probes consumed a budget: %v", limiter.classes)
	}
}

func TestUnauthenticatedCallersAreThrottledByAddress(t *testing.T) {
	limiter := &recording{refuse: admission.Authentication}
	environment := newEnvironment(t, withLimiter(limiter))
	response, problem := environment.call(t, http.MethodGet, "/v1/me", "", false)
	if response.Code != http.StatusTooManyRequests || problem["code"] != "rate_limited" {
		t.Fatalf("returned %d %v", response.Code, problem)
	}
	if len(limiter.classes) != 1 || limiter.classes[0] != admission.Authentication {
		t.Fatalf("classes = %v, want one address guard before authentication runs", limiter.classes)
	}
	if limiter.subjects[0] == travellerID {
		t.Fatalf("an unauthenticated caller was keyed as a traveller: %q", limiter.subjects[0])
	}
}
