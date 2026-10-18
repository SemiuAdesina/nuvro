-- name: GetProfile :one
SELECT * FROM user_profiles WHERE user_id = $1;

-- name: UpdateProfile :exec
UPDATE user_profiles SET
    preferred_language = $2,
    locale = $3,
    home_country = $4,
    home_timezone = $5,
    departure_airport = $6,
    preferred_currency = $7,
    units = $8,
    hour_cycle = $9,
    accessibility_preferences = $10,
    updated_at = now()
WHERE user_id = $1;

-- name: GetPreferences :one
SELECT * FROM traveler_preferences WHERE user_id = $1;

-- name: UpdatePreferences :exec
UPDATE traveler_preferences SET
    travel_style = $2,
    weather_preference = $3,
    crowd_preference = $4,
    max_flight_minutes = $5,
    trip_pace = $6,
    interests = $7,
    traveling_with_children = $8,
    scope = $9,
    entry_simplicity = $10,
    updated_at = now()
WHERE user_id = $1;
