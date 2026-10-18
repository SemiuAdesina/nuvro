package personalization

import (
	"slices"
	"strings"
	"time"
)

type Confidence string

const (
	Low              Confidence = "low"
	MediumConfidence Confidence = "medium"
	High             Confidence = "high"
)

func (c Confidence) Valid() bool {
	switch c {
	case Low, MediumConfidence, High:
		return true
	default:
		return false
	}
}

func (c Confidence) Rank() int {
	switch c {
	case High:
		return 3
	case MediumConfidence:
		return 2
	default:
		return 1
	}
}

type Status string

const (
	Active     Status = "active"
	Confirmed  Status = "confirmed"
	Dismissed  Status = "dismissed"
	Suppressed Status = "suppressed"
)

func (s Status) Valid() bool {
	switch s {
	case Active, Confirmed, Dismissed, Suppressed:
		return true
	default:
		return false
	}
}

func (s Status) Influences() bool { return s == Active || s == Confirmed }

func (s Status) CanTransitionTo(next Status) bool {
	if !next.Valid() {
		return false
	}
	switch s {
	case Active:
		return next == Confirmed || next == Dismissed || next == Suppressed
	case Confirmed:
		return next == Dismissed || next == Suppressed
	case Dismissed, Suppressed:
		return next == Active
	default:
		return false
	}
}

type Key string

const (
	InterestCoastal   Key = "interest.coastal"
	InterestFood      Key = "interest.food"
	InterestCulture   Key = "interest.culture"
	InterestNature    Key = "interest.nature"
	InterestNightlife Key = "interest.nightlife"
	StayQuietArea     Key = "stay.quiet_neighbourhood"
	StayCentralArea   Key = "stay.central_neighbourhood"
	FlightFewerStops  Key = "flight.fewer_stops"
	PaceRelaxed       Key = "pace.relaxed"
	PacePacked        Key = "pace.packed"
	TransferTolerance Key = "transfer.low_tolerance"
)

func Keys() []Key {
	return []Key{
		InterestCoastal, InterestFood, InterestCulture, InterestNature, InterestNightlife,
		StayQuietArea, StayCentralArea, FlightFewerStops, PaceRelaxed, PacePacked, TransferTolerance,
	}
}

func (k Key) Valid() bool { return slices.Contains(Keys(), k) }

func (k Key) Category() string {
	category, _, found := strings.Cut(string(k), ".")
	if !found {
		return "other"
	}
	return category
}

func (k Key) Opposes(other Key) bool {
	pairs := map[Key]Key{
		StayQuietArea: StayCentralArea, StayCentralArea: StayQuietArea,
		PaceRelaxed: PacePacked, PacePacked: PaceRelaxed,
	}
	return pairs[k] == other
}

type ExplanationCode string

const (
	RepeatedSaves      ExplanationCode = "repeated_destination_saves"
	RepeatedChoices    ExplanationCode = "repeated_activity_choice"
	PostTripFeedback   ExplanationCode = "post_trip_feedback"
	UserConfirmation   ExplanationCode = "explicit_confirmation"
	RepeatedReplanning ExplanationCode = "repeated_replan_pattern"
)

type LearnedPreference struct {
	ID              string
	UserID          string
	Key             Key
	Confidence      Confidence
	Score           int
	EvidenceCount   int
	Explanation     ExplanationCode
	Status          Status
	RulesVersion    string
	FirstObservedAt time.Time
	LastObservedAt  time.Time
	UpdatedAt       time.Time
}

func (p LearnedPreference) Validate() error {
	if p.UserID == "" || !p.Key.Valid() || !p.Confidence.Valid() || !p.Status.Valid() {
		return ErrInvalid
	}
	if p.EvidenceCount < 0 || p.Score < 0 || p.RulesVersion == "" {
		return ErrInvalid
	}
	if p.FirstObservedAt.IsZero() || p.LastObservedAt.Before(p.FirstObservedAt) {
		return ErrInvalid
	}
	return nil
}

func (p LearnedPreference) Influences() bool { return p.Status.Influences() }

type Suggestion struct {
	ID            string
	UserID        string
	Key           Key
	Confidence    Confidence
	EvidenceCount int
	Status        Status
	CreatedAt     time.Time
	ResolvedAt    time.Time
}
