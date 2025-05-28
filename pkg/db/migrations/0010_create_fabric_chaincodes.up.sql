CREATE TABLE IF NOT EXISTS fabric_chaincodes (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  network_id INTEGER NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX idx_fabric_chaincodes_network_id ON fabric_chaincodes(network_id);