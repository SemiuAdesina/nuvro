package middleware_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/SemiuAdesina/nuvro/backend/internal/platform/admission"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clientip"
	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/middleware"
)

type limiter struct {
	allowed  bool
	err      error
	classes  []admission.Class
	subjects []string
}

func (l *limiter) Allow(_ context.Context, class admission.Class, subject string) (bool, error) {
	l.classes = append(l.classes, class)
	l.subjects = append(l.subjects, subject)
	return l.allowed, l.err
}

func ok() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) })
}

func TestRateAdmission(t *testing.T) {
	for _, test := range []struct {
		limiter *limiter
		status  int
	}{
		{&limiter{allowed: true}, http.StatusNoContent},
		{&limiter{}, http.StatusTooManyRequests},
		{&limiter{err: errors.New("private")}, http.StatusServiceUnavailable},
	} {
		handler := middleware.RateLimit(test.limiter, admission.StandardRead,
			func(*http.Request) string { return "principal" })(ok())
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil))
		if response.Code != test.status {
			t.Fatalf("returned %d, want %d", response.Code, test.status)
		}
	}
}

func TestAdmissionKeysAuthenticatedTravellersByIdentity(t *testing.T) {
	recorder := &limiter{allowed: true}
	handler := middleware.Admit(recorder, func(*http.Request) (string, bool) { return "traveller-1", true })(ok())
	request := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/v1/discover/matches", nil)
	request.RemoteAddr = "203.0.113.9:51000"
	handler.ServeHTTP(httptest.NewRecorder(), request)
	if len(recorder.subjects) != 1 || recorder.subjects[0] != "traveller-1" {
		t.Fatalf("subjects = %v, want the traveller identity", recorder.subjects)
	}
	if recorder.classes[0] != admission.ExpensiveSearch {
		t.Fatalf("class = %q, want the expensive search class", recorder.classes[0])
	}
}

func TestAdmissionLeavesUnidentifiedCallersToTheGuard(t *testing.T) {
	recorder := &limiter{allowed: true}
	handler := middleware.Admit(recorder, func(*http.Request) (string, bool) { return "", false })(ok())
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/v1/me", nil))
	if response.Code != http.StatusNoContent || len(recorder.classes) != 0 {
		t.Fatalf("an unidentified caller was charged a per-traveller budget: %v", recorder.classes)
	}
}

func TestTheGuardKeysEveryCallerByAddress(t *testing.T) {
	recorder := &limiter{allowed: true}
	request := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/v1/discover/matches", nil)
	request.RemoteAddr = "203.0.113.9:51000"
	middleware.Guard(recorder, nil)(ok()).ServeHTTP(httptest.NewRecorder(), request)
	if recorder.subjects[0] != "203.0.113.9" {
		t.Fatalf("subject = %q, want the client address without its port", recorder.subjects[0])
	}
	if recorder.classes[0] != admission.Authentication {
		t.Fatalf("class = %q, want the at-the-door budget", recorder.classes[0])
	}
}

func TestTheGuardSkipsHealthProbes(t *testing.T) {
	recorder := &limiter{}
	response := httptest.NewRecorder()
	middleware.Guard(recorder, nil)(ok()).
		ServeHTTP(response, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/health/live", nil))
	if response.Code != http.StatusNoContent || len(recorder.classes) != 0 {
		t.Fatalf("a health probe was charged a budget: %d %v", response.Code, recorder.classes)
	}
}

func TestAdmissionNeverThrottlesHealthProbes(t *testing.T) {
	recorder := &limiter{}
	handler := middleware.Admit(recorder, func(*http.Request) (string, bool) { return "", false })(ok())
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/health/ready", nil))
	if response.Code != http.StatusNoContent || len(recorder.classes) != 0 {
		t.Fatalf("a health probe was admitted through the limiter: %d %v", response.Code, recorder.classes)
	}
}

func TestAdmissionWithoutALimiterLetsEverythingThrough(t *testing.T) {
	response := httptest.NewRecorder()
	middleware.Admit(nil, func(*http.Request) (string, bool) { return "traveller", true })(ok()).
		ServeHTTP(response, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/v1/me", nil))
	if response.Code != http.StatusNoContent {
		t.Fatalf("returned %d with no limiter configured", response.Code)
	}
}

func TestRouteClassification(t *testing.T) {
	for _, test := range []struct {
		method, pattern string
		class           admission.Class
	}{
		{http.MethodGet, "/health/live", admission.Exempt},
		{http.MethodGet, "/health/ready", admission.Exempt},
		{http.MethodPost, "/v1/discover/matches", admission.ExpensiveSearch},
		{http.MethodPost, "/v1/trips/{tripId}/nearby/search", admission.ExpensiveSearch},
		{http.MethodPost, "/v1/trips/{tripId}/flights/search", admission.ExpensiveSearch},
		{http.MethodPost, "/v1/trips/{tripId}/itinerary/generate", admission.ExpensiveSearch},
		{http.MethodPost, "/v1/trips/{tripId}/translate", admission.ExpensiveSearch},
		{http.MethodGet, "/v1/trips/{tripId}/offline-pack", admission.ExpensiveSearch},
		{http.MethodPost, "/v1/commerce/handoffs", admission.CommerceWrite},
		{http.MethodPost, "/v1/me/price-watches", admission.CommerceWrite},
		{http.MethodDelete, "/v1/me/price-watches/{watchId}", admission.CommerceWrite},
		{http.MethodGet, "/v1/me/price-watches", admission.StandardRead},
		{http.MethodGet, "/v1/me", admission.StandardRead},
		{http.MethodPatch, "/v1/me/personalization", admission.StandardRead},
	} {
		if class := middleware.ClassOf(test.method, test.pattern); class != test.class {
			t.Fatalf("%s %s = %q, want %q", test.method, test.pattern, class, test.class)
		}
	}
}

func TestTheGuardKeysCallersBehindATrustedProxyByTheForwardedClient(t *testing.T) {
	addresses, err := clientip.NewResolver([]string{"10.0.0.0/8"}, 1)
	if err != nil {
		t.Fatal(err)
	}
	for name, test := range map[string]struct{ remote, forwarded, want string }{
		"trusted proxy":   {"10.0.0.4:443", "203.0.113.9", "203.0.113.9"},
		"spoofed prefix":  {"10.0.0.4:443", "9.9.9.9, 203.0.113.9", "203.0.113.9"},
		"untrusted peer":  {"198.51.100.7:51000", "203.0.113.9", "198.51.100.7"},
		"malformed chain": {"10.0.0.4:443", "not-an-address", "10.0.0.4"},
	} {
		recorder := &limiter{allowed: true}
		request := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/v1/discover/matches", nil)
		request.RemoteAddr = test.remote
		request.Header.Set(clientip.ForwardedHeader, test.forwarded)
		middleware.Guard(recorder, addresses)(ok()).ServeHTTP(httptest.NewRecorder(), request)
		if recorder.subjects[0] != test.want {
			t.Fatalf("%s keyed %q, want %q", name, recorder.subjects[0], test.want)
		}
	}
}
