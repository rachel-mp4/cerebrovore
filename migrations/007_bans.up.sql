CREATE TABLE bans (
  id SERIAL PRIMARY KEY,
  username TEXT NOT NULL,
  post_id INTEGER,
  FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE SET NULL,
  reason TEXT,
  comment TEXT,
  response TEXT,
  banned_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  until TIMESTAMPTZ NOT NULL,
  moderator TEXT NOT NULL,
  repealed BOOLEAN
);

CREATE INDEX idx_bans_username ON bans(username);

CREATE INDEX idx_posts_username ON posts(username);
