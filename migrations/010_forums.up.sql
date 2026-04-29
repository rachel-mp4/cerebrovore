ALTER TABLE threads ADD COLUMN dead BOOLEAN NOT NULL DEFAULT FALSE;
CREATE INDEX idx_posts_username ON posts(username);
