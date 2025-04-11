package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/chainlaunch/chainlaunch/pkg/config"
	"github.com/chainlaunch/chainlaunch/pkg/db"
	"github.com/chainlaunch/chainlaunch/pkg/errors"
	fabricservice "github.com/chainlaunch/chainlaunch/pkg/fabric/service"
	keymanagement "github.com/chainlaunch/chainlaunch/pkg/keymanagement/service"
	"github.com/chainlaunch/chainlaunch/pkg/logger"
	"github.com/chainlaunch/chainlaunch/pkg/nodes/besu"
	"github.com/chainlaunch/chainlaunch/pkg/nodes/orderer"
	"github.com/chainlaunch/chainlaunch/pkg/nodes/peer"
	"github.com/chainlaunch/chainlaunch/pkg/nodes/types"
	"github.com/chainlaunch/chainlaunch/pkg/nodes/utils"
)

// NodeService handles business logic for node management
type NodeService struct {
	db                   *db.Queries
	logger               *logger.Logger
	keymanagementService *keymanagement.KeyManagementService
	orgService           *fabricservice.OrganizationService
	eventService         *NodeEventService
	configService        *config.ConfigService
}

// CreateNodeRequest represents the service-layer request to create a node
type CreateNodeRequest struct {
	Name               string
	DeploymentMode     types.DeploymentMode
	BlockchainPlatform types.BlockchainPlatform
	FabricPeer         *types.FabricPeerConfig
	FabricOrderer      *types.FabricOrdererConfig
	BesuNode           *types.BesuNodeConfig
}

// NewNodeService creates a new NodeService instance
func NewNodeService(
	db *db.Queries,
	logger *logger.Logger,
	keymanagementService *keymanagement.KeyManagementService,
	orgService *fabricservice.OrganizationService,
	eventService *NodeEventService,
	configService *config.ConfigService,
) *NodeService {
	return &NodeService{
		db:                   db,
		logger:               logger,
		keymanagementService: keymanagementService,
		orgService:           orgService,
		eventService:         eventService,
		configService:        configService,
	}
}

func (s *NodeService) validateCreateNodeRequest(req CreateNodeRequest) error {
	if req.Name == "" {
		return fmt.Errorf("name is required")
	}

	switch req.BlockchainPlatform {
	case types.PlatformFabric:
		if req.FabricPeer == nil && req.FabricOrderer == nil {
			return fmt.Errorf("fabric configuration is required")
		}
		if req.FabricPeer != nil && req.FabricOrderer != nil {
			return fmt.Errorf("cannot specify both peer and orderer configurations")
		}
	case types.PlatformBesu:
		if req.BesuNode == nil {
			return fmt.Errorf("besu configuration is required")
		}
	default:
		return fmt.Errorf("unsupported blockchain platform: %s", req.BlockchainPlatform)
	}

	return nil
}

func (s *NodeService) determineNodeType(req CreateNodeRequest) types.NodeType {
	switch req.BlockchainPlatform {
	case types.PlatformFabric:
		if req.FabricPeer != nil {
			return types.NodeTypeFabricPeer
		}
		return types.NodeTypeFabricOrderer
	case types.PlatformBesu:
		return types.NodeTypeBesuFullnode
	}
	return ""
}

// validateAddress checks if an address:port is valid and available
func (s *NodeService) validateAddress(address string) error {
	host, portStr, err := net.SplitHostPort(address)
	if err != nil {
		return fmt.Errorf("invalid address format %s: %w", address, err)
	}

	// Validate port
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return fmt.Errorf("invalid port number %s: %w", portStr, err)
	}
	if port < 1 || port > 65535 {
		return fmt.Errorf("port number %d out of range (1-65535)", port)
	}

	// Check if port is in use
	addr := fmt.Sprintf("%s:%d", host, port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("address %s is not available: %w", addr, err)
	}
	listener.Close()

	return nil
}

// validateFabricPeerAddresses validates all addresses used by a Fabric peer
func (s *NodeService) validateFabricPeerAddresses(config *types.FabricPeerConfig) error {
	// Get current addresses to compare against
	currentAddresses := map[string]string{
		"listen":     config.ListenAddress,
		"chaincode":  config.ChaincodeAddress,
		"events":     config.EventsAddress,
		"operations": config.OperationsListenAddress,
	}

	// Check for port conflicts between addresses
	usedPorts := make(map[string]string)
	for addrType, addr := range currentAddresses {
		_, port, err := net.SplitHostPort(addr)
		if err != nil {
			return fmt.Errorf("invalid %s address format: %w", addrType, err)
		}

		if existingType, exists := usedPorts[port]; exists {
			// If the port is already used by the same address type, it's okay
			if existingType == addrType {
				continue
			}
			return fmt.Errorf("port conflict: %s and %s addresses use the same port %s", existingType, addrType, port)
		}
		usedPorts[port] = addrType

		// Only validate port availability if it's not already in use by this peer
		if err := s.validateAddress(addr); err != nil {
			// Check if the error is due to the port being in use by this peer
			if strings.Contains(err.Error(), "address already in use") {
				continue
			}
			return fmt.Errorf("invalid %s address: %w", addrType, err)
		}
	}

	return nil
}

// validateFabricOrdererAddresses validates all addresses used by a Fabric orderer
func (s *NodeService) validateFabricOrdererAddresses(config *types.FabricOrdererConfig) error {
	// Validate listen address
	if err := s.validateAddress(config.ListenAddress); err != nil {
		return fmt.Errorf("invalid listen address: %w", err)
	}

	// Validate admin address
	if err := s.validateAddress(config.AdminAddress); err != nil {
		return fmt.Errorf("invalid admin address: %w", err)
	}

	// Validate operations listen address
	if err := s.validateAddress(config.OperationsListenAddress); err != nil {
		return fmt.Errorf("invalid operations listen address: %w", err)
	}

	// Check for port conflicts between addresses
	addresses := map[string]string{
		"listen":     config.ListenAddress,
		"admin":      config.AdminAddress,
		"operations": config.OperationsListenAddress,
	}

	usedPorts := make(map[string]string)
	for addrType, addr := range addresses {
		_, port, err := net.SplitHostPort(addr)
		if err != nil {
			return fmt.Errorf("invalid %s address format: %w", addrType, err)
		}

		if existingType, exists := usedPorts[port]; exists {
			return fmt.Errorf("port conflict: %s and %s addresses use the same port %s", existingType, addrType, port)
		}
		usedPorts[port] = addrType
	}

	return nil
}

// generateSlug creates a URL-friendly slug from a string
func (s *NodeService) generateSlug(name string) string {
	// Convert to lowercase
	slug := strings.ToLower(name)

	// Replace spaces and underscores with hyphens
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "_", "-")

	// Remove all characters except letters, numbers, and hyphens
	reg := regexp.MustCompile("[^a-z0-9-]")
	slug = reg.ReplaceAllString(slug, "")

	// Replace multiple hyphens with a single hyphen
	reg = regexp.MustCompile("-+")
	slug = reg.ReplaceAllString(slug, "-")

	// Trim hyphens from start and end‰
	slug = strings.Trim(slug, "-")

	return slug
}

// GetAllNodes retrieves all nodes without pagination
func (s *NodeService) GetAllNodes(ctx context.Context) (*PaginatedNodes, error) {
	// Get all nodes from the database
	dbNodes, err := s.db.ListNodes(ctx, &db.ListNodesParams{
		Limit:  1000, // Use a high limit to get all nodes
		Offset: 0,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}

	// Get total count
	total, err := s.db.CountNodes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count nodes: %w", err)
	}

	// Map database nodes to service nodes
	nodes := make([]NodeResponse, len(dbNodes))
	for i, dbNode := range dbNodes {
		_, nodeResponse := s.mapDBNodeToServiceNode(dbNode)
		nodes[i] = *nodeResponse
	}

	return &PaginatedNodes{
		Items:       nodes,
		Total:       total,
		Page:        1,
		PageCount:   len(nodes),
		HasNextPage: false,
	}, nil
}

// GetNodeByID retrieves a node by its ID
func (s *NodeService) GetNodeByID(ctx context.Context, id int64) (*NodeResponse, error) {
	node, err := s.db.GetNode(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get node: %w", err)
	}
	_, nodeResponse := s.mapDBNodeToServiceNode(node)
	return nodeResponse, nil
}

