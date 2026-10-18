package place

import (
	"errors"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/destination"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/evidence"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/geo"
)

var ErrInvalid = errors.New("place: invalid place")

type Reference struct {
	Provider   string
	ExternalID string
}

type Place struct {
	ID             string
	References     []Reference
	Name           string
	Category       Category
	RawTypes       []string
	Location       geo.Point
	Address        string
	Timezone       string
	Rating         float64
	ReviewCount    int
	PriceLevel     int
	Status         OpeningStatus
	Hours          OpeningHours
	Media          []destination.Media
	Attribution    []evidence.Attribution
	Freshness      evidence.Freshness
	TypicalMinutes int
}

func (p Place) Validate() error {
	if p.ID == "" || p.Name == "" || !p.Category.Valid() {
		return ErrInvalid
	}
	if err := p.Location.Validate(); err != nil {
		return ErrInvalid
	}
	if p.Rating < 0 || p.Rating > 5 || p.ReviewCount < 0 || p.PriceLevel < 0 || p.PriceLevel > 4 {
		return ErrInvalid
	}
	switch p.Status {
	case Open, Closed, UnknownStatus:
		return nil
	default:
		return ErrInvalid
	}
}

func (p Place) Reference(provider string) (Reference, bool) {
	for _, reference := range p.References {
		if reference.Provider == provider {
			return reference, true
		}
	}
	return Reference{}, false
}

func (p Place) Duration() time.Duration {
	if p.TypicalMinutes > 0 {
		return time.Duration(p.TypicalMinutes) * time.Minute
	}
	return defaultDuration(p.Category)
}

func defaultDuration(category Category) time.Duration {
	switch category {
	case Restaurant:
		return 90 * time.Minute
	case Cafe:
		return 45 * time.Minute
	case Museum, Attraction, Experience:
		return 2 * time.Hour
	case Beach:
		return 3 * time.Hour
	case Park, Market, Shopping:
		return 90 * time.Minute
	case Viewpoint:
		return 45 * time.Minute
	case Nightlife:
		return 2 * time.Hour
	default:
		return time.Hour
	}
}
