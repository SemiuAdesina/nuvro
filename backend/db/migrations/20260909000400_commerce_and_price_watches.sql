-- +goose Up
CREATE TABLE commerce_handoffs (
    id uuid PRIMARY KEY,
    offer_id text NOT NULL CHECK (length(offer_id) BETWEEN 1 AND 200),
    trip_id uuid REFERENCES trips(id) ON DELETE SET NULL,
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    partner text NOT NULL CHECK (length(partner) BETWEEN 1 AND 64),
    product_type text NOT NULL CHECK (product_type IN ('flight','stay','activity','transfer','car','esim','insurance')),
    destination_url text NOT NULL CHECK (destination_url ~ '^https://'),
    title text NOT NULL CHECK (length(title) BETWEEN 1 AND 300),
    price_minor bigint NOT NULL CHECK (price_minor >= 0),
    price_currency text NOT NULL CHECK (price_currency ~ '^[A-Z]{3}$'),
    attribution_token text NOT NULL UNIQUE CHECK (length(attribution_token) BETWEEN 8 AND 128),
    created_at timestamptz NOT NULL DEFAULT now(),
    expires_at timestamptz NOT NULL,
    CHECK (expires_at > created_at)
);
CREATE INDEX commerce_handoffs_by_user ON commerce_handoffs(user_id, created_at DESC);

CREATE TABLE commerce_clicks (
    id uuid PRIMARY KEY,
    handoff_id uuid NOT NULL UNIQUE REFERENCES commerce_handoffs(id) ON DELETE CASCADE,
    partner text NOT NULL,
    product_type text NOT NULL CHECK (product_type IN ('flight','stay','activity','transfer','car','esim','insurance')),
    occurred_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE commerce_conversions (
    id uuid PRIMARY KEY,
    partner text NOT NULL CHECK (length(partner) BETWEEN 1 AND 64),
    provider_conversion_id text NOT NULL CHECK (length(provider_conversion_id) BETWEEN 1 AND 200),
    handoff_id uuid REFERENCES commerce_handoffs(id) ON DELETE SET NULL,
    product_type text NOT NULL CHECK (product_type IN ('flight','stay','activity','transfer','car','esim','insurance')),
    status text NOT NULL CHECK (status IN (
        'booking_created','booking_completed','booking_cancelled','refunded',
        'commission_pending','commission_approved','commission_rejected')),
    booking_reference text NOT NULL DEFAULT '',
    gross_value_minor bigint CHECK (gross_value_minor >= 0),
    gross_value_currency text NOT NULL DEFAULT '' CHECK (gross_value_currency = '' OR gross_value_currency ~ '^[A-Z]{3}$'),
    commission_minor bigint CHECK (commission_minor >= 0),
    commission_currency text NOT NULL DEFAULT '' CHECK (commission_currency = '' OR commission_currency ~ '^[A-Z]{3}$'),
    occurred_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (partner, provider_conversion_id)
);

CREATE TABLE provider_webhook_events (
    id uuid PRIMARY KEY,
    partner text NOT NULL CHECK (length(partner) BETWEEN 1 AND 64),
    provider_event_id text NOT NULL CHECK (length(provider_event_id) BETWEEN 1 AND 200),
    event_type text NOT NULL DEFAULT '',
    normalized_status text NOT NULL DEFAULT '',
    received_at timestamptz NOT NULL DEFAULT now(),
    processed_at timestamptz,
    UNIQUE (partner, provider_event_id)
);

CREATE TABLE price_watches (
    id uuid PRIMARY KEY,
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    trip_id uuid REFERENCES trips(id) ON DELETE CASCADE,
    product_type text NOT NULL CHECK (product_type IN ('flight','stay','activity')),
    origin text NOT NULL DEFAULT '',
    destination text NOT NULL DEFAULT '',
    property_id text NOT NULL DEFAULT '',
    place_id text NOT NULL DEFAULT '',
    start_date date NOT NULL,
    end_date date NOT NULL,
    adults integer NOT NULL CHECK (adults BETWEEN 1 AND 20),
    children integer NOT NULL DEFAULT 0 CHECK (children BETWEEN 0 AND 20),
    cabin text NOT NULL DEFAULT '',
    baseline_minor bigint NOT NULL CHECK (baseline_minor > 0),
    baseline_currency text NOT NULL CHECK (baseline_currency ~ '^[A-Z]{3}$'),
    threshold_percentage integer NOT NULL CHECK (threshold_percentage BETWEEN 1 AND 90),
    threshold_absolute_minor bigint NOT NULL DEFAULT 0 CHECK (threshold_absolute_minor >= 0),
    status text NOT NULL DEFAULT 'active' CHECK (status IN ('active','paused','completed','ended')),
    created_at timestamptz NOT NULL DEFAULT now(),
    last_checked_at timestamptz,
    next_check_at timestamptz,
    expires_at timestamptz NOT NULL,
    CHECK (end_date >= start_date),
    CHECK (expires_at > created_at)
);
CREATE INDEX price_watches_due ON price_watches(next_check_at) WHERE status = 'active';
CREATE INDEX price_watches_by_user ON price_watches(user_id, created_at DESC);

CREATE TABLE price_observations (
    id uuid PRIMARY KEY,
    watch_id uuid NOT NULL REFERENCES price_watches(id) ON DELETE CASCADE,
    price_minor bigint NOT NULL CHECK (price_minor >= 0),
    price_currency text NOT NULL CHECK (price_currency ~ '^[A-Z]{3}$'),
    available boolean NOT NULL DEFAULT true,
    observed_at timestamptz NOT NULL
);
CREATE INDEX price_observations_by_watch ON price_observations(watch_id, observed_at DESC);

-- +goose Down
DROP TABLE price_observations;
DROP TABLE price_watches;
DROP TABLE provider_webhook_events;
DROP TABLE commerce_conversions;
DROP TABLE commerce_clicks;
DROP TABLE commerce_handoffs;
