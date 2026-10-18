package personalization

import (
	"errors"
	"slices"
	"time"
)

var ErrInvalid = errors.New("personalization: invalid record")

const RulesVersion = "personalization-v1"

type SignalType string

const (
	DestinationSaved      SignalType = "destination_saved"
	DestinationUnsaved    SignalType = "destination_unsaved"
	DestinationSelected   SignalType = "destination_selected"
	ActivityAdded         SignalType = "activity_added"
	ActivityRemoved       SignalType = "activity_removed"
	ActivityCompleted     SignalType = "activity_completed"
	StaySelected          SignalType = "stay_selected"
	NeighbourhoodSelected SignalType = "neighbourhood_selected"
	ReplanRequested       SignalType = "replan_requested"
	TripCompleted         SignalType = "trip_completed"
	FeedbackPositive      SignalType = "feedback_positive"
	FeedbackNegative      SignalType = "feedback_negative"
	ExplicitConfirmation  SignalType = "explicit_confirmation"
)

func SignalTypes() []SignalType {
	return []SignalType{
		DestinationSaved, DestinationUnsaved, DestinationSelected,
		ActivityAdded, ActivityRemoved, ActivityCompleted,
		StaySelected, NeighbourhoodSelected, ReplanRequested, TripCompleted,
		FeedbackPositive, FeedbackNegative, ExplicitConfirmation,
	}
}

func (s SignalType) Valid() bool { return slices.Contains(SignalTypes(), s) }

func (s SignalType) Negative() bool {
	switch s {
	case DestinationUnsaved, ActivityRemoved, FeedbackNegative:
		return true
	default:
		return false
	}
}

type Strength string

const (
	Weak   Strength = "weak"
	Medium Strength = "medium"
	Strong Strength = "strong"
)

func (s Strength) Valid() bool {
	switch s {
	case Weak, Medium, Strong:
		return true
	default:
		return false
	}
}

func (s Strength) Weight() int {
	switch s {
	case Strong:
		return 6
	case Medium:
		return 3
	default:
		return 1
	}
}

func StrengthOf(signalType SignalType) Strength {
	switch signalType {
	case ExplicitConfirmation, FeedbackPositive, FeedbackNegative:
		return Strong
	case DestinationSelected, ActivityCompleted, StaySelected, NeighbourhoodSelected, TripCompleted:
		return Medium
	default:
		return Weak
	}
}

type SubjectType string

const (
	Destination  SubjectType = "destination"
	ActivityKind SubjectType = "activity_category"
	StayArea     SubjectType = "stay_area"
	TripPace     SubjectType = "trip_pace"
	TravelBurden SubjectType = "travel_burden"
	Interest     SubjectType = "interest"
)

func (s SubjectType) Valid() bool {
	switch s {
	case Destination, ActivityKind, StayArea, TripPace, TravelBurden, Interest:
		return true
	default:
		return false
	}
}

type Signal struct {
	ID          string
	UserID      string
	Type        SignalType
	Subject     SubjectType
	SubjectID   string
	TripID      string
	Strength    Strength
	OccurredAt  time.Time
	SourceEvent string
}

func (s Signal) Valid() bool { return s.Validate() == nil }

func (s Signal) Validate() error {
	if s.UserID == "" || s.SubjectID == "" || s.OccurredAt.IsZero() {
		return ErrInvalid
	}
	if !s.Type.Valid() || !s.Subject.Valid() || !s.Strength.Valid() {
		return ErrInvalid
	}
	if len(s.SubjectID) > 120 || len(s.SourceEvent) > 200 {
		return ErrInvalid
	}
	return nil
}

func NewSignal(userID string, signalType SignalType, subject SubjectType, subjectID string, occurredAt time.Time) Signal {
	return Signal{
		UserID: userID, Type: signalType, Subject: subject, SubjectID: subjectID,
		Strength: StrengthOf(signalType), OccurredAt: occurredAt.UTC(),
	}
}
