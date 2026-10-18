-- name: InsertTrip :exec
INSERT INTO trips (id, owner_id, destination_id, origin_airport, start_date, end_date, adults, children, preferred_currency, status, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12);

-- name: InsertOwner :exec
INSERT INTO trip_members (trip_id, user_id, role) VALUES ($1, $2, 'owner');

-- name: GetTrip :one
SELECT * FROM trips WHERE id = $1;

-- name: GetMemberRole :one
SELECT role FROM trip_members WHERE trip_id = $1 AND user_id = $2;

-- name: UpdateTrip :exec
UPDATE trips SET
    destination_id = $2,
    origin_airport = $3,
    start_date = $4,
    end_date = $5,
    adults = $6,
    children = $7,
    preferred_currency = $8,
    status = $9,
    updated_at = $10
WHERE id = $1;

-- name: ListTripsForMember :many
SELECT t.* FROM trips t
JOIN trip_members m ON m.trip_id = t.id
WHERE m.user_id = $1
  AND (
      sqlc.narg('cursor_created_at')::timestamptz IS NULL
      OR (t.created_at, t.id) < (sqlc.narg('cursor_created_at')::timestamptz, sqlc.narg('cursor_id')::uuid)
  )
ORDER BY t.created_at DESC, t.id DESC
LIMIT sqlc.arg('page_size');

-- name: ListTripsNeedingEvaluation :many
SELECT t.id, t.owner_id, t.destination_id, t.origin_airport, t.start_date, t.end_date,
    t.adults, t.children, t.preferred_currency, t.status, t.created_at, t.updated_at
FROM trips t
JOIN itineraries i ON i.trip_id = t.id
WHERE t.status IN ('upcoming', 'active')
  AND t.start_date IS NOT NULL AND t.end_date IS NOT NULL
  AND t.start_date <= $2::date AND t.end_date >= $1::date
ORDER BY t.start_date, t.id
LIMIT $3;