// CreateNode creates a new node
func (s *NodeService) CreateNode(ctx context.Context, req CreateNodeRequest) (*NodeResponse, error) {
	if err := s.validateCreateNodeRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Generate slug from name
	slug := s.generateSlug(req.Name)

	// Check if slug already exists
	_, err := s.db.GetNodeBySlug(ctx, slug)
	if err == nil {
		return nil, fmt.Errorf("node with slug %s already exists", slug)
	} else if err != sql.ErrNoRows {
		return nil, fmt.Errorf("error checking slug existence: %w", err)
	}

	// Create node config based on request
	nodeConfig, err := s.createNodeConfig(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create node config: %w", err)
	}

	// Store node config
	configBytes, err := utils.StoreNodeConfig(nodeConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to store node config: %w", err)
	}

	nodeType := s.determineNodeType(req)

	// Determine endpoint based on node type and config
	var endpoint sql.NullString
	switch nodeConfig := nodeConfig.(type) {
	case *types.FabricPeerConfig:
		endpoint = sql.NullString{
			String: nodeConfig.ExternalEndpoint, // Use ExternalEndpoint instead of ListenAddress
			Valid:  true,
		}
	case *types.FabricOrdererConfig:
		endpoint = sql.NullString{
			String: nodeConfig.ExternalEndpoint, // Use ExternalEndpoint instead of ListenAddress
			Valid:  true,
		}
	case *types.BesuNodeConfig:
		endpoint = sql.NullString{
			String: fmt.Sprintf("%s:%d", nodeConfig.ExternalIP, nodeConfig.P2PPort), // Use ExternalIP instead of P2PHost
			Valid:  true,
		}
	}

	// Create node in database
	node, err := s.db.CreateNode(ctx, &db.CreateNodeParams{
		Name:       req.Name,
		Slug:       slug,
		Platform:   string(req.BlockchainPlatform),
		NodeType:   sql.NullString{String: string(nodeType), Valid: true},
		Status:     string(types.NodeStatusPending),
		NodeConfig: sql.NullString{String: string(configBytes), Valid: true},
		Endpoint:   endpoint, // Add endpoint here
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create node: %w", err)
	}

	// Initialize the node based on its type
	deploymentConfig, err := s.initializeNode(ctx, node, req)
	if err != nil {
		// Update node status to failed if initialization fails
		s.updateNodeStatus(ctx, node.ID, types.NodeStatusError)
		return nil, fmt.Errorf("failed to initialize node: %w", err)
	}

	// Store deployment config
	deploymentConfigJSON, err := json.Marshal(deploymentConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal deployment config: %w", err)
	}

	// Update node with deployment config
	node, err = s.db.UpdateNodeDeploymentConfig(ctx, &db.UpdateNodeDeploymentConfigParams{
		ID:               node.ID,
		DeploymentConfig: sql.NullString{String: string(deploymentConfigJSON), Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update node deployment config: %w", err)
	}

	// Start the node
	if err := s.startNode(ctx, node); err != nil {
		return nil, fmt.Errorf("failed to start node: %w", err)
	}
	node, err = s.db.GetNode(ctx, node.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get node: %w", err)
	}
	_, nodeResponse := s.mapDBNodeToServiceNode(node)

	return nodeResponse, nil
}

// Add new function to create node config
func (s *NodeService) createNodeConfig(req CreateNodeRequest) (types.NodeConfig, error) {
	switch req.BlockchainPlatform {
	case types.PlatformFabric:
		if req.FabricPeer != nil {
			return &types.FabricPeerConfig{
				BaseNodeConfig: types.BaseNodeConfig{
					Type: "fabric-peer",
					Mode: "service",
				},
				Name:                    req.FabricPeer.Name,
				OrganizationID:          req.FabricPeer.OrganizationID,
				MSPID:                   req.FabricPeer.MSPID,
				ListenAddress:           req.FabricPeer.ListenAddress,
				ChaincodeAddress:        req.FabricPeer.ChaincodeAddress,
				EventsAddress:           req.FabricPeer.EventsAddress,
				OperationsListenAddress: req.FabricPeer.OperationsListenAddress,
				ExternalEndpoint:        req.FabricPeer.ExternalEndpoint,
				DomainNames:             req.FabricPeer.DomainNames,
				Env:                     req.FabricPeer.Env,
				Version:                 req.FabricPeer.Version,
			}, nil
		} else if req.FabricOrderer != nil {
			return &types.FabricOrdererConfig{
				BaseNodeConfig: types.BaseNodeConfig{
					Type: "fabric-orderer",
					Mode: "service",
				},
				Name:                    req.FabricOrderer.Name,
				OrganizationID:          req.FabricOrderer.OrganizationID,
				MSPID:                   req.FabricOrderer.MSPID,
				ListenAddress:           req.FabricOrderer.ListenAddress,
				AdminAddress:            req.FabricOrderer.AdminAddress,
				OperationsListenAddress: req.FabricOrderer.OperationsListenAddress,
				ExternalEndpoint:        req.FabricOrderer.ExternalEndpoint,
				DomainNames:             req.FabricOrderer.DomainNames,
				Env:                     req.FabricOrderer.Env,
				Version:                 req.FabricOrderer.Version,
			}, nil
		}
	case types.PlatformBesu:
		if req.BesuNode != nil {
			return &types.BesuNodeConfig{
				BaseNodeConfig: types.BaseNodeConfig{
					Type: "besu",
					Mode: req.BesuNode.Mode,
				},
				P2PPort:    req.BesuNode.P2PPort,
				RPCPort:    req.BesuNode.RPCPort,
				NetworkID:  req.BesuNode.NetworkID,
				ExternalIP: req.BesuNode.ExternalIP,
				Env:        req.BesuNode.Env,
				KeyID:      req.BesuNode.KeyID,
				P2PHost:    req.BesuNode.P2PHost,
				RPCHost:    req.BesuNode.RPCHost,
				InternalIP: req.BesuNode.InternalIP,
				BootNodes:  req.BesuNode.BootNodes,
			}, nil
		}
	}
	return nil, fmt.Errorf("invalid node configuration")
}

// initializeNode initializes a node and returns its deployment config
func (s *NodeService) initializeNode(ctx context.Context, dbNode *db.Node, req CreateNodeRequest) (types.NodeDeploymentConfig, error) {
	switch types.BlockchainPlatform(dbNode.Platform) {
	case types.PlatformFabric:
		if req.FabricPeer != nil {
			config, err := s.initializeFabricPeer(ctx, dbNode, req.FabricPeer)
			if err != nil {
				return nil, fmt.Errorf("failed to initialize fabric peer: %w", err)
			}
			return config, nil
		} else if req.FabricOrderer != nil {
			config, err := s.initializeFabricOrderer(ctx, dbNode, req.FabricOrderer)
			if err != nil {
				return nil, fmt.Errorf("failed to initialize fabric orderer: %w", err)
			}
			return config, nil
		}
	case types.PlatformBesu:
		if req.BesuNode != nil {
			config, err := s.initializeBesuNode(ctx, dbNode, req.BesuNode)
			if err != nil {
				return nil, fmt.Errorf("failed to initialize besu node: %w", err)
			}
			return config, nil
		}
	}
	return nil, fmt.Errorf("unsupported platform: %s", dbNode.Platform)
}

// getPeerFromConfig creates a peer instance from the given configuration and database node
func (s *NodeService) getPeerFromConfig(dbNode *db.Node, org *fabricservice.OrganizationDTO, config *types.FabricPeerConfig) *peer.LocalPeer {
	return peer.NewLocalPeer(
		org.MspID,
		s.db,
		peer.StartPeerOpts{
			ID:                      dbNode.Slug,
			ListenAddress:           config.ListenAddress,
			ChaincodeAddress:        config.ChaincodeAddress,
			EventsAddress:           config.EventsAddress,
			OperationsListenAddress: config.OperationsListenAddress,
			ExternalEndpoint:        config.ExternalEndpoint,
			DomainNames:             config.DomainNames,
			Env:                     config.Env,
			Version:                 config.Version,
			AddressOverrides:        config.AddressOverrides,
		},
		config.Mode,
		org,
		org.ID,
		s.orgService,
		s.keymanagementService,
		dbNode.ID,
		s.logger,
		s.configService,
	)
}

// initializeFabricPeer initializes a Fabric peer node
func (s *NodeService) initializeFabricPeer(ctx context.Context, dbNode *db.Node, req *types.FabricPeerConfig) (types.NodeDeploymentConfig, error) {
	org, err := s.orgService.GetOrganization(ctx, req.OrganizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	localPeer := s.getPeerFromConfig(dbNode, org, req)

	// Get deployment config from initialization
	peerConfig, err := localPeer.Init()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize peer: %w", err)
	}

	return peerConfig, nil
}

// getOrdererFromConfig creates a LocalOrderer instance from configuration
func (s *NodeService) getOrdererFromConfig(dbNode *db.Node, org *fabricservice.OrganizationDTO, config *types.FabricOrdererConfig) *orderer.LocalOrderer {
	return orderer.NewLocalOrderer(
		org.MspID,
		s.db,
		orderer.StartOrdererOpts{
			ID:                      dbNode.Name,
			ListenAddress:           config.ListenAddress,
			OperationsListenAddress: config.OperationsListenAddress,
			AdminListenAddress:      config.AdminAddress,
			ExternalEndpoint:        config.ExternalEndpoint,
			DomainNames:             config.DomainNames,
			Env:                     config.Env,
			Version:                 config.Version,
			AddressOverrides:        config.AddressOverrides,
		},
		config.Mode,
		org,
		config.OrganizationID,
		s.orgService,
		s.keymanagementService,
		dbNode.ID,
		s.logger,
		s.configService,
	)
}

// initializeFabricOrderer initializes a Fabric orderer node
func (s *NodeService) initializeFabricOrderer(ctx context.Context, dbNode *db.Node, req *types.FabricOrdererConfig) (*types.FabricOrdererDeploymentConfig, error) {
	org, err := s.orgService.GetOrganization(ctx, req.OrganizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	localOrderer := s.getOrdererFromConfig(dbNode, org, req)

	// Get deployment config from initialization
	config, err := localOrderer.Init()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize orderer: %w", err)
	}

	// Type assert the config
	ordererConfig, ok := config.(*types.FabricOrdererDeploymentConfig)
	if !ok {
		return nil, fmt.Errorf("invalid orderer config type")
	}

	return ordererConfig, nil
}

// initializeBesuNode initializes a Besu node
func (s *NodeService) initializeBesuNode(ctx context.Context, dbNode *db.Node, config *types.BesuNodeConfig) (types.NodeDeploymentConfig, error) {
	// Validate key exists
	key, err := s.keymanagementService.GetKey(ctx, int(config.KeyID))
	if err != nil {
		return nil, fmt.Errorf("failed to get key: %w", err)
	}
	if key.EthereumAddress == "" {
		return nil, fmt.Errorf("key %d has no ethereum address", config.KeyID)
	}
	enodeURL := fmt.Sprintf("enode://%s@%s:%d", key.PublicKey[2:], config.ExternalIP, config.P2PPort)

	// Validate ports
	if err := s.validatePort(config.P2PHost, int(config.P2PPort)); err != nil {
		return nil, fmt.Errorf("invalid P2P port: %w", err)
	}
	if err := s.validatePort(config.RPCHost, int(config.RPCPort)); err != nil {
		return nil, fmt.Errorf("invalid RPC port: %w", err)
	}

	// Create deployment config
	deploymentConfig := &types.BesuNodeDeploymentConfig{
		BaseDeploymentConfig: types.BaseDeploymentConfig{
			Type: "besu",
			Mode: string(config.Mode),
		},
		KeyID:      config.KeyID,
		P2PPort:    config.P2PPort,
		RPCPort:    config.RPCPort,
		NetworkID:  config.NetworkID,
		ExternalIP: config.ExternalIP,
		P2PHost:    config.P2PHost,
		RPCHost:    config.RPCHost,
		InternalIP: config.InternalIP,
		EnodeURL:   enodeURL,
	}

	// Update node endpoint
	endpoint := fmt.Sprintf("%s:%d", config.P2PHost, config.P2PPort)
	_, err = s.db.UpdateNodeEndpoint(ctx, &db.UpdateNodeEndpointParams{
		ID: dbNode.ID,
		Endpoint: sql.NullString{
			String: endpoint,
			Valid:  true,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update node endpoint: %w", err)
	}

	// Update node public endpoint if external IP is set
	if config.ExternalIP != "" {
		publicEndpoint := fmt.Sprintf("%s:%d", config.ExternalIP, config.P2PPort)
		_, err = s.db.UpdateNodePublicEndpoint(ctx, &db.UpdateNodePublicEndpointParams{
			ID: dbNode.ID,
			PublicEndpoint: sql.NullString{
				String: publicEndpoint,
				Valid:  true,
			},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to update node public endpoint: %w", err)
		}
	}

	return deploymentConfig, nil
}

// validatePort checks if a port is valid and available
func (s *NodeService) validatePort(host string, port int) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("port number %d out of range (1-65535)", port)
	}

	// Check if port is in use
	addr := fmt.Sprintf("%s:%d", host, port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("address %s is not available: %w", addr, err)
	}
	listener.Close()
	return nil
}

// updateNodeStatus updates the status of a node in the database
func (s *NodeService) updateNodeStatus(ctx context.Context, nodeID int64, status types.NodeStatus) error {
	_, err := s.db.UpdateNodeStatus(ctx, &db.UpdateNodeStatusParams{
		ID:     nodeID,
		Status: string(status),
	})
	if err != nil {
		return fmt.Errorf("failed to update node status: %w", err)
	}
	dataBytes, err := json.Marshal(map[string]string{"status": string(status)})
	if err != nil {
		return fmt.Errorf("failed to marshal node status: %w", err)
	}
	// Add node status change to event history
	_, err = s.db.CreateNodeEvent(ctx, &db.CreateNodeEventParams{
		NodeID:      nodeID,
		EventType:   string(status),
		Data:        sql.NullString{String: string(dataBytes), Valid: true},
		Description: "status changed",
		Status:      string(status),
	})
	if err != nil {
		return fmt.Errorf("failed to create node event: %w", err)
	}

	// Log the status change
	s.logger.Info("Node status updated",
		"nodeID", nodeID,
		"status", status,
	)

	return nil
}

// GetNode retrieves a node by ID
func (s *NodeService) GetNode(ctx context.Context, id int64) (*NodeResponse, error) {
	node, err := s.db.GetNode(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError("node not found", map[string]interface{}{
				"id": id,
			})
		}
		return nil, fmt.Errorf("failed to get node: %w", err)
	}
	_, nodeResponse := s.mapDBNodeToServiceNode(node)
	return nodeResponse, nil
}

// ListNodes retrieves a paginated list of nodes
func (s *NodeService) ListNodes(ctx context.Context, platform *types.BlockchainPlatform, page, limit int) (*PaginatedNodes, error) {
	var dbNodes []*db.Node
	var err error
	var total int64

	offset := (page - 1) * limit

	if platform != nil {
		// Get nodes filtered by platform
		dbNodes, err = s.db.ListNodesByPlatform(ctx, &db.ListNodesByPlatformParams{
			Platform: string(*platform),
			Limit:    int64(limit),
			Offset:   int64(offset),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list nodes: %w", err)
		}
		total, err = s.db.CountNodesByPlatform(ctx, string(*platform))
	} else {
		// Get all nodes
		dbNodes, err = s.db.ListNodes(ctx, &db.ListNodesParams{
			Limit:  int64(limit),
			Offset: int64(offset),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list nodes: %w", err)
		}
		total, err = s.db.CountNodes(ctx)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to count nodes: %w", err)
	}

	nodes := make([]NodeResponse, len(dbNodes))
	for i, dbNode := range dbNodes {
		_, nodeResponse := s.mapDBNodeToServiceNode(dbNode)
		nodes[i] = *nodeResponse
	}

	return &PaginatedNodes{
		Items:       nodes,
		Total:       total,
		Page:        page,
		PageCount:   (int(total) + limit - 1) / limit,
		HasNextPage: (int(total)+limit-1)/limit > page,
	}, nil
}

// Update mapDBNodeToServiceNode to include deployment config and MSPID
func (s *NodeService) mapDBNodeToServiceNode(dbNode *db.Node) (*Node, *NodeResponse) {
	ctx := context.Background()
	var nodeConfig types.NodeConfig
	var deploymentConfig types.NodeDeploymentConfig

	// Load node config
	if dbNode.NodeConfig.Valid {
		var err error
		nodeConfig, err = utils.LoadNodeConfig([]byte(dbNode.NodeConfig.String))
		if err != nil {
			s.logger.Error("failed to load node config", "error", err)
		}
	}

	// Load deployment config
	if dbNode.DeploymentConfig.Valid {
		var err error
		deploymentConfig, err = utils.DeserializeDeploymentConfig(dbNode.DeploymentConfig.String)
		if err != nil {
			s.logger.Error("failed to deserialize deployment config", "error", err)
		}
	}

	// Create base node
	node := &Node{
		ID:                 dbNode.ID,
		Name:               dbNode.Name,
		BlockchainPlatform: types.BlockchainPlatform(dbNode.Platform),
		NodeType:           types.NodeType(dbNode.NodeType.String),
		Status:             types.NodeStatus(dbNode.Status),
		Endpoint:           dbNode.Endpoint.String,
		PublicEndpoint:     dbNode.PublicEndpoint.String,
		NodeConfig:         nodeConfig,
		DeploymentConfig:   deploymentConfig,
		CreatedAt:          dbNode.CreatedAt,
		UpdatedAt:          dbNode.UpdatedAt.Time,
	}

	// Create node response
	nodeResponse := &NodeResponse{
		ID:        dbNode.ID,
		Name:      dbNode.Name,
		Platform:  dbNode.Platform,
		Status:    dbNode.Status,
		NodeType:  types.NodeType(dbNode.NodeType.String),
		Endpoint:  dbNode.Endpoint.String,
		CreatedAt: dbNode.CreatedAt,
		UpdatedAt: dbNode.UpdatedAt.Time,
	}

	// Add type-specific properties
	if nodeConfig != nil {
		switch config := nodeConfig.(type) {
		case *types.FabricPeerConfig:
			node.MSPID = config.MSPID
			nodeResponse.FabricPeer = &FabricPeerProperties{
				MSPID:             config.MSPID,
				OrganizationID:    config.OrganizationID,
				ExternalEndpoint:  config.ExternalEndpoint,
				ChaincodeAddress:  config.ChaincodeAddress,
				EventsAddress:     config.EventsAddress,
				OperationsAddress: config.OperationsListenAddress,
				ListenAddress:     config.ListenAddress,
				DomainNames:       config.DomainNames,
				Version:           config.Version,
			}
			// Enrich with deployment config if available
			if peerDeployConfig, ok := deploymentConfig.(*types.FabricPeerDeploymentConfig); ok {
				nodeResponse.FabricPeer.ExternalEndpoint = peerDeployConfig.ExternalEndpoint
				nodeResponse.FabricPeer.ListenAddress = peerDeployConfig.ListenAddress
				nodeResponse.FabricPeer.ChaincodeAddress = peerDeployConfig.ChaincodeAddress
				nodeResponse.FabricPeer.EventsAddress = peerDeployConfig.EventsAddress
				nodeResponse.FabricPeer.OperationsAddress = peerDeployConfig.OperationsListenAddress
				nodeResponse.FabricPeer.TLSKeyID = peerDeployConfig.TLSKeyID
				nodeResponse.FabricPeer.SignKeyID = peerDeployConfig.SignKeyID
				nodeResponse.FabricPeer.Mode = peerDeployConfig.Mode
			}
			// Add certificate information
			peerConfig, ok := nodeConfig.(*types.FabricPeerConfig)
			peerDeployConfig, ok := deploymentConfig.(*types.FabricPeerDeploymentConfig)
			if ok && peerConfig != nil {
				nodeResponse.FabricPeer.AddressOverrides = peerDeployConfig.AddressOverrides
				// Get certificates from key service
				signKey, err := s.keymanagementService.GetKey(ctx, int(peerDeployConfig.SignKeyID))
				if err == nil && signKey.Certificate != nil {
					nodeResponse.FabricPeer.SignCert = *signKey.Certificate
					nodeResponse.FabricPeer.SignKeyID = peerDeployConfig.SignKeyID
				}

				tlsKey, err := s.keymanagementService.GetKey(ctx, int(peerDeployConfig.TLSKeyID))
				if err == nil && tlsKey.Certificate != nil {
					nodeResponse.FabricPeer.TLSCert = *tlsKey.Certificate
					nodeResponse.FabricPeer.TLSKeyID = peerDeployConfig.TLSKeyID
				}

				// Get CA certificates from organization
				org, err := s.orgService.GetOrganization(ctx, peerConfig.OrganizationID)
				if err == nil {
					if org.SignKeyID.Valid {
						signCAKey, err := s.keymanagementService.GetKey(ctx, int(org.SignKeyID.Int64))
						if err == nil && signCAKey.Certificate != nil {
							nodeResponse.FabricPeer.SignCACert = *signCAKey.Certificate
						}
					}

					if org.TlsRootKeyID.Valid {
						tlsCAKey, err := s.keymanagementService.GetKey(ctx, int(org.TlsRootKeyID.Int64))
						if err == nil && tlsCAKey.Certificate != nil {
							nodeResponse.FabricPeer.TLSCACert = *tlsCAKey.Certificate
						}
					}
				}
			}

		case *types.FabricOrdererConfig:
			node.MSPID = config.MSPID
			nodeResponse.FabricOrderer = &FabricOrdererProperties{
				MSPID:             config.MSPID,
				OrganizationID:    config.OrganizationID,
				ExternalEndpoint:  config.ExternalEndpoint,
				AdminAddress:      config.AdminAddress,
				OperationsAddress: config.OperationsListenAddress,
				ListenAddress:     config.ListenAddress,
				DomainNames:       config.DomainNames,
				Version:           config.Version,
			}
			// Enrich with deployment config if available
			if ordererDeployConfig, ok := deploymentConfig.(*types.FabricOrdererDeploymentConfig); ok {
				nodeResponse.FabricOrderer.ExternalEndpoint = ordererDeployConfig.ExternalEndpoint
				nodeResponse.FabricOrderer.ListenAddress = ordererDeployConfig.ListenAddress
				nodeResponse.FabricOrderer.AdminAddress = ordererDeployConfig.AdminAddress
				nodeResponse.FabricOrderer.OperationsAddress = ordererDeployConfig.OperationsListenAddress
				nodeResponse.FabricOrderer.TLSKeyID = ordererDeployConfig.TLSKeyID
				nodeResponse.FabricOrderer.SignKeyID = ordererDeployConfig.SignKeyID
				nodeResponse.FabricOrderer.Mode = ordererDeployConfig.Mode
			}
			// Add certificate information
			ordererConfig, ok := nodeConfig.(*types.FabricOrdererConfig)
			ordererDeployConfig, ok := deploymentConfig.(*types.FabricOrdererDeploymentConfig)
			if ok && ordererConfig != nil {
				// Get certificates from key service
				signKey, err := s.keymanagementService.GetKey(ctx, int(ordererDeployConfig.SignKeyID))
				if err == nil && signKey.Certificate != nil {
					nodeResponse.FabricOrderer.SignCert = *signKey.Certificate
				}

				tlsKey, err := s.keymanagementService.GetKey(ctx, int(ordererDeployConfig.TLSKeyID))
				if err == nil && tlsKey.Certificate != nil {
					nodeResponse.FabricOrderer.TLSCert = *tlsKey.Certificate
				}

				// Get CA certificates from organization
				org, err := s.orgService.GetOrganization(ctx, ordererConfig.OrganizationID)
				if err == nil {
					if org.SignKeyID.Valid {
						signCAKey, err := s.keymanagementService.GetKey(ctx, int(org.SignKeyID.Int64))
						if err == nil && signCAKey.Certificate != nil {
							nodeResponse.FabricOrderer.SignCACert = *signCAKey.Certificate
						}
					}

					if org.TlsRootKeyID.Valid {
						tlsCAKey, err := s.keymanagementService.GetKey(ctx, int(org.TlsRootKeyID.Int64))
						if err == nil && tlsCAKey.Certificate != nil {
							nodeResponse.FabricOrderer.TLSCACert = *tlsCAKey.Certificate
						}
					}
				}
			}
		}
	}

	if deploymentConfig != nil {
		switch config := deploymentConfig.(type) {
		case *types.BesuNodeDeploymentConfig:
			nodeResponse.BesuNode = &BesuNodeProperties{
				NetworkID:  config.NetworkID,
				P2PPort:    config.P2PPort,
				RPCPort:    config.RPCPort,
				ExternalIP: config.ExternalIP,
				InternalIP: config.InternalIP,
				EnodeURL:   config.EnodeURL,
				P2PHost:    config.P2PHost,
				RPCHost:    config.RPCHost,
				KeyID:      config.KeyID,
				Mode:       config.Mode,
			}
		}
	}

	return node, nodeResponse
}

// StartNode starts a node by ID
func (s *NodeService) StartNode(ctx context.Context, id int64) (*NodeResponse, error) {
	node, err := s.db.GetNode(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get node: %w", err)
	}

	if err := s.startNode(ctx, node); err != nil {
		return nil, err
	}
	_, nodeResponse := s.mapDBNodeToServiceNode(node)
	return nodeResponse, nil
}

// StopNode stops a node by ID
func (s *NodeService) StopNode(ctx context.Context, id int64) (*NodeResponse, error) {
	node, err := s.db.GetNode(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get node: %w", err)
	}

	// Update status to stopping
	if err := s.updateNodeStatus(ctx, id, types.NodeStatusStopping); err != nil {
		return nil, fmt.Errorf("failed to update node status: %w", err)
	}

	var stopErr error
	switch types.NodeType(node.NodeType.String) {
	case types.NodeTypeFabricPeer:
		stopErr = s.stopFabricPeer(ctx, node)
	case types.NodeTypeFabricOrderer:
		stopErr = s.stopFabricOrderer(ctx, node)
	case types.NodeTypeBesuFullnode:
		stopErr = s.stopBesuNode(ctx, node)
	default:
		stopErr = fmt.Errorf("unsupported node type: %s", node.NodeType.String)
	}

	if stopErr != nil {
		s.logger.Error("Failed to stop node", "error", stopErr)
		// Update status to error if stop failed
		if err := s.updateNodeStatus(ctx, id, types.NodeStatusError); err != nil {
			s.logger.Error("Failed to update node status after stop error", "error", err)
		}
		return nil, fmt.Errorf("failed to stop node: %w", stopErr)
	}

	// Update status to stopped if stop succeeded
	if err := s.updateNodeStatus(ctx, id, types.NodeStatusStopped); err != nil {
		return nil, fmt.Errorf("failed to update node status: %w", err)
	}
	_, nodeResponse := s.mapDBNodeToServiceNode(node)

	return nodeResponse, nil
}

// startNode starts a node based on its type and configuration
func (s *NodeService) startNode(ctx context.Context, dbNode *db.Node) error {
	// Update status to starting
	if err := s.updateNodeStatus(ctx, dbNode.ID, types.NodeStatusStarting); err != nil {
		return fmt.Errorf("failed to update node status: %w", err)
	}

	var startErr error
	switch types.NodeType(dbNode.NodeType.String) {
	case types.NodeTypeFabricPeer:
		startErr = s.startFabricPeer(ctx, dbNode)
	case types.NodeTypeFabricOrderer:
		startErr = s.startFabricOrderer(ctx, dbNode)
	case types.NodeTypeBesuFullnode:
		startErr = s.startBesuNode(ctx, dbNode)
	default:
		startErr = fmt.Errorf("unsupported node type: %s", dbNode.NodeType.String)
	}

	if startErr != nil {
		s.logger.Error("Failed to start node", "error", startErr)
		// Update status to error if start failed
		if err := s.updateNodeStatus(ctx, dbNode.ID, types.NodeStatusError); err != nil {
			s.logger.Error("Failed to update node status after start error", "error", err)
		}
		return fmt.Errorf("failed to start node: %w", startErr)
	}

	// Update status to running if start succeeded
	if err := s.updateNodeStatus(ctx, dbNode.ID, types.NodeStatusRunning); err != nil {
		return fmt.Errorf("failed to update node status: %w", err)
	}

	return nil
}

// startFabricPeer starts a Fabric peer node
func (s *NodeService) startFabricPeer(ctx context.Context, dbNode *db.Node) error {

	nodeConfig, err := utils.LoadNodeConfig([]byte(dbNode.NodeConfig.String))
	if err != nil {
		return fmt.Errorf("failed to deserialize node config: %w", err)
	}
	peerNodeConfig, ok := nodeConfig.(*types.FabricPeerConfig)
	if !ok {
		return fmt.Errorf("failed to assert node config to FabricPeerConfig")
	}

	deploymentConfig, err := utils.DeserializeDeploymentConfig(dbNode.DeploymentConfig.String)
	if err != nil {
		return fmt.Errorf("failed to deserialize deployment config: %w", err)
	}
	s.logger.Info("Starting fabric peer", "deploymentConfig", deploymentConfig)

	peerConfig := deploymentConfig.ToFabricPeerConfig()

	org, err := s.orgService.GetOrganization(ctx, peerConfig.OrganizationID)
	if err != nil {
		return fmt.Errorf("failed to get organization: %w", err)
	}

	localPeer := s.getPeerFromConfig(dbNode, org, peerNodeConfig)

	_, err = localPeer.Start()
	if err != nil {
		return fmt.Errorf("failed to start peer: %w", err)
	}

	return nil
}

// stopFabricPeer stops a Fabric peer node
func (s *NodeService) stopFabricPeer(ctx context.Context, dbNode *db.Node) error {
	deploymentConfig, err := utils.DeserializeDeploymentConfig(dbNode.NodeConfig.String)
	if err != nil {
		return fmt.Errorf("failed to deserialize deployment config: %w", err)
	}
	nodeConfig, err := utils.LoadNodeConfig([]byte(dbNode.NodeConfig.String))
	if err != nil {
		return fmt.Errorf("failed to deserialize node config: %w", err)
	}
	peerNodeConfig, ok := nodeConfig.(*types.FabricPeerConfig)
	if !ok {
		return fmt.Errorf("failed to assert node config to FabricPeerConfig")
	}
	s.logger.Debug("peerNodeConfig", "peerNodeConfig", peerNodeConfig)
	peerConfig := deploymentConfig.ToFabricPeerConfig()
	s.logger.Debug("peerConfig", "peerConfig", peerConfig)
	org, err := s.orgService.GetOrganization(ctx, peerNodeConfig.OrganizationID)
	if err != nil {
		return fmt.Errorf("failed to get organization: %w", err)
	}

	localPeer := s.getPeerFromConfig(dbNode, org, peerNodeConfig)

	err = localPeer.Stop()
	if err != nil {
		return fmt.Errorf("failed to stop peer: %w", err)
	}

	return nil
}

// startFabricOrderer starts a Fabric orderer node
func (s *NodeService) startFabricOrderer(ctx context.Context, dbNode *db.Node) error {
	nodeConfig, err := utils.LoadNodeConfig([]byte(dbNode.NodeConfig.String))
	if err != nil {
		return fmt.Errorf("failed to deserialize node config: %w", err)
	}
	ordererNodeConfig, ok := nodeConfig.(*types.FabricOrdererConfig)
	if !ok {
		return fmt.Errorf("failed to assert node config to FabricOrdererConfig")
	}

	org, err := s.orgService.GetOrganization(ctx, ordererNodeConfig.OrganizationID)
	if err != nil {
		return fmt.Errorf("failed to get organization: %w", err)
	}

	localOrderer := s.getOrdererFromConfig(dbNode, org, ordererNodeConfig)

	_, err = localOrderer.Start()
	if err != nil {
		return fmt.Errorf("failed to start orderer: %w", err)
	}

	return nil
}

// stopFabricOrderer stops a Fabric orderer node
func (s *NodeService) stopFabricOrderer(ctx context.Context, dbNode *db.Node) error {
	nodeConfig, err := utils.LoadNodeConfig([]byte(dbNode.NodeConfig.String))
	if err != nil {
		return fmt.Errorf("failed to deserialize node config: %w", err)
	}
	ordererNodeConfig, ok := nodeConfig.(*types.FabricOrdererConfig)
	if !ok {
		return fmt.Errorf("failed to assert node config to FabricOrdererConfig")
	}

	org, err := s.orgService.GetOrganization(ctx, ordererNodeConfig.OrganizationID)
	if err != nil {
		return fmt.Errorf("failed to get organization: %w", err)
	}

	localOrderer := s.getOrdererFromConfig(dbNode, org, ordererNodeConfig)

	err = localOrderer.Stop()
	if err != nil {
		return fmt.Errorf("failed to stop orderer: %w", err)
	}

	return nil
}

func (s *NodeService) getBesuFromConfig(ctx context.Context, dbNode *db.Node, config *types.BesuNodeConfig, deployConfig *types.BesuNodeDeploymentConfig) (*besu.LocalBesu, error) {
	network, err := s.db.GetNetwork(ctx, deployConfig.NetworkID)
	if err != nil {
		return nil, fmt.Errorf("failed to get network: %w", err)
	}
	key, err := s.keymanagementService.GetKey(ctx, int(config.KeyID))
	if err != nil {
		return nil, fmt.Errorf("failed to get key: %w", err)
	}
	privateKeyDecrypted, err := s.keymanagementService.GetDecryptedPrivateKey(int(config.KeyID))
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt key: %w", err)
	}
	localBesu := besu.NewLocalBesu(
		besu.StartBesuOpts{
			ID:             dbNode.Slug,
			GenesisFile:    network.GenesisBlockB64.String,
			NetworkID:      deployConfig.NetworkID,
			P2PPort:        fmt.Sprintf("%d", deployConfig.P2PPort),
			RPCPort:        fmt.Sprintf("%d", deployConfig.RPCPort),
			ListenAddress:  deployConfig.P2PHost,
			MinerAddress:   key.EthereumAddress,
			ConsensusType:  "qbft", // TODO: get consensus type from network
			BootNodes:      config.BootNodes,
			Version:        "25.2.0", // TODO: get version from network
			NodePrivateKey: strings.TrimPrefix(privateKeyDecrypted, "0x"),
			Env:            config.Env,
		},
		string(config.Mode),
		dbNode.ID,
		s.logger,
		s.configService,
	)

	return localBesu, nil
}

// stopBesuNode stops a Besu node
func (s *NodeService) stopBesuNode(ctx context.Context, dbNode *db.Node) error {
	// Load node configuration
	nodeConfig, err := utils.LoadNodeConfig([]byte(dbNode.NodeConfig.String))
	if err != nil {
		return fmt.Errorf("failed to deserialize node config: %w", err)
	}
	besuNodeConfig, ok := nodeConfig.(*types.BesuNodeConfig)
	if !ok {
		return fmt.Errorf("failed to assert node config to BesuNodeConfig")
	}

	// Load deployment configuration
	deploymentConfig, err := utils.DeserializeDeploymentConfig(dbNode.DeploymentConfig.String)
	if err != nil {
		return fmt.Errorf("failed to deserialize deployment config: %w", err)
	}
	besuDeployConfig, ok := deploymentConfig.(*types.BesuNodeDeploymentConfig)
	if !ok {
		return fmt.Errorf("failed to assert deployment config to BesuNodeDeploymentConfig")
	}

	// Get Besu instance
	localBesu, err := s.getBesuFromConfig(ctx, dbNode, besuNodeConfig, besuDeployConfig)
	if err != nil {
		return fmt.Errorf("failed to get besu instance: %w", err)
	}

	// Stop the node
	err = localBesu.Stop()
	if err != nil {
		return fmt.Errorf("failed to stop besu node: %w", err)
	}

	return nil
}

// startBesuNode starts a Besu node
func (s *NodeService) startBesuNode(ctx context.Context, dbNode *db.Node) error {
	// Load node configuration
	nodeConfig, err := utils.LoadNodeConfig([]byte(dbNode.NodeConfig.String))
	if err != nil {
		return fmt.Errorf("failed to deserialize node config: %w", err)
	}
	besuNodeConfig, ok := nodeConfig.(*types.BesuNodeConfig)
	if !ok {
		return fmt.Errorf("failed to assert node config to BesuNodeConfig")
	}

	// Load deployment configuration
	deploymentConfig, err := utils.DeserializeDeploymentConfig(dbNode.DeploymentConfig.String)
	if err != nil {
		return fmt.Errorf("failed to deserialize deployment config: %w", err)
	}
	besuDeployConfig, ok := deploymentConfig.(*types.BesuNodeDeploymentConfig)
	if !ok {
		return fmt.Errorf("failed to assert deployment config to BesuNodeDeploymentConfig")
	}

	// Get key for node
	key, err := s.keymanagementService.GetKey(ctx, int(besuNodeConfig.KeyID))
	if err != nil {
		return fmt.Errorf("failed to get key: %w", err)
	}
	network, err := s.db.GetNetwork(ctx, besuDeployConfig.NetworkID)
	if err != nil {
		return fmt.Errorf("failed to get network: %w", err)
	}
	privateKeyDecrypted, err := s.keymanagementService.GetDecryptedPrivateKey(int(besuNodeConfig.KeyID))
	if err != nil {
		return fmt.Errorf("failed to decrypt key: %w", err)
	}

	// Create LocalBesu instance
	localBesu := besu.NewLocalBesu(
		besu.StartBesuOpts{
			ID:             dbNode.Slug,
			GenesisFile:    network.GenesisBlockB64.String,
			NetworkID:      besuDeployConfig.NetworkID,
			P2PPort:        fmt.Sprintf("%d", besuDeployConfig.P2PPort),
			RPCPort:        fmt.Sprintf("%d", besuDeployConfig.RPCPort),
			ListenAddress:  besuDeployConfig.P2PHost,
			MinerAddress:   key.EthereumAddress,
			ConsensusType:  "qbft", // TODO: get consensus type from network
			BootNodes:      besuNodeConfig.BootNodes,
			Version:        "25.2.0", // TODO: get version from network
			NodePrivateKey: strings.TrimPrefix(privateKeyDecrypted, "0x"),
			Env:            besuNodeConfig.Env,
		},
		string(besuNodeConfig.Mode),
		dbNode.ID,
		s.logger,
		s.configService,
	)

	// Start the node
	_, err = localBesu.Start()
	if err != nil {
		return fmt.Errorf("failed to start besu node: %w", err)
	}

	s.logger.Info("Started Besu node",
		"nodeID", dbNode.ID,
		"name", dbNode.Name,
		"networkID", besuDeployConfig.NetworkID,
	)

	return nil
}

// Helper function to format arguments for launchd plist
func (s *NodeService) formatPlistArgs(args []string) string {
	var plistArgs strings.Builder
	for _, arg := range args {
		plistArgs.WriteString(fmt.Sprintf("        <string>%s</string>\n", arg))
	}
	return plistArgs.String()
}

// DeleteNode deletes a node by ID
func (s *NodeService) DeleteNode(ctx context.Context, id int64) error {
	// Get the node first to check its type and deployment config
	node, err := s.db.GetNode(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			s.logger.Info("Node not found, already deleted", "id", id)
			return nil
		}
		return fmt.Errorf("failed to get node: %w", err)
	}

	// Stop the node first
	if types.NodeStatus(node.Status) == types.NodeStatusRunning {
		_, err := s.StopNode(ctx, id)
		if err != nil {
			s.logger.Warn("Failed to stop node during deletion", "error", err)
			// Continue with deletion even if stop fails
		}
	}

	// Clean up node-specific resources based on type
	if err := s.cleanupNodeResources(ctx, node); err != nil {
		s.logger.Warn("Failed to cleanup node resources", "error", err)
		// Continue with deletion even if cleanup fails
	}

	// Delete the node from the database
	if err := s.db.DeleteNode(ctx, id); err != nil {
		if err == sql.ErrNoRows {
			s.logger.Info("Node not found during deletion, already deleted", "id", id)
			return nil
		}
		return fmt.Errorf("failed to delete node from database: %w", err)
	}

	return nil
}

// cleanupPeerResources cleans up resources specific to a Fabric peer node
func (s *NodeService) cleanupPeerResources(ctx context.Context, node *db.Node) error {
	// Clean up peer-specific directories
	dirsToClean := []string{
		filepath.Join(s.configService.GetDataPath(), "nodes", node.Slug),
		filepath.Join(s.configService.GetDataPath(), "peers", node.Slug),
		filepath.Join(s.configService.GetDataPath(), "fabric", "peers", node.Slug),
	}

	for _, dir := range dirsToClean {
		if err := os.RemoveAll(dir); err != nil {
			if !os.IsNotExist(err) {
				s.logger.Warn("Failed to remove peer directory",
					"path", dir,
					"error", err)
			}
		} else {
			s.logger.Info("Successfully removed peer directory",
				"path", dir)
		}
	}

	return nil
}

// cleanupOrdererResources cleans up resources specific to a Fabric orderer node
func (s *NodeService) cleanupOrdererResources(ctx context.Context, node *db.Node) error {

	// Clean up orderer-specific directories
	dirsToClean := []string{
		filepath.Join(s.configService.GetDataPath(), "nodes", node.Slug),
		filepath.Join(s.configService.GetDataPath(), "orderers", node.Slug),
		filepath.Join(s.configService.GetDataPath(), "fabric", "orderers", node.Slug),
	}

	for _, dir := range dirsToClean {
		if err := os.RemoveAll(dir); err != nil {
			if !os.IsNotExist(err) {
				s.logger.Warn("Failed to remove orderer directory",
					"path", dir,
					"error", err)
			}
		} else {
			s.logger.Info("Successfully removed orderer directory",
				"path", dir)
		}
	}

	return nil
}

// cleanupBesuResources cleans up resources specific to a Besu node
func (s *NodeService) cleanupBesuResources(ctx context.Context, node *db.Node) error {

	// Load node configuration
	nodeConfig, err := utils.LoadNodeConfig([]byte(node.NodeConfig.String))
	if err != nil {
		s.logger.Warn("Failed to load node config during cleanup", "error", err)
		// Continue with cleanup even if config loading fails
	}

	// Load deployment configuration
	deploymentConfig, err := utils.DeserializeDeploymentConfig(node.DeploymentConfig.String)
	if err != nil {
		s.logger.Warn("Failed to load deployment config during cleanup", "error", err)
		// Continue with cleanup even if config loading fails
	}

	// Create Besu instance for cleanup
	var localBesu *besu.LocalBesu
	if nodeConfig != nil && deploymentConfig != nil {
		besuNodeConfig, ok := nodeConfig.(*types.BesuNodeConfig)
		if !ok {
			s.logger.Warn("Invalid node config type during cleanup")
		}
		besuDeployConfig, ok := deploymentConfig.(*types.BesuNodeDeploymentConfig)
		if !ok {
			s.logger.Warn("Invalid deployment config type during cleanup")
		}
		if besuNodeConfig != nil && besuDeployConfig != nil {
			localBesu, err = s.getBesuFromConfig(ctx, node, besuNodeConfig, besuDeployConfig)
			if err != nil {
				s.logger.Warn("Failed to create Besu instance during cleanup", "error", err)
			}
		}
	}

	// Stop the service if it's running and we have a valid Besu instance
	if localBesu != nil {
		if err := localBesu.Stop(); err != nil {
			s.logger.Warn("Failed to stop Besu service during cleanup", "error", err)
			// Continue with cleanup even if stop fails
		}
	}

	// Clean up Besu-specific directories
	dirsToClean := []string{
		filepath.Join(s.configService.GetDataPath(), "nodes", node.Slug),
		filepath.Join(s.configService.GetDataPath(), "besu", node.Slug),
		filepath.Join(s.configService.GetDataPath(), "besu", "nodes", node.Slug),
	}

	for _, dir := range dirsToClean {
		if err := os.RemoveAll(dir); err != nil {
			if !os.IsNotExist(err) {
				s.logger.Warn("Failed to remove Besu directory",
					"path", dir,
					"error", err)
			}
		} else {
			s.logger.Info("Successfully removed Besu directory",
				"path", dir)
		}
	}

	// Clean up service files based on platform
	switch runtime.GOOS {
	case "linux":
		// Remove systemd service file
		if localBesu != nil {
			serviceFile := fmt.Sprintf("/etc/systemd/system/besu-%s.service", node.Slug)
			if err := os.Remove(serviceFile); err != nil {
				if !os.IsNotExist(err) {
					s.logger.Warn("Failed to remove systemd service file", "error", err)
				}
			}
		}

	case "darwin":
		// Remove launchd plist file
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		if localBesu != nil {
			plistFile := filepath.Join(homeDir, "Library/LaunchAgents", fmt.Sprintf("dev.chainlaunch.besu.%s.plist", node.Slug))
			if err := os.Remove(plistFile); err != nil {
				if !os.IsNotExist(err) {
					s.logger.Warn("Failed to remove launchd plist file", "error", err)
				}
			}
		}
	}

	// Clean up any data directories
	dataDir := filepath.Join(s.configService.GetDataPath(), "data", "besu", node.Slug)
	if err := os.RemoveAll(dataDir); err != nil {
		if !os.IsNotExist(err) {
			s.logger.Warn("Failed to remove Besu data directory",
				"path", dataDir,
				"error", err)
		}
	} else {
		s.logger.Info("Successfully removed Besu data directory",
			"path", dataDir)
	}

	return nil
}

// Update cleanupNodeResources to use the new function
func (s *NodeService) cleanupNodeResources(ctx context.Context, node *db.Node) error {
	// Get the home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	deploymentConfig, err := utils.DeserializeDeploymentConfig(node.DeploymentConfig.String)
	if err != nil {
		return fmt.Errorf("failed to deserialize deployment config: %w", err)
	}

	// Clean up service files based on platform
	switch runtime.GOOS {
	case "linux":
		// Remove systemd service file
		serviceFile := fmt.Sprintf("/etc/systemd/system/%s.service", deploymentConfig.GetServiceName())
		if err := os.Remove(serviceFile); err != nil {
			if !os.IsNotExist(err) {
				s.logger.Warn("Failed to remove systemd service file", "error", err)
			}
		}

	case "darwin":
		// Remove launchd plist file
		plistFile := filepath.Join(homeDir, "Library/LaunchAgents", fmt.Sprintf("dev.chainlaunch.%s.plist", deploymentConfig.GetServiceName()))
		if err := os.Remove(plistFile); err != nil {
			if !os.IsNotExist(err) {
				s.logger.Warn("Failed to remove launchd plist file", "error", err)
			}
		}
	}

	// Clean up node-specific resources based on type
	switch types.NodeType(node.NodeType.String) {
	case types.NodeTypeFabricPeer:
		if err := s.cleanupPeerResources(ctx, node); err != nil {
			s.logger.Warn("Failed to cleanup peer resources", "error", err)
		}
	case types.NodeTypeFabricOrderer:
		if err := s.cleanupOrdererResources(ctx, node); err != nil {
			s.logger.Warn("Failed to cleanup orderer resources", "error", err)
		}
	case types.NodeTypeBesuFullnode:
		if err := s.cleanupBesuResources(ctx, node); err != nil {
			s.logger.Warn("Failed to cleanup besu resources", "error", err)
		}
	default:
		s.logger.Warn("Unknown node type for cleanup", "type", node.NodeType.String)
	}

	return nil
}

// GetFabricPeerDefaults returns default values for a Fabric peer node
func (s *NodeService) GetFabricPeerDefaults() *NodeDefaults {
	// Get available ports for peer services
	listen, chaincode, events, operations, err := GetPeerPorts(7051)
	if err != nil {
		// If we can't get the preferred ports, try from a higher range
		listen, chaincode, events, operations, err = GetPeerPorts(10000)
		if err != nil {
			s.logger.Error("Failed to get available ports for peer", "error", err)
			// Fall back to default ports if all attempts fail
			return &NodeDefaults{
				ListenAddress:           "0.0.0.0:7051",
				ExternalEndpoint:        "localhost:7051",
				ChaincodeAddress:        "0.0.0.0:7052",
				EventsAddress:           "0.0.0.0:7053",
				OperationsListenAddress: "0.0.0.0:9443",
				Mode:                    ModeService,
				ServiceName:             "fabric-peer",
				LogPath:                 "/var/log/fabric/peer.log",
				ErrorLogPath:            "/var/log/fabric/peer.err",
			}
		}
	}

	return &NodeDefaults{
		ListenAddress:           fmt.Sprintf("0.0.0.0:%d", listen),
		ExternalEndpoint:        fmt.Sprintf("localhost:%d", listen),
		ChaincodeAddress:        fmt.Sprintf("0.0.0.0:%d", chaincode),
		EventsAddress:           fmt.Sprintf("0.0.0.0:%d", events),
		OperationsListenAddress: fmt.Sprintf("0.0.0.0:%d", operations),
		Mode:                    ModeService,
		ServiceName:             "fabric-peer",
		LogPath:                 "/var/log/fabric/peer.log",
		ErrorLogPath:            "/var/log/fabric/peer.err",
	}
}

// GetFabricOrdererDefaults returns default values for a Fabric orderer node
func (s *NodeService) GetFabricOrdererDefaults() *NodeDefaults {
	// Get available ports for orderer services
	listen, admin, operations, err := GetOrdererPorts(7050)
	if err != nil {
		// If we can't get the preferred ports, try from a higher range
		listen, admin, operations, err = GetOrdererPorts(10000)
		if err != nil {
			s.logger.Error("Failed to get available ports for orderer", "error", err)
			// Fall back to default ports if all attempts fail
			return &NodeDefaults{
				ListenAddress:           "0.0.0.0:7050",
				ExternalEndpoint:        "localhost:7050",
				AdminAddress:            "0.0.0.0:7053",
				OperationsListenAddress: "0.0.0.0:8443",
				Mode:                    ModeService,
				ServiceName:             "fabric-orderer",
				LogPath:                 "/var/log/fabric/orderer.log",
				ErrorLogPath:            "/var/log/fabric/orderer.err",
			}
		}
	}

	return &NodeDefaults{
		ListenAddress:           fmt.Sprintf("0.0.0.0:%d", listen),
		ExternalEndpoint:        fmt.Sprintf("localhost:%d", listen),
		AdminAddress:            fmt.Sprintf("0.0.0.0:%d", admin),
		OperationsListenAddress: fmt.Sprintf("0.0.0.0:%d", operations),
		Mode:                    ModeService,
		ServiceName:             "fabric-orderer",
		LogPath:                 "/var/log/fabric/orderer.log",
		ErrorLogPath:            "/var/log/fabric/orderer.err",
	}
}

// Update the port offsets and base ports to prevent overlap
const (
	// Base ports for peers and orderers with sufficient spacing
	peerBasePort    = 7000 // Starting port for peers
	ordererBasePort = 9000 // Starting port for orderers with 2000 port gap

	// Port offsets to ensure no overlap within node types
	peerPortOffset    = 100 // Each peer gets a 100 port range
	ordererPortOffset = 100 // Each orderer gets a 100 port range

	maxPortAttempts = 100 // Maximum attempts to find available ports
)

// GetNodesDefaults returns default values for multiple nodes with guaranteed non-overlapping ports
func (s *NodeService) GetNodesDefaults(params NodesDefaultsParams) (*NodesDefaultsResult, error) {
	// Validate node counts
	if params.PeerCount > 15 {
		return nil, fmt.Errorf("peer count exceeds maximum supported nodes (15)")
	}
	if params.OrdererCount > 15 {
		return nil, fmt.Errorf("orderer count exceeds maximum supported nodes (15)")
	}

	result := &NodesDefaultsResult{
		Peers:              make([]NodeDefaults, params.PeerCount),
		Orderers:           make([]NodeDefaults, params.OrdererCount),
		AvailableAddresses: []string{"localhost", "0.0.0.0"},
	}

	// Generate peer defaults with incremental ports
	// Each peer needs 4 ports (listen, chaincode, events, operations)
	for i := 0; i < params.PeerCount; i++ {
		basePort := peerBasePort + (i * peerPortOffset)
		listen, chaincode, events, operations, err := GetPeerPorts(basePort)
		if err != nil {
			// Try with a higher range if initial attempt fails
			listen, chaincode, events, operations, err = GetPeerPorts(10000 + (i * peerPortOffset))
			if err != nil {
				return nil, fmt.Errorf("failed to get peer ports: %w", err)
			}
		}

		// Validate that ports don't overlap with orderer range
		if listen >= ordererBasePort || chaincode >= ordererBasePort ||
			events >= ordererBasePort || operations >= ordererBasePort {
			return nil, fmt.Errorf("peer ports would overlap with orderer port range")
		}

		result.Peers[i] = NodeDefaults{
			ListenAddress:           fmt.Sprintf("0.0.0.0:%d", listen),
			ExternalEndpoint:        fmt.Sprintf("localhost:%d", listen),
			ChaincodeAddress:        fmt.Sprintf("0.0.0.0:%d", chaincode),
			EventsAddress:           fmt.Sprintf("0.0.0.0:%d", events),
			OperationsListenAddress: fmt.Sprintf("0.0.0.0:%d", operations),
			Mode:                    params.Mode,
			ServiceName:             fmt.Sprintf("fabric-peer-%d", i+1),
			LogPath:                 fmt.Sprintf("/var/log/fabric/peer%d.log", i+1),
			ErrorLogPath:            fmt.Sprintf("/var/log/fabric/peer%d.err", i+1),
		}
	}

	// Generate orderer defaults with incremental ports
	// Each orderer needs 3 ports (listen, admin, operations)
	for i := 0; i < params.OrdererCount; i++ {
		basePort := ordererBasePort + (i * ordererPortOffset)
		listen, admin, operations, err := GetOrdererPorts(basePort)
		if err != nil {
			// Try with a higher range if initial attempt fails
			listen, admin, operations, err = GetOrdererPorts(11000 + (i * ordererPortOffset))
			if err != nil {
				return nil, fmt.Errorf("failed to get orderer ports: %w", err)
			}
		}

		// Validate that ports don't overlap with peer range
		maxPeerPort := peerBasePort + (15 * peerPortOffset) // Account for maximum possible peers
		if listen <= maxPeerPort ||
			admin <= maxPeerPort ||
			operations <= maxPeerPort {
			return nil, fmt.Errorf("orderer ports would overlap with peer port range")
		}

		result.Orderers[i] = NodeDefaults{
			ListenAddress:           fmt.Sprintf("0.0.0.0:%d", listen),
			ExternalEndpoint:        fmt.Sprintf("localhost:%d", listen),
			AdminAddress:            fmt.Sprintf("0.0.0.0:%d", admin),
			OperationsListenAddress: fmt.Sprintf("0.0.0.0:%d", operations),
			Mode:                    params.Mode,
			ServiceName:             fmt.Sprintf("fabric-orderer-%d", i+1),
			LogPath:                 fmt.Sprintf("/var/log/fabric/orderer%d.log", i+1),
			ErrorLogPath:            fmt.Sprintf("/var/log/fabric/orderer%d.err", i+1),
		}
	}

	return result, nil
}

// TailLogs returns a channel that receives log lines from the specified node
func (s *NodeService) TailLogs(ctx context.Context, nodeID int64, tail int, follow bool) (<-chan string, error) {
	// Get the node first to verify it exists
	dbNode, err := s.db.GetNode(ctx, nodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get node: %w", err)
	}

	// Get deployment config
	deploymentConfig, err := utils.DeserializeDeploymentConfig(dbNode.DeploymentConfig.String)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize deployment config: %w", err)
	}

	switch types.NodeType(dbNode.NodeType.String) {
	case types.NodeTypeFabricPeer:
		nodeConfig, err := utils.LoadNodeConfig([]byte(dbNode.NodeConfig.String))
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize node config: %w", err)
		}
		peerNodeConfig, ok := nodeConfig.(*types.FabricPeerConfig)
		if !ok {
			return nil, fmt.Errorf("failed to assert node config to FabricPeerConfig")
		}
		s.logger.Debug("Peer config", "config", peerNodeConfig, "deploymentConfig", deploymentConfig)
		// Get organization
		org, err := s.orgService.GetOrganization(ctx, peerNodeConfig.OrganizationID)
		if err != nil {
			return nil, fmt.Errorf("failed to get organization: %w", err)
		}

		// Create peer instance
		localPeer := s.getPeerFromConfig(dbNode, org, peerNodeConfig)

		// Tail logs from peer
		return localPeer.TailLogs(ctx, tail, follow)
	case types.NodeTypeFabricOrderer:
		// Convert to FabricOrdererDeploymentConfig
		nodeConfig, err := utils.LoadNodeConfig([]byte(dbNode.NodeConfig.String))
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize node config: %w", err)
		}
		ordererNodeConfig, ok := nodeConfig.(*types.FabricOrdererConfig)
		if !ok {
			return nil, fmt.Errorf("failed to assert node config to FabricOrdererConfig")
		}
		s.logger.Info("Orderer config", "config", ordererNodeConfig, "deploymentConfig", deploymentConfig)
		// Get organization
		org, err := s.orgService.GetOrganization(ctx, ordererNodeConfig.OrganizationID)
		if err != nil {
			return nil, fmt.Errorf("failed to get organization: %w", err)
		}
		// Create orderer instance
		localOrderer := s.getOrdererFromConfig(dbNode, org, ordererNodeConfig)
		// Tail logs from orderer
		return localOrderer.TailLogs(ctx, tail, follow)
	case types.NodeTypeBesuFullnode:
		nodeConfig, err := utils.LoadNodeConfig([]byte(dbNode.NodeConfig.String))
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize node config: %w", err)
		}
		besuNodeConfig, ok := nodeConfig.(*types.BesuNodeConfig)
		if !ok {
			return nil, fmt.Errorf("failed to assert node config to BesuNodeConfig")
		}
		besuDeployConfig := deploymentConfig.ToBesuNodeConfig()

		localBesu, err := s.getBesuFromConfig(ctx, dbNode, besuNodeConfig, besuDeployConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to get besu from config: %w", err)
		}
		return localBesu.TailLogs(ctx, tail, follow)
	default:
		return nil, fmt.Errorf("unsupported node type for log tailing: %s", dbNode.NodeType.String)
	}
}

