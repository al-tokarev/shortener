-- migrate -database "postgres://alexandr@localhost:5432/shortener?sslmode=disable" -path ./internal/migrations up

CREATE TABLE IF NOT EXISTS urls (
    id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    short VARCHAR(16) NOT NULL,
    original VARCHAR(2048) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);