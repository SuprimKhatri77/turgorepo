-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (
  user_id,
  token,
  expires_at
) VALUES (
  $1, $2, $3
)
RETURNING *;

-- name: GetRefreshTokenByValue :one
SELECT
  t.*,
  u.id        AS user_id,
  u.name      AS user_name,
  u.email     AS user_email,
  u.role      AS user_role,
  u.image_url AS user_image_url
FROM refresh_tokens t
JOIN users u ON u.id = t.user_id
WHERE
  t.token      = $1
  AND t.revoked_at  IS NULL
  AND t.expires_at  > NOW();

-- name: RevokeTokenByUserIDAndToken :execresult
UPDATE refresh_tokens SET
  revoked_at = NOW()
WHERE
  user_id = $1 AND token = $2
  AND revoked_at IS NULL;

-- name: RevokeAllUserTokens :execresult
UPDATE refresh_tokens SET
  revoked_at = NOW()
WHERE
  user_id    = $1
  AND revoked_at IS NULL;

-- name: DeleteExpiredTokens :execresult
DELETE FROM refresh_tokens
WHERE expires_at < NOW();

-- name: GetRefreshTokenByUserIDAndToken :one
SELECT * FROM refresh_tokens
WHERE user_id = $1 AND token = $2
AND revoked_at IS NULL
AND expires_at > NOW();