-- name: CreateToken :one
INSERT INTO tokens (
  user_id,
  token,
  session_id,
  expires_at
) VALUES (
  $1, $2, $3, $4
)
RETURNING *;

-- name: GetTokenByValue :one
SELECT
  t.*,
  u.id        AS user_id,
  u.name      AS user_name,
  u.email     AS user_email,
  u.role      AS user_role,
  u.image_url AS user_image_url
FROM tokens t
JOIN users u ON u.id = t.user_id
WHERE
  t.token      = $1
  AND t.revoked_at  IS NULL
  AND t.expires_at  > NOW();

-- name: RevokeTokenBySessionIDAndToken :execresult
UPDATE tokens SET
  revoked_at = NOW()
WHERE
  session_id = $1 AND token = $2
  AND revoked_at IS NULL;

-- name: RevokeAllUserTokens :execresult
UPDATE tokens SET
  revoked_at = NOW()
WHERE
  user_id    = $1
  AND revoked_at IS NULL;

-- name: DeleteExpiredTokens :execresult
DELETE FROM tokens
WHERE expires_at < NOW();

-- name: GetRefreshTokenBySessionIDAndToken :one
SELECT * FROM tokens
WHERE session_id = $1 AND token = $2
AND revoked_at IS NULL
AND expires_at > NOW();