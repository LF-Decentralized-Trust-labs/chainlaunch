package service

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/chainlaunch/chainlaunch/pkg/db"
	orgservicefabric "github.com/chainlaunch/chainlaunch/pkg/fabric/service"
	keymanagement "github.com/chainlaunch/chainlaunch/pkg/keymanagement/service"
	"github.com/chainlaunch/chainlaunch/pkg/logger"
	"github.com/chainlaunch/chainlaunch/pkg/networks/service/besu"
	"github.com/chainlaunch/chainlaunch/pkg/networks/service/fabric"
	"github.com/chainlaunch/chainlaunch/pkg/networks/service/types"
	nodeservice "github.com/chainlaunch/chainlaunch/pkg/nodes/service"
	nodetypes "github.com/chainlaunch/chainlaunch/pkg/nodes/types"
	nodeutils "github.com/chainlaunch/chainlaunch/pkg/nodes/utils"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// BlockchainType represents the type of blockchain network
type BlockchainType string

const (
	BlockchainTypeFabric BlockchainType = "fabric"
	BlockchainTypeBesu   BlockchainType = "besu"
	// Add other blockchain types as needed
)

// NetworkStatus represents the status of a network
type NetworkStatus string

const (
	NetworkStatusCreating            NetworkStatus = "creating"
	NetworkStatusGenesisBlockCreated NetworkStatus = "genesis_block_created"
	NetworkStatusRunning             NetworkStatus = "running"
	NetworkStatusStopped             NetworkStatus = "stopped"
	NetworkStatusError               NetworkStatus = "error"
)

// Network represents a blockchain network
type Network struct {
	ID                 int64           `json:"id"`
	Name               string          `json:"name"`
	Platform           string          `json:"platform"`
	Status             NetworkStatus   `json:"status"`
	Description        string          `json:"description"`
	Config             json.RawMessage `json:"config,omitempty"`
	DeploymentConfig   json.RawMessage `json:"deploymentConfig,omitempty"`
	ExposedPorts       json.RawMessage `json:"exposedPorts,omitempty"`
	GenesisBlock       string          `json:"genesisBlock,omitempty"`
	CurrentConfigBlock string          `json:"currentConfigBlock,omitempty"`
	Domain             string          `json:"domain,omitempty"`
	CreatedAt          time.Time       `json:"createdAt"`
	CreatedBy          *int64          `json:"createdBy,omitempty"`
	UpdatedAt          *time.Time      `json:"updatedAt,omitempty"`
}

// ListNetworksParams represents the parameters for listing networks
type ListNetworksParams struct {
	Limit    int32
	Offset   int32
	Platform BlockchainType
}

// ListNetworksResult represents the result of listing networks
type ListNetworksResult struct {
	Networks []Network
	Total    int64
}

// ConfigUpdateOperationRequest represents a configuration update operation
type ConfigUpdateOperationRequest struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// Proposal represents a configuration update proposal
type Proposal struct {
	ID                string                         `json:"id"`
	NetworkID         int64                          `json:"network_id"`
	ChannelName       string                         `json:"channel_name"`
	Status            string                         `json:"status"`
	CreatedAt         time.Time                      `json:"created_at"`
	CreatedBy         string                         `json:"created_by"`
	Operations        []ConfigUpdateOperationRequest `json:"operations"`
	PreviewJSON       string                         `json:"preview_json,omitempty"`
	ConfigUpdateBytes []byte                         `json:"config_update_bytes,omitempty"`
}

// ProposalSignature represents a signature on a proposal
type ProposalSignature struct {
	ID       int64     `json:"id"`
	MSPID    string    `json:"msp_id"`
	SignedBy string    `json:"signed_by"`
	SignedAt time.Time `json:"signed_at"`
}

// FabricNetworkService handles network operations
type NetworkService struct {
	db              *db.Queries
	deployerFactory *DeployerFactory
	nodes           *nodeservice.NodeService
	keyMgmt         *keymanagement.KeyManagementService
	logger          *logger.Logger
	orgService      *orgservicefabric.OrganizationService
}

// NewNetworkService creates a new NetworkService
func NewNetworkService(db *db.Queries, nodes *nodeservice.NodeService, keyMgmt *keymanagement.KeyManagementService, logger *logger.Logger, orgService *orgservicefabric.OrganizationService) *NetworkService {
	return &NetworkService{
		db:              db,
		deployerFactory: NewDeployerFactory(db, nodes, keyMgmt, orgService),
		nodes:           nodes,
		keyMgmt:         keyMgmt,
		logger:          logger,
		orgService:      orgService,
	}
}

// GetNetworkByName retrieves a network by its name
func (s *NetworkService) GetNetworkByName(ctx context.Context, name string) (*Network, error) {
	network, err := s.db.GetNetworkByName(ctx, name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("network with name %s not found", name)
		}
		return nil, fmt.Errorf("failed to get network by name: %w", err)
	}

	return s.mapDBNetworkToServiceNetwork(network), nil
}

