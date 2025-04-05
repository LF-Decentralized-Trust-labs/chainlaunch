package org

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"

	"github.com/chainlaunch/chainlaunch/internal/protoutil"
	"github.com/chainlaunch/chainlaunch/pkg/fabric/service"
	keymanagement "github.com/chainlaunch/chainlaunch/pkg/keymanagement/service"
	"github.com/chainlaunch/chainlaunch/pkg/logger"
	"github.com/hyperledger/fabric-admin-sdk/pkg/channel"
	"github.com/hyperledger/fabric-admin-sdk/pkg/identity"
	"github.com/hyperledger/fabric-admin-sdk/pkg/network"
	gwidentity "github.com/hyperledger/fabric-gateway/pkg/identity"
	cb "github.com/hyperledger/fabric-protos-go-apiv2/common"
)

type FabricOrg struct {
	orgService     *service.OrganizationService
	keyMgmtService *keymanagement.KeyManagementService
	logger         *logger.Logger
	mspID          string
}

func NewOrganizationService(
	orgService *service.OrganizationService,
	keyMgmtService *keymanagement.KeyManagementService,
	logger *logger.Logger,
	mspID string,
) *FabricOrg {
	return &FabricOrg{
		orgService:     orgService,
		keyMgmtService: keyMgmtService,
		logger:         logger,
		mspID:          mspID,
	}
}

// GetConfigBlockWithNetworkConfig retrieves a config block using a generated network config
func (s *FabricOrg) GetConfigBlockWithNetworkConfig(ctx context.Context, channelID, ordererURL, ordererTLSCert string) (*cb.Block, error) {
	s.logger.Info("Fetching channel config with network config",
		"mspID", s.mspID,
		"channel", channelID,
		"ordererUrl", ordererURL,
	)
	ordererNode := network.Node{
		Addr:          ordererURL,
		TLSCACertByte: []byte(ordererTLSCert),
	}
	ordererConn, err := network.DialConnection(ordererNode)
	if err != nil {
		return nil, fmt.Errorf("failed to dial orderer: %w", err)
	}
	defer ordererConn.Close()
	// Get organization details
	org, err := s.orgService.GetOrganizationByMspID(ctx, s.mspID)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	// Get signing key
	if !org.AdminSignKeyID.Valid {
		return nil, fmt.Errorf("organization has no signing key")
	}

	// Get signing key
	var privateKeyPEM string
	if !org.AdminSignKeyID.Valid {
		return nil, fmt.Errorf("organization has no admin sign key")
	}
	adminSignKey, err := s.keyMgmtService.GetKey(ctx, int(org.AdminSignKeyID.Int64))
	if err != nil {
		return nil, fmt.Errorf("failed to get admin sign key: %w", err)
	}
	if adminSignKey.Certificate == nil {
		return nil, fmt.Errorf("admin sign key has no certificate")
	}
	// Get private key from key management service
	privateKeyPEM, err = s.keyMgmtService.GetDecryptedPrivateKey(int(org.AdminSignKeyID.Int64))
	if err != nil {
		return nil, fmt.Errorf("failed to get private key: %w", err)
	}

	cert, err := gwidentity.CertificateFromPEM([]byte(*adminSignKey.Certificate))
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate: %w", err)
	}

	priv, err := gwidentity.PrivateKeyFromPEM([]byte(privateKeyPEM))
	if err != nil {
		return nil, fmt.Errorf("failed to read private key: %w", err)
	}

	ordererMSP, err := identity.NewPrivateKeySigningIdentity(s.mspID, cert, priv)
	if err != nil {
		return nil, fmt.Errorf("failed to create orderer msp: %w", err)
	}
	// Parse the orderer TLS certificate
	ordererTLSCertParsed, err := tls.X509KeyPair([]byte(*adminSignKey.Certificate), []byte(privateKeyPEM))
	if err != nil {
		return nil, fmt.Errorf("failed to parse orderer TLS certificate: %w", err)
	}

	ordererBlock, err := channel.GetConfigBlockFromOrderer(ctx, ordererConn, ordererMSP, channelID, ordererTLSCertParsed)
	if err != nil {
		return nil, fmt.Errorf("failed to get config block from orderer: %w", err)
	}

	return ordererBlock, nil
}

