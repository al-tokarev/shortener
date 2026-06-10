-- migrate -database "postgres://alexandr@localhost:5432/shortener?sslmode=disable" -path ./migrations up 

CREATE UNIQUE INDEX IF NOT EXISTS idx_original_url ON urls (original);