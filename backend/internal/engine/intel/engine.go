package intel

import (
	"context"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/destination"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/geo"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/place"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/weather"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
)

const previewPlaceLimit = 8
const mediaLimit = 5
const sectionConcurrency = 4
const weatherValidity = time.Hour
const placesValidity = 15 * time.Minute

type Engine struct {
	places  PlaceReader
	weather WeatherReader
	media   MediaSelector
	timer   clock.Clock
}

func New(places PlaceReader, weatherReader WeatherReader, media MediaSelector, timer clock.Clock) *Engine {
	return &Engine{places: places, weather: weatherReader, media: media, timer: timer}
}

type Request struct {
	Destination destination.Destination
	Start       time.Time
	End         time.Time
	Language    string
	Interests   []string
}

type Intelligence struct {
	Destination    destination.Destination
	Weather        weather.Window
	WeatherSection Section
	ThingsToDo     []place.Place
	PlacesSection  Section
	Media          []destination.Media
	MediaSection   Section
	GeneratedAt    time.Time
}

func (e *Engine) Describe(ctx context.Context, request Request) Intelligence {
	result := Intelligence{Destination: request.Destination, GeneratedAt: e.timer.Now().UTC()}
	point := geo.Point{Latitude: request.Destination.Latitude, Longitude: request.Destination.Longitude}
	group, groupCtx := errgroup.WithContext(ctx)
	group.SetLimit(sectionConcurrency)
	group.Go(func() error {
		result.Weather, result.WeatherSection = e.describeWeather(groupCtx, point, request)
		return nil
	})
	group.Go(func() error {
		result.ThingsToDo, result.PlacesSection = e.describePlaces(groupCtx, point, request)
		return nil
	})
	group.Go(func() error {
		result.Media, result.MediaSection = e.describeMedia(groupCtx, request)
		return nil
	})
	_ = group.Wait()
	return result
}
