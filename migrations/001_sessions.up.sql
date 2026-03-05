CREATE TABLE requests (
  state TEXT NOT NULL PRIMARY KEY,
  pkce_verifier TEXT NOT NULL
);

CREATE TABLE sessions (
  session_id TEXT NOT NULL PRIMARY KEY,
  username TEXT NOT NULL
);
