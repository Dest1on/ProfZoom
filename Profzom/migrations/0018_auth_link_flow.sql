-- +goose Up
ALTER TABLE users
    ALTER COLUMN phone DROP NOT NULL;

ALTER TABLE telegram_links
    ALTER COLUMN phone DROP NOT NULL;

ALTER TABLE telegram_link_tokens
    ALTER COLUMN phone DROP NOT NULL;

ALTER TABLE otp_codes
    ADD COLUMN user_id TEXT;

UPDATE otp_codes
SET user_id = phone
WHERE user_id IS NULL;

ALTER TABLE otp_codes
    ALTER COLUMN user_id SET NOT NULL;

ALTER TABLE otp_codes
    DROP CONSTRAINT otp_codes_pkey;

ALTER TABLE otp_codes
    ADD PRIMARY KEY (user_id);

DROP INDEX IF EXISTS idx_otp_codes_phone;

CREATE INDEX IF NOT EXISTS idx_otp_codes_user_id ON otp_codes(user_id);

-- +goose Down
DROP INDEX IF EXISTS idx_otp_codes_user_id;

ALTER TABLE otp_codes
    DROP CONSTRAINT otp_codes_pkey;

ALTER TABLE otp_codes
    ADD PRIMARY KEY (phone);

CREATE INDEX IF NOT EXISTS idx_otp_codes_phone ON otp_codes(phone);

ALTER TABLE otp_codes
    ALTER COLUMN user_id DROP NOT NULL;

ALTER TABLE otp_codes
    DROP COLUMN user_id;

ALTER TABLE telegram_link_tokens
    ALTER COLUMN phone SET NOT NULL;

ALTER TABLE telegram_links
    ALTER COLUMN phone SET NOT NULL;

ALTER TABLE users
    ALTER COLUMN phone SET NOT NULL;
