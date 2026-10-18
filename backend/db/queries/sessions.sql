-- name: InsertSession :exec
INSERT INTO sessions (id, user_id, family_id, refresh_hash, issued_at, expires_at)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: GetSessionByRefreshHash :one
SELECT id, user_id, family_id, issued_at, expires_at, rotated_at, revoked_at
FROM sessions WHERE refresh_hash = $1;

-- name: GetSessionAccess :one
SELECT s.revoked_at, u.status
FROM sessions s JOIN users u ON u.id = s.user_id
WHERE s.id = $1;

-- name: MarkSessionRotated :exec
UPDATE sessions SET rotated_at = $2 WHERE id = $1 AND rotated_at IS NULL;

-- name: RevokeSession :exec
UPDATE sessions SET revoked_at = $2 WHERE id = $1 AND revoked_at IS NULL;

-- name: RevokeSessionFamily :exec
UPDATE sessions SET revoked_at = $2 WHERE family_id = $1 AND revoked_at IS NULL;

-- name: RevokeUserSessions :exec
UPDATE sessions SET revoked_at = $2 WHERE user_id = $1 AND revoked_at IS NULL;

-- name: DeleteExpiredSessions :execrows
DELETE FROM sessions WHERE expires_at < $1;
