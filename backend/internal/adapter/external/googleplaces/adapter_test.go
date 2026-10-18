package googleplaces_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/adapter/external/googleplaces"
	"github.com/SemiuAdesina/nuvro/backend/internal/adapter/external/httpclient"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/geo"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/place"
	"github.com/SemiuAdesina/nuvro/backend/internal/engine/places"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
)

var observedAt = time.Date(2026, 10, 14, 10, 0, 0, 0, time.UTC)

type capture struct {
	path      string
	fieldMask string
	apiKey    string
	body      string
}

func serve(t *testing.T, status int, fixture string, seen *capture) *googleplaces.Adapter {
	t.Helper()
	body := ""
	if fixture != "" {
		raw, err := os.ReadFile("testdata/" + fixture)
		if err != nil {
			t.Fatal(err)
		}
		body = string(raw)
	}
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if seen != nil {
			seen.path = request.URL.Path
			seen.fieldMask = request.Header.Get("X-Goog-FieldMask")
			seen.apiKey = request.Header.Get("X-Goog-Api-Key")
			buffer := make([]byte, 2048)
			read, _ := request.Body.Read(buffer)
			seen.body = string(buffer[:read])
		}
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(status)
		_, _ = writer.Write([]byte(body))
	}))
	t.Cleanup(server.Close)
	return googleplaces.New(httpclient.New(httpclient.Options{BaseURL: server.URL, Timeout: time.Second}), "synthetic-test-key", clock.NewFixed(observedAt))
}

