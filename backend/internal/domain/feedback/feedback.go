package feedback

import (
	"errors"
	"slices"
	"time"
)

var ErrInvalid = errors.New("feedback: invalid submission")

type Aspect string

const (
	CoastalPlaces     Aspect = "coastal_places"
	LocalFood         Aspect = "local_food"
	CultureAndHistory Aspect = "culture_and_history"
	NatureAndOutdoors Aspect = "nature_and_outdoors"
	Nightlife         Aspect = "nightlife"
	QuietArea         Aspect = "quiet_area"
	CentralArea       Aspect = "central_area"
	LongTransfers     Aspect = "long_transfers"
	BusySchedule      Aspect = "busy_schedule"
)

func Aspects() []Aspect {
	return []Aspect{
		CoastalPlaces, LocalFood, CultureAndHistory, NatureAndOutdoors, Nightlife,
		QuietArea, CentralArea, LongTransfers, BusySchedule,
	}
}

func (a Aspect) Valid() bool { return slices.Contains(Aspects(), a) }

type PaceAssessment string

const (
	TooSlow   PaceAssessment = "too_slow"
	JustRight PaceAssessment = "just_right"
	TooBusy   PaceAssessment = "too_busy"
)

func (p PaceAssessment) Valid() bool {
	switch p {
	case TooSlow, JustRight, TooBusy, "":
		return true
	default:
		return false
	}
}

type BurdenAssessment string

const (
	LightTravel       BurdenAssessment = "light"
	ComfortableTravel BurdenAssessment = "comfortable"
	TooMuchTravel     BurdenAssessment = "too_much"
)

func (b BurdenAssessment) Valid() bool {
	switch b {
	case LightTravel, ComfortableTravel, TooMuchTravel, "":
		return true
	default:
		return false
	}
}

type ReturnIntent string

const (
	WouldReturn    ReturnIntent = "would_return"
	Undecided      ReturnIntent = "undecided"
	WouldNotReturn ReturnIntent = "would_not_return"
)

func (r ReturnIntent) Valid() bool {
	switch r {
	case WouldReturn, Undecided, WouldNotReturn, "":
		return true
	default:
		return false
	}
}

type TripFeedback struct {
	TripID       string
	UserID       string
	Positives    []Aspect
	Negatives    []Aspect
	Pace         PaceAssessment
	TravelBurden BurdenAssessment
	ReturnIntent ReturnIntent
	SubmittedAt  time.Time
	UpdatedAt    time.Time
}

const maximumAspects = 6

func (f TripFeedback) Validate() error {
	if f.TripID == "" || f.UserID == "" || f.SubmittedAt.IsZero() {
		return ErrInvalid
	}
	if !f.Pace.Valid() || !f.TravelBurden.Valid() || !f.ReturnIntent.Valid() {
		return ErrInvalid
	}
	if err := validAspects(f.Positives); err != nil {
		return err
	}
	if err := validAspects(f.Negatives); err != nil {
		return err
	}
	for _, positive := range f.Positives {
		if slices.Contains(f.Negatives, positive) {
			return ErrInvalid
		}
	}
	return nil
}

func validAspects(aspects []Aspect) error {
	if len(aspects) > maximumAspects {
		return ErrInvalid
	}
	seen := make(map[Aspect]struct{}, len(aspects))
	for _, aspect := range aspects {
		if !aspect.Valid() {
			return ErrInvalid
		}
		if _, duplicate := seen[aspect]; duplicate {
			return ErrInvalid
		}
		seen[aspect] = struct{}{}
	}
	return nil
}

func (f TripFeedback) Empty() bool {
	return len(f.Positives) == 0 && len(f.Negatives) == 0 &&
		f.Pace == "" && f.TravelBurden == "" && f.ReturnIntent == ""
}
