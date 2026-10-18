-- +goose Up
CREATE TABLE destinations (
    id text PRIMARY KEY CHECK (length(id) BETWEEN 1 AND 200),
    name text NOT NULL CHECK (length(name) BETWEEN 1 AND 200),
    region text NOT NULL DEFAULT '',
    country_code text NOT NULL CHECK (country_code ~ '^[A-Z]{2}$'),
    timezone text NOT NULL CHECK (length(timezone) BETWEEN 1 AND 64),
    latitude double precision NOT NULL CHECK (latitude BETWEEN -90 AND 90),
    longitude double precision NOT NULL CHECK (longitude BETWEEN -180 AND 180),
    primary_airport text NOT NULL CHECK (primary_airport ~ '^[A-Z]{3}$'),
    crowd_profile text NOT NULL DEFAULT 'any' CHECK (crowd_profile IN ('any', 'quiet', 'lively')),
    pace_profile text NOT NULL DEFAULT 'balanced' CHECK (pace_profile IN ('relaxed', 'balanced', 'packed')),
    is_active boolean NOT NULL DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX destinations_active_country ON destinations(country_code) WHERE is_active;

CREATE TABLE destination_tags (
    destination_id text NOT NULL REFERENCES destinations(id) ON DELETE CASCADE,
    kind text NOT NULL CHECK (kind IN ('climate', 'interest', 'style')),
    tag text NOT NULL CHECK (tag ~ '^[a-z0-9_]{1,40}$'),
    PRIMARY KEY (destination_id, kind, tag)
);
CREATE INDEX destination_tags_lookup ON destination_tags(kind, tag);

-- +goose Down
DROP TABLE destination_tags;
DROP TABLE destinations;
