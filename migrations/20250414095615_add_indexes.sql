-- +goose Up
CREATE INDEX IF NOT EXISTS idx_people_name ON people(name);
CREATE INDEX IF NOT EXISTS idx_people_age ON people(age);

-- +goose Down
DROP INDEX IF EXISTS idx_people_name;
DROP INDEX IF EXISTS idx_people_age;
