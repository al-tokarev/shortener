ALTER TABLE urls ADD COLUMN is_deleted BOOLEAN NOT NULL DEFAULT false;
CREATE INDEX idx_urls_is_deleted ON urls(is_deleted);