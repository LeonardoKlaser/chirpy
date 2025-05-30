-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, password)
VALUES (
    gen_random_uuid(), NOW(), NOW(), $1, $2
)
RETURNING *;

-- name: DeleteUsers :execresult
TRUNCATE TABLE users CASCADE;

-- name: GetUserByEmail :one 
SELECT id, created_at, updated_at, email, password FROM users WHERE email = $1;


-- name: UpdateUserById :one
UPDATE users SET email = $1, password = $2 WHERE id = $3 RETURNING id, created_at, updated_at, email;

-- name: UpgradeToRed :execresult

UPDATE users SET is_chirpy_red = TRUE WHERE id = $1; 

