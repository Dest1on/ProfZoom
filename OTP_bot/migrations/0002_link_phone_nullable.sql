-- +goose Up
ALTER TABLE telegram_links
    ALTER COLUMN phone DROP NOT NULL;

ALTER TABLE telegram_link_tokens
    ALTER COLUMN phone DROP NOT NULL;

-- +goose Down
ALTER TABLE telegram_link_tokens
    ALTER COLUMN phone SET NOT NULL;

ALTER TABLE telegram_links
    ALTER COLUMN phone SET NOT NULL;
