package commerce

import (
	"errors"
	"strings"

	domain "github.com/SemiuAdesina/nuvro/backend/internal/domain/commerce"
)

var ErrUnknownEvent = errors.New("commerce: unrecognised provider event")

var statusByProviderCode = map[string]domain.ConversionStatus{
	"booking_created":     domain.BookingCreated,
	"booking.created":     domain.BookingCreated,
	"pending":             domain.CommissionPending,
	"booking_completed":   domain.BookingCompleted,
	"booking.completed":   domain.BookingCompleted,
	"confirmed":           domain.BookingCompleted,
	"booking_cancelled":   domain.BookingCancelled,
	"booking.cancelled":   domain.BookingCancelled,
	"cancelled":           domain.BookingCancelled,
	"refunded":            domain.Refunded,
	"refund":              domain.Refunded,
	"commission_pending":  domain.CommissionPending,
	"commission_approved": domain.CommissionApproved,
	"approved":            domain.CommissionApproved,
	"commission_rejected": domain.CommissionRejected,
	"rejected":            domain.CommissionRejected,
}

func NormalizeStatus(providerCode string) (domain.ConversionStatus, error) {
	status, found := statusByProviderCode[strings.ToLower(strings.TrimSpace(providerCode))]
	if !found {
		return "", ErrUnknownEvent
	}
	return status, nil
}

func Supersedes(existing, incoming domain.ConversionStatus) bool {
	return statusRank(incoming) >= statusRank(existing)
}

func statusRank(status domain.ConversionStatus) int {
	switch status {
	case domain.BookingCreated:
		return 1
	case domain.CommissionPending:
		return 2
	case domain.BookingCompleted:
		return 3
	case domain.CommissionApproved:
		return 4
	case domain.CommissionRejected:
		return 5
	case domain.BookingCancelled:
		return 6
	case domain.Refunded:
		return 7
	default:
		return 0
	}
}
