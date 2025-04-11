-- Add CRL-related fields to fabric_organizations
ALTER TABLE fabric_organizations
ADD COLUMN crl_key_id INTEGER REFERENCES keys(id);
ALTER TABLE fabric_organizations
ADD COLUMN crl_last_update TIMESTAMP;

-- Create table for revoked certificates
CREATE TABLE fabric_revoked_certificates (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    fabric_organization_id INTEGER NOT NULL REFERENCES fabric_organizations(id) ON DELETE CASCADE,
    serial_number TEXT NOT NULL,  -- Store as hex string for compatibility
    revocation_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    reason INTEGER NOT NULL,  -- RFC 5280 revocation reason code
    issuer_certificate_id INTEGER REFERENCES keys(id),  -- Reference to the certificate that issued this one
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(fabric_organization_id, serial_number)
);

-- Create indexes for performance
CREATE INDEX idx_revoked_certs_org_id ON fabric_revoked_certificates(fabric_organization_id);
CREATE INDEX idx_revoked_certs_serial ON fabric_revoked_certificates(serial_number);
CREATE INDEX idx_revoked_certs_revocation_time ON fabric_revoked_certificates(revocation_time);