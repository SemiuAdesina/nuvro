package travel_test

import (
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/itinerary"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/travel"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/trip"
)

func planWith(items ...itinerary.Item) itinerary.Itinerary {
	return itinerary.Itinerary{TripID: "trip", Timezone: "Atlantic/Madeira", Items: items}
}

func item(id string, dayNumber, startHour int, status itinerary.Status) itinerary.Item {
	value := activity(startHour)
	value.ID = id
	value.Day = dayNumber
	value.Status = status
	return value
}

func madeiraTrip() trip.Trip {
	start := day.AddDate(0, 0, -2)
	end := day.AddDate(0, 0, 4)
	destinationID := "madeira-pt"
	return trip.Trip{
		ID: "trip", OwnerID: "owner", DestinationID: &destinationID,
		StartDate: &start, EndDate: &end, Adults: 2, Currency: "EUR", Status: trip.Active,
	}
}

func TestNextItemSkipsFinishedAndCancelledItems(t *testing.T) {
	plan := planWith(
		item("done", 3, 9, itinerary.Completed),
		item("skipped", 3, 11, itinerary.Skipped),
		item("cancelled", 3, 13, itinerary.Cancelled),
		item("upcoming", 3, 15, itinerary.Planned),
	)
	next := travel.NextItem(plan, at(10, 0))
	if next == nil || next.ID != "upcoming" {
		t.Fatalf("next = %+v", next)
	}
}

func TestNextItemIgnoresArrayOrderAndPastItems(t *testing.T) {
	plan := planWith(
		item("later", 3, 18, itinerary.Planned),
		item("soonest", 3, 15, itinerary.Planned),
		item("past", 3, 6, itinerary.Planned),
	)
	next := travel.NextItem(plan, at(12, 0))
	if next == nil || next.ID != "soonest" {
		t.Fatalf("next = %+v", next)
	}
	if travel.NextItem(planWith(item("past", 3, 6, itinerary.Planned)), at(23, 0)) != nil {
		t.Fatal("a finished day produced a next item")
	}
}

func TestAnItemInProgressIsStillTheNextItem(t *testing.T) {
	plan := planWith(item("current", 3, 9, itinerary.Planned))
	next := travel.NextItem(plan, at(10, 0))
	if next == nil || next.ID != "current" {
		t.Fatal("an activity already under way was dropped")
	}
}

func TestTripPhaseAndDayNumberUseLocalDates(t *testing.T) {
	record := madeiraTrip()
	if phase := travel.PhaseOf(record, at(12, 0)); phase != travel.DuringTrip {
		t.Fatalf("phase = %q", phase)
	}
	if travel.PhaseOf(record, day.AddDate(0, 0, -5)) != travel.BeforeTrip {
		t.Fatal("a date before the trip was not reported as before")
	}
	if travel.PhaseOf(record, day.AddDate(0, 0, 10)) != travel.AfterTrip {
		t.Fatal("a date after the trip was not reported as after")
	}
	if number := travel.DayNumber(record, at(12, 0)); number != 3 {
		t.Fatalf("day = %d, want 3", number)
	}
	incomplete := record
	incomplete.StartDate = nil
	if travel.PhaseOf(incomplete, at(12, 0)) != travel.BeforeTrip || travel.DayNumber(incomplete, at(12, 0)) != 0 {
		t.Fatal("a trip without dates was given a phase or day")
	}
}

func TestTodayIsOrderedAndScopedToTheDay(t *testing.T) {
	plan := planWith(
		item("evening", 3, 19, itinerary.Planned),
		item("morning", 3, 9, itinerary.Planned),
		item("tomorrow", 4, 10, itinerary.Planned),
	)
	today := travel.ItemsOnDay(plan, 3, time.UTC)
	if len(today) != 2 || today[0].ID != "morning" || today[1].ID != "evening" {
		t.Fatalf("today = %v", today)
	}
}

func TestDestinationTimezoneIsUsedNotTheCallerTimezone(t *testing.T) {
	location := travel.LocationOf("Atlantic/Madeira")
	if location == time.UTC {
		t.Fatal("a valid IANA identifier was not loaded")
	}
	if travel.LocationOf("UTC+1") != time.UTC {
		t.Fatal("an offset string was accepted as a timezone")
	}
	if travel.LocationOf("") != time.UTC {
		t.Fatal("an empty timezone did not fall back to UTC")
	}
}

func TestQuickActionsReflectWhatIsActuallyAvailable(t *testing.T) {
	next := item("upcoming", 3, 15, itinerary.Planned)
	actions := travel.ActionsFor(&next, false)
	byType := map[travel.ActionType]travel.Action{}
	for _, action := range actions {
		byType[action.Type] = action
	}
	if !byType[travel.Navigate].Enabled || byType[travel.Navigate].ResourceID != next.ID {
		t.Fatalf("navigate action = %+v", byType[travel.Navigate])
	}
	if byType[travel.ViewTicket].Enabled {
		t.Fatal("a ticket action was offered with no document")
	}
	if !byType[travel.Translate].Enabled || !byType[travel.ReviewPlan].Enabled {
		t.Fatal("always available actions were disabled")
	}
	without := travel.ActionsFor(nil, false)
	for _, action := range without {
		if action.Type == travel.Navigate && action.Enabled {
			t.Fatal("navigation was offered with no next activity")
		}
	}
}
