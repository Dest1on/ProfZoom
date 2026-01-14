-- +goose Up
CREATE INDEX IF NOT EXISTS idx_applications_vacancy ON applications(vacancy_id);
CREATE INDEX IF NOT EXISTS idx_messages_sender ON messages(sender_id);

-- +goose Down
DROP INDEX IF EXISTS idx_messages_sender;
DROP INDEX IF EXISTS idx_applications_vacancy;