// CreateNetwork creates a new blockchain network
func (s *NetworkService) CreateNetwork(ctx context.Context, name, description string, configData []byte) (*Network, error) {
	// Parse base config to determine type
	var baseConfig types.BaseNetworkConfig
	if err := json.Unmarshal(configData, &baseConfig); err != nil {
		return nil, fmt.Errorf("failed to parse base config: %w", err)
	}

	var config types.NetworkConfig
	switch baseConfig.Type {
	case types.NetworkTypeFabric:
		var fabricConfig types.FabricNetworkConfig
		if err := json.Unmarshal(configData, &fabricConfig); err != nil {
			return nil, fmt.Errorf("failed to unmarshal Fabric config: %w", err)
		}
		config = &fabricConfig

	case types.NetworkTypeBesu:
		var besuConfig types.BesuNetworkConfig
		if err := json.Unmarshal(configData, &besuConfig); err != nil {
			return nil, fmt.Errorf("failed to unmarshal Besu config: %w", err)
		}
		config = &besuConfig

	default:
		return nil, fmt.Errorf("unsupported network type: %s", baseConfig.Type)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid network configuration: %w", err)
	}

	// Validate external nodes exist and are of correct type

	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	// Generate a random network ID
	networkID := fmt.Sprintf("net_%s_%s", name, uuid.New().String())
	// Create network in database
	network, err := s.db.CreateNetwork(ctx, &db.CreateNetworkParams{
		Name:        name,
		Platform:    string(baseConfig.Type),
		Description: sql.NullString{String: description, Valid: description != ""},
		Config:      sql.NullString{String: string(configJSON), Valid: true},
		Status:      string(NetworkStatusCreating),
		NetworkID:   sql.NullString{String: networkID, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create network: %w", err)
	}

	deployer, err := s.deployerFactory.GetDeployer(string(baseConfig.Type))
	if err != nil {
		return nil, fmt.Errorf("failed to get deployer: %w", err)
	}

	genesisBlock, err := deployer.CreateGenesisBlock(network.ID, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create genesis block: %w", err)
	}
	genesisBlockB64 := base64.StdEncoding.EncodeToString(genesisBlock)
	network.GenesisBlockB64 = sql.NullString{String: genesisBlockB64, Valid: true}

	// Update network status to running after successful genesis block creation
	if err := s.UpdateNetworkStatus(ctx, network.ID, NetworkStatusGenesisBlockCreated); err != nil {
		return nil, fmt.Errorf("failed to update network status: %w", err)
	}

	return s.mapDBNetworkToServiceNetwork(network), nil
}

// JoinPeerToNetwork joins a peer to a Fabric network
func (s *NetworkService) JoinPeerToNetwork(networkID, peerID int64) error {
	network, err := s.db.GetNetwork(context.Background(), networkID)
	if err != nil {
		return fmt.Errorf("failed to get network: %w", err)
	}
	deployer, err := s.deployerFactory.GetDeployer(network.Platform)
	if err != nil {
		return fmt.Errorf("failed to get deployer: %w", err)
	}
	if !network.GenesisBlockB64.Valid {
		return fmt.Errorf("genesis block is not set for network %d", networkID)
	}
	genesisBlockBytes, err := base64.StdEncoding.DecodeString(network.GenesisBlockB64.String)
	if err != nil {
		return fmt.Errorf("failed to decode genesis block: %w", err)
	}
	err = deployer.JoinNode(network.ID, genesisBlockBytes, peerID)
	if err != nil {
		return fmt.Errorf("failed to join node: %w", err)
	}
	logrus.Infof("joined peer %d to network %d", peerID, networkID)

	return nil
}

// JoinOrdererToNetwork joins an orderer to a Fabric network
func (s *NetworkService) JoinOrdererToNetwork(networkID, ordererID int64) error {
	network, err := s.db.GetNetwork(context.Background(), networkID)
	if err != nil {
		return fmt.Errorf("failed to get network: %w", err)
	}
	deployer, err := s.deployerFactory.GetDeployer(network.Platform)
	if err != nil {
		return fmt.Errorf("failed to get deployer: %w", err)
	}
	if !network.GenesisBlockB64.Valid {
		return fmt.Errorf("genesis block is not set for network %d", networkID)
	}
	genesisBlockBytes, err := base64.StdEncoding.DecodeString(network.GenesisBlockB64.String)
	if err != nil {
		return fmt.Errorf("failed to decode genesis block: %w", err)
	}
	err = deployer.JoinNode(network.ID, genesisBlockBytes, ordererID)
	if err != nil {
		return fmt.Errorf("failed to join node: %w", err)
	}
	logrus.Infof("joined orderer %d to network %d", ordererID, networkID)

	return nil
}

// RemovePeerFromNetwork removes a peer from a Fabric network
func (s *NetworkService) RemovePeerFromNetwork(networkID, peerID int64) error {
	// Get the appropriate deployer
	network, err := s.db.GetNetwork(context.Background(), networkID)
	if err != nil {
		return fmt.Errorf("failed to get network: %w", err)
	}

	deployer, err := s.deployerFactory.GetDeployer(network.Platform)
	if err != nil {
		return fmt.Errorf("failed to get deployer: %w", err)
	}

	fabricDeployer, ok := deployer.(*fabric.FabricDeployer)
	if !ok {
		return fmt.Errorf("network %d is not a Fabric network", networkID)
	}

	if err := fabricDeployer.RemoveNode(networkID, peerID); err != nil {
		return fmt.Errorf("failed to remove peer: %w", err)
	}

	logrus.Infof("removed peer %d from network %d", peerID, networkID)
	return nil
}

// RemoveOrdererFromNetwork removes an orderer from a Fabric network
func (s *NetworkService) RemoveOrdererFromNetwork(networkID, ordererID int64) error {
	// Get the appropriate deployer
	network, err := s.db.GetNetwork(context.Background(), networkID)
	if err != nil {
		return fmt.Errorf("failed to get network: %w", err)
	}

	deployer, err := s.deployerFactory.GetDeployer(network.Platform)
	if err != nil {
		return fmt.Errorf("failed to get deployer: %w", err)
	}

	fabricDeployer, ok := deployer.(*fabric.FabricDeployer)
	if !ok {
		return fmt.Errorf("network %d is not a Fabric network", networkID)
	}

	if err := fabricDeployer.RemoveNode(networkID, ordererID); err != nil {
		return fmt.Errorf("failed to remove orderer: %w", err)
	}

	logrus.Infof("removed orderer %d from network %d", ordererID, networkID)
	return nil
}

// GetCurrentChannelConfig retrieves the current channel configuration for a network
func (s *NetworkService) GetCurrentChannelConfig(networkID int64) (map[string]interface{}, error) {
	// Get the appropriate deployer
	network, err := s.db.GetNetwork(context.Background(), networkID)
	if err != nil {
		return nil, fmt.Errorf("failed to get network: %w", err)
	}

	deployer, err := s.deployerFactory.GetDeployer(network.Platform)
	if err != nil {
		return nil, fmt.Errorf("failed to get deployer: %w", err)
	}

	fabricDeployer, ok := deployer.(*fabric.FabricDeployer)
	if !ok {
		return nil, fmt.Errorf("network %d is not a Fabric network", networkID)
	}

	return fabricDeployer.GetCurrentChannelConfigAsMap(networkID)
}

// GetChannelConfig retrieves the channel configuration for a network
func (s *NetworkService) GetChannelConfig(networkID int64) (map[string]interface{}, error) {
	// Get the appropriate deployer
	network, err := s.db.GetNetwork(context.Background(), networkID)
	if err != nil {
		return nil, fmt.Errorf("failed to get network: %w", err)
	}

	deployer, err := s.deployerFactory.GetDeployer(network.Platform)
	if err != nil {
		return nil, fmt.Errorf("failed to get deployer: %w", err)
	}

	fabricDeployer, ok := deployer.(*fabric.FabricDeployer)
	if !ok {
		return nil, fmt.Errorf("network %d is not a Fabric network", networkID)
	}

	return fabricDeployer.GetChannelConfig(networkID)
}

// ListNetworks retrieves a list of networks with pagination
func (s *NetworkService) ListNetworks(ctx context.Context, params ListNetworksParams) (*ListNetworksResult, error) {
	networks, err := s.db.ListNetworks(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list networks: %w", err)
	}

	result := &ListNetworksResult{
		Networks: make([]Network, len(networks)),
		Total:    int64(len(networks)), // TODO: Implement proper counting
	}

	for i, n := range networks {
		result.Networks[i] = *s.mapDBNetworkToServiceNetwork(n)
	}

	return result, nil
}

// GetNetwork retrieves a network by ID
func (s *NetworkService) GetNetwork(ctx context.Context, networkID int64) (*Network, error) {
	network, err := s.db.GetNetwork(ctx, networkID)
	if err != nil {
		return nil, fmt.Errorf("failed to get network: %w", err)
	}

	return s.mapDBNetworkToServiceNetwork(network), nil
}

// DeleteNetwork deletes a network and all associated resources
func (s *NetworkService) DeleteNetwork(ctx context.Context, networkID int64) error {
	// Get network to determine platform
	// network, err := s.db.GetNetwork(ctx, networkID)
	// if err != nil {
	// 	return fmt.Errorf("failed to get network: %w", err)
	// }

	// Get the appropriate deployer
	// deployer, err := s.deployerFactory.GetDeployer(network.Platform)
	// if err != nil {
	// 	return fmt.Errorf("failed to get deployer: %w", err)
	// }

	// Delete network resources using deployer
	// if err := deployer.DeleteNetwork(networkID); err != nil {
	// 	return fmt.Errorf("failed to delete network resources: %w", err)
	// }

	// Delete network record
	if err := s.db.DeleteNetwork(ctx, networkID); err != nil {
		return fmt.Errorf("failed to delete network record: %w", err)
	}

	return nil
}

// Helper function to map db.Network to service.Network
func (s *NetworkService) mapDBNetworkToServiceNetwork(n *db.Network) *Network {
	var config, deploymentConfig, exposedPorts json.RawMessage
	if n.Config.Valid {
		config = json.RawMessage(n.Config.String)
	}
	if n.DeploymentConfig.Valid {
		deploymentConfig = json.RawMessage(n.DeploymentConfig.String)
	}
	if n.ExposedPorts.Valid {
		exposedPorts = json.RawMessage(n.ExposedPorts.String)
	}

	network := &Network{
		ID:               n.ID,
		Name:             n.Name,
		Platform:         n.Platform,
		Status:           NetworkStatus(n.Status),
		Config:           config,
		DeploymentConfig: deploymentConfig,
		ExposedPorts:     exposedPorts,
		CreatedAt:        n.CreatedAt,
		CreatedBy:        nil,
	}

	if n.Description.Valid {
		network.Description = n.Description.String
	}
	if n.Domain.Valid {
		network.Domain = n.Domain.String
	}
	if n.CreatedBy.Valid {
		network.CreatedBy = &n.CreatedBy.Int64
	}
	if n.UpdatedAt.Valid {
		updatedAt := n.UpdatedAt.Time
		network.UpdatedAt = &updatedAt
	}
	if n.GenesisBlockB64.Valid {
		network.GenesisBlock = n.GenesisBlockB64.String
	}
	if n.CurrentConfigBlockB64.Valid {
		network.CurrentConfigBlock = n.CurrentConfigBlockB64.String
	}

	return network
}

// UpdateNetworkStatus updates the status of a network
func (s *NetworkService) UpdateNetworkStatus(ctx context.Context, networkID int64, status NetworkStatus) error {
	err := s.db.UpdateNetworkStatus(ctx, &db.UpdateNetworkStatusParams{
		ID:     networkID,
		Status: string(status),
	})
	if err != nil {
		return fmt.Errorf("failed to update network status: %w", err)
	}
	return nil
}

// GetNetworkNodes retrieves all nodes associated with a network
func (s *NetworkService) GetNetworkNodes(ctx context.Context, networkID int64) ([]NetworkNode, error) {
	// Get network nodes from database
	dbNodes, err := s.db.GetNetworkNodes(ctx, networkID)
	if err != nil {
		return nil, fmt.Errorf("failed to get network nodes: %w", err)
	}

	nodes := make([]NetworkNode, len(dbNodes))
	for i, dbNode := range dbNodes {
		deploymentConfig, err := nodeutils.DeserializeDeploymentConfig(dbNode.DeploymentConfig.String)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize deployment config: %w", err)
		}
		nodeConfig, err := nodeutils.LoadNodeConfig([]byte(dbNode.NodeConfig.String))
		if err != nil {
			return nil, fmt.Errorf("failed to load node config: %w", err)
		}
		node := nodeservice.Node{
			ID:                 dbNode.NodeID,
			Name:               dbNode.Name,
			BlockchainPlatform: nodetypes.BlockchainPlatform(dbNode.Platform),
			NodeType:           nodetypes.NodeType(dbNode.NodeType.String),
			Status:             nodetypes.NodeStatus(dbNode.Status_2),
			Endpoint:           dbNode.Endpoint.String,
			PublicEndpoint:     dbNode.PublicEndpoint.String,
			NodeConfig:         nodeConfig,
			DeploymentConfig:   deploymentConfig,
			CreatedAt:          dbNode.CreatedAt_2,
			UpdatedAt:          dbNode.UpdatedAt_2.Time,
		}
		if node.NodeType == nodetypes.NodeTypeFabricPeer {
			if peerConfig, ok := nodeConfig.(*nodetypes.FabricPeerConfig); ok {
				node.MSPID = peerConfig.MSPID
			}
		} else if node.NodeType == nodetypes.NodeTypeFabricOrderer {
			if ordererConfig, ok := nodeConfig.(*nodetypes.FabricOrdererConfig); ok {
				node.MSPID = ordererConfig.MSPID
			}
		}
		nodes[i] = NetworkNode{
			ID:        dbNode.ID,
			NetworkID: dbNode.NetworkID,
			NodeID:    dbNode.NodeID,
			Status:    dbNode.Status,
			Role:      dbNode.Role,
			CreatedAt: dbNode.CreatedAt,
			UpdatedAt: dbNode.UpdatedAt,
			Node:      node,
		}
	}

	return nodes, nil
}

// NetworkNode represents a node in a network with its full details
type NetworkNode struct {
	ID        int64            `json:"id"`
	NetworkID int64            `json:"networkId"`
	NodeID    int64            `json:"nodeId"`
	Status    string           `json:"status"`
	Role      string           `json:"role"`
	CreatedAt time.Time        `json:"createdAt"`
	UpdatedAt time.Time        `json:"updatedAt"`
	Node      nodeservice.Node `json:"node"`
}

// AddNodeToNetwork adds a node to the network with the specified role
func (s *NetworkService) AddNodeToNetwork(ctx context.Context, networkID, nodeID int64, role string) error {
	// Get the network
	network, err := s.db.GetNetwork(ctx, networkID)
	if err != nil {
		return fmt.Errorf("failed to get network: %w", err)
	}

	// Get the node
	node, err := s.nodes.GetNode(ctx, nodeID)
	if err != nil {
		return fmt.Errorf("failed to get node: %w", err)
	}

	// Validate node type matches role
	switch role {
	case "peer":
		if node.NodeType != nodetypes.NodeTypeFabricPeer {
			return fmt.Errorf("node %d is not a peer", nodeID)
		}
	case "orderer":
		if node.NodeType != nodetypes.NodeTypeFabricOrderer {
			return fmt.Errorf("node %d is not an orderer", nodeID)
		}
	default:
		return fmt.Errorf("invalid role: %s", role)
	}

	// Check if node is already in network
	exists, err := s.db.CheckNetworkNodeExists(ctx, &db.CheckNetworkNodeExistsParams{
		NetworkID: networkID,
		NodeID:    nodeID,
	})
	if err != nil {
		return fmt.Errorf("failed to check if node exists in network: %w", err)
	}
	if exists == 1 {
		return fmt.Errorf("node %d is already in network %d", nodeID, networkID)
	}

	// Create network node entry
	_, err = s.db.CreateNetworkNode(ctx, &db.CreateNetworkNodeParams{
		NetworkID: networkID,
		NodeID:    nodeID,
		Status:    "pending",
		Role:      role,
	})
	if err != nil {
		return fmt.Errorf("failed to create network node: %w", err)
	}

	// Get genesis block
	if !network.GenesisBlockB64.Valid {
		return fmt.Errorf("network %d has no genesis block", networkID)
	}

	return nil
}

// UnjoinPeerFromNetwork removes a peer from a channel but keeps it in the network
func (s *NetworkService) UnjoinPeerFromNetwork(networkID, peerID int64) error {
	network, err := s.db.GetNetwork(context.Background(), networkID)
	if err != nil {
		return fmt.Errorf("failed to get network: %w", err)
	}
	deployer, err := s.deployerFactory.GetDeployer(network.Platform)
	if err != nil {
		return fmt.Errorf("failed to get deployer: %w", err)
	}

	fabricDeployer, ok := deployer.(*fabric.FabricDeployer)
	if !ok {
		return fmt.Errorf("network %d is not a Fabric network", networkID)
	}

	if err := fabricDeployer.UnjoinNode(networkID, peerID); err != nil {
		return fmt.Errorf("failed to unjoin peer: %w", err)
	}

	logrus.Infof("unjoined peer %d from network %d", peerID, networkID)
	return nil
}

// UnjoinOrdererFromNetwork removes an orderer from a channel but keeps it in the network
func (s *NetworkService) UnjoinOrdererFromNetwork(networkID, ordererID int64) error {
	network, err := s.db.GetNetwork(context.Background(), networkID)
	if err != nil {
		return fmt.Errorf("failed to get network: %w", err)
	}
	deployer, err := s.deployerFactory.GetDeployer(network.Platform)
	if err != nil {
		return fmt.Errorf("failed to get deployer: %w", err)
	}

	fabricDeployer, ok := deployer.(*fabric.FabricDeployer)
	if !ok {
		return fmt.Errorf("network %d is not a Fabric network", networkID)
	}

	if err := fabricDeployer.UnjoinNode(networkID, ordererID); err != nil {
		return fmt.Errorf("failed to unjoin orderer: %w", err)
	}

	logrus.Infof("unjoined orderer %d from network %d", ordererID, networkID)
	return nil
}

type AnchorPeer struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

// SetAnchorPeers sets the anchor peers for an organization in a Fabric network
func (s *NetworkService) SetAnchorPeers(ctx context.Context, networkID, organizationID int64, anchorPeers []AnchorPeer) (string, error) {
	// Get network details
	network, err := s.db.GetNetwork(ctx, networkID)
	if err != nil {
		return "", fmt.Errorf("failed to get network: %w", err)
	}

	// Get deployer
	deployer, err := s.deployerFactory.GetDeployer(network.Platform)
	if err != nil {
		return "", fmt.Errorf("failed to get deployer: %w", err)
	}

	fabricDeployer, ok := deployer.(*fabric.FabricDeployer)
	if !ok {
		return "", fmt.Errorf("network %d is not a Fabric network", networkID)
	}

	// Convert anchor peers to deployer format
	deployerAnchorPeers := make([]types.HostPort, len(anchorPeers))
	for i, ap := range anchorPeers {
		deployerAnchorPeers[i] = types.HostPort{
			Host: ap.Host,
			Port: ap.Port,
		}
	}

	// Try to get orderer info from network nodes first
	networkNodes, err := s.GetNetworkNodes(ctx, networkID)
	if err != nil {
		return "", fmt.Errorf("failed to get network nodes: %w", err)
	}

	var ordererAddress, ordererTLSCert string

	// Look for orderer in our registry
	for _, node := range networkNodes {
		if node.Node.NodeType == nodetypes.NodeTypeFabricOrderer {
			ordererConfig, ok := node.Node.DeploymentConfig.(*nodetypes.FabricOrdererDeploymentConfig)
			if !ok {
				continue
			}
			ordererAddress = ordererConfig.ExternalEndpoint
			ordererTLSCert = ordererConfig.TLSCACert
			break
		}
	}

	// If no orderer found in registry, try to get from current config block
	if ordererAddress == "" {
		// Get current config block
		configBlock, err := fabricDeployer.GetCurrentChannelConfig(networkID)
		if err != nil {
			return "", fmt.Errorf("failed to get current config block: %w", err)
		}

		// Extract orderer info from config block
		ordererInfo, err := fabricDeployer.GetOrderersFromConfigBlock(ctx, configBlock)
		if err != nil {
			return "", fmt.Errorf("failed to get orderer info from config: %w", err)
		}
		if len(ordererInfo) == 0 {
			return "", fmt.Errorf("no orderer found in config block")
		}
		ordererAddress = ordererInfo[0].URL
		ordererTLSCert = ordererInfo[0].TLSCert
	}

	if ordererAddress == "" {
		return "", fmt.Errorf("no orderer found in network or config block")
	}

	// Set anchor peers using deployer with the found orderer info
	txID, err := fabricDeployer.SetAnchorPeersWithOrderer(ctx, networkID, organizationID, deployerAnchorPeers, ordererAddress, ordererTLSCert)
	if err != nil {
		return "", err
	}

	logrus.Info("Reloading network block after setting anchor peers, waiting 3 seconds")
	time.Sleep(3 * time.Second)

	// Reload network block
	if err := s.ReloadNetworkBlock(ctx, networkID); err != nil {
		logrus.Errorf("Failed to reload network block after setting anchor peers: %v", err)
	}

	return txID, nil
}

// ReloadNetworkBlock reloads the network block for a given network ID
func (s *NetworkService) ReloadNetworkBlock(ctx context.Context, networkID int64) error {
	// Get the network
	network, err := s.db.GetNetwork(ctx, networkID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("network with id %d not found", networkID)
		}
		return fmt.Errorf("failed to get network: %w", err)
	}

	// Get the deployer for this network type
	deployer, err := s.deployerFactory.GetDeployer(network.Platform)
	if err != nil {
		return fmt.Errorf("failed to get deployer: %w", err)
	}
	fabricDeployer, ok := deployer.(*fabric.FabricDeployer)
	if !ok {
		return fmt.Errorf("network %d is not a Fabric network", networkID)
	}

	// Get the current config block
	configBlock, err := fabricDeployer.FetchCurrentChannelConfig(ctx, networkID)
	if err != nil {
		return fmt.Errorf("failed to get current config block: %w", err)
	}
	configBlockB64 := base64.StdEncoding.EncodeToString(configBlock)

	err = s.db.UpdateNetworkCurrentConfigBlock(ctx, &db.UpdateNetworkCurrentConfigBlockParams{
		ID:                    networkID,
		CurrentConfigBlockB64: sql.NullString{String: configBlockB64, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("failed to update network config block: %w", err)
	}

	return nil
}

// GetNetworkConfig retrieves the network configuration as YAML
func (s *NetworkService) GetNetworkConfig(ctx context.Context, networkID, orgID int64) (string, error) {
	// Get the network
	network, err := s.db.GetNetwork(ctx, networkID)
	if err != nil {
		return "", fmt.Errorf("failed to get network: %w", err)
	}

	// Get the deployer
	deployer, err := s.deployerFactory.GetDeployer(network.Platform)
	if err != nil {
		return "", fmt.Errorf("failed to get deployer: %w", err)
	}

	fabricDeployer, ok := deployer.(*fabric.FabricDeployer)
	if !ok {
		return "", fmt.Errorf("network %d is not a Fabric network", networkID)
	}

	// Generate network config YAML
	configYAML, err := fabricDeployer.GenerateNetworkConfig(ctx, networkID, orgID)
	if err != nil {
		return "", fmt.Errorf("failed to generate network config: %w", err)
	}

	return configYAML, nil
}

// GetGenesisBlock retrieves the genesis block for a network
func (s *NetworkService) GetGenesisBlock(ctx context.Context, networkID int64) ([]byte, error) {
	network, err := s.db.GetNetwork(ctx, networkID)
	if err != nil {
		return nil, fmt.Errorf("failed to get network: %w", err)
	}
	genesisBlockB64 := network.GenesisBlockB64.String
	if genesisBlockB64 == "" {
		return nil, fmt.Errorf("no genesis block found for network")
	}
	genesisBlock, err := base64.StdEncoding.DecodeString(genesisBlockB64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode genesis block: %w", err)
	}

	return genesisBlock, nil
}

func (s *NetworkService) ImportNetwork(ctx context.Context, params ImportNetworkParams) (*ImportNetworkResult, error) {
	switch params.NetworkType {
	case "fabric":
		return s.importFabricNetwork(ctx, params)
	case "besu":
		return s.importBesuNetwork(ctx, params)
	default:
		return nil, fmt.Errorf("unsupported network type: %s", params.NetworkType)
	}
}

// ImportNetworkWithOrgParams contains parameters for importing a network with organization details
type ImportNetworkWithOrgParams struct {
	ChannelID      string
	OrganizationID int64
	OrdererURL     string
	OrdererTLSCert []byte
	Description    string
}

// ImportNetworkWithOrg imports a Fabric network using organization details
func (s *NetworkService) ImportNetworkWithOrg(ctx context.Context, params ImportNetworkWithOrgParams) (*ImportNetworkResult, error) {
	// Get the Fabric deployer
	deployer, err := s.deployerFactory.GetDeployer("fabric")
	if err != nil {
		return nil, fmt.Errorf("failed to get Fabric deployer: %w", err)
	}
	fabricDeployer, ok := deployer.(*fabric.FabricDeployer)
	if !ok {
		return nil, fmt.Errorf("invalid deployer type")
	}

	// Import the network using the Fabric deployer
	networkID, err := fabricDeployer.ImportNetworkWithOrg(ctx, params.ChannelID, params.OrganizationID, params.OrdererURL, params.OrdererTLSCert, params.Description)
	if err != nil {
		return nil, fmt.Errorf("failed to import Fabric network with org: %w", err)
	}

	return &ImportNetworkResult{
		NetworkID: networkID,
		Message:   "Fabric network imported successfully with organization",
	}, nil
}

func (s *NetworkService) importFabricNetwork(ctx context.Context, params ImportNetworkParams) (*ImportNetworkResult, error) {
	// Get the Fabric deployer
	deployer, err := s.deployerFactory.GetDeployer("fabric")
	if err != nil {
		return nil, fmt.Errorf("failed to get Fabric deployer: %w", err)
	}

	fabricDeployer, ok := deployer.(*fabric.FabricDeployer)
	if !ok {
		return nil, fmt.Errorf("invalid deployer type")
	}

	// Import the network using the Fabric deployer
	networkID, err := fabricDeployer.ImportNetwork(ctx, params.GenesisFile, params.Description)
	if err != nil {
		return nil, fmt.Errorf("failed to import Fabric network: %w", err)
	}

	return &ImportNetworkResult{
		NetworkID: networkID,
		Message:   "Fabric network imported successfully",
	}, nil
}

func (s *NetworkService) importBesuNetwork(ctx context.Context, params ImportNetworkParams) (*ImportNetworkResult, error) {
	// Get the Besu deployer
	deployer, err := s.deployerFactory.GetDeployer("besu")
	if err != nil {
		return nil, fmt.Errorf("failed to get Besu deployer: %w", err)
	}

	besuDeployer, ok := deployer.(*besu.BesuDeployer)
	if !ok {
		return nil, fmt.Errorf("invalid deployer type")
	}

	// Import the network using the Besu deployer
	networkID, err := besuDeployer.ImportNetwork(ctx, params.GenesisFile, params.Name, params.Description)
	if err != nil {
		return nil, fmt.Errorf("failed to import Besu network: %w", err)
	}

	return &ImportNetworkResult{
		NetworkID: networkID,
		Message:   "Besu network imported successfully",
	}, nil
}

// UpdateFabricNetwork prepares a config update proposal for a Fabric network
func (s *NetworkService) UpdateFabricNetwork(ctx context.Context, networkID int64, operations []fabric.ConfigUpdateOperation) (*fabric.ConfigUpdateProposal, error) {
	// Get deployer for the network
	deployer, err := s.deployerFactory.GetDeployer(string(BlockchainTypeFabric))
	if err != nil {
		return nil, fmt.Errorf("failed to get deployer: %w", err)
	}

	// Assert that it's a Fabric deployer
	fabricDeployer, ok := deployer.(*fabric.FabricDeployer)
	if !ok {
		return nil, fmt.Errorf("network %d is not a Fabric network", networkID)
	}

	// Prepare the config update
	proposal, err := fabricDeployer.PrepareConfigUpdate(ctx, networkID, operations)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare config update: %w", err)
	}

	// Get organizations managed by us that can sign the config update
	orgs, err := s.db.ListFabricOrganizations(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get network organizations: %w", err)
	}
	var signingOrgIDs []string
	for _, org := range orgs {
		signingOrgIDs = append(signingOrgIDs, org.MspID)
	}

	ordererAddress, ordererTLSCert, err := s.getOrdererAddressAndCertForNetwork(ctx, networkID, fabricDeployer)
	if err != nil {
		return nil, fmt.Errorf("failed to get orderer address and TLS certificate: %w", err)
	}

	res, err := fabricDeployer.UpdateChannelConfig(ctx, networkID, proposal.ConfigUpdateEnvelope, signingOrgIDs, ordererAddress, ordererTLSCert)
	if err != nil {
		return nil, fmt.Errorf("failed to update channel config: %w", err)
	}
	s.logger.Info("Channel config updated", "txID", res)
	return proposal, nil
}

func (s *NetworkService) getOrdererAddressAndCertForNetwork(ctx context.Context, networkID int64, fabricDeployer *fabric.FabricDeployer) (string, string, error) {

	// Try to get orderer info from network nodes first
	networkNodes, err := s.GetNetworkNodes(ctx, networkID)
	if err != nil {
		return "", "", fmt.Errorf("failed to get network nodes: %w", err)
	}

	var ordererAddress, ordererTLSCert string

	// Look for orderer in our registry
	for _, node := range networkNodes {
		if node.Node.NodeType == nodetypes.NodeTypeFabricOrderer {
			ordererConfig, ok := node.Node.DeploymentConfig.(*nodetypes.FabricOrdererDeploymentConfig)
			if !ok {
				continue
			}
			ordererAddress = ordererConfig.ExternalEndpoint
			ordererTLSCert = ordererConfig.TLSCACert
			break
		}
	}

	// If no orderer found in registry, try to get from current config block
	if ordererAddress == "" {
		// Get current config block
		configBlock, err := fabricDeployer.GetCurrentChannelConfig(networkID)
		if err != nil {
			return "", "", fmt.Errorf("failed to get current config block: %w", err)
		}

		// Extract orderer info from config block
		ordererInfo, err := fabricDeployer.GetOrderersFromConfigBlock(ctx, configBlock)
		if err != nil {
			return "", "", fmt.Errorf("failed to get orderer info from config: %w", err)
		}
		if len(ordererInfo) == 0 {
			return "", "", fmt.Errorf("no orderer found in config block")
		}
		ordererAddress = ordererInfo[0].URL
		ordererTLSCert = ordererInfo[0].TLSCert
	}

	if ordererAddress == "" {
		return "", "", fmt.Errorf("no orderer found in network or config block")
	}

	return ordererAddress, ordererTLSCert, nil
}

// GetBlocks retrieves a paginated list of blocks from the network
func (s *NetworkService) GetBlocks(ctx context.Context, networkID int64, limit, offset int32, reverse bool) ([]Block, int64, error) {
	// Get the fabric deployer for this network
	fabricDeployer, err := s.getFabricDeployerForNetwork(ctx, networkID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get fabric deployer: %w", err)
	}

	// Use the fabric deployer to get blocks
	fabricBlocks, total, err := fabricDeployer.GetBlocks(ctx, networkID, limit, offset, reverse)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get blocks: %w", err)
	}

	// Map fabric.Block to service.Block
	blocks := make([]Block, len(fabricBlocks))
	for i, fb := range fabricBlocks {
		blocks[i] = Block{
			Number:       fb.Number,
			Hash:         fb.Hash,
			PreviousHash: fb.PreviousHash,
			Timestamp:    fb.Timestamp,
			TxCount:      fb.TxCount,
			Data:         fb.Data,
		}
	}

	return blocks, total, nil
}

