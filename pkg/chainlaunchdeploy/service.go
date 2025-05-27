package chainlaunchdeploy

import (
	"context"
	"database/sql"

	"github.com/chainlaunch/chainlaunch/pkg/db"
)

type ChaincodeService struct {
	db *db.Queries
}

func NewChaincodeService(dbq *db.Queries) *ChaincodeService {
	return &ChaincodeService{db: dbq}
}

func (s *ChaincodeService) ListChaincodes(ctx context.Context) ([]*db.FabricChaincode, error) {
	return s.db.ListFabricChaincodes(ctx)
}

func (s *ChaincodeService) GetChaincodeBySlug(ctx context.Context, slug string) (*db.FabricChaincode, error) {
	return s.db.GetFabricChaincodeBySlug(ctx, slug)
}

func (s *ChaincodeService) InsertChaincode(ctx context.Context, name, slug, packageID, dockerImage, hostPort, containerPort, status string) (*db.FabricChaincode, error) {
	return s.db.InsertFabricChaincode(ctx, &db.InsertFabricChaincodeParams{
		Name:          name,
		Slug:          slug,
		PackageID:     packageID,
		DockerImage:   dockerImage,
		HostPort:      sql.NullString{String: hostPort, Valid: hostPort != ""},
		ContainerPort: sql.NullString{String: containerPort, Valid: containerPort != ""},
		Status:        status,
	})
}

func (s *ChaincodeService) UpdateChaincodeBySlug(ctx context.Context, slug, dockerImage, packageID, hostPort, containerPort, status string) (*db.FabricChaincode, error) {
	return s.db.UpdateFabricChaincodeBySlug(ctx, &db.UpdateFabricChaincodeBySlugParams{
		DockerImage:   dockerImage,
		PackageID:     packageID,
		HostPort:      sql.NullString{String: hostPort, Valid: hostPort != ""},
		ContainerPort: sql.NullString{String: containerPort, Valid: containerPort != ""},
		Status:        status,
		Slug:          slug,
	})
}
