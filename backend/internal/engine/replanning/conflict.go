package replanning

import (
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/itinerary"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/place"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/replan"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/weather"
)

type conflict struct {
	Item   itinerary.Item
	Reason replan.Reason
}

func (e *Engine) conflicts(request Request) []conflict {
	found := make([]conflict, 0)
	for _, item := range affected(request) {
		if item.Locked() {
			continue
		}
		if reason, blocked := e.blocked(request, item); blocked {
			found = append(found, conflict{Item: item, Reason: reason})
		}
	}
	return found
}

func affected(request Request) []itinerary.Item {
	items := make([]itinerary.Item, 0, len(request.Itinerary.Items))
	for _, item := range request.Itinerary.Items {
		if item.Status == itinerary.Completed || item.Status == itinerary.Cancelled || item.Status == itinerary.Skipped {
			continue
		}
		if !inWindow(item, request) {
			continue
		}
		items = append(items, item)
	}
	return items
}

func inWindow(item itinerary.Item, request Request) bool {
	if !request.WindowStart.IsZero() && item.EndAt.Before(request.WindowStart.UTC()) {
		return false
	}
	if !request.WindowEnd.IsZero() && item.StartAt.After(request.WindowEnd.UTC()) {
		return false
	}
	return !item.EndAt.Before(request.Now.UTC())
}

func (e *Engine) blocked(request Request, item itinerary.Item) (replan.Reason, bool) {
	switch request.Trigger {
	case replan.Weather:
		return replan.WeatherConflict, e.weatherBlocks(request, item)
	case replan.Closed:
		return replan.PlaceClosed, item.Category != "" && !opensDuring(item)
	case replan.Late, replan.Missed:
		return lateReason(request.Trigger), item.StartAt.Before(request.Now.UTC())
	case replan.Tired:
		return replan.TravelerRequestedLowerPace, densityCandidate(request, item)
	default:
		return replan.BetterTimeWindow, false
	}
}

func (e *Engine) weatherBlocks(request Request, item itinerary.Item) bool {
	if e.weather == nil || !request.Weather.Known() {
		return item.WeatherImpact == weather.HighImpact
	}
	impact := e.weather.Impact(request.Weather, item.StartAt, weather.SensitivityOf(item.Category))
	return impact == weather.HighImpact
}

func lateReason(trigger replan.Trigger) replan.Reason {
	if trigger == replan.Missed {
		return replan.MissedStartTime
	}
	return replan.InsufficientTravelTime
}

func densityCandidate(request Request, item itinerary.Item) bool {
	day := 0
	for _, candidate := range request.Itinerary.Items {
		if candidate.Day == item.Day && !candidate.Locked() {
			day++
		}
	}
	if day < 2 {
		return false
	}
	return item.StartAt.After(request.Now.UTC()) && lastOfDay(request, item)
}

func lastOfDay(request Request, item itinerary.Item) bool {
	for _, candidate := range request.Itinerary.Items {
		if candidate.Day == item.Day && candidate.StartAt.After(item.StartAt) && !candidate.Locked() {
			return false
		}
	}
	return true
}

func opensDuring(item itinerary.Item) bool {
	return item.Status != itinerary.Planned || item.PlaceID == ""
}

func sensitivityFor(category place.Category) weather.Sensitivity {
	return weather.SensitivityOf(category)
}
