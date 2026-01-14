-- +goose Up
ALTER TABLE vacancies
    ADD COLUMN role TEXT NOT NULL DEFAULT '',
    ADD COLUMN tasks TEXT[] NOT NULL DEFAULT ARRAY[]::TEXT[],
    ADD COLUMN hours_week INT NOT NULL DEFAULT 0,
    ADD COLUMN format TEXT NOT NULL DEFAULT '',
    ADD COLUMN city TEXT NOT NULL DEFAULT '',
    ADD COLUMN compensation TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE vacancies
    DROP COLUMN role,
    DROP COLUMN tasks,
    DROP COLUMN hours_week,
    DROP COLUMN format,
    DROP COLUMN city,
    DROP COLUMN compensation;
