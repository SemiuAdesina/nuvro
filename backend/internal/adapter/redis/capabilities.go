package redis

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/place"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/route"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/weather"
)

const placeCapability = "place"
const routeCapability = "route"
const weatherCapability = "weather"

type PlaceCache struct{ cache *Cache }
type RouteCache struct{ cache *Cache }
type WeatherCache struct{ cache *Cache }

func NewPlaceCache(cache *Cache) *PlaceCache     { return &PlaceCache{cache: cache} }
func NewRouteCache(cache *Cache) *RouteCache     { return &RouteCache{cache: cache} }
func NewWeatherCache(cache *Cache) *WeatherCache { return &WeatherCache{cache: cache} }

func (p *PlaceCache) Places(ctx context.Context, key string) ([]place.Place, bool) {
	var values []place.Place
	if !p.cache.read(ctx, placeCapability, digest(key), &values) || len(values) == 0 {
		return nil, false
	}
	return values, true
}

func (p *PlaceCache) StorePlaces(ctx context.Context, key string, values []place.Place, ttl time.Duration) {
	p.cache.write(ctx, placeCapability, digest(key), values, ttl)
}

func (r *RouteCache) Route(ctx context.Context, key string) (route.Evidence, bool) {
	var evidence route.Evidence
	if !r.cache.read(ctx, routeCapability, digest(key), &evidence) || !evidence.Known() {
		return route.Evidence{}, false
	}
	return evidence, true
}

func (r *RouteCache) StoreRoute(ctx context.Context, key string, evidence route.Evidence, ttl time.Duration) {
	r.cache.write(ctx, routeCapability, digest(key), evidence, ttl)
}

func (w *WeatherCache) Window(ctx context.Context, key string) (weather.Window, bool) {
	var window weather.Window
	if !w.cache.read(ctx, weatherCapability, digest(key), &window) || !window.Known() {
		return weather.Window{}, false
	}
	return window, true
}

func (w *WeatherCache) StoreWindow(ctx context.Context, key string, window weather.Window, ttl time.Duration) {
	w.cache.write(ctx, weatherCapability, digest(key), window, ttl)
}

func digest(key string) string {
	sum := sha256.Sum256([]byte(key))
	return hex.EncodeToString(sum[:16])
}
