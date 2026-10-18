-- +goose Up
CREATE TABLE sessions (
    id uuid PRIMARY KEY,
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    family_id uuid NOT NULL,
    refresh_hash bytea NOT NULL UNIQUE,
    issued_at timestamptz NOT NULL,
    expires_at timestamptz NOT NULL,
    rotated_at timestamptz,
    revoked_at timestamptz,
    CHECK (expires_at > issued_at)
);
CREATE INDEX sessions_by_user ON sessions(user_id);
CREATE INDEX sessions_by_family ON sessions(family_id);

-- +goose Down
DROP TABLE sessions;
