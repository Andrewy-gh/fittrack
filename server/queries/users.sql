-- name: GetUser :one
SELECT id, user_id, created_at FROM users WHERE id = $1;

-- name: GetUserByUserID :one
SELECT id, user_id, created_at FROM users WHERE user_id = $1 LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (user_id)
VALUES ($1)
RETURNING id;

-- name: DeleteUser :execrows
DELETE FROM users
WHERE user_id = $1;

-- Feature access queries
