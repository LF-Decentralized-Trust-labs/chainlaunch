package projects

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/chainlaunch/chainlaunch/pkg/db"
	"go.uber.org/zap"
)

// FabricLifecycle implements PlatformLifecycle for Hyperledger Fabric
type FabricLifecycle struct {
	queries *db.Queries
	logger  *zap.Logger
}

// NewFabricLifecycle creates a new FabricLifecycle instance
func NewFabricLifecycle(queries *db.Queries, logger *zap.Logger) *FabricLifecycle {
	return &FabricLifecycle{
		queries: queries,
		logger:  logger,
	}
}

// PreStart is called before starting the project container
func (f *FabricLifecycle) PreStart(ctx context.Context, params PreStartParams) error {
	f.logger.Info("PreStart hook for Fabric project",
		zap.Int64("projectID", params.ProjectID),
		zap.String("projectName", params.ProjectName),
		zap.String("boilerplate", params.Boilerplate),
	)

	// Validate that the project is associated with a Fabric network
	if params.Platform != "fabric" {
		return fmt.Errorf("project is not associated with a Fabric network")
	}

	// Get network details
	network, err := f.queries.GetNetwork(ctx, params.NetworkID)
	if err != nil {
		return fmt.Errorf("failed to get network details: %w", err)
	}

	// Validate network is ready
	if network.Status != "running" {
		return fmt.Errorf("network is not running: %s", network.Status)
	}

	return nil
}

// PostStart is called after the project container has started
func (f *FabricLifecycle) PostStart(ctx context.Context, params PostStartParams) error {
	f.logger.Info("PostStart hook for Fabric project",
		zap.Int64("projectID", params.ProjectID),
		zap.String("projectName", params.ProjectName),
		zap.String("containerID", params.ContainerID),
	)

	// Get network details
	network, err := f.queries.GetNetwork(ctx, params.NetworkID)
	if err != nil {
		return fmt.Errorf("failed to get network details: %w", err)
	}

	// TODO: Implement chaincode installation and approval
	// This will involve:
	// 1. Getting the chaincode package from the container
	// 2. Installing it on the peers
	// 3. Approving the chaincode definition
	// 4. Committing the chaincode definition
	_ = network // TODO: Use network details for chaincode installation

	f.logger.Info("Chaincode setup completed",
		zap.Int64("projectID", params.ProjectID),
		zap.String("projectName", params.ProjectName),
	)

	return nil
}

// PreStop is called before stopping the project container
func (f *FabricLifecycle) PreStop(ctx context.Context, params PreStopParams) error {
	f.logger.Info("PreStop hook for Fabric project",
		zap.Int64("projectID", params.ProjectID),
		zap.String("projectName", params.ProjectName),
		zap.String("containerID", params.ContainerID),
	)

	// TODO: Implement any necessary cleanup before stopping
	// This might include:
	// 1. Saving chaincode state
	// 2. Cleaning up temporary files
	// 3. Updating project status

	return nil
}

// PostStop is called after the project container has stopped
func (f *FabricLifecycle) PostStop(ctx context.Context, params PostStopParams) error {
	f.logger.Info("PostStop hook for Fabric project",
		zap.Int64("projectID", params.ProjectID),
		zap.String("projectName", params.ProjectName),
		zap.String("containerID", params.ContainerID),
	)

	// Update project status in database
	err := f.queries.UpdateProjectContainerInfo(ctx, &db.UpdateProjectContainerInfoParams{
		ID:            params.ProjectID,
		Status:        sql.NullString{String: "stopped", Valid: true},
		LastStoppedAt: sql.NullTime{Time: params.StoppedAt, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("failed to update project status: %w", err)
	}

	return nil
}
