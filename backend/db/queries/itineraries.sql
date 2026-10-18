-- name: UpsertItinerary :one
INSERT INTO itineraries (id, trip_id, timezone, version, algorithm_version, generated_at)
VALUES ($1, $2, $3, 1, $4, $5)
ON CONFLICT (trip_id) DO UPDATE SET
    timezone = EXCLUDED.timezone,
    version = itineraries.version + 1,
    algorithm_version = EXCLUDED.algorithm_version,
    generated_at = EXCLUDED.generated_at,
    updated_at = now()
RETURNING id, trip_id, timezone, version, algorithm_version, generated_at;

-- name: GetItineraryByTrip :one
SELECT id, trip_id, timezone, version, algorithm_version, generated_at
FROM itineraries WHERE trip_id = $1;

-- name: DeleteUnlockedItems :exec
DELETE FROM itinerary_items
WHERE itinerary_id = $1 AND NOT user_locked AND NOT reservation_locked;

-- name: InsertItineraryItem :exec
INSERT INTO itinerary_items (
    id, itinerary_id, trip_id, day_number, starts_at, ends_at, kind, title, place_id, category,
    latitude, longitude, route_seconds, route_meters, route_mode, weather_impact,
    estimated_cost_minor, cost_currency, status, source, user_locked, reservation_locked
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22
)
ON CONFLICT (id) DO NOTHING;

-- name: ListItineraryItems :many
SELECT
    i.id, i.itinerary_id, i.trip_id, i.day_number, i.starts_at, i.ends_at, i.kind, i.title,
    i.place_id, i.category, i.latitude, i.longitude, i.route_seconds, i.route_meters, i.route_mode,
    i.weather_impact, i.estimated_cost_minor, i.cost_currency, i.status, i.source,
    i.user_locked, i.reservation_locked
FROM itinerary_items i
WHERE i.itinerary_id = $1
ORDER BY i.day_number, i.starts_at, i.id;

-- name: GetItineraryItem :one
SELECT
    i.id, i.itinerary_id, i.trip_id, i.day_number, i.starts_at, i.ends_at, i.kind, i.title,
    i.place_id, i.category, i.latitude, i.longitude, i.route_seconds, i.route_meters, i.route_mode,
    i.weather_impact, i.estimated_cost_minor, i.cost_currency, i.status, i.source,
    i.user_locked, i.reservation_locked
FROM itinerary_items i WHERE i.id = $1;

-- name: UpdateItineraryItem :exec
UPDATE itinerary_items SET
    starts_at = $2,
    ends_at = $3,
    title = $4,
    status = $5,
    user_locked = $6,
    updated_at = now()
WHERE id = $1;

-- name: UpsertPlaceReference :exec
INSERT INTO place_references (id, provider, provider_place_id, name, category, latitude, longitude, address, timezone, observed_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    category = EXCLUDED.category,
    latitude = EXCLUDED.latitude,
    longitude = EXCLUDED.longitude,
    address = EXCLUDED.address,
    timezone = EXCLUDED.timezone,
    observed_at = EXCLUDED.observed_at,
    updated_at = now();

-- name: DeleteAllItems :exec
DELETE FROM itinerary_items WHERE itinerary_id = $1;

-- name: BumpItineraryVersion :one
UPDATE itineraries SET version = version + 1, updated_at = now()
WHERE trip_id = $1 AND version = $2
RETURNING id, trip_id, timezone, version, algorithm_version, generated_at;
