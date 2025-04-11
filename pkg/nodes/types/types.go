package types

import "encoding/json"

type DeploymentMode string

const (
	DeploymentModeService DeploymentMode = "SERVICE"
	DeploymentModeDocker  DeploymentMode = "DOCKER"
)

// BlockchainPlatform represents the type of blockchain platform
type BlockchainPlatform string

const (
	PlatformFabric BlockchainPlatform = "FABRIC"
	PlatformBesu   BlockchainPlatform = "BESU"
)

// NodeType represents the type of node
type NodeType string

const (
	// Fabric node types
	NodeTypeFabricPeer    NodeType = "FABRIC_PEER"
	NodeTypeFabricOrderer NodeType = "FABRIC_ORDERER"

	// Besu node types
	NodeTypeBesuFullnode NodeType = "BESU_FULLNODE"
)

// NodeStatus represents the status of a node
type NodeStatus string

const (
	NodeStatusPending  NodeStatus = "PENDING"
	NodeStatusRunning  NodeStatus = "RUNNING"
	NodeStatusStopped  NodeStatus = "STOPPED"
	NodeStatusStopping NodeStatus = "STOPPING"
	NodeStatusStarting NodeStatus = "STARTING"
	NodeStatusUpdating NodeStatus = "UPDATING"
	NodeStatusError    NodeStatus = "ERROR"
)

// StoredConfig represents the stored configuration with type information
type StoredConfig struct {
	Type   string          `json:"type"`
	Config json.RawMessage `json:"config"`
}

type AddressOverride struct {
	From      string `json:"from"`
	To        string `json:"to"`
	TLSCACert string `json:"tlsCACert"`
}