// getAdminIdentity retrieves the admin identity for the organization
func (s *FabricOrg) getAdminIdentity(ctx context.Context) (identity.SigningIdentity, error) {
	// Get organization details
	org, err := s.orgService.GetOrganizationByMspID(ctx, s.mspID)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	if !org.AdminSignKeyID.Valid {
		return nil, fmt.Errorf("organization has no signing key")
	}

	// Get admin signing key
	adminSignKey, err := s.keyMgmtService.GetKey(ctx, int(org.AdminSignKeyID.Int64))
	if err != nil {
		return nil, fmt.Errorf("failed to get admin sign key: %w", err)
	}
	if adminSignKey.Certificate == nil {
		return nil, fmt.Errorf("admin sign key has no certificate")
	}

	// Get private key from key management service
	privateKeyPEM, err := s.keyMgmtService.GetDecryptedPrivateKey(int(org.AdminSignKeyID.Int64))
	if err != nil {
		return nil, fmt.Errorf("failed to get private key: %w", err)
	}

	cert, err := gwidentity.CertificateFromPEM([]byte(*adminSignKey.Certificate))
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate: %w", err)
	}

	priv, err := gwidentity.PrivateKeyFromPEM([]byte(privateKeyPEM))
	if err != nil {
		return nil, fmt.Errorf("failed to read private key: %w", err)
	}

	signingIdentity, err := identity.NewPrivateKeySigningIdentity(s.mspID, cert, priv)
	if err != nil {
		return nil, fmt.Errorf("failed to create signing identity: %w", err)
	}

	return signingIdentity, nil
}

// getOrdererMSP creates a signing identity for interacting with the orderer
func (s *FabricOrg) getOrdererMSP(ctx context.Context) (identity.SigningIdentity, error) {
	// Get organization details
	org, err := s.orgService.GetOrganizationByMspID(ctx, s.mspID)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	if !org.AdminSignKeyID.Valid {
		return nil, fmt.Errorf("organization has no signing key")
	}

	// Get admin signing key
	adminSignKey, err := s.keyMgmtService.GetKey(ctx, int(org.AdminSignKeyID.Int64))
	if err != nil {
		return nil, fmt.Errorf("failed to get admin sign key: %w", err)
	}
	if adminSignKey.Certificate == nil {
		return nil, fmt.Errorf("admin sign key has no certificate")
	}

	// Get private key from key management service
	privateKeyPEM, err := s.keyMgmtService.GetDecryptedPrivateKey(int(org.AdminSignKeyID.Int64))
	if err != nil {
		return nil, fmt.Errorf("failed to get private key: %w", err)
	}

	cert, err := identity.ReadCertificate(*adminSignKey.Certificate)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate: %w", err)
	}

	priv, err := identity.ReadPrivateKey(privateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key: %w", err)
	}

	ordererMSP, err := identity.NewPrivateKeySigningIdentity(s.mspID, cert, priv)
	if err != nil {
		return nil, fmt.Errorf("failed to create orderer msp: %w", err)
	}

	return ordererMSP, nil
}

// getOrdererConnection establishes a gRPC connection to the orderer
func (s *FabricOrg) getOrdererConnection(ctx context.Context, ordererURL string, ordererTLSCert string) (*grpc.ClientConn, error) {

	// Create orderer connection
	ordererConn, err := network.DialConnection(network.Node{
		Addr:          ordererURL,
		TLSCACertByte: []byte(ordererTLSCert),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create orderer connection: %w", err)
	}

	return ordererConn, nil
}

// getOrdererTLSKeyPair creates a TLS key pair for secure communication with the orderer
func (s *FabricOrg) getOrdererTLSKeyPair(ctx context.Context, ordererTLSCert string) (tls.Certificate, error) {
	// Get organization details
	org, err := s.orgService.GetOrganizationByMspID(ctx, s.mspID)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to get organization: %w", err)
	}

	if !org.AdminSignKeyID.Valid {
		return tls.Certificate{}, fmt.Errorf("organization has no admin sign key")
	}

	// Get private key from key management service
	privateKeyPEM, err := s.keyMgmtService.GetDecryptedPrivateKey(int(org.AdminSignKeyID.Int64))
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to get private key: %w", err)
	}

	// Parse the orderer TLS certificate
	ordererTLSCertParsed, err := tls.X509KeyPair([]byte(ordererTLSCert), []byte(privateKeyPEM))
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to parse orderer TLS certificate: %w", err)
	}

	return ordererTLSCertParsed, nil
}

