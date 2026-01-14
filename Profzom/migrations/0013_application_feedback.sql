-- +goose Up
ALTER TABLE applications
    ADD COLUMN feedback TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE applications
    DROP COLUMN feedback;
