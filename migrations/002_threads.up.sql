CREATE TABLE threads (
  id INTEGER PRIMARY KEY,
	posted_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	bumped_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	topic TEXT
);

CREATE INDEX idx_threads_bumped_at ON threads(bumped_at DESC);

CREATE TABLE posts (
  id INTEGER PRIMARY KEY,
  thread_id INTEGER NOT NULL,
	FOREIGN KEY (thread_id) REFERENCES threads(id) ON DELETE CASCADE DEFERRABLE INITIALLY DEFERRED,
  username TEXT NOT NULL,
  anon BOOLEAN NOT NULL,
  nick TEXT,
	color INTEGER CHECK (color BETWEEN 0 AND 16777215),
	posted_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_posts_thread_id_id ON posts(thread_id, id);

CREATE TABLE text_posts (
  post_id INTEGER PRIMARY KEY,
  FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
  body TEXT NOT NULL
);

CREATE TABLE post_replies (
  from_id INTEGER NOT NULL,
  FOREIGN KEY (from_id) REFERENCES posts(id) ON DELETE CASCADE,
  to_id INTEGER NOT NULL,
  FOREIGN KEY (to_id) REFERENCES posts(id) ON DELETE CASCADE,
  PRIMARY KEY (from_id, to_id)
);

CREATE INDEX idx_post_replies_to_id ON post_replies(to_id);

CREATE TABLE pending_post_replies (
  from_id INTEGER NOT NULL,
  FOREIGN KEY (from_id) REFERENCES posts(id) ON DELETE CASCADE,
  to_id INTEGER NOT NULL,
  PRIMARY KEY (from_id, to_id)
);

CREATE INDEX idx_pending_reply_to ON pending_post_replies(to_id);

CREATE TABLE watched_threads (
  thread_id INTEGER NOT NULL,
  FOREIGN KEY (thread_id) REFERENCES threads(id) ON DELETE CASCADE,
  username TEXT NOT NULL,
  PRIMARY KEY (thread_id, username)
);

CREATE INDEX idx_watched_threads_username ON watched_threads(username);
