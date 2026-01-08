-- name: CreateUser :exec
INSERT INTO users (id, name, email, status, password_hash, created_at, updated_at, email_confirmed_at, blocked_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);

-- name: FindUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: FindUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: ExistsByEmail :one
SELECT EXISTS(SELECT 1 FROM users WHERE email = $1);

-- name: UpdateUserEmail :exec
UPDATE users
SET email = $1, updated_at = $2
WHERE id = $3;

-- name: UpdateUserPassword :exec
UPDATE users
SET password_hash = $1, updated_at = $2
WHERE id = $3;

-- name: VerifyUserEmail :exec
UPDATE users
SET status = 'ACTIVE',
    updated_at = $1,
    email_confirmed_at = $2
WHERE id = $3;

-- name: BlockUser :exec
UPDATE users
SET status = 'BLOCKED',
    updated_at = $1,
    blocked_at = $2
WHERE id = $3;

-- name: UpdateUserName :exec
UPDATE users
SET name = $1, updated_at = $2
WHERE id = $3;
