package replanning

import (
	"context"
	"slices"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/geo"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/itinerary"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/place"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/replan"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/route"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/weather"
)

const candidateLimit = 12
const nearRadiusMeters = 4000

func (e *Engine) candidates(ctx context.Context, request Request, blocked conflict, intent Intent) []place.Place {
	if e.places == nil {
		return nil
	}
	centre := searchCentre(request, blocked)
	found, err := e.places.Nearby(ctx, centre, replacementCategories(request, blocked, intent), request.Language, candidateLimit)
	if err != nil {
		return nil
	}
	return e.acceptable(request, blocked, found, centre)
}

func searchCentre(request Request, blocked conflict) geo.Point {
	if request.CurrentLocation != nil && !request.CurrentLocation.IsZero() {
		return *request.CurrentLocation
	}
	if !blocked.Item.Location.IsZero() {
		return blocked.Item.Location
	}
	for _, item := range request.Itinerary.Items {
		if !item.Location.IsZero() {
			return item.Location
		}
	}
	return geo.Point{}
}

func replacementCategories(request Request, blocked conflict, intent Intent) []place.Category {
	if len(intent.Categories) > 0 {
		return intent.Categories
	}
	if request.Trigger == replan.Weather {
		return []place.Category{place.Museum, place.Restaurant, place.Cafe, place.Shopping}
	}
	if blocked.Item.Category.Valid() && blocked.Item.Category != place.Other {
		return []place.Category{blocked.Item.Category, place.Attraction}
	}
	return []place.Category{place.Attraction, place.Museum}
}

func (e *Engine) acceptable(request Request, blocked conflict, found []place.Place, centre geo.Point) []place.Place {
	accepted := make([]place.Place, 0, len(found))
	for _, candidate := range found {
		if candidate.Validate() != nil || candidate.Status == place.Closed {
			continue
		}
		if candidate.ID == blocked.Item.PlaceID {
			continue
		}
		if !e.suitable(request, blocked, candidate) {
			continue
		}
		if !withinWalkingLimit(request, centre, candidate) {
			continue
		}
		accepted = append(accepted, candidate)
	}
	slices.SortStableFunc(accepted, func(left, right place.Place) int {
		leftDistance := int(centre.DistanceMeters(left.Location))
		rightDistance := int(centre.DistanceMeters(right.Location))
		if request.ReducedWalking || leftDistance/500 != rightDistance/500 {
			if order := leftDistance - rightDistance; order != 0 {
				return order
			}
		}
		if order := int(right.Rating*10) - int(left.Rating*10); order != 0 {
			return order
		}
		return slices.Compare([]string{left.ID}, []string{right.ID})
	})
	return accepted
}

func (e *Engine) suitable(request Request, blocked conflict, candidate place.Place) bool {
	if candidate.Hours.Known && candidate.Hours.OpenDuring(blocked.Item.StartAt, blocked.Item.EndAt) == place.Closed {
		return false
	}
	if e.weather == nil || !request.Weather.Known() {
		return true
	}
	impact := e.weather.Impact(request.Weather, blocked.Item.StartAt, weather.SensitivityOf(candidate.Category))
	return impact != weather.HighImpact
}

func withinWalkingLimit(request Request, centre geo.Point, candidate place.Place) bool {
	if !request.ReducedWalking || centre.IsZero() {
		return true
	}
	limit := request.MaximumWalkMeters
	if limit <= 0 {
		limit = nearRadiusMeters
	}
	return int(centre.DistanceMeters(candidate.Location)) <= limit
}

func (e *Engine) routeFor(ctx context.Context, request Request, from geo.Point, item itinerary.Item) route.Evidence {
	if e.routes == nil || from.IsZero() || item.Location.IsZero() {
		return route.UnavailableEvidence(from, item.Location, travelMode(request))
	}
	return e.routes.Compute(ctx, from, item.Location, travelMode(request), item.StartAt)
}

func travelMode(request Request) route.Mode {
	if request.ReducedWalking {
		return route.Drive
	}
	return route.Walk
}
