-- Drop session tables
DROP INDEX IF EXISTS idx_sessions_expires_at;
DROP INDEX IF EXISTS idx_sessions_token;
DROP INDEX IF EXISTS idx_sessions_user_id;
DROP INDEX IF EXISTS idx_sessions_session_id;
DROP TABLE IF EXISTS sessions;

-- Drop notification tables
DROP INDEX IF EXISTS idx_notification_providers_name;
DROP INDEX IF EXISTS idx_notification_providers_type;
DROP TABLE IF EXISTS notification_providers;

-- Drop backup tables
DROP INDEX IF EXISTS idx_backups_status;
DROP INDEX IF EXISTS idx_backups_target_id;
DROP INDEX IF EXISTS idx_backups_schedule_id;
DROP INDEX IF EXISTS idx_backup_schedules_target_id;
DROP TABLE IF EXISTS backups;
DROP TABLE IF EXISTS backup_schedules;
DROP TABLE IF EXISTS backup_targets;

-- Drop node event tables
DROP INDEX IF EXISTS idx_node_events_event_type;
DROP INDEX IF EXISTS idx_node_events_created_at;
DROP INDEX IF EXISTS idx_node_events_node_id;
DROP TABLE IF EXISTS node_events;

-- Drop node and network tables
DROP TABLE IF EXISTS node_keys;
DROP INDEX IF EXISTS idx_network_nodes_network_node;
DROP INDEX IF EXISTS idx_network_nodes_node_id;
DROP INDEX IF EXISTS idx_network_nodes_network_id;
DROP TABLE IF EXISTS network_nodes;
DROP INDEX IF EXISTS idx_nodes_slug;
DROP TABLE IF EXISTS nodes;
DROP TABLE IF EXISTS fabric_organizations;
DROP TABLE IF EXISTS networks;

-- Drop user tables
DROP INDEX IF EXISTS idx_users_username;
DROP TABLE IF EXISTS users;

-- Drop key tables
DROP TABLE IF EXISTS keys;
DROP TABLE IF EXISTS key_providers;

-- Drop enum-like tables
DROP TABLE IF EXISTS node_key_types;
DROP TABLE IF EXISTS node_statuses;
DROP TABLE IF EXISTS node_types;
DROP TABLE IF EXISTS key_provider_types;
DROP TABLE IF EXISTS blockchain_platforms;