// GetBlockTransactions retrieves all transactions from a specific block
func (s *NetworkService) GetBlockTransactions(ctx context.Context, networkID int64, blockNum uint64) ([]Transaction, error) {
	// Get the fabric deployer for this network
	fabricDeployer, err := s.getFabricDeployerForNetwork(ctx, networkID)
	if err != nil {
		return nil, fmt.Errorf("failed to get fabric deployer: %w", err)
	}

	// Use the fabric deployer to get block transactions
	fabricTransactions, err := fabricDeployer.GetBlockTransactions(ctx, networkID, blockNum)
	if err != nil {
		return nil, fmt.Errorf("failed to get block transactions: %w", err)
	}

	// Map fabric.Transaction to service.Transaction
	transactions := make([]Transaction, len(fabricTransactions))
	for i, ft := range fabricTransactions {
		transactions[i] = Transaction{
			TxID:        ft.ID,
			BlockNumber: ft.BlockNum,
			Timestamp:   ft.Timestamp,
			Type:        ft.Type,
			Creator:     ft.Creator,
		}
	}

	return transactions, nil
}

// GetTransaction retrieves a specific transaction by its ID
func (s *NetworkService) GetTransaction(ctx context.Context, networkID int64, txID string) (Transaction, error) {
	// Get the fabric deployer for this network
	fabricDeployer, err := s.getFabricDeployerForNetwork(ctx, networkID)
	if err != nil {
		return Transaction{}, fmt.Errorf("failed to get fabric deployer: %w", err)
	}

	// Use the fabric deployer to get transaction
	ft, err := fabricDeployer.GetTransaction(ctx, networkID, txID)
	if err != nil {
		return Transaction{}, fmt.Errorf("failed to get transaction: %w", err)
	}

	// Map fabric.Transaction to service.Transaction
	transaction := Transaction{
		TxID:        ft.ID,
		BlockNumber: ft.BlockNum,
		Timestamp:   ft.Timestamp,
		Type:        ft.Type,
		Creator:     ft.Creator,
	}

	return transaction, nil
}

