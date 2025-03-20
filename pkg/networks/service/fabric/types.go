package fabric

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
