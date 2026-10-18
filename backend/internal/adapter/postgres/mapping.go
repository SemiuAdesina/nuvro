package postgres

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/SemiuAdesina/nuvro/backend/internal/apperror"
)

type pgtypeUUID = pgtype.UUID

func tags(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}

func identifier(value string) (pgtype.UUID, error) {
	var result pgtype.UUID
	if err := result.Scan(value); err != nil {
		return result, apperror.New(apperror.CodeInvalidInput, "Invalid identifier.")
	}
	return result, nil
}

func uuidText(value pgtype.UUID) string { return value.String() }

func pgtypeDate(value time.Time) pgtype.Date {
	return pgtype.Date{Time: value.UTC(), Valid: true}
}

func dateValue(value *time.Time) pgtype.Date {
	if value == nil {
		return pgtype.Date{}
	}
	return pgtype.Date{Time: *value, Valid: true}
}

func datePointer(value pgtype.Date) *time.Time {
	if !value.Valid {
		return nil
	}
	instant := value.Time.UTC()
	return &instant
}
