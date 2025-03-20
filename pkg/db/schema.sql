-- Enum-like tables (since SQLite doesn't support enums)
CREATE TABLE blockchain_platforms (
    name TEXT PRIMARY KEY
);
INSERT INTO blockchain_platforms (name) VALUES 
    ('FABRIC'),
    ('BESU');

CREATE TABLE key_provider_types (
    name TEXT PRIMARY KEY
);
INSERT INTO key_provider_types (name) VALUES 
    ('database'),
    ('vault'),
    ('hsm');

CREATE TABLE node_types (
    name TEXT PRIMARY KEY
);
INSERT INTO node_types (name) VALUES 
    ('FABRIC_PEER'),
    ('FABRIC_ORDERER'),
    ('FABRIC_CA'),
    ('BESU_VALIDATOR'),
    ('BESU_BOOTNODE'),
    ('BESU_FULLNODE');

CREATE TABLE node_statuses (
    name TEXT PRIMARY KEY
);
INSERT INTO node_statuses (name) VALUES 
    ('creating'),
    ('running'),
    ('stopped'),
    ('error'),
    ('deleted');

CREATE TABLE node_key_types (
    name TEXT PRIMARY KEY
);
INSERT INTO node_key_types (name) VALUES 
    ('signing'),
    ('tls');

-- Main tables
CREATE TABLE key_providers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    type TEXT NOT NULL REFERENCES key_provider_types(name),
    is_default INTEGER NOT NULL DEFAULT FALSE,
    config JSON,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE keys (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT,
    provider_id INTEGER NOT NULL REFERENCES key_providers(id),
    private_key TEXT NOT NULL,
    public_key TEXT NOT NULL,
    sha256_fingerprint TEXT,
    sha1_fingerprint TEXT,
    certificate TEXT,
    is_ca INTEGER,
    signing_key_id INTEGER REFERENCES keys(id),
    algorithm TEXT NOT NULL,
    key_size INTEGER,
    curve TEXT,
    format TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP,
    status TEXT NOT NULL DEFAULT 'active',
    last_rotated_at TIMESTAMP
);

CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    username TEXT,
    role TEXT NOT NULL DEFAULT 'user',
    provider TEXT NOT NULL,
    provider_id TEXT NOT NULL,
    avatar_url TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE networks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    platform TEXT NOT NULL REFERENCES blockchain_platforms(name),
    status TEXT NOT NULL REFERENCES node_statuses(name),
    description TEXT,
    config JSON,
    deployment_config JSON,
    exposed_ports JSON,
    domain TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER REFERENCES users(id),
    updated_at TIMESTAMP
);

CREATE TABLE fabric_organizations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    msp_id TEXT NOT NULL,
    description TEXT,
    config JSON,
    ca_config JSON,
    sign_key_id INTEGER REFERENCES keys(id),
    tls_root_key_id INTEGER REFERENCES keys(id),
    provider_id INTEGER REFERENCES key_providers(id),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER REFERENCES users(id),
    updated_at TIMESTAMP
);

CREATE TABLE nodes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    platform TEXT NOT NULL REFERENCES blockchain_platforms(name),
    status TEXT NOT NULL REFERENCES node_statuses(name),
    description TEXT,
    network_id INTEGER REFERENCES networks(id),
    config JSON,
    resources JSON,
    endpoint TEXT,
    public_endpoint TEXT,
    p2p_address TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER REFERENCES users(id),
    updated_at TIMESTAMP,
    fabric_organization_id INTEGER REFERENCES fabric_organizations(id),
    node_type TEXT REFERENCES node_types(name),
    node_config JSON,
    deployment_config JSON
);

CREATE TABLE node_keys (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    node_id INTEGER NOT NULL REFERENCES nodes(id),
    key_id INTEGER NOT NULL REFERENCES keys(id),
    key_type TEXT NOT NULL REFERENCES node_key_types(name),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE fabric_org_nodes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    organization_id INTEGER NOT NULL REFERENCES fabric_organizations(id),
    node_id INTEGER NOT NULL REFERENCES nodes(id),
    role TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE proposals (
    id TEXT PRIMARY KEY,
    network_id INTEGER NOT NULL REFERENCES networks(id),
    channel_name TEXT NOT NULL,
    status TEXT NOT NULL,
    operations JSON NOT NULL,
    preview_json TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by TEXT NOT NULL,
    updated_at TIMESTAMP
);

CREATE TABLE proposal_signatures (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    proposal_id TEXT NOT NULL REFERENCES proposals(id),
    msp_id TEXT NOT NULL,
    signed_by TEXT NOT NULL,
    signed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    signature BLOB NOT NULL
);

CREATE INDEX idx_proposals_network_id ON proposals(network_id);
CREATE INDEX idx_proposal_signatures_proposal_id ON proposal_signatures(proposal_id);

-- Table for storing proposal submission notifications
CREATE TABLE IF NOT EXISTS proposal_submission_notifications (
  id SERIAL PRIMARY KEY,
  proposal_id TEXT NOT NULL,
  network_id BIGINT NOT NULL,
  tx_id TEXT NOT NULL,
  submitted_by TEXT NOT NULL,
  submitted_at TIMESTAMP NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  FOREIGN KEY (proposal_id) REFERENCES governance_proposals(proposal_id) ON DELETE CASCADE,
  FOREIGN KEY (network_id) REFERENCES networks(id) ON DELETE CASCADE
); 