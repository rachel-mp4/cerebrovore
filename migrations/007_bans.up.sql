CREATE TABLE bans (
  id SERIAL PRIMARY KEY,
  username TEXT NOT NULL,
  for INTEGER,
  reason TEXT,
  until TIMESTAMPTZ NOT NULL,
  moderator TEXT NOT NULL
);
