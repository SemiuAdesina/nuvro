package app

import (
	"github.com/SemiuAdesina/nuvro/backend/internal/adapter/external/googleplaces"
	"github.com/SemiuAdesina/nuvro/backend/internal/adapter/external/googleroutes"
	"github.com/SemiuAdesina/nuvro/backend/internal/adapter/external/googleweather"
	"github.com/SemiuAdesina/nuvro/backend/internal/adapter/external/httpclient"
	redisadapter "github.com/SemiuAdesina/nuvro/backend/internal/adapter/redis"
	"github.com/SemiuAdesina/nuvro/backend/internal/engine/intel"
	itineraryengine "github.com/SemiuAdesina/nuvro/backend/internal/engine/itinerary"
	"github.com/SemiuAdesina/nuvro/backend/internal/engine/media"
	"github.com/SemiuAdesina/nuvro/backend/internal/engine/places"
	"github.com/SemiuAdesina/nuvro/backend/internal/engine/routing"
	weatherengine "github.com/SemiuAdesina/nuvro/backend/internal/engine/weather"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/config"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/id"
)

type intelligence struct {
	places    *places.Engine
	routes    *routing.Engine
	weather   *weatherengine.Engine
	media     *media.Selector
	describer *intel.Engine
	planner   *itineraryengine.Engine
}

func newIntelligence(cfg config.Config, cache *redisadapter.Cache, catalog googleplaces.DestinationResolver,
	timer clock.Clock, metrics httpclient.Recorder) intelligence {
	search, details, routeProvider, weatherProvider := providers(cfg, timer, metrics)
	placesEngine := places.New(search, details, placeCache(cache), timer)
	routesEngine := routing.New(routeProvider, routeCache(cache), timer)
	weatherEngine := weatherengine.New(weatherProvider, weatherCache(cache), timer)
	selector := media.NewSelector(mediaReader(search, catalog))
	return intelligence{
		places: placesEngine, routes: routesEngine, weather: weatherEngine, media: selector,
		describer: intel.New(placeReader{engine: placesEngine}, weatherEngine, selector, timer),
		planner: itineraryengine.New(
			placeFinder{engine: placesEngine}, routeFinder{engine: routesEngine},
			weatherEngine, id.NewUUID(), timer,
		),
	}
}

func providers(cfg config.Config, timer clock.Clock, metrics httpclient.Recorder) (places.SearchProvider, places.DetailsProvider, routing.Provider, weatherengine.Provider) {
	if !cfg.Providers.UsesGoogle() {
		return nil, nil, nil, nil
	}
	key := cfg.Providers.GoogleMapsAPIKey
	connect := func(name, baseURL string) *httpclient.Client {
		return httpclient.New(httpclient.Options{
			Provider: name, BaseURL: baseURL, Timeout: cfg.Providers.RequestTimeout,
			Retry: httpclient.DefaultRetry(), Breaker: httpclient.DefaultBreaker(),
			Metrics: metrics, Clock: timer,
		})
	}
	placesAdapter := googleplaces.New(connect("google_places", cfg.Providers.PlacesBaseURL), key, timer)
	routesAdapter := googleroutes.New(connect("google_routes", cfg.Providers.RoutesBaseURL), key, timer)
	weatherAdapter := googleweather.New(connect("google_weather", cfg.Providers.WeatherBaseURL), key, timer)
	return placesAdapter, placesAdapter, routesAdapter, weatherAdapter
}

func mediaReader(search places.SearchProvider, catalog googleplaces.DestinationResolver) media.CandidateReader {
	if search == nil || catalog == nil {
		return nil
	}
	return googleplaces.NewMediaAdapter(search, catalog)
}

func placeCache(cache *redisadapter.Cache) places.Cache {
	if cache == nil {
		return nil
	}
	return redisadapter.NewPlaceCache(cache)
}

func routeCache(cache *redisadapter.Cache) routing.Cache {
	if cache == nil {
		return nil
	}
	return redisadapter.NewRouteCache(cache)
}

func weatherCache(cache *redisadapter.Cache) weatherengine.Cache {
	if cache == nil {
		return nil
	}
	return redisadapter.NewWeatherCache(cache)
}
