-- name: InsertPriceWatch :exec
INSERT INTO price_watches (
    id, user_id, trip_id, product_type, origin, destination, property_id, place_id,
    start_date, end_date, adults, children, cabin, baseline_minor, baseline_currency,
    threshold_percentage, threshold_absolute_minor, status, next_check_at, expires_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20);

-- name: ListPriceWatches :many
SELECT id, user_id, trip_id, product_type, origin, destination, property_id, place_id,
    start_date, end_date, adults, children, cabin, baseline_minor, baseline_currency,
    threshold_percentage, threshold_absolute_minor, status, created_at, last_checked_at,
    next_check_at, expires_at
FROM price_watches WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2;

-- name: GetPriceWatch :one
SELECT id, user_id, trip_id, product_type, origin, destination, property_id, place_id,
    start_date, end_date, adults, children, cabin, baseline_minor, baseline_currency,
    threshold_percentage, threshold_absolute_minor, status, created_at, last_checked_at,
    next_check_at, expires_at
FROM price_watches WHERE id = $1;

-- name: CountActivePriceWatches :one
SELECT count(*) FROM price_watches WHERE user_id = $1 AND status = 'active';

-- name: EndPriceWatch :exec
UPDATE price_watches SET status = $2 WHERE id = $1;

-- name: ListDuePriceWatches :many
SELECT id, user_id, trip_id, product_type, origin, destination, property_id, place_id,
    start_date, end_date, adults, children, cabin, baseline_minor, baseline_currency,
    threshold_percentage, threshold_absolute_minor, status, created_at, last_checked_at,
    next_check_at, expires_at
FROM price_watches
WHERE status = 'active' AND expires_at > $1 AND (next_check_at IS NULL OR next_check_at <= $1)
ORDER BY next_check_at NULLS FIRST, id
LIMIT $2;

-- name: RecordPriceObservation :exec
INSERT INTO price_observations (id, watch_id, price_minor, price_currency, available, observed_at)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: ReschedulePriceWatch :exec
UPDATE price_watches SET last_checked_at = $2, next_check_at = $3, baseline_minor = $4 WHERE id = $1;

-- name: ExpirePriceWatches :exec
UPDATE price_watches SET status = 'ended' WHERE status = 'active' AND expires_at <= $1;
