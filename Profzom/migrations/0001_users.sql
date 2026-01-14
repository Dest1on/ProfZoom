-- +goose Up
CREATE TABLE users (
    id UUID PRIMARY KEY,
    phone TEXT NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE users;
