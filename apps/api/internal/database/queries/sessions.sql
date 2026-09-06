-- name: CreateSession :one
INSERT INTO sessions (id, user_id, current_token_hash, user_agent, ip_address, expires_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetActiveSessionByID :one
SELECT * FROM sessions
WHERE id = $1
  AND revoked_at IS NULL
  AND expires_at > now();

-- name: RotateSessionToken :one
UPDATE sessions SET
  previous_token_hash = current_token_hash,
  previous_rotated_at = now(),
  current_token_hash = sqlc.arg(new_token_hash),
  expires_at = sqlc.arg(expires_at),
  last_seen_at = now()
WHERE id = sqlc.arg(id)
  AND current_token_hash = sqlc.arg(expected_token_hash)
  AND revoked_at IS NULL
  AND expires_at > now()
RETURNING *;

-- name: TouchSessionLastSeen :exec
UPDATE sessions SET last_seen_at = now()
WHERE id = $1;

-- name: RevokeSession :exec
UPDATE sessions SET revoked_at = now()
WHERE id = $1;

-- name: RevokeAllUserSessions :exec
UPDATE sessions SET revoked_at = now()
WHERE user_id = $1 AND revoked_at IS NULL;

-- name: ListActiveSessionsByUser :many
SELECT id, user_agent, ip_address, created_at, last_seen_at
FROM sessions
WHERE user_id = $1
  AND revoked_at IS NULL
ORDER BY last_seen_at DESC;

-- name: DeleteExpiredSessions :execresult
DELETE FROM sessions
WHERE expires_at < now();
