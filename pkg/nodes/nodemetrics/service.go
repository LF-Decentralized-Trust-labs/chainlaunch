package nodemetrics

import (
	"context"
	"fmt"

	"github.com/chainlaunch/chainlaunch/pkg/logger"
	"github.com/chainlaunch/chainlaunch/pkg/metrics"
	"github.com/chainlaunch/chainlaunch/pkg/nodes/events"
	"github.com/chainlaunch/chainlaunch/pkg/nodes/eventtypes"
	"github.com/chainlaunch/chainlaunch/pkg/nodes/nodetypes"
	"github.com/chainlaunch/chainlaunch/pkg/nodes/types"
)

// Service handles node metrics operations
type Service struct {
	logger         *logger.Logger
	metricsService *metrics.Service
}

// NewService creates a new Service instance
func NewService(
	logger *logger.Logger,
	metricsService *metrics.Service,
	eventBus *events.EventBus,
) *Service {
	s := &Service{
		logger:         logger,
		metricsService: metricsService,
	}

	// Subscribe to node events
	if eventBus != nil {
		eventBus.Subscribe(s.handleNodeEvent)
	}

	return s
}

// handleNodeEvent handles node events
func (s *Service) handleNodeEvent(event eventtypes.Event) {
	ctx := context.Background()

	switch event.Type {
	case eventtypes.EventTypeNodeCreated:
		if event.Node != nil {
			if err := s.OnNodeCreated(ctx, event.Node); err != nil {
				s.logger.Error("failed to handle node created event", "error", err)
			}
		}
	case eventtypes.EventTypeNodeUpdated:
		if event.Node != nil {
			if err := s.OnNodeUpdated(ctx, event.Node); err != nil {
				s.logger.Error("failed to handle node updated event", "error", err)
			}
		}
	case eventtypes.EventTypeNodeDeleted:
		if err := s.OnNodeDeleted(ctx, event.NodeID); err != nil {
			s.logger.Error("failed to handle node deleted event", "error", err)
		}
	}
}

// OnNodeCreated is called when a node is created
func (s *Service) OnNodeCreated(ctx context.Context, node *nodetypes.Node) error {
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
	if err := s.metricsService.Reload(ctx); err != nil {
		return fmt.Errorf("failed to reload metrics configuration: %w", err)
	}

	return nil
}

// OnNodeUpdated is called when a node is updated
func (s *Service) OnNodeUpdated(ctx context.Context, node *nodetypes.Node) error {
	// Reload Prometheus configuration to reflect the changes
	if err := s.metricsService.Reload(ctx); err != nil {
		return fmt.Errorf("failed to reload metrics configuration: %w", err)
	}

	return nil
}

// OnNodeDeleted is called when a node is deleted
func (s *Service) OnNodeDeleted(ctx context.Context, nodeID int64) error {
	// Reload Prometheus configuration to remove the deleted node
	if err := s.metricsService.Reload(ctx); err != nil {
		return fmt.Errorf("failed to reload metrics configuration: %w", err)
	}

	return nil
}
