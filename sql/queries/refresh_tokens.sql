-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens(
    token,
    created_at,
    updated_at,
    user_id,
    expires_at,
    revoked_at
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
) RETURNING *;


-- name: GetUserFromRefreshToken :one
SELECT * FROM refresh_tokens 
WHERE token = $1 AND revoked_at IS NULL;


-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens 
SET updated_at = Now(), revoked_at=Now()
WHERE token=$1; 