// GetGenesisBlock fetches the genesis block for a channel from the orderer
func (s *FabricOrg) GetGenesisBlock(ctx context.Context, channelID string, ordererURL string, ordererTLSCert []byte) ([]byte, error) {
	s.logger.Info("Fetching genesis block with network config",
		"mspID", s.mspID,
		"channel", channelID,
		"ordererUrl", ordererURL)

	ordererConn, err := s.getOrdererConnection(ctx, ordererURL, string(ordererTLSCert))
	if err != nil {
		return nil, fmt.Errorf("failed to get orderer connection: %w", err)
	}
	defer ordererConn.Close()

	ordererMSP, err := s.getOrdererMSP(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get orderer msp: %w", err)
	}

	ordererTLSKeyPair, err := s.getOrdererTLSKeyPair(ctx, string(ordererTLSCert))
	if err != nil {
		return nil, fmt.Errorf("failed to get orderer tls key pair: %w", err)
	}
	genesisBlock, err := channel.GetGenesisBlock(ctx, ordererConn, ordererMSP, channelID, ordererTLSKeyPair)
	if err != nil {
		return nil, fmt.Errorf("failed to get genesis block: %w", err)
	}
	genesisBlockBytes, err := proto.Marshal(genesisBlock)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal genesis block: %w", err)
	}

	return genesisBlockBytes, nil
}

// CreateConfigSignature creates a signature for a config update using the organization's admin credentials
func (s *FabricOrg) CreateConfigSignature(ctx context.Context, channelID string, configUpdateBytes *cb.Envelope) (*cb.Envelope, error) {
	s.logger.Info("Creating config signature",
		"mspID", s.mspID,
		"channel", channelID)

	// Get organization details
	org, err := s.orgService.GetOrganizationByMspID(ctx, s.mspID)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	// Verify admin signing key exists
	if !org.AdminSignKeyID.Valid {
		return nil, fmt.Errorf("organization has no admin sign key")
	}

	// Get admin signing key and certificate
	adminSignKey, err := s.keyMgmtService.GetKey(ctx, int(org.AdminSignKeyID.Int64))
	if err != nil {
		return nil, fmt.Errorf("failed to get admin sign key: %w", err)
	}
	if adminSignKey.Certificate == nil {
		return nil, fmt.Errorf("admin sign key has no certificate")
	}

	// Create signing identity
	signingIdentity, err := s.getAdminIdentity(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create signing identity: %w", err)
	}

	// Create config signature from the config update bytes
	signature, err := SignConfigTx(channelID, configUpdateBytes, signingIdentity)
	if err != nil {
		return nil, fmt.Errorf("failed to create config signature: %w", err)
	}
	return signature, nil
}

const (
	msgVersion = int32(0)
	epoch      = 0
)

func SignConfigTx(channelID string, envConfigUpdate *cb.Envelope, signer identity.SigningIdentity) (*cb.Envelope, error) {
	payload, err := protoutil.UnmarshalPayload(envConfigUpdate.Payload)
	if err != nil {
		return nil, errors.New("bad payload")
	}

	if payload.Header == nil || payload.Header.ChannelHeader == nil {
		return nil, errors.New("bad header")
	}

	ch, err := protoutil.UnmarshalChannelHeader(payload.Header.ChannelHeader)
	if err != nil {
		return nil, errors.New("could not unmarshall channel header")
	}

	if ch.Type != int32(cb.HeaderType_CONFIG_UPDATE) {
		return nil, errors.New("bad type")
	}

	if ch.ChannelId == "" {
		return nil, errors.New("empty channel id")
	}

	configUpdateEnv, err := protoutil.UnmarshalConfigUpdateEnvelope(payload.Data)
	if err != nil {
		return nil, errors.New("bad config update env")
	}

	sigHeader, err := protoutil.NewSignatureHeader(signer)
	if err != nil {
		return nil, err
	}

	configSig := &cb.ConfigSignature{
		SignatureHeader: protoutil.MarshalOrPanic(sigHeader),
	}

	configSig.Signature, err = signer.Sign(Concatenate(configSig.SignatureHeader, configUpdateEnv.ConfigUpdate))
	if err != nil {
		return nil, err
	}

	configUpdateEnv.Signatures = append(configUpdateEnv.Signatures, configSig)

	return protoutil.CreateSignedEnvelope(cb.HeaderType_CONFIG_UPDATE, channelID, signer, configUpdateEnv, msgVersion, epoch)
}

func Concatenate[T any](slices ...[]T) []T {
	size := 0
	for _, slice := range slices {
		size += len(slice)
	}

	result := make([]T, size)
	i := 0
	for _, slice := range slices {
		copy(result[i:], slice)
		i += len(slice)
	}

	return result
}
