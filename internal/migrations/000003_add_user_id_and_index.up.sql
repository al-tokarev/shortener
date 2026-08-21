ALTER TABLE urls ADD COLUMN user_id VARCHAR(255) NOT NULL DEFAULT '';
CREATE INDEX idx_urls_user_id ON urls(user_id);