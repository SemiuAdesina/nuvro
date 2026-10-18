package itinerary

import (
	"slices"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/place"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/weather"
)

type ViolationCode string

const (
	Overlap          ViolationCode = "overlapping_items"
	TravelTime       ViolationCode = "travel_time_not_feasible"
	OutsideTripDates ViolationCode = "outside_trip_dates"
	PlaceClosed      ViolationCode = "place_closed"
	WeatherBlocked   ViolationCode = "weather_blocked"
	MissingPlace     ViolationCode = "place_not_resolved"
	InvalidItem      ViolationCode = "invalid_item"
	LockedItemMoved  ViolationCode = "locked_item_moved"
)

type Violation struct {
	Code   ViolationCode
	ItemID string
	Detail string
}

type Constraints struct {
	Start         time.Time
	End           time.Time
	BufferMinutes int
	Openings      map[string]place.OpeningHours
	Locked        map[string]Item
}

func Validate(plan Itinerary, constraints Constraints) []Violation {
	violations := make([]Violation, 0)
	violations = append(violations, itemViolations(plan, constraints)...)
	for _, day := range plan.Days() {
		violations = append(violations, dayViolations(day, constraints)...)
	}
	slices.SortStableFunc(violations, func(left, right Violation) int {
		if order := int(left.Code[0]) - int(right.Code[0]); order != 0 {
			return order
		}
		return len(left.ItemID) - len(right.ItemID)
	})
	return violations
}

func itemViolations(plan Itinerary, constraints Constraints) []Violation {
	violations := make([]Violation, 0)
	for _, item := range plan.Items {
		if err := item.Validate(); err != nil {
			violations = append(violations, Violation{Code: InvalidItem, ItemID: item.ID, Detail: err.Error()})
			continue
		}
		if outsideWindow(item, constraints) {
			violations = append(violations, Violation{Code: OutsideTripDates, ItemID: item.ID})
		}
		if item.Kind == Activity && item.PlaceID == "" {
			violations = append(violations, Violation{Code: MissingPlace, ItemID: item.ID})
		}
		if item.WeatherImpact == weather.HighImpact {
			violations = append(violations, Violation{Code: WeatherBlocked, ItemID: item.ID})
		}
		violations = append(violations, closureViolation(item, constraints)...)
		violations = append(violations, lockViolation(item, constraints)...)
	}
	return violations
}

func outsideWindow(item Item, constraints Constraints) bool {
	if constraints.Start.IsZero() || constraints.End.IsZero() {
		return false
	}
	return item.StartAt.Before(constraints.Start) || item.EndAt.After(constraints.End)
}

func closureViolation(item Item, constraints Constraints) []Violation {
	hours, found := constraints.Openings[item.PlaceID]
	if !found || item.PlaceID == "" {
		return nil
	}
	if hours.OpenDuring(item.StartAt, item.EndAt) == place.Closed {
		return []Violation{{Code: PlaceClosed, ItemID: item.ID}}
	}
	return nil
}

func lockViolation(item Item, constraints Constraints) []Violation {
	original, found := constraints.Locked[item.ID]
	if !found {
		return nil
	}
	if !original.StartAt.Equal(item.StartAt) || !original.EndAt.Equal(item.EndAt) {
		return []Violation{{Code: LockedItemMoved, ItemID: item.ID}}
	}
	return nil
}

func dayViolations(day Day, constraints Constraints) []Violation {
	violations := make([]Violation, 0)
	buffer := time.Duration(constraints.BufferMinutes) * time.Minute
	for index := 1; index < len(day.Items); index++ {
		previous := day.Items[index-1]
		current := day.Items[index]
		if previous.Overlaps(current) {
			violations = append(violations, Violation{Code: Overlap, ItemID: current.ID})
			continue
		}
		if !travelFits(previous, current, buffer) {
			violations = append(violations, Violation{Code: TravelTime, ItemID: current.ID})
		}
	}
	return violations
}

func travelFits(previous, current Item, buffer time.Duration) bool {
	if current.RouteFromPrevious == nil || !current.RouteFromPrevious.Known() {
		return true
	}
	required := current.RouteFromPrevious.Duration() + buffer
	return current.StartAt.Sub(previous.EndAt) >= required
}
