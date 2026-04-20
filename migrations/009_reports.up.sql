ALTER TABLE profiles ADD COLUMN deleted BOOLEAN NOT NULL DEFAULT FALSE;

CREATE TABLE reports (
  id SERIAL PRIMARY KEY,
  reporter TEXT NOT NULL,
  reported TEXT NOT NULL,
  post_id INTEGER,
  FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE SET NULL,
  for_profile BOOLEAN NOT NULL,
  reason TEXT,
  reviewed_by TEXT
);

CREATE INDEX idx_reports_reported ON reports(reported);

CREATE TABLE moderators (
  username TEXT NOT NULL PRIMARY KEY
);
