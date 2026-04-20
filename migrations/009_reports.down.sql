DROP TABLE IF EXISTS moderators;
DROP INDEX IF EXISTS idx_reports_reported;
DROP TABLE IF EXISTS reports;
ALTER TABLE profiles DROP COLUMN IF EXISTS deleted;
