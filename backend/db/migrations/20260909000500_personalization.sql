-- +goose Up
ALTER TABLE traveler_preferences
    ADD COLUMN history_learning_enabled boolean NOT NULL DEFAULT true;

CREATE TABLE personalization_signals (
    id uuid PRIMARY KEY,
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    trip_id uuid REFERENCES trips(id) ON DELETE SET NULL,
    signal_type text NOT NULL CHECK (signal_type IN (
        'destination_saved','destination_unsaved','destination_selected',
        'activity_added','activity_removed','activity_completed',
        'stay_selected','neighbourhood_selected','replan_requested','trip_completed',
        'feedback_positive','feedback_negative','explicit_confirmation')),
    subject_type text NOT NULL CHECK (subject_type IN (
        'destination','activity_category','stay_area','trip_pace','travel_burden','interest')),
    subject_id text NOT NULL CHECK (length(subject_id) BETWEEN 1 AND 120),
    strength text NOT NULL CHECK (strength IN ('weak','medium','strong')),
    source_event text NOT NULL DEFAULT '' CHECK (length(source_event) <= 200),
    occurred_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX personalization_signals_by_user ON personalization_signals(user_id, occurred_at DESC);
CREATE UNIQUE INDEX personalization_signals_source
    ON personalization_signals(user_id, source_event) WHERE source_event <> '';

CREATE TABLE learned_preferences (
    id uuid PRIMARY KEY,
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    preference_key text NOT NULL CHECK (length(preference_key) BETWEEN 1 AND 120),
    confidence text NOT NULL CHECK (confidence IN ('low','medium','high')),
    score integer NOT NULL CHECK (score >= 0),
    evidence_count integer NOT NULL CHECK (evidence_count >= 0),
    explanation_code text NOT NULL DEFAULT '',
    status text NOT NULL DEFAULT 'active' CHECK (status IN ('active','confirmed','dismissed','suppressed')),
    rules_version text NOT NULL,
    first_observed_at timestamptz NOT NULL,
    last_observed_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (user_id, preference_key),
    CHECK (last_observed_at >= first_observed_at)
);
CREATE INDEX learned_preferences_active ON learned_preferences(user_id)
    WHERE status IN ('active', 'confirmed');

CREATE TABLE trip_feedback (
    trip_id uuid PRIMARY KEY REFERENCES trips(id) ON DELETE CASCADE,
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    positives text[] NOT NULL DEFAULT '{}',
    negatives text[] NOT NULL DEFAULT '{}',
    pace text NOT NULL DEFAULT '' CHECK (pace IN ('', 'too_slow','just_right','too_busy')),
    travel_burden text NOT NULL DEFAULT '' CHECK (travel_burden IN ('', 'light','comfortable','too_much')),
    return_intent text NOT NULL DEFAULT '' CHECK (return_intent IN ('', 'would_return','undecided','would_not_return')),
    submitted_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE trip_feedback;
DROP TABLE learned_preferences;
DROP TABLE personalization_signals;
ALTER TABLE traveler_preferences DROP COLUMN history_learning_enabled;
