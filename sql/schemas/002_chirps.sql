-- +goose Up
CREATE TABLE chirps (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    body VARCHAR(141) NOT NULL,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE
);
