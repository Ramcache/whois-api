-- +goose Up
CREATE TABLE IF NOT EXISTS people (
                                      id SERIAL PRIMARY KEY,
                                      name TEXT NOT NULL,
                                      surname TEXT NOT NULL,
                                      patronymic TEXT,
                                      age INT,
                                      gender TEXT,
                                      nationality TEXT
);


-- +goose Down
DROP TABLE IF EXISTS people;
