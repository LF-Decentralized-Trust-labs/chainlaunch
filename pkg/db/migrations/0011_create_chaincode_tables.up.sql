-- 0011_create_chaincode_tables.up.sql
-- Migration: Create tables for fabric_chaincodes, fabric_chaincode_definitions, and fabric_chaincode_definition_peer_status


CREATE TABLE IF NOT EXISTS fabric_chaincode_definitions (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  chaincode_id INTEGER NOT NULL,
  version TEXT NOT NULL,
  sequence INTEGER NOT NULL,
  docker_image TEXT NOT NULL,
  endorsement_policy TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (chaincode_id) REFERENCES fabric_chaincodes(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS fabric_chaincode_definition_peer_status (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  definition_id INTEGER NOT NULL,
  peer_id INTEGER NOT NULL,
  status TEXT NOT NULL, -- e.g. 'installed', 'approved', 'committed'
  last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(definition_id, peer_id),
  FOREIGN KEY (definition_id) REFERENCES fabric_chaincode_definitions(id) ON DELETE CASCADE
); 