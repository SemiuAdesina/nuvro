-- name: ListActiveDestinations :many
SELECT
    d.id, d.name, d.region, d.country_code, d.timezone, d.latitude, d.longitude,
    d.primary_airport, d.crowd_profile, d.pace_profile,
    COALESCE(ARRAY_AGG(DISTINCT t.tag) FILTER (WHERE t.kind = 'climate'), '{}')::text[] AS climate_tags,
    COALESCE(ARRAY_AGG(DISTINCT t.tag) FILTER (WHERE t.kind = 'interest'), '{}')::text[] AS interest_tags,
    COALESCE(ARRAY_AGG(DISTINCT t.tag) FILTER (WHERE t.kind = 'style'), '{}')::text[] AS style_tags
FROM destinations d
LEFT JOIN destination_tags t ON t.destination_id = d.id
WHERE d.is_active
GROUP BY d.id
ORDER BY d.id
LIMIT sqlc.arg('page_size');

-- name: UpsertDestination :exec
INSERT INTO destinations (id, name, region, country_code, timezone, latitude, longitude, primary_airport, crowd_profile, pace_profile)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    region = EXCLUDED.region,
    country_code = EXCLUDED.country_code,
    timezone = EXCLUDED.timezone,
    latitude = EXCLUDED.latitude,
    longitude = EXCLUDED.longitude,
    primary_airport = EXCLUDED.primary_airport,
    crowd_profile = EXCLUDED.crowd_profile,
    pace_profile = EXCLUDED.pace_profile,
    updated_at = now();

-- name: ReplaceDestinationTags :exec
INSERT INTO destination_tags (destination_id, kind, tag)
SELECT $1, $2, UNNEST($3::text[])
ON CONFLICT DO NOTHING;

-- name: DeleteDestinationTags :exec
DELETE FROM destination_tags WHERE destination_id = $1 AND kind = $2;
