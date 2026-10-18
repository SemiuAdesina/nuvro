//go:build integration

package integration_test

import (
	"os"
	"testing"
	"time"

	redisclient "github.com/redis/go-redis/v9"

	redisadapter "github.com/SemiuAdesina/nuvro/backend/internal/adapter/redis"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/commerce"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/geo"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/money"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/place"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/route"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/weather"
	commerceengine "github.com/SemiuAdesina/nuvro/backend/internal/engine/commerce"
)

func redisURL(t *testing.T) string {
	t.Helper()
	url := os.Getenv("NUVRO_TEST_REDIS_URL")
	if url == "" {
		t.Fatal("NUVRO_TEST_REDIS_URL is required")
	}
	return url
}

func newCache(t *testing.T) *redisadapter.Cache {
	t.Helper()
	client, err := redisadapter.Open(redisURL(t))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	cache := redisadapter.NewCache(client)
	if cache == nil {
		t.Fatal("cache was not created")
	}
	return cache
}

func inspector(t *testing.T) *redisclient.Client {
	t.Helper()
	options, err := redisclient.ParseURL(redisURL(t))
	if err != nil {
		t.Fatal(err)
	}
	client := redisclient.NewClient(options)
	t.Cleanup(func() { _ = client.Close() })
	return client
}

func TestPlaceCacheStoresAndExpires(t *testing.T) {
	cache := redisadapter.NewPlaceCache(newCache(t))
	key := "integration-place-" + time.Now().UTC().Format("150405.000000")
	if _, found := cache.Places(t.Context(), key); found {
		t.Fatal("a cold cache reported a hit")
	}
	values := []place.Place{{
		ID: "fixture:place", Name: "Museu", Category: place.Museum,
		Location: geo.Point{Latitude: 32.6478, Longitude: -16.9048}, Status: place.Open,
	}}
	cache.StorePlaces(t.Context(), key, values, 2*time.Second)
	stored, found := cache.Places(t.Context(), key)
	if !found || len(stored) != 1 || stored[0].ID != "fixture:place" {
		t.Fatalf("cache read = %v %v", stored, found)
	}
	if stored[0].Category != place.Museum || stored[0].Location.Latitude != 32.6478 {
		t.Fatalf("round trip lost data: %+v", stored[0])
	}
	cache.StorePlaces(t.Context(), key+"-expiring", values, time.Second)
	time.Sleep(1200 * time.Millisecond)
	if _, found := cache.Places(t.Context(), key+"-expiring"); found {
		t.Fatal("an expired entry was still served")
	}
}

func TestRouteCacheOnlyServesKnownEvidence(t *testing.T) {
	cache := redisadapter.NewRouteCache(newCache(t))
	key := "integration-route-" + time.Now().UTC().Format("150405.000000")
	known := route.Evidence{
		Status: route.Found, Mode: route.Walk, DurationSeconds: 900, DistanceMeters: 1200, Confidence: route.High,
	}
	cache.StoreRoute(t.Context(), key, known, 5*time.Second)
	stored, found := cache.Route(t.Context(), key)
	if !found || stored.DurationSeconds != 900 || stored.Mode != route.Walk {
		t.Fatalf("route cache = %+v %v", stored, found)
	}
	unknown := route.UnavailableEvidence(geo.Point{}, geo.Point{}, route.Walk)
	cache.StoreRoute(t.Context(), key+"-unknown", unknown, 5*time.Second)
	if _, found := cache.Route(t.Context(), key+"-unknown"); found {
		t.Fatal("unavailable route evidence was served from cache as usable")
	}
}

func TestWeatherCacheOnlyServesAvailableWindows(t *testing.T) {
	cache := redisadapter.NewWeatherCache(newCache(t))
	key := "integration-weather-" + time.Now().UTC().Format("150405.000000")
	window := weather.Window{
		Availability: weather.Available,
		Daily: []weather.Day{{
			Date: time.Now().UTC().Truncate(24 * time.Hour), Condition: weather.Clear,
			HighCelsius: 24, LowCelsius: 18,
		}},
	}
	cache.StoreWindow(t.Context(), key, window, 5*time.Second)
	stored, found := cache.Window(t.Context(), key)
	if !found || len(stored.Daily) != 1 || stored.Daily[0].Condition != weather.Clear {
		t.Fatalf("weather cache = %+v %v", stored, found)
	}
	cache.StoreWindow(t.Context(), key+"-beyond", weather.Unavailable(weather.BeyondForecast), 5*time.Second)
	if _, found := cache.Window(t.Context(), key+"-beyond"); found {
		t.Fatal("an unavailable window was cached and served")
	}
}

func TestCacheKeysStayInsideTheNuvroNamespace(t *testing.T) {
	cache := redisadapter.NewPlaceCache(newCache(t))
	key := "integration-namespace-" + time.Now().UTC().Format("150405.000000")
	cache.StorePlaces(t.Context(), key, []place.Place{{
		ID: "fixture:place", Name: "Museu", Category: place.Museum,
		Location: geo.Point{Latitude: 1, Longitude: 1}, Status: place.Open,
	}}, 5*time.Second)
	client := inspector(t)
	found, err := client.Keys(t.Context(), "nuvro:v1:place:*").Result()
	if err != nil {
		t.Fatal(err)
	}
	if len(found) == 0 {
		t.Fatal("nothing was written under the documented namespace")
	}
	stray, err := client.Keys(t.Context(), "integration-namespace-*").Result()
	if err != nil {
		t.Fatal(err)
	}
	if len(stray) != 0 {
		t.Fatalf("keys were written outside the namespace: %v", stray)
	}
	for _, name := range found {
		if len(name) < len("nuvro:v1:place:") {
			t.Fatalf("malformed key %q", name)
		}
	}
}

func TestOfferCacheExpiresWithTheOffer(t *testing.T) {
	cache := redisadapter.NewOfferCache(newCache(t))
	key := "integration-offer-" + time.Now().UTC().Format("150405.000000")
	price, err := money.New(22_000, money.Currency("EUR"))
	if err != nil {
		t.Fatal(err)
	}
	stored := commerceengine.StoredOffer{
		Offer: commerce.Offer{
			ID: key, Partner: "fixture", Product: commerce.Flight, Title: "Lisbon to Funchal",
			Price: price, Available: true, ObservedAt: time.Now().UTC(),
		},
		Commercials: commerce.Commercials{
			OfferID: key, Partner: "fixture", CommissionMinor: 900, CommissionCurrency: "EUR",
		},
		Destination: "https://partner.invalid/book/" + key,
	}
	if err := cache.SaveOffers(t.Context(), []commerceengine.StoredOffer{stored}, 2*time.Second); err != nil {
		t.Fatal(err)
	}
	loaded, found := cache.Offer(t.Context(), key)
	if !found || loaded.Destination != stored.Destination {
		t.Fatalf("cache read = %+v %v", loaded, found)
	}
	if loaded.Commercials.CommissionMinor != 900 {
		t.Fatal("commission was lost from the internal record")
	}
	if loaded.Offer.Price.Minor() != 22_000 || loaded.Offer.Price.Currency().String() != "EUR" {
		t.Fatalf("price round trip lost data: %+v", loaded.Offer.Price)
	}
	if _, found := cache.Offer(t.Context(), key+"-missing"); found {
		t.Fatal("a cold key reported a hit")
	}
	time.Sleep(2200 * time.Millisecond)
	if _, found := cache.Offer(t.Context(), key); found {
		t.Fatal("an expired offer was still served")
	}
}
