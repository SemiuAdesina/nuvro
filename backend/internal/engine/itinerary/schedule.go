package itinerary

import (
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/geo"
	domain "github.com/SemiuAdesina/nuvro/backend/internal/domain/itinerary"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/place"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/route"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/weather"
)

type slot struct {
	day     int
	date    time.Time
	cursor  time.Time
	closeAt time.Time
}

func (e *Engine) schedule(request Request, candidates []place.Place, matrix route.Matrix, identifiers []string) []domain.Item {
	shape := ShapeFor(request.Pace)
	location := locationOf(request.Timezone)
	items := make([]domain.Item, 0, len(candidates))
	items = append(items, request.Locked...)
	index := 0
	for _, current := range daySlots(request, shape, location) {
		placed := 0
		var previous *place.Place
		var previousIndex int
		for placed < shape.ActivitiesPerDay && index < len(candidates) {
			candidate := candidates[index]
			item, ok := e.place(request, shape, &current, candidate, previous, previousIndex, index, matrix, identifiers)
			index++
			if !ok {
				continue
			}
			items = append(items, item)
			previous = &candidate
			previousIndex = index - 1
			placed++
		}
	}
	return items
}

func (e *Engine) place(request Request, shape Shape, current *slot, candidate place.Place, previous *place.Place, previousIndex, candidateIndex int, matrix route.Matrix, identifiers []string) (domain.Item, bool) {
	travel := travelEvidence(matrix, previous, previousIndex, candidateIndex)
	start := current.cursor
	if previous != nil {
		start = start.Add(travel.Duration() + shape.Buffer())
	}
	end := start.Add(candidate.Duration())
	if end.After(current.closeAt) {
		return domain.Item{}, false
	}
	if candidate.Hours.Known && candidate.Hours.OpenDuring(start, end) == place.Closed {
		return domain.Item{}, false
	}
	impact := e.impact(request, start, candidate)
	if impact == weather.HighImpact {
		return domain.Item{}, false
	}
	current.cursor = end
	return domain.Item{
		ID: identifiers[candidateIndex], TripID: request.TripID, Day: current.day,
		StartAt: start.UTC(), EndAt: end.UTC(), Kind: kindOf(candidate.Category),
		Title: candidate.Name, PlaceID: candidate.ID, Category: candidate.Category,
		Location: candidate.Location, RouteFromPrevious: routeOrNil(previous, travel),
		WeatherImpact: impact, CostCurrency: request.Currency,
		Status: domain.Planned, Source: domain.Generated,
	}, true
}

func (e *Engine) impact(request Request, at time.Time, candidate place.Place) weather.Impact {
	if e.weather == nil {
		return weather.UnknownImpact
	}
	return e.weather.Impact(request.Weather, at, weather.SensitivityOf(candidate.Category))
}

func daySlots(request Request, shape Shape, location *time.Location) []slot {
	days := int(request.End.Sub(request.Start).Hours()/24) + 1
	slots := make([]slot, 0, days)
	for offset := range days {
		date := request.Start.In(location).AddDate(0, 0, offset)
		start := time.Date(date.Year(), date.Month(), date.Day(), 0, shape.DayStartMinutes, 0, 0, location)
		end := time.Date(date.Year(), date.Month(), date.Day(), 0, shape.DayEndMinutes, 0, 0, location)
		slots = append(slots, slot{day: offset + 1, date: date, cursor: start, closeAt: end})
	}
	return slots
}

func travelEvidence(matrix route.Matrix, previous *place.Place, previousIndex, candidateIndex int) route.Evidence {
	if previous == nil {
		return route.Evidence{}
	}
	if evidence, found := matrix.Between(previousIndex, candidateIndex); found {
		return evidence
	}
	return route.UnavailableEvidence(previous.Location, geo.Point{}, matrix.Mode)
}

func routeOrNil(previous *place.Place, evidence route.Evidence) *route.Evidence {
	if previous == nil {
		return nil
	}
	value := evidence
	return &value
}

func kindOf(category place.Category) domain.Kind {
	switch category {
	case place.Restaurant, place.Cafe:
		return domain.Meal
	default:
		return domain.Activity
	}
}

func locationOf(timezone string) *time.Location {
	if timezone == "" {
		return time.UTC
	}
	location, err := time.LoadLocation(timezone)
	if err != nil {
		return time.UTC
	}
	return location
}
