-- name: CreateEmailVerificationToken :exec
-- Only the newest link for a user works.
WITH invalidated AS (
  UPDATE email_verification_tokens SET used_at = now()
  WHERE user_id = sqlc.arg(user_id) AND used_at IS NULL
)
INSERT INTO email_verification_tokens (user_id, email, token_hash, expires_at)
VALUES (sqlc.arg(user_id), sqlc.arg(email), sqlc.arg(token_hash), sqlc.arg(expires_at));

-- name: VerifyEmailWithToken :one
-- One statement, so a token can't be used without the address being verified.
-- The token only counts while the account still has the address it was sent to.
WITH consumed AS (
  UPDATE email_verification_tokens AS t SET used_at = now()
  FROM users AS u
  WHERE t.token_hash = sqlc.arg(token_hash)
    AND t.used_at IS NULL
    AND t.expires_at > now()
    AND u.id = t.user_id
    AND u.email = t.email
  RETURNING t.user_id
)
UPDATE users SET email_verified_at = COALESCE(email_verified_at, now())
WHERE users.id IN (SELECT consumed.user_id FROM consumed)
RETURNING *;

-- name: MarkEmailVerified :exec
UPDATE users SET email_verified_at = COALESCE(email_verified_at, now())
WHERE id = $1;

-- name: DeleteExpiredEmailVerificationTokens :execresult
DELETE FROM email_verification_tokens
WHERE expires_at < now();