// GetNodeEvents retrieves a paginated list of node events
func (s *NodeService) GetNodeEvents(ctx context.Context, nodeID int64, page, limit int) ([]NodeEvent, error) {
	return s.eventService.GetEvents(ctx, nodeID, page, limit)
}

// GetLatestNodeEvent retrieves the latest event for a node
func (s *NodeService) GetLatestNodeEvent(ctx context.Context, nodeID int64) (*NodeEvent, error) {
	return s.eventService.GetLatestEvent(ctx, nodeID)
}

// GetEventsByType retrieves a paginated list of node events of a specific type
func (s *NodeService) GetEventsByType(ctx context.Context, nodeID int64, eventType NodeEventType, page, limit int) ([]NodeEvent, error) {
	return s.eventService.GetEventsByType(ctx, nodeID, eventType, page, limit)
}

// GetFabricPeer gets a Fabric peer node configuration
func (s *NodeService) GetFabricPeer(ctx context.Context, id int64) (*peer.LocalPeer, error) {
	// Get the node from database
	node, err := s.db.GetNode(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("peer node not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get peer node: %w", err)
	}

	// Verify node type
	if types.NodeType(node.NodeType.String) != types.NodeTypeFabricPeer {
		return nil, fmt.Errorf("node %d is not a Fabric peer", id)
	}

	// Load node config
	nodeConfig, err := utils.LoadNodeConfig([]byte(node.NodeConfig.String))
	if err != nil {
		return nil, fmt.Errorf("failed to load peer config: %w", err)
	}

	// Type assert to FabricPeerConfig
	peerConfig, ok := nodeConfig.(*types.FabricPeerConfig)
	if !ok {
		return nil, fmt.Errorf("invalid peer config type")
	}

	// Get deployment config if available
	if node.DeploymentConfig.Valid {
		deploymentConfig, err := utils.DeserializeDeploymentConfig(node.DeploymentConfig.String)
		if err != nil {
			s.logger.Warn("Failed to deserialize deployment config", "error", err)
		} else {
			// Update config with deployment values
			if deployConfig, ok := deploymentConfig.(*types.FabricPeerDeploymentConfig); ok {
				peerConfig.ExternalEndpoint = deployConfig.ExternalEndpoint
				// Add any other deployment-specific fields that should be included
			}
		}
	}

	// Get organization
	org, err := s.orgService.GetOrganization(ctx, peerConfig.OrganizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	// Create and return local peer
	localPeer := s.getPeerFromConfig(node, org, peerConfig)
	return localPeer, nil
}

// GetFabricOrderer gets a Fabric orderer node configuration
func (s *NodeService) GetFabricOrderer(ctx context.Context, id int64) (*orderer.LocalOrderer, error) {
	// Get the node from database
	node, err := s.db.GetNode(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("orderer node not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get orderer node: %w", err)
	}

	// Verify node type
	if types.NodeType(node.NodeType.String) != types.NodeTypeFabricOrderer {
		return nil, fmt.Errorf("node %d is not a Fabric orderer", id)
	}

	// Load node config
	nodeConfig, err := utils.LoadNodeConfig([]byte(node.NodeConfig.String))
	if err != nil {
		return nil, fmt.Errorf("failed to load orderer config: %w", err)
	}

	// Type assert to FabricOrdererConfig
	ordererConfig, ok := nodeConfig.(*types.FabricOrdererConfig)
	if !ok {
		return nil, fmt.Errorf("invalid orderer config type")
	}

	// Get deployment config if available
	if node.DeploymentConfig.Valid {
		deploymentConfig, err := utils.DeserializeDeploymentConfig(node.DeploymentConfig.String)
		if err != nil {
			s.logger.Warn("Failed to deserialize deployment config", "error", err)
		} else {
			// Update config with deployment values
			if deployConfig, ok := deploymentConfig.(*types.FabricOrdererDeploymentConfig); ok {
				ordererConfig.ExternalEndpoint = deployConfig.ExternalEndpoint
				// Add any other deployment-specific fields that should be included
			}
		}
	}

	// Get organization
	org, err := s.orgService.GetOrganization(ctx, ordererConfig.OrganizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	// Create and return local orderer
	localOrderer := s.getOrdererFromConfig(node, org, ordererConfig)
	return localOrderer, nil
}

// GetFabricNodesByOrganization gets all Fabric nodes (peers and orderers) for an organization
func (s *NodeService) GetFabricNodesByOrganization(ctx context.Context, orgID int64) ([]NodeResponse, error) {
	// Get all nodes
	nodes, err := s.GetAllNodes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get nodes: %w", err)
	}

	// Filter nodes by organization
	var orgNodes []NodeResponse
	for _, node := range nodes.Items {
		// Check node type and config
		switch node.NodeType {
		case types.NodeTypeFabricPeer:
			if node.FabricPeer != nil {
				if node.FabricPeer.OrganizationID == orgID {
					orgNodes = append(orgNodes, node)
				}
			}
		case types.NodeTypeFabricOrderer:
			if node.FabricOrderer != nil {
				if node.FabricOrderer.OrganizationID == orgID {
					orgNodes = append(orgNodes, node)
				}
			}
		}
	}

	return orgNodes, nil
}

// GetBesuPorts attempts to find available ports for P2P and RPC, starting from default ports
func GetBesuPorts(baseP2PPort, baseRPCPort uint) (p2pPort uint, rpcPort uint, err error) {
	maxAttempts := 100
	// Try to find available ports for P2P and RPC
	p2pPorts, err := findConsecutivePorts(int(baseP2PPort), 1, int(baseP2PPort)+maxAttempts)
	if err != nil {
		return 0, 0, fmt.Errorf("could not find available P2P port: %w", err)
	}
	p2pPort = uint(p2pPorts[0])

	rpcPorts, err := findConsecutivePorts(int(baseRPCPort), 1, int(baseRPCPort)+maxAttempts)
	if err != nil {
		return 0, 0, fmt.Errorf("could not find available RPC port: %w", err)
	}
	rpcPort = uint(rpcPorts[0])

	return p2pPort, rpcPort, nil
}

// GetBesuNodeDefaults returns the default configuration for a Besu node
func (s *NodeService) GetBesuNodeDefaults() (*BesuNodeDefaults, error) {
	// Try to get available ports starting from default Besu ports
	p2pPort, rpcPort, err := GetBesuPorts(30303, 8545)
	if err != nil {
		// If we can't get the preferred ports, try from a higher range
		p2pPort, rpcPort, err = GetBesuPorts(40303, 18545)
		if err != nil {
			return nil, fmt.Errorf("failed to find available ports: %w", err)
		}
	}
	externalIP := "127.0.0.1"
	internalIP := "127.0.0.1"

	return &BesuNodeDefaults{
		P2PAddress: fmt.Sprintf("%s:%d", externalIP, p2pPort),
		RPCAddress: fmt.Sprintf("%s:%d", externalIP, rpcPort),
		NetworkID:  1337, // Default private network ID
		Mode:       ModeService,
		ExternalIP: externalIP,
		InternalIP: internalIP,
	}, nil
}

// Add a method to get full node details when needed
func (s *NodeService) GetNodeWithConfig(ctx context.Context, id int64) (*Node, error) {
	dbNode, err := s.db.GetNode(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get node: %w", err)
	}

	node, _ := s.mapDBNodeToServiceNode(dbNode)
	return node, nil
}

// Update the fabric deployer to use this method
func (s *NodeService) GetNodeForDeployment(ctx context.Context, id int64) (*Node, error) {
	return s.GetNodeWithConfig(ctx, id)
}

// Channel represents a Fabric channel
type Channel struct {
	Name      string    `json:"name"`
	BlockNum  int64     `json:"blockNum"`
	CreatedAt time.Time `json:"createdAt"`
}

// GetNodeChannels retrieves the list of channels for a Fabric node
func (s *NodeService) GetNodeChannels(ctx context.Context, id int64) ([]Channel, error) {
	// Get the node first
	node, err := s.db.GetNode(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError("node not found", nil)
		}
		return nil, fmt.Errorf("failed to get node: %w", err)
	}

	// Verify node type
	nodeType := types.NodeType(node.NodeType.String)
	if nodeType != types.NodeTypeFabricPeer && nodeType != types.NodeTypeFabricOrderer {
		return nil, errors.NewValidationError("node is not a Fabric node", nil)
	}

	switch nodeType {
	case types.NodeTypeFabricPeer:
		// Get peer instance
		peer, err := s.GetFabricPeer(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("failed to get peer: %w", err)
		}
		peerChannels, err := peer.GetChannels(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get peer channels: %w", err)
		}
		channels := make([]Channel, len(peerChannels))
		for i, channel := range peerChannels {
			channels[i] = Channel{
				Name:      channel.Name,
				BlockNum:  channel.BlockNum,
				CreatedAt: channel.CreatedAt,
			}
		}
		return channels, nil

	case types.NodeTypeFabricOrderer:
		// Get orderer instance
		orderer, err := s.GetFabricOrderer(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("failed to get orderer: %w", err)
		}
		ordererChannels, err := orderer.GetChannels(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get orderer channels: %w", err)
		}
		channels := make([]Channel, len(ordererChannels))
		for i, channel := range ordererChannels {
			channels[i] = Channel{
				Name:      channel.Name,
				BlockNum:  channel.BlockNum,
				CreatedAt: channel.CreatedAt,
			}
		}
		return channels, nil
	}

	return nil, fmt.Errorf("unsupported node type: %s", nodeType)
}

