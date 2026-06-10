-- migrate -database "postgres://alexandr@localhost:5432/shortener?sslmode=disable" -path ./internal/migrations down 

DROP TABLE IF EXISTS urls; 