-- name: CreateUser :one 
INSERT INTO users (user_id, username, hashed_pw, admin, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetUsers :many
SELECT * FROM users ORDER BY username;

-- name: GetUserByID :one
SELECT * FROM users WHERE user_id = $1;

-- name: GetUserByName :one
SELECT * FROM users WHERE username = $1;

-- name: UpdateUser :one
UPDATE users SET username = $1, updated_at = $2 WHERE user_id = $3
RETURNING *;

-- name: DeleteUserByID :exec
DELETE FROM users WHERE user_id = $1;


