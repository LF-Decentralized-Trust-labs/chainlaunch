package chainlaunchdeploy

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/chainlaunch/chainlaunch/pkg/db"
	"github.com/chainlaunch/chainlaunch/pkg/logger"
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
	logger       *logger.Logger
}

func NewChaincodeService(dbq *db.Queries, logger *logger.Logger) *ChaincodeService {
	return &ChaincodeService{db: dbq, logger: logger}
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
	definition, err := s.GetChaincodeDefinition(ctx, definitionID)
	if err != nil {
		return err
	}
	label := definition.DockerImage
	chaincodeAddress := fmt.Sprintf("localhost:%d", 17056)
	codeTarGz, err := s.getCodeTarGz(chaincodeAddress, "", "", "", "")
	if err != nil {
		return err
	}
	pkg, err := s.getChaincodePackage(label, codeTarGz)
	if err != nil {
		return err
	}
	for _, peerID := range peerIDs {
		peerService, err := s.nodesService.GetFabricPeerService(ctx, peerID)
		if err != nil {
			return err
		}
		res, err := peerService.Install(ctx, bytes.NewReader(pkg))
		if err != nil {
			return err
		}
		s.logger.Debugf("Install result: %+v", res)
	}

	return nil
}

func (s *ChaincodeService) getCodeTarGz(
	address string,
	rootCert string,
	clientKey string,
	clientCert string,
	metaInfPath string,
) ([]byte, error) {
	var err error
	// Determine if TLS is required based on certificate presence
	tlsRequired := rootCert != ""
	clientAuthRequired := clientCert != "" && clientKey != ""

	// Read certificate files if provided
	var rootCertContent, clientKeyContent, clientCertContent string
	if tlsRequired {
		rootCertBytes, err := os.ReadFile(rootCert)
		if err != nil {
			return nil, fmt.Errorf("failed to read root certificate: %w", err)
		}
		rootCertContent = string(rootCertBytes)
	}

	if clientAuthRequired {
		clientKeyBytes, err := os.ReadFile(clientKey)
		if err != nil {
			return nil, fmt.Errorf("failed to read client key: %w", err)
		}
		clientKeyContent = string(clientKeyBytes)

		clientCertBytes, err := os.ReadFile(clientCert)
		if err != nil {
			return nil, fmt.Errorf("failed to read client certificate: %w", err)
		}
		clientCertContent = string(clientCertBytes)
	}

	connMap := map[string]interface{}{
		"address":              address,
		"dial_timeout":         "10s",
		"tls_required":         tlsRequired,
		"root_cert":            rootCertContent,
		"client_auth_required": clientAuthRequired,
		"client_key":           clientKeyContent,
		"client_cert":          clientCertContent,
	}
	connJsonBytes, err := json.Marshal(connMap)
	if err != nil {
		return nil, err
	}
	s.logger.Debugf("Conn=%s", string(connJsonBytes))
	// set up the output file
	buf := &bytes.Buffer{}
	// set up the gzip writer
	gw := gzip.NewWriter(buf)
	tw := tar.NewWriter(gw)
	header := new(tar.Header)
	header.Name = "connection.json"
	header.Size = int64(len(connJsonBytes))
	header.Mode = 0755
	err = tw.WriteHeader(header)
	if err != nil {
		return nil, err
	}
	r := bytes.NewReader(connJsonBytes)
	_, err = io.Copy(tw, r)
	if err != nil {
		return nil, err
	}
	if metaInfPath != "" {
		src := metaInfPath
		// walk through 3 file in the folder
		err = filepath.Walk(src, func(file string, fi os.FileInfo, err error) error {
			// generate tar header
			header, err := tar.FileInfoHeader(fi, file)
			if err != nil {
				return err
			}

			// must provide real name
			// (see https://golang.org/src/archive/tar/common.go?#L626)
			relname, err := filepath.Rel(src, file)
			if err != nil {
				return err
			}
			if relname == "." {
				return nil
			}
			header.Name = "META-INF/" + filepath.ToSlash(relname)

			// write header
			if err := tw.WriteHeader(header); err != nil {
				return err
			}
			// if not a dir, write file content
			if !fi.IsDir() {
				data, err := os.Open(file)
				if err != nil {
					return err
				}
				if _, err := io.Copy(tw, data); err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	err = tw.Close()
	if err != nil {
		return nil, err
	}
	err = gw.Close()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (s *ChaincodeService) getChaincodePackage(label string, codeTarGz []byte) ([]byte, error) {
	var err error
	metadataJson := fmt.Sprintf(`
{
  "type": "ccaas",
  "label": "%s"
}
`, label)
	// set up the output file
	buf := &bytes.Buffer{}

	// set up the gzip writer
	gw := gzip.NewWriter(buf)
	defer func(gw *gzip.Writer) {
		err := gw.Close()
		if err != nil {
			s.logger.Warnf("gzip.Writer.Close() failed: %s", err)
		}
	}(gw)
	tw := tar.NewWriter(gw)
	defer func(tw *tar.Writer) {
		err := tw.Close()
		if err != nil {
			s.logger.Warnf("tar.Writer.Close() failed: %s", err)
		}
	}(tw)
	header := new(tar.Header)
	header.Name = "metadata.json"
	metadataJsonBytes := []byte(metadataJson)
	header.Size = int64(len(metadataJsonBytes))
	header.Mode = 0777
	err = tw.WriteHeader(header)
	if err != nil {
		return nil, err
	}
	r := bytes.NewReader(metadataJsonBytes)
	_, err = io.Copy(tw, r)
	if err != nil {
		return nil, err
	}
	headerCode := new(tar.Header)
	headerCode.Name = "code.tar.gz"
	headerCode.Size = int64(len(codeTarGz))
	headerCode.Mode = 0777
	err = tw.WriteHeader(headerCode)
	if err != nil {
		return nil, err
	}
	r = bytes.NewReader(codeTarGz)
	_, err = io.Copy(tw, r)
	if err != nil {
		return nil, err
	}
	err = tw.Close()
	if err != nil {
		return nil, err
	}
	err = gw.Close()
	if err != nil {
		s.logger.Warnf("gzip.Writer.Close() failed: %s", err)
		return nil, err
	}
	return buf.Bytes(), nil
}

// ApproveChaincodeByDefinition approves a chaincode definition using the given peer
func (s *ChaincodeService) ApproveChaincodeByDefinition(ctx context.Context, definitionID int64, peerID int64) error {
	peerGateway, err := s.nodesService.GetFabricPeerGateway(ctx, peerID)
	if err != nil {
		return err
	}
	res, err := peerGateway.Approve(ctx, bytes.NewReader(pkg))
	if err != nil {
		return err
	}
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
