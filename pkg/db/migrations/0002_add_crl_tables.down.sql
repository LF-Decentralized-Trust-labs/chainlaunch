-- Drop indexes
DROP INDEX IF EXISTS idx_revoked_certs_org_id;
DROP INDEX IF EXISTS idx_revoked_certs_serial;
DROP INDEX IF EXISTS idx_revoked_certs_revocation_time;

-- Drop the revoked certificates table
DROP TABLE IF EXISTS fabric_revoked_certificates;

-- Remove CRL-related columns from fabric_organizations
-- SQLite doesn't support DROP COLUMN directly, need to recreate table
PRAGMA foreign_keys=off;

CREATE TABLE fabric_organizations_new AS 
SELECT id, name, msp_id, sign_key_id, tls_root_key_id, admin_sign_key_id, provider_id, created_at, updated_at
FROM fabric_organizations;

DROP TABLE fabric_organizations;
ALTER TABLE fabric_organizations_new RENAME TO fabric_organizations;

-- Recreate foreign key constraints
CREATE INDEX IF NOT EXISTS idx_fabric_organizations_sign_key_id ON fabric_organizations(sign_key_id);
CREATE INDEX IF NOT EXISTS idx_fabric_organizations_tls_root_key_id ON fabric_organizations(tls_root_key_id);
CREATE INDEX IF NOT EXISTS idx_fabric_organizations_admin_sign_key_id ON fabric_organizations(admin_sign_key_id);
CREATE INDEX IF NOT EXISTS idx_fabric_organizations_provider_id ON fabric_organizations(provider_id);

PRAGMA foreign_keys=on;