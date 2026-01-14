-- +goose Up
CREATE EXTENSION IF NOT EXISTS pgcrypto;

ALTER TABLE refresh_tokens
    ADD COLUMN token_hash TEXT;

UPDATE refresh_tokens SET token_hash = encode(digest(token, 'sha256'), 'hex');

ALTER TABLE refresh_tokens
    DROP COLUMN token;

CREATE UNIQUE INDEX IF NOT EXISTS idx_refresh_tokens_hash ON refresh_tokens(token_hash);

-- +goose Down
ALTER TABLE refresh_tokens
    ADD COLUMN token TEXT;

ALTER TABLE refresh_tokens
    DROP COLUMN token_hash;

DROP INDEX IF EXISTS idx_refresh_tokens_hash;
DROP INDEX IF EXISTS idx_refresh_tokens_token;
