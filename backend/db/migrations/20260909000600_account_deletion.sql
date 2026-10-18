-- +goose Up
CREATE TABLE account_deletions (
    id uuid PRIMARY KEY,
    user_id uuid NOT NULL UNIQUE,
    status text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'purged', 'failed')),
    reason text NOT NULL DEFAULT '' CHECK (length(reason) <= 64),
    requested_at timestamptz NOT NULL,
    purge_after timestamptz NOT NULL,
    purged_at timestamptz,
    trips_transferred integer NOT NULL DEFAULT 0 CHECK (trips_transferred >= 0),
    trips_deleted integer NOT NULL DEFAULT 0 CHECK (trips_deleted >= 0),
    CHECK (purge_after >= requested_at),
    CHECK ((status = 'purged') = (purged_at IS NOT NULL))
);
CREATE INDEX account_deletions_due ON account_deletions(purge_after) WHERE status = 'pending';

-- +goose Down
DROP TABLE account_deletions;
