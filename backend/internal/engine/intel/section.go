package intel

import (
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/evidence"
)

type Status string

const (
	Available     Status = "available"
	Unavailable   Status = "unavailable"
	NotApplicable Status = "not_applicable"
)

type Section struct {
	Status    Status
	Reason    string
	Freshness evidence.Freshness
}

func available(observedAt time.Time, validFor time.Duration) Section {
	return Section{
		Status: Available,
		Freshness: evidence.Freshness{
			ObservedAt: observedAt.UTC(), ValidUntil: observedAt.UTC().Add(validFor), Status: evidence.Live,
		},
	}
}

func unavailable(reason string) Section {
	return Section{Status: Unavailable, Reason: reason}
}
