CREATE TABLE fabric_chaincodes (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  slug TEXT NOT NULL UNIQUE,
  package_id TEXT NOT NULL,
  docker_image TEXT NOT NULL,
  host_port TEXT,
  container_port TEXT,
  status TEXT NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_fabric_chaincodes_slug ON fabric_chaincodes(slug); 