package commerce

import (
	"slices"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/money"
)

type ConversionStatus string

const (
	BookingCreated     ConversionStatus = "booking_created"
	BookingCompleted   ConversionStatus = "booking_completed"
	BookingCancelled   ConversionStatus = "booking_cancelled"
	Refunded           ConversionStatus = "refunded"
	CommissionPending  ConversionStatus = "commission_pending"
	CommissionApproved ConversionStatus = "commission_approved"
	CommissionRejected ConversionStatus = "commission_rejected"
)

func ConversionStatuses() []ConversionStatus {
	return []ConversionStatus{
		BookingCreated, BookingCompleted, BookingCancelled, Refunded,
		CommissionPending, CommissionApproved, CommissionRejected,
	}
}

func (s ConversionStatus) Valid() bool { return slices.Contains(ConversionStatuses(), s) }

func (s ConversionStatus) ConfirmsBooking() bool {
	return s == BookingCreated || s == BookingCompleted
}

func (s ConversionStatus) EndsBooking() bool {
	return s == BookingCancelled || s == Refunded
}

type Conversion struct {
	ID                   string
	Partner              string
	ProviderConversionID string
	HandoffID            string
	Product              ProductType
	Status               ConversionStatus
	BookingReference     string
	GrossValue           money.Amount
	HasGrossValue        bool
	Commission           money.Amount
	HasCommission        bool
	OccurredAt           time.Time
	UpdatedAt            time.Time
}

func (c Conversion) Validate() error {
	if c.ID == "" {
		return ErrInvalid
	}
	return c.ValidateInbound()
}

func (c Conversion) ValidateInbound() error {
	if c.Partner == "" || c.ProviderConversionID == "" {
		return ErrInvalid
	}
	if !c.Product.Valid() || !c.Status.Valid() || c.OccurredAt.IsZero() {
		return ErrInvalid
	}
	if c.HasGrossValue && c.GrossValue.IsNegative() {
		return ErrInvalid
	}
	if c.HasCommission && c.Commission.IsNegative() {
		return ErrInvalid
	}
	return nil
}

func MaskReference(reference string) string {
	if len(reference) <= 4 {
		return "****"
	}
	return "****" + reference[len(reference)-4:]
}

type WebhookEvent struct {
	Partner    string
	EventID    string
	EventType  string
	ReceivedAt time.Time
	Conversion Conversion
}

func (e WebhookEvent) Validate() error {
	if e.Partner == "" || e.EventID == "" || e.ReceivedAt.IsZero() {
		return ErrInvalid
	}
	return e.Conversion.ValidateInbound()
}
