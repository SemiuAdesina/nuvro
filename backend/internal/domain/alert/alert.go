package alert

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"slices"
	"strings"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/evidence"
)

var ErrInvalid = errors.New("alert: invalid alert")

type Type string

const (
	WeatherImpact       Type = "weather_impact"
	LeaveReminder       Type = "leave_reminder"
	ScheduleConflict    Type = "schedule_conflict"
	PlaceChange         Type = "place_change"
	ReservationReminder Type = "reservation_reminder"
	DocumentReminder    Type = "document_reminder"
	EntryReminder       Type = "entry_reminder"
	PriceDropped        Type = "price_dropped"
	PriceIncreased      Type = "price_increased"
	AvailabilityChanged Type = "availability_changed"
)

func Types() []Type {
	return []Type{
		WeatherImpact, LeaveReminder, ScheduleConflict, PlaceChange,
		ReservationReminder, DocumentReminder, EntryReminder,
		PriceDropped, PriceIncreased, AvailabilityChanged,
	}
}

func (t Type) Valid() bool { return slices.Contains(Types(), t) }

type Severity string

const (
	Info      Severity = "info"
	Attention Severity = "attention"
	Important Severity = "important"
)

func (s Severity) Valid() bool {
	switch s {
	case Info, Attention, Important:
		return true
	default:
		return false
	}
}

type Status string

const (
	Unread   Status = "unread"
	Read     Status = "read"
	Resolved Status = "resolved"
	Expired  Status = "expired"
)

func (s Status) Valid() bool {
	switch s {
	case Unread, Read, Resolved, Expired:
		return true
	default:
		return false
	}
}

func (s Status) CanTransitionTo(next Status) bool {
	if !next.Valid() {
		return false
	}
	switch s {
	case Unread:
		return next == Read || next == Resolved
	case Read:
		return next == Resolved
	default:
		return false
	}
}

type Action string

const (
	AdjustPlan   Action = "adjust_plan"
	NavigateTo   Action = "navigate"
	ViewActivity Action = "view_activity"
	ViewDocument Action = "view_document"
	ViewOffer    Action = "view_offer"
	NoAction     Action = "none"
)

func (a Action) Valid() bool {
	switch a {
	case AdjustPlan, NavigateTo, ViewActivity, ViewDocument, ViewOffer, NoAction:
		return true
	default:
		return false
	}
}

type Alert struct {
	ID            string
	TripID        string
	Type          Type
	Severity      Severity
	TitleCode     string
	MessageData   map[string]string
	RelatedItemID string
	Action        Action
	Status        Status
	Fingerprint   string
	CreatedAt     time.Time
	EffectiveAt   time.Time
	ExpiresAt     time.Time
	Freshness     evidence.Freshness
}

func (a Alert) Validate() error {
	if a.TripID == "" || a.TitleCode == "" || a.Fingerprint == "" || a.EffectiveAt.IsZero() {
		return ErrInvalid
	}
	if !a.vocabularyValid() {
		return ErrInvalid
	}
	if !a.ExpiresAt.IsZero() && !a.ExpiresAt.After(a.EffectiveAt) {
		return ErrInvalid
	}
	return a.dataValid()
}

func (a Alert) vocabularyValid() bool {
	return a.Type.Valid() && a.Severity.Valid() && a.Status.Valid() && a.Action.Valid()
}

func (a Alert) dataValid() error {
	for key, value := range a.MessageData {
		if key == "" || len(value) > 200 {
			return ErrInvalid
		}
	}
	return nil
}

func (a Alert) ExpiredAt(now time.Time) bool {
	return !a.ExpiresAt.IsZero() && !a.ExpiresAt.After(now.UTC())
}

func (a Alert) Visible(now time.Time) bool {
	if a.Status == Resolved || a.Status == Expired {
		return false
	}
	return !a.ExpiredAt(now) && !a.EffectiveAt.After(now.UTC())
}

func Fingerprint(tripID string, kind Type, resourceID string, window string) string {
	parts := []string{tripID, string(kind), resourceID, window}
	sum := sha256.Sum256([]byte(strings.Join(parts, "|")))
	return hex.EncodeToString(sum[:16])
}
