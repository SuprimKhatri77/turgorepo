-- name: CreatePasswordResetToken :exec
-- Only the newest link for a user works.
WITH invalidated AS (
  UPDATE password_reset_tokens SET used_at = now()
  WHERE user_id = sqlc.arg(user_id) AND used_at IS NULL
)
INSERT INTO password_reset_tokens (user_id, token_hash, expires_at)
VALUES (sqlc.arg(user_id), sqlc.arg(token_hash), sqlc.arg(expires_at));

-- name: ResetPasswordWithToken :one
-- One statement, so a token can't be used without the rest of the reset happening.
-- A concurrent request waits on the token's row lock, then finds used_at set.
WITH consumed AS (
  UPDATE password_reset_tokens AS t SET used_at = now()
  WHERE t.token_hash = sqlc.arg(token_hash)
    AND t.used_at IS NULL
    AND t.expires_at > now()
  RETURNING t.user_id
), other_tokens AS (
  UPDATE password_reset_tokens AS t SET used_at = now()
  WHERE t.user_id IN (SELECT consumed.user_id FROM consumed)
    AND t.used_at IS NULL
    AND t.token_hash <> sqlc.arg(token_hash)
), revoked_sessions AS (
  UPDATE sessions AS s SET revoked_at = now()
  WHERE s.user_id IN (SELECT consumed.user_id FROM consumed)
    AND s.revoked_at IS NULL
)
-- Following the emailed link proves the user owns the address.
UPDATE users SET
  password_hash = sqlc.arg(password_hash),
  email_verified_at = COALESCE(users.email_verified_at, now())
WHERE users.id IN (SELECT consumed.user_id FROM consumed)
RETURNING *;

-- name: DeleteExpiredPasswordResetTokens :execresult
DELETE FROM password_reset_tokens
WHERE expires_at < now();
