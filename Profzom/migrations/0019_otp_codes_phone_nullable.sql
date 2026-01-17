-- +goose Up
ALTER TABLE otp_codes
    ALTER COLUMN phone DROP NOT NULL;

-- +goose Down
ALTER TABLE otp_codes
    ALTER COLUMN phone SET NOT NULL;
