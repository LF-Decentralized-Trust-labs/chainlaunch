ALTER TABLE plugins
ADD COLUMN deployment_metadata JSON DEFAULT NULL;

ALTER TABLE plugins
ADD COLUMN deployment_status TEXT DEFAULT 'pending';
