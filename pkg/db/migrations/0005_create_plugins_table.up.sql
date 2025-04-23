CREATE TABLE IF NOT EXISTS plugins (
    name VARCHAR(255) PRIMARY KEY,
    api_version VARCHAR(50) NOT NULL,
    kind VARCHAR(50) NOT NULL,
    metadata JSON NOT NULL,
    spec JSON NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_plugins_kind ON plugins(kind);
CREATE INDEX idx_plugins_created_at ON plugins(created_at);
CREATE INDEX idx_plugins_updated_at ON plugins(updated_at); 