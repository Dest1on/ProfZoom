-- +goose Up
CREATE TABLE auth_codes (
    phone TEXT NOT NULL,
    code_hash TEXT NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    attempts INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE INDEX idx_auth_codes_phone ON auth_codes(phone);

-- +goose Down
DROP TABLE auth_codes;
