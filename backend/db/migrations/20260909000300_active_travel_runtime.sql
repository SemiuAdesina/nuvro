-- +goose Up
CREATE TABLE trip_alerts (
    id uuid PRIMARY KEY,
    trip_id uuid NOT NULL REFERENCES trips(id) ON DELETE CASCADE,
    alert_type text NOT NULL CHECK (alert_type IN (
        'weather_impact','leave_reminder','schedule_conflict','place_change',
        'reservation_reminder','document_reminder','entry_reminder')),
    severity text NOT NULL CHECK (severity IN ('info','attention','important')),
    title_code text NOT NULL CHECK (length(title_code) BETWEEN 1 AND 120),
    message_data jsonb NOT NULL DEFAULT '{}'::jsonb,
    related_item_id text NOT NULL DEFAULT '',
    action text NOT NULL DEFAULT 'none' CHECK (action IN (
        'adjust_plan','navigate','view_activity','view_document','none')),
    status text NOT NULL DEFAULT 'unread' CHECK (status IN ('unread','read','resolved','expired')),
    fingerprint text NOT NULL CHECK (length(fingerprint) BETWEEN 8 AND 128),
    observed_at timestamptz,
    valid_until timestamptz,
    freshness_status text CHECK (freshness_status IN ('live','cached','stale')),
    created_at timestamptz NOT NULL DEFAULT now(),
    effective_at timestamptz NOT NULL,
    expires_at timestamptz,
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (trip_id, fingerprint),
    CHECK (expires_at IS NULL OR expires_at > effective_at)
);
CREATE INDEX trip_alerts_visible ON trip_alerts(trip_id, effective_at DESC)
    WHERE status IN ('unread', 'read');

CREATE TABLE replan_proposals (
    id uuid PRIMARY KEY,
    trip_id uuid NOT NULL REFERENCES trips(id) ON DELETE CASCADE,
    trigger text NOT NULL CHECK (trigger IN ('weather','late','tired','missed','closed','free_time','manual')),
    itinerary_version integer NOT NULL CHECK (itinerary_version >= 1),
    changes jsonb NOT NULL,
    reasons text[] NOT NULL DEFAULT '{}',
    preserved text[] NOT NULL DEFAULT '{}',
    budget_delta_minor bigint,
    budget_currency text NOT NULL DEFAULT '' CHECK (budget_currency = '' OR budget_currency ~ '^[A-Z]{3}$'),
    summary text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT now(),
    expires_at timestamptz NOT NULL,
    applied_at timestamptz,
    applied_by uuid REFERENCES users(id) ON DELETE SET NULL,
    CHECK (expires_at > created_at)
);
CREATE INDEX replan_proposals_by_trip ON replan_proposals(trip_id, created_at DESC);

-- +goose Down
DROP TABLE replan_proposals;
DROP TABLE trip_alerts;
