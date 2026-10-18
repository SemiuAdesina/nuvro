package commerce

import (
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/money"
)

const HandoffLifetime = 30 * time.Minute

type DisclosureCode string

const EligibleBookingDisclosure DisclosureCode = "affiliate.eligible_booking"

type Disclosure struct {
	Code    DisclosureCode
	Partner string
}

type Handoff struct {
	ID               string
	OfferID          string
	TripID           string
	UserID           string
	Partner          string
	Product          ProductType
	DestinationURL   string
	Price            money.Amount
	Title            string
	Disclosure       Disclosure
	AttributionToken string
	CreatedAt        time.Time
	ExpiresAt        time.Time
}

func (h Handoff) Validate() error {
	if !h.identityValid() || !h.Product.Valid() {
		return ErrInvalid
	}
	if h.CreatedAt.IsZero() || !h.ExpiresAt.After(h.CreatedAt) {
		return ErrInvalid
	}
	if h.Disclosure.Code != EligibleBookingDisclosure || h.Disclosure.Partner == "" {
		return ErrInvalid
	}
	return nil
}

func (h Handoff) identityValid() bool {
	return h.ID != "" && h.OfferID != "" && h.UserID != "" && h.Partner != "" &&
		h.Title != "" && h.DestinationURL != "" && h.AttributionToken != ""
}

func (h Handoff) Expired(now time.Time) bool { return !h.ExpiresAt.After(now.UTC()) }

type Click struct {
	ID         string
	HandoffID  string
	Partner    string
	Product    ProductType
	OccurredAt time.Time
}

func (c Click) Validate() error {
	if c.ID == "" || c.HandoffID == "" || c.Partner == "" || !c.Product.Valid() || c.OccurredAt.IsZero() {
		return ErrInvalid
	}
	return nil
}
