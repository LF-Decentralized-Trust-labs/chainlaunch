package fabric

import "time"

// NodeType represents the type of a Fabric node
type NodeType string

const (
	NodeTypePeer    NodeType = "peer"
	NodeTypeOrderer NodeType = "orderer"
)

// NodeStatus represents the status of a Fabric node
type NodeStatus string

const (
	NodeStatusCreating NodeStatus = "creating"
	NodeStatusRunning  NodeStatus = "running"
	NodeStatusStopped  NodeStatus = "stopped"
	NodeStatusError    NodeStatus = "error"
)

// FabricNode represents a node in a Fabric network
type FabricNode struct {
	ID              int64            `json:"id"`
	Type            NodeType         `json:"type"`
	Status          NodeStatus       `json:"status"`
	OrganizationID  int64            `json:"organizationId"`
	EndpointURL     string           `json:"endpointUrl"`
	ConsenterConfig *ConsenterConfig `json:"consenterConfig,omitempty"`
}

// ConsenterConfig represents the configuration for a Fabric orderer node
type ConsenterConfig struct {
	Host      string `json:"host"`
	Port      int    `json:"port"`
	TLSCA     string `json:"tlsCA"`
	ServerTLS string `json:"serverTLS"`
	ClientTLS string `json:"clientTLS"`
}

// MSPConfig represents the configuration for a Fabric MSP
type MSPConfig struct {
	MSPID       string   `json:"mspId"`
	CACerts     []string `json:"caCerts"`
	Admins      []string `json:"admins"`
	TLSRootCert string   `json:"tlsRootCert"`
}

// ChannelConfig represents the configuration for a Fabric channel
type ChannelConfig struct {
	Name          string                 `json:"name"`
	Organizations []string               `json:"organizations"`
	Capabilities  map[string][]string    `json:"capabilities"`
	Policies      map[string]interface{} `json:"policies"`
}

// NetworkConfig represents the configuration for a Fabric network
type NetworkConfig struct {
	ChannelName      string                 `json:"channelName"`
	ConsortiumName   string                 `json:"consortiumName"`
	Organizations    []string               `json:"organizations"`
	OrdererEndpoints []string               `json:"ordererEndpoints"`
	Capabilities     map[string][]string    `json:"capabilities"`
	Policies         map[string]interface{} `json:"policies"`
}

// Block represents a block in the Fabric blockchain
type Block struct {
	Number       uint64    `json:"number"`
	Hash         string    `json:"hash"`
	PreviousHash string    `json:"previousHash"`
	DataHash     string    `json:"dataHash"`
	Timestamp    time.Time `json:"timestamp"`
	TxCount      int       `json:"txCount"`
	Data         []byte    `json:"data"`
}

// Transaction represents a transaction in a Fabric block
type Transaction struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	Creator   string    `json:"creator"`
	Status    string    `json:"status"`
	BlockNum  uint64    `json:"blockNum"`
}

// BlockInfo represents information about the blockchain
type BlockInfo struct {
	Height            uint64 `json:"height"`
	CurrentBlockHash  string `json:"currentBlockHash"`
	PreviousBlockHash string `json:"previousBlockHash"`
}

// DeployerOptions contains options for the Fabric deployer
type DeployerOptions struct {
	NetworkID       int64  `json:"networkId"`
	ChannelID       string `json:"channelId"`
	ConsortiumName  string `json:"consortiumName"`
	OrdererEndpoint string `json:"ordererEndpoint"`
	PeerEndpoint    string `json:"peerEndpoint"`
}

// NetworkQueryOptions contains options for querying network data
type NetworkQueryOptions struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

// BlockQueryOptions contains options for querying blocks
type BlockQueryOptions struct {
	StartBlock uint64 `json:"startBlock"`
	EndBlock   uint64 `json:"endBlock"`
	Limit      int32  `json:"limit"`
	Offset     int32  `json:"offset"`
}

// PaginatedBlocks represents a paginated list of blocks
type PaginatedBlocks struct {
	Items      []Block `json:"items"`
	TotalCount int64   `json:"totalCount"`
}

// PaginatedTransactions represents a paginated list of transactions
type PaginatedTransactions struct {
	Items      []Transaction `json:"items"`
	TotalCount int64         `json:"totalCount"`
}
