DROP INDEX IF EXISTS idx_watched_threads_username;
DROP TABLE IF EXISTS watched_threads;
DROP INDEX IF EXISTS idx_pending_reply_to;
DROP TABLE IF EXISTS pending_post_replies; 
DROP INDEX IF EXISTS idx_post_replies_to_id;
DROP TABLE IF EXISTS post_replies;
DROP TABLE IF EXISTS text_posts;
DROP INDEX IF EXISTS idx_posts_thread_id_id;
DROP TABLE IF EXISTS posts;
DROP INDEX IF EXISTS idx_threads_bumped_at;
DROP TABLE IF EXISTS threads;

