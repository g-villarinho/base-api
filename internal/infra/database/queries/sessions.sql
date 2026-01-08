-- name: CreateSession :exec
INSERT INTO sessions (id, token, device_name, ip_address, user_agent, expires_at, created_at, user_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: FindSessionByID :one
SELECT * FROM sessions WHERE id = $1;

-- name: FindSessionByToken :one
SELECT * FROM sessions WHERE token = $1;

-- name: FindSessionsByUserID :many
SELECT * FROM sessions
WHERE user_id = $1 AND expires_at > $2;

-- name: DeleteSessionByID :exec
DELETE FROM sessions WHERE id = $1;

-- name: DeleteSessionsByUserID :exec
DELETE FROM sessions WHERE user_id = $1;

-- name: DeleteSessionsByUserExceptID :exec
DELETE FROM sessions
WHERE user_id = $1 AND id != $2;
