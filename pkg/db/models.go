// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0

package db

import (
	"database/sql"
	"time"
)

type Backup struct {
	ID               int64          `json:"id"`
	ScheduleID       sql.NullInt64  `json:"schedule_id"`
	TargetID         int64          `json:"target_id"`
	Status           string         `json:"status"`
	SizeBytes        sql.NullInt64  `json:"size_bytes"`
	StartedAt        time.Time      `json:"started_at"`
	CompletedAt      sql.NullTime   `json:"completed_at"`
	ErrorMessage     sql.NullString `json:"error_message"`
	CreatedAt        time.Time      `json:"created_at"`
	NotificationSent int64          `json:"notification_sent"`
}

type BackupSchedule struct {
	ID             int64          `json:"id"`
	Name           string         `json:"name"`
	Description    sql.NullString `json:"description"`
	CronExpression string         `json:"cron_expression"`
	TargetID       int64          `json:"target_id"`
	RetentionDays  int64          `json:"retention_days"`
	Enabled        bool           `json:"enabled"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      sql.NullTime   `json:"updated_at"`
	LastRunAt      sql.NullTime   `json:"last_run_at"`
	NextRunAt      sql.NullTime   `json:"next_run_at"`
}

type BackupTarget struct {
	ID             int64          `json:"id"`
	Name           string         `json:"name"`
	BucketName     sql.NullString `json:"bucket_name"`
	Region         sql.NullString `json:"region"`
	Endpoint       sql.NullString `json:"endpoint"`
	BucketPath     sql.NullString `json:"bucket_path"`
	AccessKeyID    sql.NullString `json:"access_key_id"`
	SecretKey      sql.NullString `json:"secret_key"`
	S3PathStyle    sql.NullBool   `json:"s3_path_style"`
	ResticPassword sql.NullString `json:"restic_password"`
	Type           string         `json:"type"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      sql.NullTime   `json:"updated_at"`
}

type BlockchainPlatform struct {
	Name string `json:"name"`
}

type FabricOrganization struct {
	ID              int64          `json:"id"`
	MspID           string         `json:"msp_id"`
	Description     sql.NullString `json:"description"`
	Config          sql.NullString `json:"config"`
	CaConfig        sql.NullString `json:"ca_config"`
	SignKeyID       sql.NullInt64  `json:"sign_key_id"`
	TlsRootKeyID    sql.NullInt64  `json:"tls_root_key_id"`
	AdminTlsKeyID   sql.NullInt64  `json:"admin_tls_key_id"`
	AdminSignKeyID  sql.NullInt64  `json:"admin_sign_key_id"`
	ClientSignKeyID sql.NullInt64  `json:"client_sign_key_id"`
	ProviderID      sql.NullInt64  `json:"provider_id"`
	CreatedAt       time.Time      `json:"created_at"`
	CreatedBy       sql.NullInt64  `json:"created_by"`
	UpdatedAt       sql.NullTime   `json:"updated_at"`
}

