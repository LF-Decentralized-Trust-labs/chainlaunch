package metrics

import (
	"context"
	"fmt"

	"github.com/chainlaunch/chainlaunch/pkg/logger"
	"github.com/chainlaunch/chainlaunch/pkg/metrics"
	nodeservice "github.com/chainlaunch/chainlaunch/pkg/nodes/service"
	"github.com/chainlaunch/chainlaunch/pkg/nodes/types"
)

// Service defines the interface for node metrics operations
type Service interface {
	// OnNodeCreated handles the creation of a new node
	OnNodeCreated(ctx context.Context, node *nodeservice.Node) error
	// OnNodeUpdated handles the update of an existing node
	OnNodeUpdated(ctx context.Context, node *nodeservice.Node) error
	// OnNodeDeleted handles the deletion of a node
	OnNodeDeleted(ctx context.Context, nodeID int64) error
}

// service implements the Service interface
type service struct {
	logger      *logger.Logger
	metricsSvc  metrics.Service
	nodeService *nodeservice.NodeService
}

// NewService creates a new node metrics service
func NewService(
	logger *logger.Logger,
	metricsSvc metrics.Service,
	nodeService *nodeservice.NodeService,
) Service {
	return &service{
		logger:      logger,
		metricsSvc:  metricsSvc,
		nodeService: nodeService,
	}
}

// OnNodeCreated handles the creation of a new node
func (s *service) OnNodeCreated(ctx context.Context, node *nodeservice.Node) error {
	// Check if the node has metrics enabled
	if node.NodeConfig == nil {
		return nil
	}

	// Handle different node types
	switch node.NodeType {
	case types.NodeTypeBesuFullnode:
		// For Besu nodes, check if metrics are enabled
		if besuConfig, ok := node.NodeConfig.(*types.BesuNodeConfig); ok {
			if !besuConfig.MetricsEnabled {
				return nil
			}
		}
	case types.NodeTypeFabricPeer, types.NodeTypeFabricOrderer:
		// For Fabric nodes, we always want to monitor them
		break
	default:
		return nil
	}

	// Reload Prometheus configuration to include the new node
	if err := s.metricsSvc.Reload(ctx); err != nil {
		return fmt.Errorf("failed to reload metrics configuration: %w", err)
	}

	return nil
}

// OnNodeUpdated handles the update of an existing node
func (s *service) OnNodeUpdated(ctx context.Context, node *nodeservice.Node) error {
	// Reload Prometheus configuration to reflect the changes
	if err := s.metricsSvc.Reload(ctx); err != nil {
		return fmt.Errorf("failed to reload metrics configuration: %w", err)
	}

	return nil
}

// OnNodeDeleted handles the deletion of a node
func (s *service) OnNodeDeleted(ctx context.Context, nodeID int64) error {
	// Reload Prometheus configuration to remove the deleted node
	if err := s.metricsSvc.Reload(ctx); err != nil {
		return fmt.Errorf("failed to reload metrics configuration: %w", err)
	}

	return nil
}
