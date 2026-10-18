package travel

import (
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/itinerary"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/place"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/route"
)

type LeaveStatus string

const (
	OnTime           LeaveStatus = "on_time"
	LeaveSoon        LeaveStatus = "leave_soon"
	LeaveNow         LeaveStatus = "leave_now"
	LikelyLate       LeaveStatus = "likely_late"
	RouteUnavailable LeaveStatus = "route_unavailable"
	NoUpcomingItem   LeaveStatus = "no_upcoming_item"
)

type BufferPolicy struct {
	Default    time.Duration
	Meal       time.Duration
	Reserved   time.Duration
	Transfer   time.Duration
	SoonWindow time.Duration
	NowWindow  time.Duration
}

func DefaultBufferPolicy() BufferPolicy {
	return BufferPolicy{
		Default:    5 * time.Minute,
		Meal:       8 * time.Minute,
		Reserved:   15 * time.Minute,
		Transfer:   10 * time.Minute,
		SoonWindow: 30 * time.Minute,
		NowWindow:  5 * time.Minute,
	}
}

func (p BufferPolicy) For(item itinerary.Item) time.Duration {
	if item.ReservationLocked || item.Status == itinerary.Reserved {
		return p.Reserved
	}
	switch item.Kind {
	case itinerary.Meal:
		return p.Meal
	case itinerary.Transfer:
		return p.Transfer
	default:
		return arrivalFor(item.Category, p.Default)
	}
}

func arrivalFor(category place.Category, fallback time.Duration) time.Duration {
	switch category {
	case place.Museum, place.Attraction, place.Experience:
		return 10 * time.Minute
	case place.Transport:
		return 15 * time.Minute
	default:
		return fallback
	}
}

type Departure struct {
	Status            LeaveStatus
	LeaveBy           time.Time
	HasLeaveBy        bool
	MinutesUntilLeave int
	ArrivalBuffer     time.Duration
	Route             route.Evidence
	HasRoute          bool
}

func EvaluateDeparture(now time.Time, item *itinerary.Item, evidence route.Evidence, policy BufferPolicy) Departure {
	if item == nil {
		return Departure{Status: NoUpcomingItem, MinutesUntilLeave: -1}
	}
	buffer := policy.For(*item)
	if !evidence.Known() {
		return Departure{Status: RouteUnavailable, MinutesUntilLeave: -1, ArrivalBuffer: buffer, Route: evidence}
	}
	leaveBy := item.StartAt.Add(-evidence.Duration()).Add(-buffer)
	remaining := leaveBy.Sub(now.UTC())
	return Departure{
		Status: statusFor(remaining, policy), LeaveBy: leaveBy.UTC(), HasLeaveBy: true,
		MinutesUntilLeave: int(remaining.Round(time.Minute).Minutes()),
		ArrivalBuffer:     buffer, Route: evidence, HasRoute: true,
	}
}

func statusFor(remaining time.Duration, policy BufferPolicy) LeaveStatus {
	switch {
	case remaining < 0:
		return LikelyLate
	case remaining <= policy.NowWindow:
		return LeaveNow
	case remaining <= policy.SoonWindow:
		return LeaveSoon
	default:
		return OnTime
	}
}
