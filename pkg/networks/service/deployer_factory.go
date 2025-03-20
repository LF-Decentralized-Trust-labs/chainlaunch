package service

import (
	"fmt"

	"github.com/chainlaunch/chainlaunch/pkg/db"
	orgservicefabric "github.com/chainlaunch/chainlaunch/pkg/fabric/service"
	keymanagement "github.com/chainlaunch/chainlaunch/pkg/keymanagement/service"
	"github.com/chainlaunch/chainlaunch/pkg/networks/service/besu"
	"github.com/chainlaunch/chainlaunch/pkg/networks/service/fabric"
	"github.com/chainlaunch/chainlaunch/pkg/networks/service/types"
	nodeservice "github.com/chainlaunch/chainlaunch/pkg/nodes/service"
)

// DeployerFactory creates network deployers based on blockchain platform
type DeployerFactory struct {
	db         *db.Queries
	nodes      *nodeservice.NodeService
	keyMgmt    *keymanagement.KeyManagementService
	orgService *orgservicefabric.OrganizationService
}

// NewDeployerFactory creates a new deployer factory
func NewDeployerFactory(db *db.Queries, nodes *nodeservice.NodeService, keyMgmt *keymanagement.KeyManagementService, orgService *orgservicefabric.OrganizationService) *DeployerFactory {
	return &DeployerFactory{
		db:         db,
		nodes:      nodes,
		keyMgmt:    keyMgmt,
		orgService: orgService,
	}
}

// GetDeployer returns a deployer for the specified blockchain platform
func (f *DeployerFactory) GetDeployer(platform string) (types.NetworkDeployer, error) {
	switch platform {
	case string(BlockchainTypeFabric):
		return fabric.NewFabricDeployer(f.db, f.nodes, f.keyMgmt, f.orgService), nil
	case string(BlockchainTypeBesu):
		return besu.NewBesuDeployer(f.db, f.nodes, f.keyMgmt), nil
	default:
		return nil, fmt.Errorf("unsupported blockchain platform: %s", platform)
	}
}
