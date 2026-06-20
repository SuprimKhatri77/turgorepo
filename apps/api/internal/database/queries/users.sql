-- name: CreateUser :one
INSERT INTO users (
  name,
  email,
  password_hash,
  role,
  image_url
) VALUES (
  $1, $2, $3, $4, $5
)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1;

-- name: ListUsers :many
SELECT
  id, name, email, role, image_url, created_at, updated_at
FROM users
WHERE
  (
    sqlc.narg('q')::TEXT IS NULL
    OR name  ILIKE '%' || sqlc.narg('q')::TEXT || '%'
    OR email ILIKE '%' || sqlc.narg('q')::TEXT || '%'
  )
  AND (
    sqlc.narg('role')::TEXT IS NULL
    OR role = sqlc.narg('role')::TEXT
  )
ORDER BY created_at DESC
LIMIT  sqlc.narg('limit')::INT
OFFSET sqlc.narg('offset')::INT;

-- name: CountUsers :one
SELECT COUNT(*) FROM users
WHERE
  (
    sqlc.narg('q')::TEXT IS NULL
    OR name  ILIKE '%' || sqlc.narg('q')::TEXT || '%'
    OR email ILIKE '%' || sqlc.narg('q')::TEXT || '%'
  )
  AND (
    sqlc.narg('role')::TEXT IS NULL
    OR role = sqlc.narg('role')::TEXT
  );

-- name: UpdateUser :one
UPDATE users SET
  name      = COALESCE(sqlc.narg('name')::TEXT,      name),
  email     = COALESCE(sqlc.narg('email')::TEXT,     email),
  role      = COALESCE(sqlc.narg('role')::TEXT,      role),
  image_url = COALESCE(sqlc.narg('image_url')::TEXT, image_url)
WHERE id = $1
RETURNING *;

-- name: UpdateUserPassword :one
UPDATE users SET
  password_hash = $2
WHERE id = $1
RETURNING id;

-- name: DeleteUser :execresult
DELETE FROM users
WHERE id = $1;



