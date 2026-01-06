-- name: CreateSession :exec
INSERT INTO sessions (id, token, device_name, ip_address, user_agent, expires_at, created_at, user_id)
VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: FindSessionByID :one
SELECT * FROM sessions WHERE id = ?;

-- name: FindSessionByToken :one
SELECT * FROM sessions WHERE token = ?;

-- name: FindSessionsByUserID :many
SELECT * FROM sessions
WHERE user_id = ? AND expires_at > ?;

-- name: DeleteSessionByID :exec
DELETE FROM sessions WHERE id = ?;

-- name: DeleteSessionsByUserID :exec
DELETE FROM sessions WHERE user_id = ?;

-- name: DeleteSessionsByUserExceptID :exec
DELETE FROM sessions
WHERE user_id = ? AND id != ?;
