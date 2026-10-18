package travel

import (
	"slices"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/evidence"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/itinerary"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/trip"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/weather"
)

type Phase string

const (
	BeforeTrip Phase = "before_trip"
	DuringTrip Phase = "during_trip"
	AfterTrip  Phase = "after_trip"
)

type ActionType string

const (
	Navigate   ActionType = "navigate"
	ViewTicket ActionType = "ticket"
	Translate  ActionType = "translate"
	ReviewPlan ActionType = "review_plan"
)

type Action struct {
	Type       ActionType
	ResourceID string
	Enabled    bool
}

type Snapshot struct {
	Trip           trip.Trip
	Timezone       string
	LocalNow       time.Time
	Phase          Phase
	CurrentDay     int
	Today          []itinerary.Item
	NextItem       *itinerary.Item
	Departure      Departure
	Weather        weather.Window
	WeatherSection evidence.Freshness
	HasWeather     bool
	Actions        []Action
	GeneratedAt    time.Time
}

func LocationOf(timezone string) *time.Location {
	if timezone == "" {
		return time.UTC
	}
	location, err := time.LoadLocation(timezone)
	if err != nil {
		return time.UTC
	}
	return location
}

func PhaseOf(record trip.Trip, localNow time.Time) Phase {
	if record.StartDate == nil || record.EndDate == nil {
		return BeforeTrip
	}
	day := localNow.Format(time.DateOnly)
	switch {
	case day < record.StartDate.UTC().Format(time.DateOnly):
		return BeforeTrip
	case day > record.EndDate.UTC().Format(time.DateOnly):
		return AfterTrip
	default:
		return DuringTrip
	}
}

func DayNumber(record trip.Trip, localNow time.Time) int {
	if record.StartDate == nil {
		return 0
	}
	start := record.StartDate.UTC().Truncate(24 * time.Hour)
	current := time.Date(localNow.Year(), localNow.Month(), localNow.Day(), 0, 0, 0, 0, time.UTC)
	days := int(current.Sub(start).Hours()/24) + 1
	if days < 1 {
		return 0
	}
	return days
}

func ItemsOnDay(plan itinerary.Itinerary, day int, location *time.Location) []itinerary.Item {
	items := make([]itinerary.Item, 0, len(plan.Items))
	for _, item := range plan.Items {
		if item.Day == day {
			items = append(items, item)
		}
	}
	slices.SortStableFunc(items, func(left, right itinerary.Item) int {
		return left.StartAt.Compare(right.StartAt)
	})
	return items
}

func NextItem(plan itinerary.Itinerary, now time.Time) *itinerary.Item {
	var next *itinerary.Item
	for index := range plan.Items {
		candidate := plan.Items[index]
		if !upcoming(candidate, now) {
			continue
		}
		if next == nil || candidate.StartAt.Before(next.StartAt) ||
			(candidate.StartAt.Equal(next.StartAt) && candidate.ID < next.ID) {
			chosen := candidate
			next = &chosen
		}
	}
	return next
}

func upcoming(item itinerary.Item, now time.Time) bool {
	switch item.Status {
	case itinerary.Completed, itinerary.Skipped, itinerary.Cancelled:
		return false
	}
	return item.EndAt.After(now.UTC())
}

func ActionsFor(next *itinerary.Item, hasDocument bool) []Action {
	actions := []Action{
		{Type: Navigate, Enabled: next != nil && !next.Location.IsZero()},
		{Type: ViewTicket, Enabled: hasDocument},
		{Type: Translate, Enabled: true},
		{Type: ReviewPlan, Enabled: true},
	}
	if next != nil {
		actions[0].ResourceID = next.ID
	}
	return actions
}
