-- name: UpsertTripAlert :one
INSERT INTO trip_alerts (
    id, trip_id, alert_type, severity, title_code, message_data, related_item_id, action,
    status, fingerprint, observed_at, valid_until, freshness_status, effective_at, expires_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
ON CONFLICT (trip_id, fingerprint) DO UPDATE SET
    severity = EXCLUDED.severity,
    message_data = EXCLUDED.message_data,
    observed_at = EXCLUDED.observed_at,
    valid_until = EXCLUDED.valid_until,
    freshness_status = EXCLUDED.freshness_status,
    expires_at = EXCLUDED.expires_at,
    updated_at = now()
RETURNING id, trip_id, alert_type, severity, title_code, message_data, related_item_id, action,
    status, fingerprint, observed_at, valid_until, freshness_status, created_at, effective_at, expires_at;

-- name: ListTripAlerts :many
SELECT id, trip_id, alert_type, severity, title_code, message_data, related_item_id, action,
    status, fingerprint, observed_at, valid_until, freshness_status, created_at, effective_at, expires_at
FROM trip_alerts
WHERE trip_id = $1
ORDER BY effective_at DESC, id
LIMIT $2;

-- name: GetTripAlert :one
SELECT id, trip_id, alert_type, severity, title_code, message_data, related_item_id, action,
    status, fingerprint, observed_at, valid_until, freshness_status, created_at, effective_at, expires_at
FROM trip_alerts WHERE id = $1;

-- name: UpdateTripAlertStatus :exec
UPDATE trip_alerts SET status = $2, updated_at = now() WHERE id = $1;

-- name: ExpireTripAlerts :exec
UPDATE trip_alerts SET status = 'expired', updated_at = now()
WHERE trip_id = $1 AND status IN ('unread','read') AND expires_at IS NOT NULL AND expires_at <= $2;

-- name: InsertReplanProposal :exec
INSERT INTO replan_proposals (
    id, trip_id, trigger, itinerary_version, changes, reasons, preserved,
    budget_delta_minor, budget_currency, summary, expires_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);

-- name: GetReplanProposal :one
SELECT id, trip_id, trigger, itinerary_version, changes, reasons, preserved,
    budget_delta_minor, budget_currency, summary, created_at, expires_at, applied_at
FROM replan_proposals WHERE id = $1;

-- name: MarkReplanProposalApplied :exec
UPDATE replan_proposals SET applied_at = $2, applied_by = $3
WHERE id = $1 AND applied_at IS NULL;
