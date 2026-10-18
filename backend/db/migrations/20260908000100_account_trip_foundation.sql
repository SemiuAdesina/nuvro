-- +goose Up
CREATE TABLE users (
    id uuid PRIMARY KEY,
    status text NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'suspended', 'deleted')),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE user_identities (
    id uuid PRIMARY KEY,
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider text NOT NULL CHECK (length(provider) BETWEEN 1 AND 80),
    provider_subject text NOT NULL CHECK (length(provider_subject) BETWEEN 1 AND 512),
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (provider, provider_subject)
);
CREATE INDEX user_identities_user ON user_identities(user_id);

CREATE TABLE user_profiles (
    user_id uuid PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    preferred_language text NOT NULL DEFAULT 'en',
    locale text NOT NULL DEFAULT 'en',
    home_country text CHECK (home_country ~ '^[A-Z]{2}$'),
    home_timezone text,
    departure_airport text CHECK (departure_airport ~ '^[A-Z]{3}$'),
    preferred_currency text NOT NULL CHECK (preferred_currency ~ '^[A-Z]{3}$'),
    units text NOT NULL DEFAULT 'metric' CHECK (units IN ('metric', 'imperial')),
    hour_cycle text NOT NULL DEFAULT '24' CHECK (hour_cycle IN ('12', '24')),
    accessibility_preferences text[] NOT NULL DEFAULT '{}',
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE traveler_preferences (
    user_id uuid PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    travel_style text[] NOT NULL DEFAULT '{}',
    weather_preference text NOT NULL DEFAULT 'any' CHECK (weather_preference IN ('any', 'warm', 'mild', 'cold')),
    crowd_preference text NOT NULL DEFAULT 'any' CHECK (crowd_preference IN ('any', 'quiet', 'lively')),
    max_flight_minutes integer CHECK (max_flight_minutes BETWEEN 0 AND 2880),
    trip_pace text NOT NULL DEFAULT 'balanced' CHECK (trip_pace IN ('relaxed', 'balanced', 'packed')),
    interests text[] NOT NULL DEFAULT '{}',
    traveling_with_children boolean NOT NULL DEFAULT false,
    scope text NOT NULL DEFAULT 'any' CHECK (scope IN ('any', 'domestic', 'international')),
    entry_simplicity text NOT NULL DEFAULT 'any' CHECK (entry_simplicity IN ('any', 'simple')),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE trips (
    id uuid PRIMARY KEY,
    owner_id uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    destination_id text CHECK (length(destination_id) BETWEEN 1 AND 200),
    origin_airport text CHECK (origin_airport ~ '^[A-Z]{3}$'),
    start_date date,
    end_date date,
    adults integer NOT NULL CHECK (adults BETWEEN 1 AND 100),
    children integer NOT NULL DEFAULT 0 CHECK (children BETWEEN 0 AND 100),
    preferred_currency text NOT NULL CHECK (preferred_currency ~ '^[A-Z]{3}$'),
    status text NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'planning', 'upcoming', 'active', 'completed', 'cancelled', 'archived')),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CHECK ((start_date IS NULL AND end_date IS NULL) OR (start_date IS NOT NULL AND end_date IS NOT NULL AND end_date >= start_date))
);
CREATE INDEX trips_owner_created ON trips(owner_id, created_at DESC, id DESC);

CREATE TABLE trip_members (
    trip_id uuid NOT NULL REFERENCES trips(id) ON DELETE CASCADE,
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    role text NOT NULL CHECK (role IN ('owner', 'editor', 'viewer')),
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (trip_id, user_id)
);
CREATE UNIQUE INDEX trip_members_one_owner ON trip_members(trip_id) WHERE role = 'owner';
CREATE INDEX trip_members_user ON trip_members(user_id, trip_id);

CREATE TABLE saved_destinations (
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    destination_id text NOT NULL CHECK (length(destination_id) BETWEEN 1 AND 200),
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, destination_id)
);
CREATE INDEX saved_destinations_created ON saved_destinations(user_id, created_at DESC, destination_id DESC);

-- +goose Down
DROP TABLE saved_destinations;
DROP TABLE trip_members;
DROP TABLE trips;
DROP TABLE traveler_preferences;
DROP TABLE user_profiles;
DROP TABLE user_identities;
DROP TABLE users;