// RenewCertificates renews the certificates for a node
func (s *NodeService) RenewCertificates(ctx context.Context, id int64) (*NodeResponse, error) {
	// Get the node from database
	node, err := s.db.GetNode(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError("node not found", nil)
		}
		return nil, fmt.Errorf("failed to get node: %w", err)
	}

	// Update status to indicate certificate renewal is in progress
	if err := s.updateNodeStatus(ctx, id, types.NodeStatusUpdating); err != nil {
		return nil, fmt.Errorf("failed to update node status: %w", err)
	}

	// Get deployment config
	deploymentConfig, err := utils.DeserializeDeploymentConfig(node.DeploymentConfig.String)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize deployment config: %w", err)
	}

	var renewErr error
	switch types.NodeType(node.NodeType.String) {
	case types.NodeTypeFabricPeer:
		renewErr = s.renewPeerCertificates(ctx, node, deploymentConfig)
	case types.NodeTypeFabricOrderer:
		renewErr = s.renewOrdererCertificates(ctx, node, deploymentConfig)
	default:
		renewErr = fmt.Errorf("certificate renewal not supported for node type: %s", node.NodeType.String)
	}

	if renewErr != nil {
		// Update status to error if renewal failed
		if err := s.updateNodeStatus(ctx, id, types.NodeStatusError); err != nil {
			s.logger.Error("Failed to update node status after renewal error", "error", err)
		}
		return nil, fmt.Errorf("failed to renew certificates: %w", renewErr)
	}

	// Update status to running after successful renewal
	if err := s.updateNodeStatus(ctx, id, types.NodeStatusRunning); err != nil {
		return nil, fmt.Errorf("failed to update node status: %w", err)
	}

	// Get updated node
	updatedNode, err := s.GetNode(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated node: %w", err)
	}

	return updatedNode, nil
}

