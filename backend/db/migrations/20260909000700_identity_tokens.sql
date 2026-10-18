-- +goose Up
ALTER TABLE user_identities
    ADD COLUMN sealed_refresh_token bytea,
    ADD COLUMN linked_at timestamptz,
    ADD COLUMN revoked_at timestamptz;

CREATE INDEX user_identities_revocable ON user_identities(user_id)
    WHERE sealed_refresh_token IS NOT NULL AND revoked_at IS NULL;

-- +goose Down
DROP INDEX user_identities_revocable;
ALTER TABLE user_identities
    DROP COLUMN sealed_refresh_token,
    DROP COLUMN linked_at,
    DROP COLUMN revoked_at;
