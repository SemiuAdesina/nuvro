package journey

import (
	"context"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/apperror"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/geo"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/itinerary"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/place"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
	"github.com/SemiuAdesina/nuvro/backend/internal/engine/places"
)

const defaultRadiusMeters = 2500
const nearbyLimit = 20

type NearbyRequest struct {
	TripID           string
	Location         geo.Point
	Categories       []place.Category
	RadiusMeters     int
	OpenNow          bool
	AvailableMinutes int
	Language         string
	Limit            int
}

type NearbyResult struct {
	Places           []places.Ranked
	AvailableMinutes int
	NextItem         *itinerary.Item
}

func (s *Service) Nearby(ctx context.Context, principal user.Principal, request NearbyRequest) (NearbyResult, error) {
	if s.finder == nil {
		return NearbyResult{}, apperror.New(apperror.CodeUnavailable, "Place evidence is unavailable.")
	}
	if _, err := s.trips.Read(ctx, principal, request.TripID); err != nil {
		return NearbyResult{}, err
	}
	if err := request.Location.Validate(); err != nil {
		return NearbyResult{}, apperror.New(apperror.CodeInvalidInput, "A valid location is required.")
	}
	now := s.timer.Now().UTC()
	next, available := s.freeTime(ctx, request, now)
	found, err := s.finder.Nearby(ctx, places.NearbyQuery{
		Centre: request.Location, RadiusMeters: radius(request.RadiusMeters),
		Categories: request.Categories, OpenNow: request.OpenNow,
		Language: request.Language, Limit: limit(request.Limit),
	})
	if err != nil {
		return NearbyResult{}, apperror.Wrap(apperror.CodeUnavailable, "Place evidence is unavailable.", err)
	}
	ranked := places.Rank(found, places.Context{
		Now: now, Origin: request.Location, AvailableMinutes: available,
		Interests: s.interests(ctx, principal),
	})
	return NearbyResult{Places: ranked, AvailableMinutes: available, NextItem: next}, nil
}

func (s *Service) freeTime(ctx context.Context, request NearbyRequest, now time.Time) (*itinerary.Item, int) {
	if request.AvailableMinutes > 0 {
		return nil, request.AvailableMinutes
	}
	plan, err := s.store.Itinerary(ctx, request.TripID)
	if err != nil {
		return nil, 0
	}
	var next *itinerary.Item
	for index := range plan.Items {
		item := plan.Items[index]
		if item.StartAt.Before(now) || item.Status == itinerary.Cancelled {
			continue
		}
		if next == nil || item.StartAt.Before(next.StartAt) {
			candidate := item
			next = &candidate
		}
	}
	if next == nil {
		return nil, 0
	}
	return next, int(next.StartAt.Sub(now).Minutes())
}

func radius(requested int) int {
	if requested < 100 || requested > 20000 {
		return defaultRadiusMeters
	}
	return requested
}

func limit(requested int) int {
	if requested < 1 || requested > nearbyLimit {
		return nearbyLimit
	}
	return requested
}
