package itinerary

import (
	"context"
	"slices"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/place"
)

const CandidateCap = 24
const searchRadiusMeters = 15000

func categoriesFor(interests []string) []place.Category {
	categories := []place.Category{place.Attraction, place.Restaurant}
	for _, interest := range interests {
		switch interest {
		case "beach":
			categories = append(categories, place.Beach)
		case "food":
			categories = append(categories, place.Cafe, place.Market)
		case "culture", "history":
			categories = append(categories, place.Museum)
		case "nature":
			categories = append(categories, place.Park, place.Viewpoint)
		case "shopping":
			categories = append(categories, place.Shopping)
		case "nightlife":
			categories = append(categories, place.Nightlife)
		case "adventure":
			categories = append(categories, place.Experience)
		}
	}
	slices.Sort(categories)
	return slices.Compact(categories)
}

func (e *Engine) candidates(ctx context.Context, request Request) ([]place.Place, error) {
	if e.places == nil {
		return nil, ErrPlacesUnavailable
	}
	found, err := e.places.Nearby(ctx, NearbyRequest{
		Centre: request.Destination, RadiusMeters: searchRadiusMeters,
		Categories: categoriesFor(request.Interests), Language: request.Language, Limit: CandidateCap,
	})
	if err != nil {
		return nil, err
	}
	return prune(found, request), nil
}

func prune(found []place.Place, request Request) []place.Place {
	accepted := make([]place.Place, 0, len(found))
	for _, candidate := range found {
		if candidate.Validate() != nil || candidate.Status == place.Closed {
			continue
		}
		accepted = append(accepted, candidate)
	}
	slices.SortStableFunc(accepted, func(left, right place.Place) int {
		if order := interestRank(right, request.Interests) - interestRank(left, request.Interests); order != 0 {
			return order
		}
		if order := int(right.Rating*10) - int(left.Rating*10); order != 0 {
			return order
		}
		return slices.Compare([]string{left.ID}, []string{right.ID})
	})
	if len(accepted) > CandidateCap {
		accepted = accepted[:CandidateCap]
	}
	return accepted
}

var preferredByInterest = map[string][]place.Category{
	"beach":   {place.Beach},
	"food":    {place.Restaurant, place.Cafe},
	"culture": {place.Museum},
	"history": {place.Museum},
	"nature":  {place.Park, place.Viewpoint},
}

func interestRank(candidate place.Place, interests []string) int {
	for _, interest := range interests {
		if slices.Contains(preferredByInterest[interest], candidate.Category) {
			return 2
		}
	}
	if candidate.Category == place.Attraction {
		return 1
	}
	return 0
}
