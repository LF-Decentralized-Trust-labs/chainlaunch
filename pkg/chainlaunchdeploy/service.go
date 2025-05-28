package chainlaunchdeploy

import (
	"context"
	"database/sql"
	"fmt"
	"io"

	"github.com/chainlaunch/chainlaunch/pkg/db"
	"github.com/chainlaunch/chainlaunch/pkg/nodes/service"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

// --- Service-layer structs ---
type Chaincode struct {
	ID          int64
	Name        string
	NetworkID   int64
	CreatedAt   string // ISO8601
	Definitions []ChaincodeDefinition
}

type ChaincodeDefinition struct {
	ID                int64
	ChaincodeID       int64
	Version           string
	Sequence          int64
	DockerImage       string
	EndorsementPolicy string
	CreatedAt         string // ISO8601
	PeerStatuses      []PeerStatus
}

type PeerStatus struct {
	ID           int64
	DefinitionID int64
	PeerID       int64
	Status       string
	LastUpdated  string // ISO8601
}

type ChaincodeService struct {
	db           *db.Queries
	nodesService *service.NodeService
}

func NewChaincodeService(dbq *db.Queries) *ChaincodeService {
	return &ChaincodeService{db: dbq}
}

// --- Chaincode CRUD ---
func (s *ChaincodeService) CreateChaincode(ctx context.Context, name string, networkID int64) (*Chaincode, error) {
	cc, err := s.db.CreateChaincode(ctx, &db.CreateChaincodeParams{
		Name:      name,
		NetworkID: networkID,
	})
	if err != nil {
		return nil, err
	}
	return &Chaincode{
		ID:        cc.ID,
		Name:      cc.Name,
		NetworkID: cc.NetworkID,
		CreatedAt: nullTimeToString(cc.CreatedAt),
	}, nil
}

func (s *ChaincodeService) ListChaincodes(ctx context.Context) ([]*Chaincode, error) {
	dbChaincodes, err := s.db.ListChaincodes(ctx)
	if err != nil {
		return nil, err
	}
	var result []*Chaincode
	for _, cc := range dbChaincodes {
		result = append(result, &Chaincode{
			ID:        cc.ID,
			Name:      cc.Name,
			NetworkID: cc.NetworkID,
			CreatedAt: nullTimeToString(cc.CreatedAt),
		})
	}
	return result, nil
}

func (s *ChaincodeService) GetChaincode(ctx context.Context, id int64) (*Chaincode, error) {
	cc, err := s.db.GetChaincode(ctx, id)
	if err != nil {
		return nil, err
	}
	return &Chaincode{
		ID:        cc.ID,
		Name:      cc.Name,
		NetworkID: cc.NetworkID,
		CreatedAt: nullTimeToString(cc.CreatedAt),
	}, nil
}

func (s *ChaincodeService) UpdateChaincode(ctx context.Context, id int64, name string, networkID int64) (*Chaincode, error) {
	cc, err := s.db.UpdateChaincode(ctx, &db.UpdateChaincodeParams{
		ID:        id,
		Name:      name,
		NetworkID: networkID,
	})
	if err != nil {
		return nil, err
	}
	return &Chaincode{
		ID:        cc.ID,
		Name:      cc.Name,
		NetworkID: cc.NetworkID,
		CreatedAt: nullTimeToString(cc.CreatedAt),
	}, nil
}

func (s *ChaincodeService) DeleteChaincode(ctx context.Context, id int64) error {
	return s.db.DeleteChaincode(ctx, id)
}

// --- ChaincodeDefinition CRUD ---
func (s *ChaincodeService) CreateChaincodeDefinition(ctx context.Context, chaincodeID int64, version string, sequence int64, dockerImage, endorsementPolicy string) (*ChaincodeDefinition, error) {
	def, err := s.db.CreateChaincodeDefinition(ctx, &db.CreateChaincodeDefinitionParams{
		ChaincodeID:       chaincodeID,
		Version:           version,
		Sequence:          sequence,
		DockerImage:       dockerImage,
		EndorsementPolicy: sql.NullString{String: endorsementPolicy, Valid: endorsementPolicy != ""},
	})
	if err != nil {
		return nil, err
	}
	return &ChaincodeDefinition{
		ID:                def.ID,
		ChaincodeID:       def.ChaincodeID,
		Version:           def.Version,
		Sequence:          def.Sequence,
		DockerImage:       def.DockerImage,
		EndorsementPolicy: nullStringToString(def.EndorsementPolicy),
		CreatedAt:         nullTimeToString(def.CreatedAt),
	}, nil
}

func (s *ChaincodeService) ListChaincodeDefinitions(ctx context.Context, chaincodeID int64) ([]*ChaincodeDefinition, error) {
	defs, err := s.db.ListChaincodeDefinitions(ctx, chaincodeID)
	if err != nil {
		return nil, err
	}
	var result []*ChaincodeDefinition
	for _, def := range defs {
		result = append(result, &ChaincodeDefinition{
			ID:                def.ID,
			ChaincodeID:       def.ChaincodeID,
			Version:           def.Version,
			Sequence:          def.Sequence,
			DockerImage:       def.DockerImage,
			EndorsementPolicy: nullStringToString(def.EndorsementPolicy),
			CreatedAt:         nullTimeToString(def.CreatedAt),
		})
	}
	return result, nil
}

func (s *ChaincodeService) GetChaincodeDefinition(ctx context.Context, id int64) (*ChaincodeDefinition, error) {
	def, err := s.db.GetChaincodeDefinition(ctx, id)
	if err != nil {
		return nil, err
	}
	return &ChaincodeDefinition{
		ID:                def.ID,
		ChaincodeID:       def.ChaincodeID,
		Version:           def.Version,
		Sequence:          def.Sequence,
		DockerImage:       def.DockerImage,
		EndorsementPolicy: nullStringToString(def.EndorsementPolicy),
		CreatedAt:         nullTimeToString(def.CreatedAt),
	}, nil
}

func (s *ChaincodeService) UpdateChaincodeDefinition(ctx context.Context, id int64, version string, sequence int64, dockerImage, endorsementPolicy string) (*ChaincodeDefinition, error) {
	def, err := s.db.UpdateChaincodeDefinition(ctx, &db.UpdateChaincodeDefinitionParams{
		ID:                id,
		Version:           version,
		Sequence:          sequence,
		DockerImage:       dockerImage,
		EndorsementPolicy: sql.NullString{String: endorsementPolicy, Valid: endorsementPolicy != ""},
	})
	if err != nil {
		return nil, err
	}
	return &ChaincodeDefinition{
		ID:                def.ID,
		ChaincodeID:       def.ChaincodeID,
		Version:           def.Version,
		Sequence:          def.Sequence,
		DockerImage:       def.DockerImage,
		EndorsementPolicy: nullStringToString(def.EndorsementPolicy),
		CreatedAt:         nullTimeToString(def.CreatedAt),
	}, nil
}

func (s *ChaincodeService) DeleteChaincodeDefinition(ctx context.Context, id int64) error {
	return s.db.DeleteChaincodeDefinition(ctx, id)
}

// --- PeerStatus operations ---
func (s *ChaincodeService) SetPeerStatus(ctx context.Context, definitionID, peerID int64, status string) (*PeerStatus, error) {
	ps, err := s.db.SetPeerStatus(ctx, &db.SetPeerStatusParams{
		DefinitionID: definitionID,
		PeerID:       peerID,
		Status:       status,
	})
	if err != nil {
		return nil, err
	}
	return &PeerStatus{
		ID:           ps.ID,
		DefinitionID: ps.DefinitionID,
		PeerID:       ps.PeerID,
		Status:       ps.Status,
		LastUpdated:  nullTimeToString(ps.LastUpdated),
	}, nil
}

func (s *ChaincodeService) ListPeerStatuses(ctx context.Context, definitionID int64) ([]*PeerStatus, error) {
	pss, err := s.db.ListPeerStatuses(ctx, definitionID)
	if err != nil {
		return nil, err
	}
	var result []*PeerStatus
	for _, ps := range pss {
		result = append(result, &PeerStatus{
			ID:           ps.ID,
			DefinitionID: ps.DefinitionID,
			PeerID:       ps.PeerID,
			Status:       ps.Status,
			LastUpdated:  nullTimeToString(ps.LastUpdated),
		})
	}
	return result, nil
}

// --- Utility functions for sql.NullTime and sql.NullString ---
func nullTimeToString(nt sql.NullTime) string {
	if nt.Valid {
		return nt.Time.Format("2006-01-02T15:04:05Z07:00")
	}
	return ""
}

func nullStringToString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

// --- Docker utility functions (as before, using unified types from fabric.go) ---
// (Assume DockerContainerInfo and FabricChaincodeDetail are imported from fabric.go)

func ListChaincodesWithDockerInfo(ctx context.Context, chaincodes []*Chaincode) ([]FabricChaincodeDetail, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	defer cli.Close()
	var details []FabricChaincodeDetail
	for _, cc := range chaincodes {
		if cc == nil {
			continue
		}
		containerName := fmt.Sprintf("chaincode-%d", cc.ID) // Example: use ID for unique container name
		containers, err := cli.ContainerList(ctx, container.ListOptions{All: true})
		var dockerInfo *DockerContainerInfo
		if err == nil {
			for _, c := range containers {
				for _, name := range c.Names {
					if name == containerName {
						ports := []string{}
						for _, p := range c.Ports {
							ports = append(ports, fmt.Sprintf("%s:%d->%d/%s", p.IP, p.PublicPort, p.PrivatePort, p.Type))
						}
						dockerInfo = &DockerContainerInfo{
							ID:      c.ID,
							Name:    name,
							Image:   c.Image,
							State:   c.State,
							Status:  c.Status,
							Ports:   ports,
							Created: c.Created,
						}
						break
					}
				}
				if dockerInfo != nil {
					break
				}
			}
		}
		details = append(details, FabricChaincodeDetail{
			Chaincode:  cc,
			DockerInfo: dockerInfo,
		})
	}
	return details, nil
}

func GetChaincodeWithDockerInfo(ctx context.Context, cc *Chaincode) (FabricChaincodeDetail, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return FabricChaincodeDetail{}, err
	}
	defer cli.Close()
	containerName := fmt.Sprintf("chaincode-%d", cc.ID)
	containers, err := cli.ContainerList(ctx, container.ListOptions{All: true})
	var dockerInfo *DockerContainerInfo
	if err == nil {
		for _, c := range containers {
			for _, name := range c.Names {
				if name == containerName {
					ports := []string{}
					for _, p := range c.Ports {
						ports = append(ports, fmt.Sprintf("%s:%d->%d/%s", p.IP, p.PublicPort, p.PrivatePort, p.Type))
					}
					dockerInfo = &DockerContainerInfo{
						ID:      c.ID,
						Name:    name,
						Image:   c.Image,
						State:   c.State,
						Status:  c.Status,
						Ports:   ports,
						Created: c.Created,
					}
					break
				}
			}
			if dockerInfo != nil {
				break
			}
		}
	}
	return FabricChaincodeDetail{
		Chaincode:  cc,
		DockerInfo: dockerInfo,
	}, nil
}