// renewPeerCertificates handles certificate renewal for a Fabric peer
func (s *NodeService) renewPeerCertificates(ctx context.Context, dbNode *db.Node, deploymentConfig types.NodeDeploymentConfig) error {
	nodeConfig, err := utils.LoadNodeConfig([]byte(dbNode.NodeConfig.String))
	if err != nil {
		return fmt.Errorf("failed to load node config: %w", err)
	}

	peerConfig, ok := nodeConfig.(*types.FabricPeerConfig)
	if !ok {
		return fmt.Errorf("invalid peer config type")
	}

	peerDeployConfig, ok := deploymentConfig.(*types.FabricPeerDeploymentConfig)
	if !ok {
		return fmt.Errorf("invalid peer deployment config type")
	}

	org, err := s.orgService.GetOrganization(ctx, peerConfig.OrganizationID)
	if err != nil {
		return fmt.Errorf("failed to get organization: %w", err)
	}

	localPeer := s.getPeerFromConfig(dbNode, org, peerConfig)
	err = localPeer.RenewCertificates(peerDeployConfig)
	if err != nil {
		return fmt.Errorf("failed to renew peer certificates: %w", err)
	}

	return nil
}

// renewOrdererCertificates handles certificate renewal for a Fabric orderer
func (s *NodeService) renewOrdererCertificates(ctx context.Context, dbNode *db.Node, deploymentConfig types.NodeDeploymentConfig) error {
	nodeConfig, err := utils.LoadNodeConfig([]byte(dbNode.NodeConfig.String))
	if err != nil {
		return fmt.Errorf("failed to load node config: %w", err)
	}

	ordererConfig, ok := nodeConfig.(*types.FabricOrdererConfig)
	if !ok {
		return fmt.Errorf("invalid orderer config type")
	}

	ordererDeployConfig, ok := deploymentConfig.(*types.FabricOrdererDeploymentConfig)
	if !ok {
		return fmt.Errorf("invalid orderer deployment config type")
	}

	org, err := s.orgService.GetOrganization(ctx, ordererConfig.OrganizationID)
	if err != nil {
		return fmt.Errorf("failed to get organization: %w", err)
	}

	localOrderer := s.getOrdererFromConfig(dbNode, org, ordererConfig)
	err = localOrderer.RenewCertificates(ordererDeployConfig)
	if err != nil {
		return fmt.Errorf("failed to renew orderer certificates: %w", err)
	}

	return nil
}

