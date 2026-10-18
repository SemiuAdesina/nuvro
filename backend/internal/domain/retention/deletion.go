package retention

import (
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/apperror"
)

const GracePeriod = 30 * day

type DeletionStatus string

const (
	Pending DeletionStatus = "pending"
	Purged  DeletionStatus = "purged"
	Failed  DeletionStatus = "failed"
)

type Deletion struct {
	ID               string
	UserID           string
	Status           DeletionStatus
	Reason           string
	RequestedAt      time.Time
	PurgeAfter       time.Time
	PurgedAt         *time.Time
	TripsTransferred int
	TripsDeleted     int
}

func Request(deletionID, userID, reason string, now time.Time) (Deletion, error) {
	if deletionID == "" || userID == "" {
		return Deletion{}, apperror.New(apperror.CodeInvalidInput, "A deletion needs an account.")
	}
	if len(reason) > 64 {
		return Deletion{}, apperror.New(apperror.CodeInvalidInput, "That reason is too long.")
	}
	return Deletion{
		ID: deletionID, UserID: userID, Status: Pending, Reason: reason,
		RequestedAt: now.UTC(), PurgeAfter: now.Add(GracePeriod).UTC(),
	}, nil
}

func (d Deletion) Due(now time.Time) bool {
	return d.Status == Pending && !now.UTC().Before(d.PurgeAfter)
}

func (d Deletion) Complete(now time.Time, transferred, deleted int) Deletion {
	completed := now.UTC()
	d.Status = Purged
	d.PurgedAt = &completed
	d.TripsTransferred = transferred
	d.TripsDeleted = deleted
	return d
}
