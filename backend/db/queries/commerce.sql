-- name: InsertHandoff :exec
INSERT INTO commerce_handoffs (
    id, offer_id, trip_id, user_id, partner, product_type, destination_url,
    title, price_minor, price_currency, attribution_token, expires_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12);

-- name: GetHandoff :one
SELECT id, offer_id, trip_id, user_id, partner, product_type, destination_url,
    title, price_minor, price_currency, attribution_token, created_at, expires_at
FROM commerce_handoffs WHERE id = $1;

-- name: RecordClick :exec
INSERT INTO commerce_clicks (id, handoff_id, partner, product_type, occurred_at)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (handoff_id) DO NOTHING;

-- name: ClickCount :one
SELECT count(*) FROM commerce_clicks WHERE handoff_id = $1;

-- name: ClaimWebhookEvent :one
INSERT INTO provider_webhook_events (id, partner, provider_event_id, event_type, normalized_status)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (partner, provider_event_id) DO NOTHING
RETURNING id;

-- name: MarkWebhookProcessed :exec
UPDATE provider_webhook_events SET processed_at = $3
WHERE partner = $1 AND provider_event_id = $2;

-- name: UpsertConversion :one
INSERT INTO commerce_conversions (
    id, partner, provider_conversion_id, handoff_id, product_type, status,
    booking_reference, gross_value_minor, gross_value_currency,
    commission_minor, commission_currency, occurred_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
ON CONFLICT (partner, provider_conversion_id) DO UPDATE SET
    status = EXCLUDED.status,
    booking_reference = EXCLUDED.booking_reference,
    gross_value_minor = EXCLUDED.gross_value_minor,
    gross_value_currency = EXCLUDED.gross_value_currency,
    commission_minor = EXCLUDED.commission_minor,
    commission_currency = EXCLUDED.commission_currency,
    occurred_at = EXCLUDED.occurred_at,
    updated_at = now()
RETURNING id, partner, provider_conversion_id, handoff_id, product_type, status,
    booking_reference, gross_value_minor, gross_value_currency,
    commission_minor, commission_currency, occurred_at, updated_at;

-- name: GetConversion :one
SELECT id, partner, provider_conversion_id, handoff_id, product_type, status,
    booking_reference, gross_value_minor, gross_value_currency,
    commission_minor, commission_currency, occurred_at, updated_at
FROM commerce_conversions WHERE partner = $1 AND provider_conversion_id = $2;
