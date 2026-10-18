package postgres

import (
	"context"
	"encoding/json"
	"time"

	queries "github.com/SemiuAdesina/nuvro/backend/internal/adapter/postgres/sqlc"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/alert"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/evidence"
)

const alertPageSize = 100

func (s *Store) SaveAlert(ctx context.Context, value alert.Alert) (alert.Alert, error) {
	if err := value.Validate(); err != nil {
		return alert.Alert{}, err
	}
	alertID, err := identifier(value.ID)
	if err != nil {
		return alert.Alert{}, err
	}
	tripID, err := identifier(value.TripID)
	if err != nil {
		return alert.Alert{}, err
	}
	payload, err := json.Marshal(messageData(value.MessageData))
	if err != nil {
		return alert.Alert{}, err
	}
	row, err := s.queries(ctx).UpsertTripAlert(ctx, queries.UpsertTripAlertParams{
		ID: alertID, TripID: tripID, AlertType: string(value.Type), Severity: string(value.Severity),
		TitleCode: value.TitleCode, MessageData: payload, RelatedItemID: value.RelatedItemID,
		Action: string(value.Action), Status: string(value.Status), Fingerprint: value.Fingerprint,
		ObservedAt: instantOrNil(value.Freshness.ObservedAt), ValidUntil: instantOrNil(value.Freshness.ValidUntil),
		FreshnessStatus: freshnessStatus(value.Freshness), EffectiveAt: value.EffectiveAt.UTC(),
		ExpiresAt: instantOrNil(value.ExpiresAt),
	})
	if err != nil {
		return alert.Alert{}, databaseError(err)
	}
	return toAlert(queries.ListTripAlertsRow(row))
}

func (s *Store) Alerts(ctx context.Context, tripID string) ([]alert.Alert, error) {
	identity, err := identifier(tripID)
	if err != nil {
		return nil, err
	}
	rows, err := s.queries(ctx).ListTripAlerts(ctx, queries.ListTripAlertsParams{TripID: identity, Limit: alertPageSize})
	if err != nil {
		return nil, databaseError(err)
	}
	alerts := make([]alert.Alert, 0, len(rows))
	for _, row := range rows {
		value, convertErr := toAlert(row)
		if convertErr != nil {
			continue
		}
		alerts = append(alerts, value)
	}
	return alerts, nil
}

func (s *Store) Alert(ctx context.Context, alertID string) (alert.Alert, error) {
	identity, err := identifier(alertID)
	if err != nil {
		return alert.Alert{}, err
	}
	row, err := s.queries(ctx).GetTripAlert(ctx, identity)
	if err != nil {
		return alert.Alert{}, databaseError(err)
	}
	return toAlert(queries.ListTripAlertsRow(row))
}

func (s *Store) UpdateAlertStatus(ctx context.Context, alertID string, status alert.Status) error {
	identity, err := identifier(alertID)
	if err != nil {
		return err
	}
	return databaseError(s.queries(ctx).UpdateTripAlertStatus(ctx, queries.UpdateTripAlertStatusParams{
		ID: identity, Status: string(status),
	}))
}

func (s *Store) ExpireAlerts(ctx context.Context, tripID string, now time.Time) error {
	identity, err := identifier(tripID)
	if err != nil {
		return err
	}
	return databaseError(s.queries(ctx).ExpireTripAlerts(ctx, queries.ExpireTripAlertsParams{
		TripID: identity, ExpiresAt: instantOrNil(now.UTC()),
	}))
}

func toAlert(row queries.ListTripAlertsRow) (alert.Alert, error) {
	data := map[string]string{}
	if len(row.MessageData) > 0 {
		if err := json.Unmarshal(row.MessageData, &data); err != nil {
			return alert.Alert{}, err
		}
	}
	value := alert.Alert{
		ID: uuidText(row.ID), TripID: uuidText(row.TripID), Type: alert.Type(row.AlertType),
		Severity: alert.Severity(row.Severity), TitleCode: row.TitleCode, MessageData: data,
		RelatedItemID: row.RelatedItemID, Action: alert.Action(row.Action), Status: alert.Status(row.Status),
		Fingerprint: row.Fingerprint, CreatedAt: row.CreatedAt.UTC(), EffectiveAt: row.EffectiveAt.UTC(),
	}
	if row.ExpiresAt != nil {
		value.ExpiresAt = row.ExpiresAt.UTC()
	}
	value.Freshness = toFreshness(row.ObservedAt, row.ValidUntil, row.FreshnessStatus)
	return value, nil
}

func toFreshness(observedAt, validUntil *time.Time, status *string) evidence.Freshness {
	freshness := evidence.Freshness{}
	if observedAt != nil {
		freshness.ObservedAt = observedAt.UTC()
	}
	if validUntil != nil {
		freshness.ValidUntil = validUntil.UTC()
	}
	if status != nil {
		freshness.Status = evidence.Status(*status)
	}
	return freshness
}

func freshnessStatus(freshness evidence.Freshness) *string {
	if freshness.Status == "" {
		return nil
	}
	status := string(freshness.Status)
	return &status
}

func instantOrNil(value time.Time) *time.Time {
	if value.IsZero() {
		return nil
	}
	moment := value.UTC()
	return &moment
}

func messageData(data map[string]string) map[string]string {
	if data == nil {
		return map[string]string{}
	}
	return data
}
