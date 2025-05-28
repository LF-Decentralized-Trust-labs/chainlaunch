-- 0011_create_chaincode_tables.down.sql
-- Migration: Drop tables for fabric_chaincodes, fabric_chaincode_definitions, and fabric_chaincode_definition_peer_status

DROP TABLE IF EXISTS fabric_chaincode_definition_peer_status;
DROP TABLE IF EXISTS fabric_chaincode_definitions;
DROP TABLE IF EXISTS fabric_chaincodes; 