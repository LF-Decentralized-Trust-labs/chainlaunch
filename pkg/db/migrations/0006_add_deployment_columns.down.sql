-- SQLite doesn't support multiple column drops in a single ALTER TABLE statement
-- We need to drop each column separately
ALTER TABLE plugins DROP COLUMN deployment_metadata;
ALTER TABLE plugins DROP COLUMN deployment_status;