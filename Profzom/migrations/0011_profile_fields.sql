-- +goose Up
ALTER TABLE student_profiles
    ADD COLUMN skills TEXT[] NOT NULL DEFAULT ARRAY[]::TEXT[],
    ADD COLUMN hours_week INT NOT NULL DEFAULT 0,
    ADD COLUMN format TEXT NOT NULL DEFAULT '',
    ADD COLUMN resume_link TEXT NOT NULL DEFAULT '';

ALTER TABLE company_profiles
    ADD COLUMN contact_name TEXT NOT NULL DEFAULT '',
    ADD COLUMN contact_role TEXT NOT NULL DEFAULT '',
    ADD COLUMN contact_email TEXT NOT NULL DEFAULT '',
    ADD COLUMN contact_phone TEXT NOT NULL DEFAULT '',
    ADD COLUMN contact_telegram TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE company_profiles
    DROP COLUMN contact_name,
    DROP COLUMN contact_role,
    DROP COLUMN contact_email,
    DROP COLUMN contact_phone,
    DROP COLUMN contact_telegram;

ALTER TABLE student_profiles
    DROP COLUMN skills,
    DROP COLUMN hours_week,
    DROP COLUMN format,
    DROP COLUMN resume_link;
