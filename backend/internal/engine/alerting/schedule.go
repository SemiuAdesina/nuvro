package alerting

import (
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/alert"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/itinerary"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/travel"
)

const reminderWindow = 45 * time.Minute
const reminderBucket = 15 * time.Minute

type LeaveTime struct{}

func (LeaveTime) Name() string { return "leave_time" }

func (l LeaveTime) Evaluate(input Input) []alert.Alert {
	if input.NextItem == nil || !input.Departure.HasLeaveBy {
		return nil
	}
	remaining := input.Departure.LeaveBy.Sub(input.Now.UTC())
	if remaining > reminderWindow {
		return nil
	}
	if input.Departure.Status == travel.OnTime {
		return nil
	}
	value := base(input, alert.LeaveReminder, severityFor(input.Departure.Status),
		"alert.leave_reminder", input.NextItem.ID, alert.NavigateTo)
	value.MessageData = map[string]string{
		"activity":    input.NextItem.Title,
		"leaveBy":     input.Departure.LeaveBy.UTC().Format(time.RFC3339),
		"leaveStatus": string(input.Departure.Status),
	}
	value.ExpiresAt = input.NextItem.StartAt.UTC()
	value.Freshness = input.Departure.Route.Source.Freshness
	value.Fingerprint = alert.Fingerprint(input.TripID, alert.LeaveReminder, input.NextItem.ID, bucket(remaining))
	return []alert.Alert{value}
}

func severityFor(status travel.LeaveStatus) alert.Severity {
	switch status {
	case travel.LikelyLate, travel.LeaveNow:
		return alert.Important
	default:
		return alert.Attention
	}
}

func bucket(remaining time.Duration) string {
	if remaining < 0 {
		return "late"
	}
	return minutes(remaining.Truncate(reminderBucket))
}

type ScheduleConflict struct{}

func (ScheduleConflict) Name() string { return "schedule_conflict" }

func (s ScheduleConflict) Evaluate(input Input) []alert.Alert {
	constraints := itinerary.Constraints{BufferMinutes: 15}
	alerts := make([]alert.Alert, 0)
	for _, violation := range itinerary.Validate(input.Itinerary, constraints) {
		if !reportable(violation.Code) {
			continue
		}
		item, found := itemByID(input.Itinerary, violation.ItemID)
		if !found || item.EndAt.Before(input.Now.UTC()) {
			continue
		}
		value := base(input, alert.ScheduleConflict, alert.Attention, "alert.schedule_conflict", item.ID, alert.AdjustPlan)
		value.MessageData = map[string]string{"activity": item.Title, "issue": string(violation.Code)}
		value.ExpiresAt = item.EndAt.UTC()
		value.Fingerprint = alert.Fingerprint(input.TripID, alert.ScheduleConflict, item.ID, string(violation.Code))
		alerts = append(alerts, value)
	}
	return alerts
}

func reportable(code itinerary.ViolationCode) bool {
	switch code {
	case itinerary.Overlap, itinerary.TravelTime, itinerary.PlaceClosed:
		return true
	default:
		return false
	}
}

func itemByID(plan itinerary.Itinerary, id string) (itinerary.Item, bool) {
	for _, item := range plan.Items {
		if item.ID == id {
			return item, true
		}
	}
	return itinerary.Item{}, false
}
