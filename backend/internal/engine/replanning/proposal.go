package replanning

import (
	"context"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/itinerary"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/place"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/replan"
)

func (e *Engine) changeFor(ctx context.Context, request Request, blocked conflict, intent Intent) replan.Change {
	if request.Trigger == replan.Tired && intent.LowerPace {
		return replan.Change{
			Kind: replan.Remove, Original: itemPointer(blocked.Item),
			Reasons: []replan.Reason{replan.TravelerRequestedLowerPace},
		}
	}
	candidates := e.candidates(ctx, request, blocked, intent)
	if len(candidates) == 0 {
		if moved, ok := e.movedLater(request, blocked); ok {
			return moved
		}
		return replan.Change{
			Kind: replan.Remove, Original: itemPointer(blocked.Item), Reasons: []replan.Reason{blocked.Reason},
		}
	}
	replacement := e.replacement(ctx, request, blocked, candidates[0])
	return replan.Change{
		Kind: replan.Replace, Original: itemPointer(blocked.Item), Proposed: &replacement,
		Reasons: []replan.Reason{blocked.Reason},
	}
}

func (e *Engine) replacement(ctx context.Context, request Request, blocked conflict, candidate place.Place) itinerary.Item {
	item := blocked.Item
	item.PlaceID = candidate.ID
	item.Title = candidate.Name
	item.Category = candidate.Category
	item.Location = candidate.Location
	item.Source = itinerary.Generated
	item.Status = itinerary.Planned
	item.EstimatedCostMinor = 0
	item.EndAt = item.StartAt.Add(candidate.Duration())
	if e.weather != nil && request.Weather.Known() {
		item.WeatherImpact = e.weather.Impact(request.Weather, item.StartAt, sensitivityFor(candidate.Category))
	}
	if previous, found := precedingItem(request, blocked.Item); found {
		evidence := e.routeFor(ctx, request, previous.Location, item)
		item.RouteFromPrevious = &evidence
	}
	return item
}

func (e *Engine) movedLater(request Request, blocked conflict) (replan.Change, bool) {
	day, startAt, found := laterSlot(request, blocked.Item)
	if !found {
		return replan.Change{}, false
	}
	moved := blocked.Item
	duration := blocked.Item.EndAt.Sub(blocked.Item.StartAt)
	moved.Day = day
	moved.StartAt = startAt
	moved.EndAt = startAt.Add(duration)
	return replan.Change{
		Kind: replan.Move, Original: itemPointer(blocked.Item), Proposed: &moved,
		Reasons: []replan.Reason{blocked.Reason, replan.BetterTimeWindow},
	}, true
}

func itemPointer(item itinerary.Item) *itinerary.Item {
	value := item
	return &value
}

func precedingItem(request Request, item itinerary.Item) (itinerary.Item, bool) {
	var previous itinerary.Item
	found := false
	for _, candidate := range request.Itinerary.Items {
		if candidate.Day != item.Day || !candidate.EndAt.Before(item.StartAt) {
			continue
		}
		if !found || candidate.EndAt.After(previous.EndAt) {
			previous = candidate
			found = true
		}
	}
	return previous, found
}
