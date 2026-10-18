package trip

import (
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/change"
)

type Dates struct {
	Start time.Time
	End   time.Time
}

type Patch struct {
	DestinationID change.Change[string]
	OriginAirport change.Change[string]
	Dates         change.Change[Dates]
	Adults        change.Change[int32]
	Children      change.Change[int32]
	Currency      change.Change[string]
}

func (t Trip) Apply(patch Patch, at time.Time) (Trip, error) {
	updated := t
	patch.DestinationID.ApplyToOptional(&updated.DestinationID)
	patch.OriginAirport.ApplyToOptional(&updated.OriginAirport)
	applyDates(&updated, patch.Dates)
	if err := patch.Adults.ApplyTo(&updated.Adults); err != nil {
		return Trip{}, ErrInvalid
	}
	if err := patch.Children.ApplyTo(&updated.Children); err != nil {
		return Trip{}, ErrInvalid
	}
	if err := patch.Currency.ApplyTo(&updated.Currency); err != nil {
		return Trip{}, ErrInvalid
	}
	if err := updated.Validate(); err != nil {
		return Trip{}, err
	}
	updated.UpdatedAt = at.UTC()
	return updated, nil
}

func applyDates(target *Trip, dates change.Change[Dates]) {
	if dates.IsUnchanged() {
		return
	}
	if dates.IsCleared() {
		target.StartDate = nil
		target.EndDate = nil
		return
	}
	value, _ := dates.Value()
	start := value.Start.UTC()
	end := value.End.UTC()
	target.StartDate = &start
	target.EndDate = &end
}
