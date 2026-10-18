package contract_test

import (
	"net/http"
	"testing"
)

func TestMalformedParametersReturnAProblemAndLeakNothing(t *testing.T) {
	environment := newEnvironment(t)
	for _, test := range []struct{ name, method, path string }{
		{"malformed trip", http.MethodGet, "/v1/trips/not-a-uuid/itinerary"},
		{"malformed proposal", http.MethodPost, "/v1/trips/not-a-uuid/replan/also-not-a-uuid/apply"},
		{"malformed watch", http.MethodDelete, "/v1/me/price-watches/not-a-uuid"},
	} {
		name := test.name
		for _, authenticated := range []bool{true, false} {
			response, problem := environment.call(t, test.method, test.path, "", authenticated)
			if response.Header().Get("Content-Type") != "application/problem+json" {
				t.Fatalf("%s returned %q", name, response.Header().Get("Content-Type"))
			}
			if problem["code"] == nil || problem["requestId"] == nil {
				t.Fatalf("%s returned %v", name, problem)
			}
			if detail, _ := problem["detail"].(string); detail != "Request is invalid." {
				t.Fatalf("%s leaked parser detail: %q", name, detail)
			}
		}
	}
}
