-- name: CreateUser :exec
INSERT INTO users (id, name, email, status, password_hash, created_at, updated_at, email_confirmed_at, blocked_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: FindUserByID :one
SELECT * FROM users WHERE id = ?;

-- name: FindUserByEmail :one
SELECT * FROM users WHERE email = ?;

-- name: ExistsByEmail :one
SELECT COUNT(*) > 0 FROM users WHERE email = ?;

-- name: UpdateUserEmail :exec
UPDATE users
SET email = ?, updated_at = ?
WHERE id = ?;

-- name: UpdateUserPassword :exec
UPDATE users
SET password_hash = ?, updated_at = ?
WHERE id = ?;

-- name: VerifyUserEmail :exec
UPDATE users
SET status = 'ACTIVE',
    updated_at = ?,
    email_confirmed_at = ?
WHERE id = ?;

-- name: BlockUser :exec
UPDATE users
SET status = 'BLOCKED',
    updated_at = ?,
    blocked_at = ?
WHERE id = ?;

-- name: UpdateUserName :exec
UPDATE users
SET name = ?, updated_at = ?
WHERE id = ?;
