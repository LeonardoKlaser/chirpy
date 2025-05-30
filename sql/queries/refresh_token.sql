-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (token, created_at, updated_at, user_id, expires_at)
VALUES (
    $1, NOW(), NOW(), $2, $3
)
RETURNING *;


-- name: GetValidRefreshToken :one 
SELECT EXISTS (SELECT 1 FROM refresh_tokens WHERE token = $1 AND expires_at > NOW() AND revoked_at IS NULL);

-- name: RevokeRefreshToken :one
UPDATE refresh_tokens SET revoked_at = NOW(), updated_at = NOW() WHERE token = $1;

-- name: GetUserForValidRefreshToken :one
SELECT u.*
FROM users u
INNER JOIN refresh_tokens rt ON u.id = rt.user_id
WHERE rt.token = $1
  AND rt.expires_at > NOW()      
  AND rt.revoked_at IS NULL;    
