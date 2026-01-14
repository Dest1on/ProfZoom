-- +goose Up
CREATE EXTENSION IF NOT EXISTS pgcrypto;

ALTER TABLE otp_codes
    ADD COLUMN code_hash TEXT,
    ADD COLUMN attempts INT NOT NULL DEFAULT 0,
    ADD COLUMN requested_at TIMESTAMP NOT NULL DEFAULT now();

UPDATE otp_codes SET code_hash = encode(digest(code, 'sha256'), 'hex');

ALTER TABLE otp_codes
    DROP COLUMN code;

-- +goose Down
ALTER TABLE otp_codes
    ADD COLUMN code TEXT;

ALTER TABLE otp_codes
    DROP COLUMN code_hash,
    DROP COLUMN attempts,
    DROP COLUMN requested_at;