func TestNearbySearchNormalisesIntoTheCanonicalPlaceModel(t *testing.T) {
	seen := &capture{}
	adapter := serve(t, http.StatusOK, "nearby.json", seen)
	found, err := adapter.Nearby(t.Context(), places.NearbyQuery{
		Centre: geo.Point{Latitude: 32.65, Longitude: -16.91}, RadiusMeters: 5000,
		Categories: []place.Category{place.Market, place.Museum}, Limit: 10, Language: "en",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(found) != 2 {
		t.Fatalf("accepted %d places, want the two valid ones", len(found))
	}
	assertMarket(t, found[0])
	if found[1].Status != place.Closed {
		t.Fatalf("a closed place was not reported as closed: %q", found[1].Status)
	}
}

func assertMarket(t *testing.T, market place.Place) {
	t.Helper()
	if market.ID != "google_places:SYNTHETIC_PLACE_A" {
		t.Fatalf("identity was not namespaced: %q", market.ID)
	}
	reference, ok := market.Reference("google_places")
	if !ok || reference.ExternalID != "SYNTHETIC_PLACE_A" {
		t.Fatal("the provider identity was not preserved separately")
	}
	if market.Category != place.Market || market.Status != place.Open {
		t.Fatalf("normalisation wrong: %+v", market)
	}
	if market.Location.Latitude != 32.6478 || market.Address == "" {
		t.Fatal("location or address lost")
	}
	if market.PriceLevel != 1 || market.Rating != 4.3 || market.ReviewCount != 18422 {
		t.Fatalf("metadata lost: %+v", market)
	}
	if !market.Freshness.ObservedAt.Equal(observedAt) {
		t.Fatal("observation time not recorded")
	}
}

func TestNearbySearchRequestsOnlyTheFieldsItNeeds(t *testing.T) {
	seen := &capture{}
	adapter := serve(t, http.StatusOK, "nearby.json", seen)
	if _, err := adapter.Nearby(t.Context(), places.NearbyQuery{
		Centre: geo.Point{Latitude: 32.65, Longitude: -16.91}, RadiusMeters: 5000,
		Categories: []place.Category{place.Museum}, Limit: 10,
	}); err != nil {
		t.Fatal(err)
	}
	if seen.fieldMask == "" {
		t.Fatal("no field mask was sent")
	}
	for _, expensive := range []string{"reviews", "websiteUri", "nationalPhoneNumber", "editorialSummary"} {
		if contains(seen.fieldMask, expensive) {
			t.Fatalf("field mask requests %q, which the preview does not need", expensive)
		}
	}
	if seen.apiKey == "" {
		t.Fatal("the api key header was not sent")
	}
	if !contains(seen.body, "\"includedTypes\"") || !contains(seen.body, "museum") {
		t.Fatalf("category was not translated into provider types: %s", seen.body)
	}
	if !contains(seen.body, "locationRestriction") {
		t.Fatal("nearby search sent no location restriction")
	}
}

func TestPhotoAttributionAndDimensionsSurvive(t *testing.T) {
	adapter := serve(t, http.StatusOK, "details.json", nil)
	found, err := adapter.Details(t.Context(), places.DetailsQuery{
		Reference: place.Reference{Provider: "google_places", ExternalID: "SYNTHETIC_PLACE_A"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(found.Media) != 1 {
		t.Fatalf("photo without dimensions was kept: %+v", found.Media)
	}
	asset := found.Media[0]
	if asset.Width != 1600 || asset.Height != 900 || asset.AspectRatio < 1.7 {
		t.Fatalf("dimensions lost: %+v", asset)
	}
	if len(asset.Attribution) != 1 || asset.Attribution[0].Text != "First Author" {
		t.Fatalf("attribution lost: %+v", asset.Attribution)
	}
	if asset.Attribution[0].Provider != "google_places" {
		t.Fatal("attribution lost its provider")
	}
	if len(found.Attribution) != 2 {
		t.Fatalf("place level attribution = %d, want both authors", len(found.Attribution))
	}
}

func TestOutOfRangeValuesAreClampedOrRejected(t *testing.T) {
	adapter := serve(t, http.StatusOK, "details.json", nil)
	found, err := adapter.Details(t.Context(), places.DetailsQuery{
		Reference: place.Reference{Provider: "google_places", ExternalID: "SYNTHETIC_PLACE_A"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if found.Rating != 5 {
		t.Fatalf("rating = %v, want clamped to 5", found.Rating)
	}
	if found.ReviewCount != 0 {
		t.Fatalf("negative review count = %d", found.ReviewCount)
	}
	if !found.Hours.Known || len(found.Hours.Intervals) != 1 {
		t.Fatalf("opening hours = %+v, want only the valid interval", found.Hours)
	}
	if found.Hours.Intervals[0].Closes != 24*60 {
		t.Fatalf("midnight close was not normalised: %+v", found.Hours.Intervals[0])
	}
	if found.Status != place.UnknownStatus {
		t.Fatalf("absent open-now became %q", found.Status)
	}
}

func TestProviderFailuresMapToSafeErrors(t *testing.T) {
	for status, want := range map[int]error{
		http.StatusTooManyRequests:     httpclient.ErrRateLimited,
		http.StatusForbidden:           httpclient.ErrUnavailable,
		http.StatusInternalServerError: httpclient.ErrUnavailable,
	} {
		adapter := serve(t, status, "nearby.json", nil)
		_, err := adapter.Nearby(t.Context(), places.NearbyQuery{Centre: geo.Point{Latitude: 1, Longitude: 1}, Limit: 5})
		if !errors.Is(err, want) {
			t.Fatalf("status %d produced %v, want %v", status, err, want)
		}
	}
}

func TestMalformedResponsesAreRejected(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		_, _ = writer.Write([]byte("{not json"))
	}))
	t.Cleanup(server.Close)
	adapter := googleplaces.New(httpclient.New(httpclient.Options{BaseURL: server.URL, Timeout: time.Second}), "synthetic-test-key", clock.NewFixed(observedAt))
	if _, err := adapter.Nearby(t.Context(), places.NearbyQuery{Centre: geo.Point{Latitude: 1, Longitude: 1}, Limit: 5}); !errors.Is(err, httpclient.ErrMalformed) {
		t.Fatalf("malformed body produced %v", err)
	}
}

func TestEmptyResultsAreNotAnError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		_, _ = writer.Write([]byte(`{"places":[]}`))
	}))
	t.Cleanup(server.Close)
	adapter := googleplaces.New(httpclient.New(httpclient.Options{BaseURL: server.URL, Timeout: time.Second}), "synthetic-test-key", clock.NewFixed(observedAt))
	found, err := adapter.Nearby(t.Context(), places.NearbyQuery{Centre: geo.Point{Latitude: 1, Longitude: 1}, Limit: 5})
	if err != nil || len(found) != 0 {
		t.Fatalf("empty result produced %v %v", found, err)
	}
}

func TestDetailsRejectsForeignReferences(t *testing.T) {
	adapter := serve(t, http.StatusOK, "details.json", nil)
	if _, err := adapter.Details(t.Context(), places.DetailsQuery{
		Reference: place.Reference{Provider: "someone_else", ExternalID: "X"},
	}); err == nil {
		t.Fatal("a reference from another provider was accepted")
	}
}

func contains(haystack, needle string) bool {
	return len(haystack) >= len(needle) && indexOf(haystack, needle) >= 0
}

func indexOf(haystack, needle string) int {
	for index := 0; index+len(needle) <= len(haystack); index++ {
		if haystack[index:index+len(needle)] == needle {
			return index
		}
	}
	return -1
}
