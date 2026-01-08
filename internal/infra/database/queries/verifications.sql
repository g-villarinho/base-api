-- name: CreateVerification :exec
INSERT INTO verifications (id, flow, token, created_at, expires_at, payload, user_id)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: FindVerificationByID :one
SELECT * FROM verifications WHERE id = $1;

-- name: FindVerificationByToken :one
SELECT * FROM verifications WHERE token = $1;

-- name: DeleteVerification :exec
DELETE FROM verifications WHERE id = $1;

-- name: FindValidVerificationByUserIDAndFlow :one
SELECT * FROM verifications
WHERE user_id = $1
  AND flow = $2
  AND expires_at > $3
ORDER BY created_at DESC
LIMIT 1;

-- name: DeleteVerificationsByUserIDAndFlow :exec
DELETE FROM verifications
WHERE user_id = $1
  AND flow = $2
  AND expires_at > $3;
