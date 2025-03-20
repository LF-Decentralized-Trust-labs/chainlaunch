-- Consolidated migration file that includes all previous migrations

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
    ('DATABASE'),
    ('VAULT'), 
    ('HSM');

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
    ('CREATING'),
    ('RUNNING'),
    ('STOPPED'),
    ('ERROR'),
    ('DELETED');

CREATE TABLE node_key_types (
    name TEXT PRIMARY KEY
);
INSERT INTO node_key_types (name) VALUES 
    ('SIGNING'),
    ('TLS');

-- Main tables
CREATE TABLE key_providers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    is_default INTEGER NOT NULL DEFAULT 0,
    config TEXT NOT NULL DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE keys (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT,
    algorithm TEXT NOT NULL,
    key_size INTEGER,
    curve TEXT,
    format TEXT NOT NULL,
    public_key TEXT NOT NULL,
    private_key TEXT NOT NULL DEFAULT '',
    certificate TEXT,
    status TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP,
    last_rotated_at TIMESTAMP,
    signing_key_id INTEGER REFERENCES keys(id),
    sha256_fingerprint TEXT NOT NULL,
    sha1_fingerprint TEXT NOT NULL,
    provider_id INTEGER NOT NULL REFERENCES key_providers(id),
    user_id INTEGER NOT NULL,
    is_ca INTEGER NOT NULL DEFAULT 0,
    ethereum_address TEXT
);

CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    name TEXT,
    email TEXT,
    role TEXT DEFAULT 'user',
    provider TEXT,
    provider_id TEXT,
    avatar_url TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_login_at TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE INDEX idx_users_username ON users(username);

CREATE TABLE networks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    network_id TEXT,
    platform TEXT NOT NULL REFERENCES blockchain_platforms(name),
    status TEXT NOT NULL REFERENCES node_statuses(name),
    description TEXT,
    config TEXT,
    deployment_config TEXT,
    exposed_ports TEXT,
    domain TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER REFERENCES users(id),
    updated_at TIMESTAMP,
    genesis_block_b64 TEXT,
    current_config_block_b64 TEXT
);

CREATE TABLE fabric_organizations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    msp_id TEXT NOT NULL,
    description TEXT,
    config TEXT,
    ca_config TEXT,
    sign_key_id INTEGER REFERENCES keys(id),
    tls_root_key_id INTEGER REFERENCES keys(id),
    admin_tls_key_id INTEGER REFERENCES keys(id),
    admin_sign_key_id INTEGER REFERENCES keys(id),
    client_sign_key_id INTEGER REFERENCES keys(id),
    provider_id INTEGER REFERENCES key_providers(id),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER REFERENCES users(id),
    updated_at TIMESTAMP
);

CREATE TABLE nodes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    slug TEXT NOT NULL DEFAULT '',
    platform TEXT NOT NULL REFERENCES blockchain_platforms(name),
    status TEXT NOT NULL REFERENCES node_statuses(name),
    description TEXT,
    network_id INTEGER REFERENCES networks(id),
    config TEXT,
    resources TEXT,
    endpoint TEXT,
    public_endpoint TEXT,
    p2p_address TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER REFERENCES users(id),
    updated_at TIMESTAMP,
    fabric_organization_id INTEGER REFERENCES fabric_organizations(id),
    node_type TEXT REFERENCES node_types(name),
    node_config TEXT,
    deployment_config TEXT
);

CREATE UNIQUE INDEX idx_nodes_slug ON nodes(slug);

-- Network nodes table for many-to-many relationship
CREATE TABLE network_nodes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    network_id INTEGER NOT NULL REFERENCES networks(id) ON DELETE CASCADE,
    node_id INTEGER NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    role TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    config TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_network_nodes_network_id ON network_nodes(network_id);
CREATE INDEX idx_network_nodes_node_id ON network_nodes(node_id);
CREATE UNIQUE INDEX idx_network_nodes_network_node ON network_nodes(network_id, node_id);

CREATE TABLE node_keys (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    node_id INTEGER NOT NULL REFERENCES nodes(id),
    key_id INTEGER NOT NULL REFERENCES keys(id),
    key_type TEXT NOT NULL REFERENCES node_key_types(name),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);


-- Node events table for tracking node status changes
CREATE TABLE node_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    node_id INTEGER NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    event_type TEXT NOT NULL,
    description TEXT NOT NULL,
    data TEXT,
    status TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (node_id) REFERENCES nodes(id) ON DELETE CASCADE
);

CREATE INDEX idx_node_events_node_id ON node_events(node_id);
CREATE INDEX idx_node_events_created_at ON node_events(created_at);
CREATE INDEX idx_node_events_event_type ON node_events(event_type);

-- Backup tables
CREATE TABLE backup_targets (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    bucket_name TEXT,
    region TEXT,
    endpoint TEXT,
    bucket_path TEXT,
    access_key_id TEXT,
    secret_key TEXT,
    s3_path_style BOOLEAN,
    restic_password TEXT,
    type VARCHAR(50) NOT NULL, -- e.g., 'S3', 'LOCAL'
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE backup_schedules (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    cron_expression VARCHAR(100) NOT NULL,
    target_id INTEGER NOT NULL REFERENCES backup_targets(id),
    retention_days INTEGER NOT NULL DEFAULT 30,
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
    last_run_at TIMESTAMP,
    next_run_at TIMESTAMP
);

CREATE TABLE backups (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    schedule_id INTEGER REFERENCES backup_schedules(id),
    target_id INTEGER NOT NULL REFERENCES backup_targets(id),
    status VARCHAR(50) NOT NULL, -- 'PENDING', 'IN_PROGRESS', 'COMPLETED', 'FAILED'
    size_bytes BIGINT DEFAULT 0,
    started_at TIMESTAMP NOT NULL,
    completed_at TIMESTAMP,
    error_message TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    notification_sent INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX idx_backup_schedules_target_id ON backup_schedules(target_id);
CREATE INDEX idx_backups_schedule_id ON backups(schedule_id);
CREATE INDEX idx_backups_target_id ON backups(target_id);
CREATE INDEX idx_backups_status ON backups(status);

-- Notification tables
CREATE TABLE notification_providers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    type TEXT NOT NULL, -- 'EMAIL', 'SLACK', 'TELEGRAM', etc.
    config TEXT NOT NULL, -- JSON configuration
    is_default BOOLEAN NOT NULL DEFAULT false,
    is_enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    notify_node_downtime BOOLEAN NOT NULL DEFAULT true,
    notify_backup_success BOOLEAN NOT NULL DEFAULT true,
    notify_backup_failure BOOLEAN NOT NULL DEFAULT true,
    notify_s3_connection_issue BOOLEAN NOT NULL DEFAULT true,
    last_test_at TIMESTAMP,
    last_test_status TEXT,
    last_test_message TEXT
);

CREATE INDEX idx_notification_providers_type ON notification_providers(type);
CREATE UNIQUE INDEX idx_notification_providers_name ON notification_providers(name);

-- Session tables
CREATE TABLE sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT NOT NULL UNIQUE, 
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token TEXT NOT NULL,
    ip_address TEXT,
    user_agent TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    last_activity_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_sessions_session_id ON sessions(session_id);

CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_token ON sessions(token);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);