// getFabricDeployerForNetwork creates and returns a fabric deployer for the specified network
func (s *NetworkService) getFabricDeployerForNetwork(ctx context.Context, networkID int64) (*fabric.FabricDeployer, error) {
	// Get network details to verify it exists and is a Fabric network
	network, err := s.db.GetNetwork(ctx, networkID)
	if err != nil {
		return nil, fmt.Errorf("failed to get network: %w", err)
	}
	deployer, err := s.deployerFactory.GetDeployer(network.Platform)
	if err != nil {
		return nil, fmt.Errorf("failed to get deployer: %w", err)
	}

	fabricDeployer, ok := deployer.(*fabric.FabricDeployer)
	if !ok {
		return nil, fmt.Errorf("network %d is not a Fabric network", networkID)
	}

	return fabricDeployer, nil
}

// UpdateOrganizationCRL updates the CRL for an organization in the network
func (s *NetworkService) UpdateOrganizationCRL(ctx context.Context, networkID, organizationID int64) (string, error) {
	// Get network details
	network, err := s.db.GetNetwork(ctx, networkID)
	if err != nil {
		return "", fmt.Errorf("failed to get network: %w", err)
	}

	// Get deployer
	deployer, err := s.deployerFactory.GetDeployer(network.Platform)
	if err != nil {
		return "", fmt.Errorf("failed to get deployer: %w", err)
	}

	fabricDeployer, ok := deployer.(*fabric.FabricDeployer)
	if !ok {
		return "", fmt.Errorf("network %d is not a Fabric network", networkID)
	}

	// Update the CRL in the network
	txID, err := fabricDeployer.UpdateOrganizationCRL(ctx, networkID, fabric.UpdateOrganizationCRLInput{
		OrganizationID: organizationID,
	})
	if err != nil {
		return "", fmt.Errorf("failed to update CRL: %w", err)
	}

	logrus.Info("Reloading network block after updating CRL, waiting 3 seconds")
	time.Sleep(3 * time.Second)

	// Reload network block
	if err := s.ReloadNetworkBlock(ctx, networkID); err != nil {
		logrus.Errorf("Failed to reload network block after updating CRL: %v", err)
	}

	return txID, nil
}
