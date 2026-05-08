CREATE TABLE notifications (
  id SERIAL PRIMARY KEY,
  username TEXT NOT NULL
);

CREATE INDEX idx_notifications_username_id ON notifications(username, id);

CREATE TABLE read_notifications (
  username TEXT PRIMARY KEY,
  notification_id BIGINT,
  FOREIGN KEY (notification_id) REFERENCES notifications(id) ON DELETE CASCADE
);

CREATE TABLE reply_notifications (
  notification_id BIGINT PRIMARY KEY,
  FOREIGN KEY (notification_id) REFERENCES notifications(id) ON DELETE CASCADE,
  post_id INTEGER NOT NULL,
  FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE
);

CREATE TABLE mention_notifications (
  notification_id BIGINT PRIMARY KEY,
  FOREIGN KEY (notification_id) REFERENCES notifications(id) ON DELETE CASCADE,
  post_id INTEGER NOT NULL,
  FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE
);

ALTER TABLE watched_threads ADD COLUMN notified BOOLEAN NOT NULL DEFAULT FALSE;

CREATE TABLE watch_notifications (
  notification_id BIGINT PRIMARY KEY,
  FOREIGN KEY (notification_id) REFERENCES notifications(id) ON DELETE CASCADE,
  thread_id INTEGER NOT NULL,
  FOREIGN KEY (thread_id) REFERENCES threads(id) ON DELETE CASCADE
);

CREATE TABLE poke_notifications (
  notification_id BIGINT PRIMARY KEY,
  FOREIGN KEY (notification_id) REFERENCES notifications(id) ON DELETE CASCADE,
  username TEXT NOT NULL,
  message TEXT
);

CREATE TABLE mod_notifications (
  notification_id BIGINT PRIMARY KEY,
  FOREIGN KEY (notification_id) REFERENCES notifications(id) ON DELETE CASCADE,
  reason TEXT NOT NULL
);

CREATE TABLE get_notifications (
  notification_id BIGINT PRIMARY KEY,
  FOREIGN KEY (notification_id) REFERENCES notifications(id) ON DELETE CASCADE,
  post_id INTEGER NOT NULL,
  FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
  value INTEGER NOT NULL
);
