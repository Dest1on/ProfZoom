-- +goose Up
CREATE TABLE otp_codes (
    phone TEXT PRIMARY KEY,
    code TEXT NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE INDEX idx_otp_codes_phone ON otp_codes(phone);

-- +goose Down
DROP TABLE otp_codes;
