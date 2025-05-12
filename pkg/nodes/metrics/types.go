package nodemetrics

import (
	"context"

	"github.com/chainlaunch/chainlaunch/pkg/nodes/service"
)

// Service defines the interface for node metrics operations
type Service interface {
	// OnNodeCreated handles the creation of a new node
	OnNodeCreated(ctx context.Context, node *service.Node) error
	// OnNodeUpdated handles the update of an existing node
	OnNodeUpdated(ctx context.Context, node *service.Node) error
	// OnNodeDeleted handles the deletion of a node
	OnNodeDeleted(ctx context.Context, nodeID int64) error
}