// UpdateNodeEnvironment updates the environment variables for a node
func (s *NodeService) UpdateNodeEnvironment(ctx context.Context, nodeID int64, req *types.UpdateNodeEnvRequest) (*types.UpdateNodeEnvResponse, error) {
	// Get the node from the database
	dbNode, err := s.db.GetNode(ctx, nodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get node: %w", err)
	}

	// Get the node's current configuration
	switch dbNode.NodeType.String {
	case string(types.NodeTypeFabricPeer):
		var peerConfig types.FabricPeerConfig
		if err := json.Unmarshal([]byte(dbNode.Config.String), &peerConfig); err != nil {
			return nil, fmt.Errorf("failed to unmarshal peer config: %w", err)
		}
		peerConfig.Env = req.Env
		newConfig, err := json.Marshal(peerConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal updated peer config: %w", err)
		}
		if _, err := s.db.UpdateNodeConfig(ctx, &db.UpdateNodeConfigParams{
			ID:         nodeID,
			NodeConfig: sql.NullString{String: string(newConfig), Valid: true},
		}); err != nil {
			return nil, fmt.Errorf("failed to update node config: %w", err)
		}

	case string(types.NodeTypeFabricOrderer):
		var ordererConfig types.FabricOrdererConfig
		if err := json.Unmarshal([]byte(dbNode.Config.String), &ordererConfig); err != nil {
			return nil, fmt.Errorf("failed to unmarshal orderer config: %w", err)
		}
		ordererConfig.Env = req.Env
		newConfig, err := json.Marshal(ordererConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal updated orderer config: %w", err)
		}
		if _, err := s.db.UpdateNodeConfig(ctx, &db.UpdateNodeConfigParams{
			ID:         nodeID,
			NodeConfig: sql.NullString{String: string(newConfig), Valid: true},
		}); err != nil {
			return nil, fmt.Errorf("failed to update node config: %w", err)
		}

	default:
		return nil, fmt.Errorf("unsupported node type: %s", dbNode.NodeType.String)
	}

	// Return the updated environment variables and indicate that a restart is required
	return &types.UpdateNodeEnvResponse{
		Env:             req.Env,
		RequiresRestart: true,
	}, nil
}

