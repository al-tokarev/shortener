-- migrate -database "postgres://alexandr@localhost:5432/shortener?sslmode=disable" -path ./migrations up 

CREATE TABLE urls (
    id SERIAL PRIMARY KEY,
    short VARCHAR(16) NOT NULL,
    original TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);