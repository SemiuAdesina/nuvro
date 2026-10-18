package googleroutes_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/adapter/external/googleroutes"
	"github.com/SemiuAdesina/nuvro/backend/internal/adapter/external/httpclient"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/geo"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/route"
	"github.com/SemiuAdesina/nuvro/backend/internal/engine/routing"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
)

var computedAt = time.Date(2026, 10, 14, 10, 0, 0, 0, time.UTC)

var funchal = geo.Point{Latitude: 32.6478, Longitude: -16.9048}
var monte = geo.Point{Latitude: 32.6690, Longitude: -16.9000}

func adapter(t *testing.T, status int, body string, seen *string) *googleroutes.Adapter {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if seen != nil {
			buffer := make([]byte, 4096)
			read, _ := request.Body.Read(buffer)
			*seen = string(buffer[:read])
		}
		writer.WriteHeader(status)
		_, _ = writer.Write([]byte(body))
	}))
	t.Cleanup(server.Close)
	return googleroutes.New(httpclient.New(httpclient.Options{BaseURL: server.URL, Timeout: time.Second}), "synthetic-test-key", clock.NewFixed(computedAt))
}

func TestComputeRouteParsesDurationAndDistance(t *testing.T) {
	body := `{"routes":[{"duration":"1380s","distanceMeters":5400}]}`
	evidence, err := adapter(t, http.StatusOK, body, nil).Compute(t.Context(), routing.Query{
		Origin: funchal, Destination: monte, Mode: route.Walk,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !evidence.Known() || evidence.DurationSeconds != 1380 || evidence.DistanceMeters != 5400 {
		t.Fatalf("parsed evidence = %+v", evidence)
	}
	if evidence.Duration() != 23*time.Minute {
		t.Fatalf("duration = %v", evidence.Duration())
	}
	if evidence.Mode != route.Walk || evidence.Confidence != route.High {
		t.Fatalf("mode or confidence lost: %+v", evidence)
	}
	if evidence.Source.Provider != "google_routes" || evidence.Source.Freshness.ObservedAt.IsZero() {
		t.Fatal("provenance missing")
	}
}

func TestNoRouteIsReportedHonestly(t *testing.T) {
	evidence, err := adapter(t, http.StatusOK, `{"routes":[]}`, nil).Compute(t.Context(), routing.Query{
		Origin: funchal, Destination: monte, Mode: route.Transit,
	})
	if err != nil {
		t.Fatal(err)
	}
	if evidence.Status != route.NotFound || evidence.Known() {
		t.Fatalf("an absent route was not reported: %+v", evidence)
	}
}

func TestMalformedDurationsAreRejectedNotGuessed(t *testing.T) {
	for _, body := range []string{
		`{"routes":[{"duration":"twenty minutes","distanceMeters":100}]}`,
		`{"routes":[{"duration":"","distanceMeters":100}]}`,
		`{"routes":[{"duration":"1380","distanceMeters":100}]}`,
		`{"routes":[{"duration":"-90s","distanceMeters":100}]}`,
	} {
		evidence, err := adapter(t, http.StatusOK, body, nil).Compute(t.Context(), routing.Query{
			Origin: funchal, Destination: monte, Mode: route.Walk,
		})
		if err != nil {
			t.Fatal(err)
		}
		if evidence.Known() {
			t.Fatalf("body %q produced a usable duration", body)
		}
	}
}

func TestDrivingRequestsCarryDepartureAndTrafficPreference(t *testing.T) {
	var seen string
	_, err := adapter(t, http.StatusOK, `{"routes":[{"duration":"600s","distanceMeters":8000}]}`, &seen).
		Compute(t.Context(), routing.Query{
			Origin: funchal, Destination: monte, Mode: route.Drive, DepartAt: computedAt,
		})
	if err != nil {
		t.Fatal(err)
	}
	if !contains(seen, "departureTime") || !contains(seen, "TRAFFIC_AWARE") {
		t.Fatalf("driving request omitted traffic context: %s", seen)
	}
	if !contains(seen, "\"travelMode\":\"DRIVE\"") {
		t.Fatalf("travel mode not translated: %s", seen)
	}
	var walkRequest string
	_, err = adapter(t, http.StatusOK, `{"routes":[{"duration":"600s","distanceMeters":800}]}`, &walkRequest).
		Compute(t.Context(), routing.Query{Origin: funchal, Destination: monte, Mode: route.Walk, DepartAt: computedAt})
	if err != nil {
		t.Fatal(err)
	}
	if contains(walkRequest, "departureTime") {
		t.Fatal("a walking route was made time-dependent")
	}
}

func TestMatrixKeepsOnlyElementsWithRealRoutes(t *testing.T) {
	body := `[
      {"originIndex":0,"destinationIndex":1,"duration":"900s","distanceMeters":3000,"condition":"ROUTE_EXISTS"},
      {"originIndex":1,"destinationIndex":0,"duration":"880s","distanceMeters":3000,"condition":"ROUTE_EXISTS"},
      {"originIndex":0,"destinationIndex":2,"condition":"ROUTE_NOT_FOUND"},
      {"originIndex":0,"destinationIndex":0,"duration":"0s","distanceMeters":0,"condition":"ROUTE_EXISTS"},
      {"originIndex":9,"destinationIndex":1,"duration":"10s","distanceMeters":10,"condition":"ROUTE_EXISTS"}
    ]`
	matrix, err := adapter(t, http.StatusOK, body, nil).Matrix(t.Context(), routing.MatrixQuery{
		Points: []geo.Point{funchal, monte, {Latitude: 32.7, Longitude: -17.0}}, Mode: route.Walk,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(matrix.Legs) != 2 {
		t.Fatalf("kept %d legs, want the two real ones", len(matrix.Legs))
	}
	evidence, found := matrix.Between(0, 1)
	if !found || evidence.DurationSeconds != 900 {
		t.Fatalf("leg lookup failed: %+v", evidence)
	}
	if _, found := matrix.Between(0, 2); found {
		t.Fatal("a leg with no route was returned")
	}
	if _, found := matrix.Between(9, 1); found {
		t.Fatal("an out-of-range index was accepted")
	}
}

func TestFailedMatrixDegradesToEmpty(t *testing.T) {
	matrix, err := adapter(t, http.StatusTooManyRequests, "[]", nil).Matrix(t.Context(), routing.MatrixQuery{
		Points: []geo.Point{funchal, monte}, Mode: route.Walk,
	})
	if err == nil {
		t.Fatal("a rate-limited matrix reported success")
	}
	if len(matrix.Legs) != 0 {
		t.Fatal("legs were invented on failure")
	}
}

func contains(haystack, needle string) bool {
	for index := 0; index+len(needle) <= len(haystack); index++ {
		if haystack[index:index+len(needle)] == needle {
			return true
		}
	}
	return false
}
