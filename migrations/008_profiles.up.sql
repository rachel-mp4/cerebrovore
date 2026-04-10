CREATE TABLE profiles (
  username TEXT NOT NULL PRIMARY KEY,
  display_name TEXT,
  avatar TEXT,
  is_pixel BOOLEAN,
	color INTEGER CHECK (color BETWEEN 0 AND 16777215),
	status TEXT,
	bio TEXT,
	is_mono BOOLEAN,
	friends_header TEXT,
	posts_header TEXT,
	links_header TEXT,
	at_identifier TEXT
);

CREATE TABLE profile_friends (
  profile_username TEXT NOT NULL,
  FOREIGN KEY (profile_username) REFERENCES profiles(username) ON DELETE CASCADE,
  friend TEXT NOT NULL,
  FOREIGN KEY (friend) REFERENCES profiles(username) ON DELETE CASCADE,
  comment TEXT,
  id INTEGER NOT NULL CHECK (id BETWEEN 0 and 11),
  PRIMARY KEY(profile_username, id)
);

CREATE TABLE profile_posts (
  profile_username TEXT NOT NULL,
  FOREIGN KEY (profile_username) REFERENCES profiles(username) ON DELETE CASCADE,
  post_id INTEGER NOT NULL,
  FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
  comment TEXT,
  just_body BOOLEAN,
  id INTEGER NOT NULL CHECK (id BETWEEN 0 and 11),
  PRIMARY KEY(profile_username, id)
);

CREATE TABLE profile_links (
  profile_username TEXT NOT NULL,
  FOREIGN KEY (profile_username) REFERENCES profiles(username) ON DELETE CASCADE,
  link TEXT NOT NULL,
  comment TEXT,
  id INTEGER NOT NULL CHECK (id BETWEEN 0 and 11),
  PRIMARY KEY(profile_username, id)
);
