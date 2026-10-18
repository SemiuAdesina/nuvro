-- +goose Up
CREATE TABLE place_references (
    id text PRIMARY KEY CHECK (length(id) BETWEEN 1 AND 200),
    provider text NOT NULL CHECK (length(provider) BETWEEN 1 AND 64),
    provider_place_id text NOT NULL CHECK (length(provider_place_id) BETWEEN 1 AND 200),
    name text NOT NULL CHECK (length(name) BETWEEN 1 AND 300),
    category text NOT NULL CHECK (category IN (
        'restaurant','cafe','attraction','museum','park','beach','market',
        'nightlife','shopping','viewpoint','transport','hotel','experience','other')),
    latitude double precision NOT NULL CHECK (latitude BETWEEN -90 AND 90),
    longitude double precision NOT NULL CHECK (longitude BETWEEN -180 AND 180),
    address text NOT NULL DEFAULT '',
    timezone text NOT NULL DEFAULT '',
    observed_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (provider, provider_place_id)
);

CREATE TABLE itineraries (
    id uuid PRIMARY KEY,
    trip_id uuid NOT NULL UNIQUE REFERENCES trips(id) ON DELETE CASCADE,
    timezone text NOT NULL DEFAULT 'UTC',
    version integer NOT NULL DEFAULT 1 CHECK (version >= 1),
    algorithm_version text NOT NULL,
    generated_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE itinerary_items (
    id uuid PRIMARY KEY,
    itinerary_id uuid NOT NULL REFERENCES itineraries(id) ON DELETE CASCADE,
    trip_id uuid NOT NULL REFERENCES trips(id) ON DELETE CASCADE,
    day_number integer NOT NULL CHECK (day_number >= 1),
    starts_at timestamptz NOT NULL,
    ends_at timestamptz NOT NULL,
    kind text NOT NULL CHECK (kind IN ('activity','meal','rest','transfer','note')),
    title text NOT NULL CHECK (length(title) BETWEEN 1 AND 300),
    place_id text REFERENCES place_references(id) ON DELETE SET NULL,
    category text NOT NULL DEFAULT 'other',
    latitude double precision CHECK (latitude BETWEEN -90 AND 90),
    longitude double precision CHECK (longitude BETWEEN -180 AND 180),
    route_seconds integer CHECK (route_seconds >= 0),
    route_meters integer CHECK (route_meters >= 0),
    route_mode text CHECK (route_mode IN ('walk','drive','transit','bike')),
    weather_impact text NOT NULL DEFAULT 'unknown' CHECK (weather_impact IN ('none','low','moderate','high','unknown')),
    estimated_cost_minor bigint NOT NULL DEFAULT 0 CHECK (estimated_cost_minor >= 0),
    cost_currency text NOT NULL DEFAULT '' CHECK (cost_currency = '' OR cost_currency ~ '^[A-Z]{3}$'),
    status text NOT NULL DEFAULT 'planned' CHECK (status IN ('planned','reserved','completed','skipped','cancelled')),
    source text NOT NULL DEFAULT 'generated' CHECK (source IN ('generated','traveller')),
    user_locked boolean NOT NULL DEFAULT false,
    reservation_locked boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CHECK (ends_at > starts_at)
);
CREATE INDEX itinerary_items_ordered ON itinerary_items(itinerary_id, day_number, starts_at);

-- +goose Down
DROP TABLE itinerary_items;
DROP TABLE itineraries;
DROP TABLE place_references;
