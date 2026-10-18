package replanning

import (
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/itinerary"
)

const slotSearchDays = 3
const dayOpensMinutes = 9 * 60
const dayClosesMinutes = 20 * 60
const slotBuffer = 30 * time.Minute

func laterSlot(request Request, item itinerary.Item) (int, time.Time, bool) {
	location := locationOf(request.Timezone)
	duration := item.EndAt.Sub(item.StartAt)
	lastDay := lastDayNumber(request)
	for day := item.Day + 1; day <= lastDay && day <= item.Day+slotSearchDays; day++ {
		if start, found := openingOn(request, day, duration, location); found {
			return day, start, true
		}
	}
	return 0, time.Time{}, false
}

func openingOn(request Request, day int, duration time.Duration, location *time.Location) (time.Time, bool) {
	reference, found := dayReference(request, day, location)
	if !found {
		return time.Time{}, false
	}
	cursor := time.Date(reference.Year(), reference.Month(), reference.Day(), 0, dayOpensMinutes, 0, 0, location)
	closes := time.Date(reference.Year(), reference.Month(), reference.Day(), 0, dayClosesMinutes, 0, 0, location)
	for cursor.Add(duration).Before(closes) || cursor.Add(duration).Equal(closes) {
		if free(request, day, cursor, cursor.Add(duration)) {
			return cursor.UTC(), true
		}
		cursor = cursor.Add(slotBuffer)
	}
	return time.Time{}, false
}

func dayReference(request Request, day int, location *time.Location) (time.Time, bool) {
	for _, item := range request.Itinerary.Items {
		if item.Day == day {
			return item.StartAt.In(location), true
		}
	}
	for _, item := range request.Itinerary.Items {
		if item.Day < day {
			return item.StartAt.In(location).AddDate(0, 0, day-item.Day), true
		}
	}
	return time.Time{}, false
}

func free(request Request, day int, start, end time.Time) bool {
	for _, item := range request.Itinerary.Items {
		if item.Day != day {
			continue
		}
		if start.Before(item.EndAt.Add(slotBuffer)) && item.StartAt.Add(-slotBuffer).Before(end) {
			return false
		}
	}
	return true
}

func lastDayNumber(request Request) int {
	last := 0
	for _, item := range request.Itinerary.Items {
		if item.Day > last {
			last = item.Day
		}
	}
	return last
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