type Key struct {
	ID                int64          `json:"id"`
	Name              string         `json:"name"`
	Description       sql.NullString `json:"description"`
	Algorithm         string         `json:"algorithm"`
	KeySize           sql.NullInt64  `json:"key_size"`
	Curve             sql.NullString `json:"curve"`
	Format            string         `json:"format"`
	PublicKey         string         `json:"public_key"`
	PrivateKey        string         `json:"private_key"`
	Certificate       sql.NullString `json:"certificate"`
	Status            string         `json:"status"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	ExpiresAt         sql.NullTime   `json:"expires_at"`
	LastRotatedAt     sql.NullTime   `json:"last_rotated_at"`
	SigningKeyID      sql.NullInt64  `json:"signing_key_id"`
	Sha256Fingerprint string         `json:"sha256_fingerprint"`
	Sha1Fingerprint   string         `json:"sha1_fingerprint"`
	ProviderID        int64          `json:"provider_id"`
	UserID            int64          `json:"user_id"`
	IsCa              int64          `json:"is_ca"`
	EthereumAddress   sql.NullString `json:"ethereum_address"`
}

type KeyProvider struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	IsDefault int64     `json:"is_default"`
	Config    string    `json:"config"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type KeyProviderType struct {
	Name string `json:"name"`
}

type Network struct {
	ID                    int64          `json:"id"`
	Name                  string         `json:"name"`
	NetworkID             sql.NullString `json:"network_id"`
	Platform              string         `json:"platform"`
	Status                string         `json:"status"`
	Description           sql.NullString `json:"description"`
	Config                sql.NullString `json:"config"`
	DeploymentConfig      sql.NullString `json:"deployment_config"`
	ExposedPorts          sql.NullString `json:"exposed_ports"`
	Domain                sql.NullString `json:"domain"`
	CreatedAt             time.Time      `json:"created_at"`
	CreatedBy             sql.NullInt64  `json:"created_by"`
	UpdatedAt             sql.NullTime   `json:"updated_at"`
	GenesisBlockB64       sql.NullString `json:"genesis_block_b64"`
	CurrentConfigBlockB64 sql.NullString `json:"current_config_block_b64"`
}

type NetworkNode struct {
	ID        int64          `json:"id"`
	NetworkID int64          `json:"network_id"`
	NodeID    int64          `json:"node_id"`
	Role      string         `json:"role"`
	Status    string         `json:"status"`
	Config    sql.NullString `json:"config"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

type Node struct {
	ID                   int64          `json:"id"`
	Name                 string         `json:"name"`
	Slug                 string         `json:"slug"`
	Platform             string         `json:"platform"`
	Status               string         `json:"status"`
	Description          sql.NullString `json:"description"`
	NetworkID            sql.NullInt64  `json:"network_id"`
	Config               sql.NullString `json:"config"`
	Resources            sql.NullString `json:"resources"`
	Endpoint             sql.NullString `json:"endpoint"`
	PublicEndpoint       sql.NullString `json:"public_endpoint"`
	P2pAddress           sql.NullString `json:"p2p_address"`
	CreatedAt            time.Time      `json:"created_at"`
	CreatedBy            sql.NullInt64  `json:"created_by"`
	UpdatedAt            sql.NullTime   `json:"updated_at"`
	FabricOrganizationID sql.NullInt64  `json:"fabric_organization_id"`
	NodeType             sql.NullString `json:"node_type"`
	NodeConfig           sql.NullString `json:"node_config"`
	DeploymentConfig     sql.NullString `json:"deployment_config"`
}

type NodeEvent struct {
	ID          int64          `json:"id"`
	NodeID      int64          `json:"node_id"`
	EventType   string         `json:"event_type"`
	Description string         `json:"description"`
	Data        sql.NullString `json:"data"`
	Status      string         `json:"status"`
	CreatedAt   time.Time      `json:"created_at"`
}

type NodeKey struct {
	ID        int64     `json:"id"`
	NodeID    int64     `json:"node_id"`
	KeyID     int64     `json:"key_id"`
	KeyType   string    `json:"key_type"`
	CreatedAt time.Time `json:"created_at"`
}

type NodeKeyType struct {
	Name string `json:"name"`
}

type NodeStatus struct {
	Name string `json:"name"`
}

type NodeType struct {
	Name string `json:"name"`
}

type NotificationProvider struct {
	ID                      int64          `json:"id"`
	Name                    string         `json:"name"`
	Type                    string         `json:"type"`
	Config                  string         `json:"config"`
	IsDefault               bool           `json:"is_default"`
	IsEnabled               bool           `json:"is_enabled"`
	CreatedAt               time.Time      `json:"created_at"`
	UpdatedAt               time.Time      `json:"updated_at"`
	NotifyNodeDowntime      bool           `json:"notify_node_downtime"`
	NotifyBackupSuccess     bool           `json:"notify_backup_success"`
	NotifyBackupFailure     bool           `json:"notify_backup_failure"`
	NotifyS3ConnectionIssue bool           `json:"notify_s3_connection_issue"`
	LastTestAt              sql.NullTime   `json:"last_test_at"`
	LastTestStatus          sql.NullString `json:"last_test_status"`
	LastTestMessage         sql.NullString `json:"last_test_message"`
}

type Session struct {
	ID             int64          `json:"id"`
	SessionID      string         `json:"session_id"`
	UserID         int64          `json:"user_id"`
	Token          string         `json:"token"`
	IpAddress      sql.NullString `json:"ip_address"`
	UserAgent      sql.NullString `json:"user_agent"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	ExpiresAt      time.Time      `json:"expires_at"`
	LastActivityAt time.Time      `json:"last_activity_at"`
}

type User struct {
	ID          int64          `json:"id"`
	Username    string         `json:"username"`
	Password    string         `json:"password"`
	Name        sql.NullString `json:"name"`
	Email       sql.NullString `json:"email"`
	Role        sql.NullString `json:"role"`
	Provider    sql.NullString `json:"provider"`
	ProviderID  sql.NullString `json:"provider_id"`
	AvatarUrl   sql.NullString `json:"avatar_url"`
	CreatedAt   time.Time      `json:"created_at"`
	LastLoginAt sql.NullTime   `json:"last_login_at"`
	UpdatedAt   sql.NullTime   `json:"updated_at"`
}