// Add any other Docker utility functions as needed

// InstallChaincodeByDefinition installs a chaincode definition on the given peers
func (s *ChaincodeService) InstallChaincodeByDefinition(ctx context.Context, definitionID int64, peerIDs []int64) error {
	// peer, err := s.nodesService.GetPeer(ctx, peerIDs[0])
	// if err != nil {
	// 	return err
	// }
	// peer.ChaincodeID = definitionID
	for _, peerID := range peerIDs {
		peerService, err := s.nodesService.GetFabricPeerService(ctx, peer)
		if err != nil {
			return err
		}
		peerService.Install(ctx, io.Reader)
	}

	return nil
}

// ApproveChaincodeByDefinition approves a chaincode definition using the given peer
func (s *ChaincodeService) ApproveChaincodeByDefinition(ctx context.Context, definitionID int64, peerID int64) error {
	// TODO: Implement actual Fabric approve logic for the given definition and peer using fabric-admin-sdk
	return nil
}

// CommitChaincodeByDefinition commits a chaincode definition using the given peer
func (s *ChaincodeService) CommitChaincodeByDefinition(ctx context.Context, definitionID int64, peerID int64) error {
	// TODO: Implement actual Fabric commit logic for the given definition and peer using fabric-admin-sdk
	return nil
}

// DeployChaincodeByDefinition deploys a chaincode definition using Docker image
func (s *ChaincodeService) DeployChaincodeByDefinition(ctx context.Context, definitionID int64, hostPort, containerPort string) error {
	// TODO: Implement Docker deploy logic using the definition's Docker image (see fabric.go for reference)
	return nil
}
