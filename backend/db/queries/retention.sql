-- name: InsertAccountDeletion :one
INSERT INTO account_deletions (id, user_id, status, reason, requested_at, purge_after)
VALUES ($1, $2, 'pending', $3, $4, $5)
ON CONFLICT (user_id) DO UPDATE SET reason = account_deletions.reason
RETURNING id, user_id, status, reason, requested_at, purge_after, purged_at,
    trips_transferred, trips_deleted;

-- name: GetAccountDeletion :one
SELECT id, user_id, status, reason, requested_at, purge_after, purged_at,
    trips_transferred, trips_deleted
FROM account_deletions WHERE user_id = $1;

-- name: ListDueAccountDeletions :many
SELECT id, user_id, status, reason, requested_at, purge_after, purged_at,
    trips_transferred, trips_deleted
FROM account_deletions
WHERE status = 'pending' AND purge_after <= $1
ORDER BY purge_after
LIMIT $2;

-- name: CompleteAccountDeletion :exec
UPDATE account_deletions
SET status = 'purged', purged_at = $2, trips_transferred = $3, trips_deleted = $4
WHERE id = $1;

-- name: FailAccountDeletion :exec
UPDATE account_deletions SET status = 'failed' WHERE id = $1;

-- name: MarkUserDeleted :exec
UPDATE users SET status = 'deleted', updated_at = $2 WHERE id = $1;

-- name: ListSoleOwnedTrips :many
SELECT t.id FROM trips t
WHERE t.owner_id = $1
  AND NOT EXISTS (SELECT 1 FROM trip_members m WHERE m.trip_id = t.id AND m.user_id <> $1);

-- name: ListSharedOwnedTrips :many
SELECT t.id, m.user_id, m.created_at
FROM trips t
JOIN trip_members m ON m.trip_id = t.id
WHERE t.owner_id = $1 AND m.user_id <> $1
ORDER BY t.id, m.created_at, m.user_id;

-- name: TransferTripOwnership :exec
UPDATE trips SET owner_id = $2, updated_at = $3 WHERE id = $1;

-- name: PromoteTripMember :exec
UPDATE trip_members SET role = 'owner' WHERE trip_id = $1 AND user_id = $2;

-- name: DeleteTrip :exec
DELETE FROM trips WHERE id = $1;

-- name: DeleteTripMemberships :exec
DELETE FROM trip_members WHERE user_id = $1;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;

-- name: DeleteExpiredPersonalizationSignals :execrows
DELETE FROM personalization_signals WHERE occurred_at < $1;

-- name: DeleteExpiredCommerceHandoffs :execrows
DELETE FROM commerce_handoffs WHERE created_at < $1;

-- name: DeleteExpiredWebhookEvents :execrows
DELETE FROM provider_webhook_events WHERE received_at < $1;

-- name: DeleteExpiredTripAlerts :execrows
DELETE FROM trip_alerts WHERE created_at < $1;

-- name: DeleteExpiredReplanProposals :execrows
DELETE FROM replan_proposals WHERE created_at < $1;

-- name: DeleteExpiredPriceWatches :execrows
DELETE FROM price_watches WHERE expires_at < $1;

-- name: LinkIdentityToken :exec
UPDATE user_identities
SET sealed_refresh_token = $3, linked_at = $4, revoked_at = NULL
WHERE user_id = $1 AND provider = $2;

-- name: RevocableIdentities :many
SELECT i.id, i.user_id, i.provider, i.sealed_refresh_token
FROM user_identities i
JOIN account_deletions d ON d.user_id = i.user_id
WHERE i.sealed_refresh_token IS NOT NULL AND i.revoked_at IS NULL AND d.status = 'pending'
ORDER BY d.requested_at
LIMIT $1;

-- name: MarkIdentityRevoked :exec
UPDATE user_identities SET revoked_at = $2, sealed_refresh_token = NULL WHERE id = $1;

-- name: IdentitiesForUser :many
SELECT id, user_id, provider, sealed_refresh_token FROM user_identities WHERE user_id = $1;
