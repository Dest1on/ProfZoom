-- +goose Up
ALTER TABLE student_profiles
    ADD COLUMN name TEXT NOT NULL DEFAULT '',
    ADD COLUMN course INT NOT NULL DEFAULT 0;

UPDATE student_profiles
SET name = trim(concat_ws(' ', first_name, last_name))
WHERE name = '';

ALTER TABLE vacancies
    ADD COLUMN vacancy_type TEXT NOT NULL DEFAULT '',
    ADD COLUMN requirements TEXT[] NOT NULL DEFAULT ARRAY[]::TEXT[],
    ADD COLUMN conditions TEXT[] NOT NULL DEFAULT ARRAY[]::TEXT[],
    ADD COLUMN salary TEXT NOT NULL DEFAULT '';

UPDATE vacancies
SET vacancy_type = role
WHERE vacancy_type = '' AND role <> '';

UPDATE vacancies
SET requirements = tasks
WHERE array_length(requirements, 1) IS NULL AND array_length(tasks, 1) IS NOT NULL;

UPDATE vacancies
SET conditions = CASE WHEN format <> '' THEN ARRAY[format] ELSE conditions END
WHERE array_length(conditions, 1) IS NULL;

UPDATE vacancies
SET salary = compensation
WHERE salary = '' AND compensation <> '';

UPDATE vacancies
SET location = COALESCE(NULLIF(location, ''), city)
WHERE city <> '' AND location = '';

-- +goose Down
ALTER TABLE vacancies
    DROP COLUMN vacancy_type,
    DROP COLUMN requirements,
    DROP COLUMN conditions,
    DROP COLUMN salary;

ALTER TABLE student_profiles
    DROP COLUMN name,
    DROP COLUMN course;
