package commerce

import (
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/money"
)

type Freshness string

const (
	Live    Freshness = "live"
	Recent  Freshness = "recent"
	Stale   Freshness = "stale"
	Expired Freshness = "expired"
)

func (f Freshness) Valid() bool {
	switch f {
	case Live, Recent, Stale, Expired:
		return true
	default:
		return false
	}
}

func (f Freshness) Bookable() bool { return f == Live || f == Recent }

const recentWindow = 5 * time.Minute
const staleWindow = 30 * time.Minute

func FreshnessAt(observedAt, expiresAt, now time.Time) Freshness {
	if !expiresAt.IsZero() && !expiresAt.After(now.UTC()) {
		return Expired
	}
	age := now.UTC().Sub(observedAt.UTC())
	switch {
	case age < 0:
		return Live
	case age <= recentWindow:
		return Live
	case age <= staleWindow:
		return Recent
	default:
		return Stale
	}
}

type Reason string

const (
	BestPrice            Reason = "best_price"
	ShortestTrip         Reason = "shortest_trip"
	FewerStops           Reason = "fewer_stops"
	BestLocation         Reason = "best_location"
	WithinBudget         Reason = "within_budget"
	FlexibleCancellation Reason = "flexible_cancellation"
	FitsItinerary        Reason = "fits_itinerary"
	HighlyRated          Reason = "highly_rated"
)

type Offer struct {
	ID                  string
	Partner             string
	Product             ProductType
	Title               string
	Price               money.Amount
	OriginalPrice       money.Amount
	HasOriginal         bool
	Available           bool
	CancellationSummary string
	TermsURL            string
	ObservedAt          time.Time
	ExpiresAt           time.Time
	Reasons             []Reason
	Rank                int
}

func (o Offer) Validate() error {
	if o.ID == "" || o.Partner == "" || o.Title == "" || !o.Product.Valid() {
		return ErrInvalid
	}
	if o.Price.Minor() < 0 || o.ObservedAt.IsZero() {
		return ErrInvalid
	}
	if o.HasOriginal && o.OriginalPrice.Currency() != o.Price.Currency() {
		return ErrInvalid
	}
	return nil
}

func (o Offer) FreshnessAt(now time.Time) Freshness {
	return FreshnessAt(o.ObservedAt, o.ExpiresAt, now)
}

type Commercials struct {
	OfferID            string
	Partner            string
	CommissionMinor    int64
	CommissionCurrency string
	TrackingReference  string
}

func (c Commercials) Amount() (money.Amount, bool) {
	if c.CommissionCurrency == "" {
		return money.Amount{}, false
	}
	amount, err := money.New(c.CommissionMinor, money.Currency(c.CommissionCurrency))
	if err != nil {
		return money.Amount{}, false
	}
	return amount, true
}
