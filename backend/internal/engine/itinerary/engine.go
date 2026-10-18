package itinerary

import (
	"context"
	"errors"
	"slices"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/geo"
	domain "github.com/SemiuAdesina/nuvro/backend/internal/domain/itinerary"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/place"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/route"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/id"
)

var ErrPlacesUnavailable = errors.New("itinerary: place evidence unavailable")

type Engine struct {
	places      PlaceFinder
	routes      RouteFinder
	weather     WeatherReader
	identifiers id.Generator
	timer       clock.Clock
}

func New(places PlaceFinder, routes RouteFinder, weatherReader WeatherReader, identifiers id.Generator, timer clock.Clock) *Engine {
	return &Engine{places: places, routes: routes, weather: weatherReader, identifiers: identifiers, timer: timer}
}

type Result struct {
	Itinerary  domain.Itinerary
	Violations []domain.Violation
	Candidates int
}

func (e *Engine) Generate(ctx context.Context, request Request) (Result, error) {
	candidates, err := e.candidates(ctx, request)
	if err != nil {
		return Result{}, err
	}
	if len(candidates) == 0 {
		return Result{}, ErrPlacesUnavailable
	}
	matrix := e.matrix(ctx, request, candidates)
	identifiers, err := e.identify(len(candidates))
	if err != nil {
		return Result{}, err
	}
	items := e.schedule(request, candidates, matrix, identifiers)
	slices.SortStableFunc(items, func(left, right domain.Item) int { return left.StartAt.Compare(right.StartAt) })
	plan := domain.Itinerary{
		TripID: request.TripID, Timezone: request.Timezone, Version: 1,
		GeneratedAt: e.timer.Now().UTC(), Items: items, AlgorithmVersion: AlgorithmVersion,
	}
	return Result{
		Itinerary:  plan,
		Violations: domain.Validate(plan, e.constraints(request, candidates)),
		Candidates: len(candidates),
	}, nil
}

func (e *Engine) identify(count int) ([]string, error) {
	identifiers := make([]string, count)
	for index := range identifiers {
		value, err := e.identifiers.Generate()
		if err != nil {
			return nil, err
		}
		identifiers[index] = value
	}
	return identifiers, nil
}

func (e *Engine) constraints(request Request, candidates []place.Place) domain.Constraints {
	shape := ShapeFor(request.Pace)
	locked := make(map[string]domain.Item, len(request.Locked))
	for _, item := range request.Locked {
		locked[item.ID] = item
	}
	return domain.Constraints{
		Start: request.Start.UTC(), End: request.End.UTC().AddDate(0, 0, 1),
		BufferMinutes: shape.BufferMinutes, Locked: locked,
		Openings: openings(candidates),
	}
}

func openings(candidates []place.Place) map[string]place.OpeningHours {
	hours := make(map[string]place.OpeningHours, len(candidates))
	for _, candidate := range candidates {
		if candidate.Hours.Known {
			hours[candidate.ID] = candidate.Hours
		}
	}
	return hours
}

func (e *Engine) matrix(ctx context.Context, request Request, candidates []place.Place) route.Matrix {
	if e.routes == nil {
		return route.Matrix{Mode: route.Walk}
	}
	points := make([]geo.Point, 0, len(candidates))
	for _, candidate := range candidates {
		points = append(points, candidate.Location)
	}
	return e.routes.Matrix(ctx, points, route.Walk, request.Start)
}
