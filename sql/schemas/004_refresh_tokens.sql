-- +goose Up
CREATE TABLE refresh_tokens (
    token VARCHAR PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    expires_at TIMESTAMP,
    revoked_at TIMESTAMP
);

-- +goose Down
DROP TABLE refresh_tokens;