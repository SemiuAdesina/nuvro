-- name: FindIdentity :one
SELECT u.id, u.status FROM users u
JOIN user_identities i ON i.user_id = u.id
WHERE i.provider = $1 AND i.provider_subject = $2;

-- name: InsertUser :exec
INSERT INTO users (id) VALUES ($1);

-- name: InsertIdentity :exec
INSERT INTO user_identities (id, user_id, provider, provider_subject) VALUES ($1, $2, $3, $4);

-- name: InsertProfile :exec
INSERT INTO user_profiles (
    user_id, preferred_language, locale, home_country, home_timezone,
    departure_airport, preferred_currency, units, hour_cycle, accessibility_preferences
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);

-- name: InsertPreferences :exec
INSERT INTO traveler_preferences (
    user_id, travel_style, weather_preference, crowd_preference, max_flight_minutes,
    trip_pace, interests, traveling_with_children, scope, entry_simplicity
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);