// GetNodeEnvironment retrieves the current environment variables for a node
func (s *NodeService) GetNodeEnvironment(ctx context.Context, nodeID int64) (map[string]string, error) {
	// Get the node from the database
	dbNode, err := s.db.GetNode(ctx, nodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get node: %w", err)
	}

	// Get the node's current configuration
	switch dbNode.NodeType.String {
	case string(types.NodeTypeFabricPeer):
		var peerConfig types.FabricPeerConfig
		if err := json.Unmarshal([]byte(dbNode.Config.String), &peerConfig); err != nil {
			return nil, fmt.Errorf("failed to unmarshal peer config: %w", err)
		}
		return peerConfig.Env, nil

	case string(types.NodeTypeFabricOrderer):
		var ordererConfig types.FabricOrdererConfig
		if err := json.Unmarshal([]byte(dbNode.Config.String), &ordererConfig); err != nil {
			return nil, fmt.Errorf("failed to unmarshal orderer config: %w", err)
		}
		return ordererConfig.Env, nil

	default:
		return nil, fmt.Errorf("unsupported node type: %s", dbNode.NodeType.String)
	}
}

// UpdateFabricPeer updates a Fabric peer node configuration
func (s *NodeService) UpdateFabricPeer(ctx context.Context, opts UpdateFabricPeerOpts) (*NodeResponse, error) {
	// Get the node from database
	node, err := s.db.GetNode(ctx, opts.NodeID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError("peer node not found", nil)
		}
		return nil, fmt.Errorf("failed to get peer node: %w", err)
	}

	// Verify node type
	if types.NodeType(node.NodeType.String) != types.NodeTypeFabricPeer {
		return nil, fmt.Errorf("node %d is not a Fabric peer", opts.NodeID)
	}

	// Load current config
	nodeConfig, err := utils.LoadNodeConfig([]byte(node.NodeConfig.String))
	if err != nil {
		return nil, fmt.Errorf("failed to load peer config: %w", err)
	}

	peerConfig, ok := nodeConfig.(*types.FabricPeerConfig)
	if !ok {
		return nil, fmt.Errorf("invalid peer config type")
	}

	deployConfig, err := utils.DeserializeDeploymentConfig(node.DeploymentConfig.String)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize deployment config: %w", err)
	}
	deployPeerConfig, ok := deployConfig.(*types.FabricPeerDeploymentConfig)
	if !ok {
		return nil, fmt.Errorf("invalid deployment config type")
	}

	// Update configuration fields if provided
	if opts.ExternalEndpoint != "" && opts.ExternalEndpoint != peerConfig.ExternalEndpoint {
		peerConfig.ExternalEndpoint = opts.ExternalEndpoint
	}
	if opts.ListenAddress != "" && opts.ListenAddress != peerConfig.ListenAddress {
		if err := s.validateAddress(opts.ListenAddress); err != nil {
			return nil, fmt.Errorf("invalid listen address: %w", err)
		}
		peerConfig.ListenAddress = opts.ListenAddress
	}
	if opts.EventsAddress != "" && opts.EventsAddress != peerConfig.EventsAddress {
		if err := s.validateAddress(opts.EventsAddress); err != nil {
			return nil, fmt.Errorf("invalid events address: %w", err)
		}
		peerConfig.EventsAddress = opts.EventsAddress
	}
	if opts.OperationsListenAddress != "" && opts.OperationsListenAddress != peerConfig.OperationsListenAddress {
		if err := s.validateAddress(opts.OperationsListenAddress); err != nil {
			return nil, fmt.Errorf("invalid operations listen address: %w", err)
		}
		peerConfig.OperationsListenAddress = opts.OperationsListenAddress
	}
	if opts.ChaincodeAddress != "" && opts.ChaincodeAddress != peerConfig.ChaincodeAddress {
		if err := s.validateAddress(opts.ChaincodeAddress); err != nil {
			return nil, fmt.Errorf("invalid chaincode address: %w", err)
		}
		peerConfig.ChaincodeAddress = opts.ChaincodeAddress
	}
	if opts.DomainNames != nil {
		peerConfig.DomainNames = opts.DomainNames
	}
	if opts.Env != nil {
		peerConfig.Env = opts.Env
	}
	if opts.AddressOverrides != nil {
		peerConfig.AddressOverrides = opts.AddressOverrides
		deployPeerConfig.AddressOverrides = opts.AddressOverrides
	}
	if opts.Version != "" {
		peerConfig.Version = opts.Version
		deployPeerConfig.Version = opts.Version
	}

	// Validate all addresses together for port conflicts
	if err := s.validateFabricPeerAddresses(peerConfig); err != nil {
		return nil, err
	}

	configBytes, err := utils.StoreNodeConfig(nodeConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to store node config: %w", err)
	}
	node, err = s.db.UpdateNodeConfig(ctx, &db.UpdateNodeConfigParams{
		ID: opts.NodeID,
		NodeConfig: sql.NullString{
			String: string(configBytes),
			Valid:  true,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update node config: %w", err)
	}

	// Update the deployment config in the database
	deploymentConfigBytes, err := json.Marshal(deployPeerConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal updated deployment config: %w", err)
	}

	node, err = s.db.UpdateDeploymentConfig(ctx, &db.UpdateDeploymentConfigParams{
		ID: opts.NodeID,
		DeploymentConfig: sql.NullString{
			String: string(deploymentConfigBytes),
			Valid:  true,
		},
	})

	// Synchronize the peer config
	if err := s.SynchronizePeerConfig(ctx, opts.NodeID); err != nil {
		return nil, fmt.Errorf("failed to synchronize peer config: %w", err)
	}

	// Return updated node response
	_, nodeResponse := s.mapDBNodeToServiceNode(node)
	return nodeResponse, nil
}

// UpdateFabricOrderer updates a Fabric orderer node configuration
func (s *NodeService) UpdateFabricOrderer(ctx context.Context, opts UpdateFabricOrdererOpts) (*NodeResponse, error) {
	// Get the node from database
	node, err := s.db.GetNode(ctx, opts.NodeID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError("orderer node not found", nil)
		}
		return nil, fmt.Errorf("failed to get orderer node: %w", err)
	}

	// Verify node type
	if types.NodeType(node.NodeType.String) != types.NodeTypeFabricOrderer {
		return nil, fmt.Errorf("node %d is not a Fabric orderer", opts.NodeID)
	}

	// Load current config
	nodeConfig, err := utils.LoadNodeConfig([]byte(node.NodeConfig.String))
	if err != nil {
		return nil, fmt.Errorf("failed to load orderer config: %w", err)
	}

	ordererConfig, ok := nodeConfig.(*types.FabricOrdererConfig)
	if !ok {
		return nil, fmt.Errorf("invalid orderer config type")
	}

	// Update configuration fields if provided
	if opts.ExternalEndpoint != "" {
		ordererConfig.ExternalEndpoint = opts.ExternalEndpoint
	}
	if opts.ListenAddress != "" {
		if err := s.validateAddress(opts.ListenAddress); err != nil {
			return nil, fmt.Errorf("invalid listen address: %w", err)
		}
		ordererConfig.ListenAddress = opts.ListenAddress
	}
	if opts.AdminAddress != "" {
		if err := s.validateAddress(opts.AdminAddress); err != nil {
			return nil, fmt.Errorf("invalid admin address: %w", err)
		}
		ordererConfig.AdminAddress = opts.AdminAddress
	}
	if opts.OperationsListenAddress != "" {
		if err := s.validateAddress(opts.OperationsListenAddress); err != nil {
			return nil, fmt.Errorf("invalid operations listen address: %w", err)
		}
		ordererConfig.OperationsListenAddress = opts.OperationsListenAddress
	}
	if opts.DomainNames != nil {
		ordererConfig.DomainNames = opts.DomainNames
	}
	if opts.Env != nil {
		ordererConfig.Env = opts.Env
	}
	if opts.Version != "" {
		ordererConfig.Version = opts.Version
	}
	// Validate all addresses together for port conflicts
	if err := s.validateFabricOrdererAddresses(ordererConfig); err != nil {
		return nil, err
	}

	// Update the config in the database
	configBytes, err := json.Marshal(ordererConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal updated orderer config: %w", err)
	}

	node, err = s.db.UpdateNodeConfig(ctx, &db.UpdateNodeConfigParams{
		ID: opts.NodeID,
		NodeConfig: sql.NullString{
			String: string(configBytes),
			Valid:  true,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update node config: %w", err)
	}

	// Return updated node response
	_, nodeResponse := s.mapDBNodeToServiceNode(node)
	return nodeResponse, nil
}

// SynchronizePeerConfig synchronizes the peer's configuration files and service
func (s *NodeService) SynchronizePeerConfig(ctx context.Context, nodeID int64) error {
	// Get the node from database
	node, err := s.db.GetNode(ctx, nodeID)
	if err != nil {
		return fmt.Errorf("failed to get node: %w", err)
	}

	// Verify node type
	if types.NodeType(node.NodeType.String) != types.NodeTypeFabricPeer {
		return fmt.Errorf("node %d is not a Fabric peer", nodeID)
	}

	// Load node config
	nodeConfig, err := utils.LoadNodeConfig([]byte(node.NodeConfig.String))
	if err != nil {
		return fmt.Errorf("failed to load node config: %w", err)
	}

	peerConfig, ok := nodeConfig.(*types.FabricPeerConfig)
	if !ok {
		return fmt.Errorf("invalid peer config type")
	}

	// Get organization
	org, err := s.orgService.GetOrganization(ctx, peerConfig.OrganizationID)
	if err != nil {
		return fmt.Errorf("failed to get organization: %w", err)
	}

	// Get local peer instance
	localPeer := s.getPeerFromConfig(node, org, peerConfig)

	// Get deployment config
	deploymentConfig, err := utils.DeserializeDeploymentConfig(node.DeploymentConfig.String)
	if err != nil {
		return fmt.Errorf("failed to deserialize deployment config: %w", err)
	}

	peerDeployConfig, ok := deploymentConfig.(*types.FabricPeerDeploymentConfig)
	if !ok {
		return fmt.Errorf("invalid peer deployment config type")
	}

	// Synchronize configuration
	if err := localPeer.SynchronizeConfig(peerDeployConfig); err != nil {
		return fmt.Errorf("failed to synchronize peer config: %w", err)
	}

	return nil
}
