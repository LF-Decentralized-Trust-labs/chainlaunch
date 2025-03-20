package org

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"text/template"

	"github.com/golang/protobuf/proto"

	"github.com/chainlaunch/chainlaunch/pkg/fabric/service"
	keymanagement "github.com/chainlaunch/chainlaunch/pkg/keymanagement/service"
	"github.com/chainlaunch/chainlaunch/pkg/logger"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite/bccsp/sw"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	mspimpl "github.com/hyperledger/fabric-sdk-go/pkg/msp"
)

const tmplGoConfig = `
name: hlf-network
version: 1.0.0
client:
  organization: "{{ .Organization }}"
{{- if not .Organizations }}
organizations: {}
{{- else }}
organizations:
  {{ range $org := .Organizations }}
  {{ $org.MSPID }}:
    mspid: {{ $org.MSPID }}
    cryptoPath: /tmp/cryptopath
{{ if not $org.Users }}
    users: {}
{{- else }}
    users:
      {{- range $user := $org.Users }}
      {{ $user.Name }}:
        cert:
          pem: |
{{ $user.Cert | indent 12 }}
        key:
          pem: |
{{ $user.Key | indent 12 }}
{{- end }}
{{- end }}
{{- if not $org.CertAuths }}
    certificateAuthorities: []
{{- else }}
    certificateAuthorities: 
      {{- range $ca := $org.CertAuths }}
      - {{ $ca.Name }}
 	  {{- end }}
{{- end }}
{{- if not $org.Peers }}
    peers: []
{{- else }}
    peers:
      {{- range $peer := $org.Peers }}
      - {{ $peer }}
 	  {{- end }}
{{- end }}
{{- if not $org.Orderers }}
    orderers: []
{{- else }}
    orderers:
      {{- range $orderer := $org.Orderers }}
      - {{ $orderer }}
 	  {{- end }}

    {{- end }}
{{- end }}
{{- end }}

{{- if not .Orderers }}
{{- else }}
orderers:
{{- range $orderer := .Orderers }}
  {{$orderer.Name}}:
    url: {{ $orderer.URL }}
    grpcOptions:
      allow-insecure: false
    tlsCACerts:
      pem: |
{{ $orderer.TLSCACert | indent 8 }}
{{- end }}
{{- end }}

{{- if not .Peers }}
{{- else }}
peers:
  {{- range $peer := .Peers }}
  {{$peer.Name}}:
    url: {{ $peer.URL }}
    tlsCACerts:
      pem: |
{{ $peer.TLSCACert | indent 8 }}
{{- end }}
{{- end }}

{{- if not .CertAuths }}
{{- else }}
certificateAuthorities:
{{- range $ca := .CertAuths }}
  {{ $ca.Name }}:
    url: https://{{ $ca.URL }}
{{if $ca.EnrollID }}
    registrar:
        enrollId: {{ $ca.EnrollID }}
        enrollSecret: "{{ $ca.EnrollSecret }}"
{{ end }}
    caName: {{ $ca.CAName }}
    tlsCACerts:
      pem: 
       - |
{{ $ca.TLSCert | indent 12 }}

{{- end }}
{{- end }}

channels:
  _default:
{{- if not .Orderers }}
    orderers: []
{{- else }}
    orderers:
{{- range $orderer := .Orderers }}
      - {{$orderer.Name}}
{{- end }}
{{- end }}
{{- if not .Peers }}
    peers: {}
{{- else }}
    peers:
{{- range $peer := .Peers }}
       {{$peer.Name}}:
        discover: true
        endorsingPeer: true
        chaincodeQuery: true
        ledgerQuery: true
        eventSource: true
{{- end }}
{{- end }}

`

type OrgUser struct {
	Name string
	Cert string
	Key  string
}
type Org struct {
	MSPID     string
	CertAuths []string
	Peers     []string
	Orderers  []string
	Users     []OrgUser
}
type Peer struct {
	Name      string
	URL       string
	TLSCACert string
}
type CA struct {
	Name         string
	URL          string
	TLSCert      string
	EnrollID     string
	EnrollSecret string
}

type Orderer struct {
	URL       string
	Name      string
	TLSCACert string
}

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

// GenerateNetworkConfig generates a network configuration for connecting to Fabric network
func (s *FabricOrg) GenerateNetworkConfig(ctx context.Context, channelID, ordererURL, ordererTLSCert string) (string, error) {
	s.logger.Info("Generating network config",
		"mspID", s.mspID,
		"channel", channelID,
		"ordererUrl", ordererURL)

	// Get organization details
	org, err := s.orgService.GetOrganizationByMspID(ctx, s.mspID)
	if err != nil {
		return "", fmt.Errorf("failed to get organization: %w", err)
	}

	// Get signing key
	var privateKeyPEM string
	if !org.AdminSignKeyID.Valid {
		return "", fmt.Errorf("organization has no admin sign key")
	}
	adminSignKey, err := s.keyMgmtService.GetKey(ctx, int(org.AdminSignKeyID.Int64))
	if err != nil {
		return "", fmt.Errorf("failed to get admin sign key: %w", err)
	}
	if adminSignKey.Certificate == nil {
		return "", fmt.Errorf("admin sign key has no certificate")
	}
	// Get private key from key management service
	privateKeyPEM, err = s.keyMgmtService.GetDecryptedPrivateKey(int(org.AdminSignKeyID.Int64))
	if err != nil {
		return "", fmt.Errorf("failed to get private key: %w", err)
	}

	// Create template data
	orgs := []*Org{}
	var peers []*Peer
	var certAuths []*CA
	var ordererNodes []*Orderer

	// Add organization with user
	fabricOrg := &Org{
		MSPID:     org.MspID,
		CertAuths: []string{},
		Peers:     []string{},
		Orderers:  []string{},
	}

	// Add admin user if signing certificate is available
	if org.SignKeyID.Valid && org.SignCertificate != "" {
		adminUser := OrgUser{
			Name: "Admin",
			Cert: *adminSignKey.Certificate,
			Key:  privateKeyPEM,
		}
		fabricOrg.Users = []OrgUser{adminUser}
	}

	orgs = append(orgs, fabricOrg)
	if ordererURL != "" && ordererTLSCert != "" {
		fabricOrg.Orderers = []string{"orderer0"}
		// Add orderer
		orderer := &Orderer{
			URL:       ordererURL,
			Name:      "orderer0",
			TLSCACert: ordererTLSCert,
		}
		ordererNodes = append(ordererNodes, orderer)
	}

	// Parse template
	tmpl, err := template.New("networkConfig").Funcs(template.FuncMap{
		"indent": func(spaces int, v string) string {
			pad := strings.Repeat(" ", spaces)
			return pad + strings.Replace(v, "\n", "\n"+pad, -1)
		},
	}).Parse(tmplGoConfig)
	if err != nil {
		return "", fmt.Errorf("failed to parse network config template: %w", err)
	}

	// Execute template
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, map[string]interface{}{
		"Peers":         peers,
		"Orderers":      ordererNodes,
		"Organizations": orgs,
		"CertAuths":     certAuths,
		"Organization":  s.mspID,
		"Internal":      false,
	})
	if err != nil {
		return "", fmt.Errorf("failed to execute network config template: %w", err)
	}

	return buf.String(), nil
}

// GetConfigBlockWithNetworkConfig retrieves a config block using a generated network config
func (s *FabricOrg) GetConfigBlockWithNetworkConfig(ctx context.Context, channelID, ordererURL, ordererTLSCert string) (*common.Block, error) {
	s.logger.Info("Fetching channel config with network config",
		"mspID", s.mspID,
		"channel", channelID,
		"ordererUrl", ordererURL)

	// Get organization details
	org, err := s.orgService.GetOrganizationByMspID(ctx, s.mspID)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	// Get signing key
	if !org.AdminSignKeyID.Valid {
		return nil, fmt.Errorf("organization has no signing key")
	}
	// Generate network config
	networkConfig, err := s.GenerateNetworkConfig(ctx, channelID, ordererURL, ordererTLSCert)
	if err != nil {
		return nil, fmt.Errorf("failed to generate network config: %w", err)
	}

	// Initialize SDK with network config
	configBackend := config.FromRaw([]byte(networkConfig), "yaml")
	sdk, err := fabsdk.New(configBackend)
	if err != nil {
		return nil, fmt.Errorf("failed to create sdk: %w", err)
	}
	defer sdk.Close()

	// Create SDK context
	sdkContext := sdk.Context(
		fabsdk.WithOrg(s.mspID),
		fabsdk.WithUser("Admin"),
	)

	// Create resource management client
	resClient, err := resmgmt.New(sdkContext)
	if err != nil {
		return nil, fmt.Errorf("failed to create resmgmt client: %w", err)
	}

	// Fetch channel configuration
	configBlock, err := resClient.QueryConfigBlockFromOrderer(channelID)
	if err != nil {
		return nil, fmt.Errorf("failed to query channel config: %w", err)
	}

	return configBlock, nil
}

// GetGenesisBlock fetches the genesis block for a channel from the orderer
func (s *FabricOrg) GetGenesisBlock(ctx context.Context, channelID string, ordererURL string, ordererTLSCert []byte) ([]byte, error) {
	s.logger.Info("Fetching genesis block with network config",
		"mspID", s.mspID,
		"channel", channelID,
		"ordererUrl", ordererURL)

	// Get organization details
	org, err := s.orgService.GetOrganizationByMspID(ctx, s.mspID)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	// Get signing key
	if !org.AdminSignKeyID.Valid {
		return nil, fmt.Errorf("organization has no signing key")
	}

	// Generate network config
	networkConfig, err := s.GenerateNetworkConfig(ctx, channelID, ordererURL, string(ordererTLSCert))
	if err != nil {
		return nil, fmt.Errorf("failed to generate network config: %w", err)
	}

	// Initialize SDK with network config
	configBackend := config.FromRaw([]byte(networkConfig), "yaml")
	sdk, err := fabsdk.New(configBackend)
	if err != nil {
		return nil, fmt.Errorf("failed to create sdk: %w", err)
	}
	defer sdk.Close()
	resmClient, err := resmgmt.New(sdk.Context(
		fabsdk.WithOrg(s.mspID),
		fabsdk.WithUser("Admin"),
	))
	if err != nil {
		return nil, fmt.Errorf("failed to create resmgmt client: %w", err)
	}
	genesisBlock, err := resmClient.GenesisBlock(channelID)
	if err != nil {
		return nil, fmt.Errorf("failed to query genesis block: %w", err)
	}
	genesisBlockBytes, err := proto.Marshal(genesisBlock)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal genesis block: %w", err)
	}
	return genesisBlockBytes, nil
}

// createSigningIdentity creates a signing identity from the organization's admin credentials
func (s *FabricOrg) createSigningIdentity(sdk *fabsdk.FabricSDK, privateKeyPEM string, certPEM string) (msp.SigningIdentity, error) {
	sdkConfig, err := sdk.Config()
	if err != nil {
		return nil, fmt.Errorf("failed to get SDK config: %w", err)
	}

	cryptoConfig := cryptosuite.ConfigFromBackend(sdkConfig)
	cryptoSuite, err := sw.GetSuiteByConfig(cryptoConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to get crypto suite: %w", err)
	}

	userStore := mspimpl.NewMemoryUserStore()
	endpointConfig, err := fab.ConfigFromBackend(sdkConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to get endpoint config: %w", err)
	}

	identityManager, err := mspimpl.NewIdentityManager(s.mspID, userStore, cryptoSuite, endpointConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create identity manager: %w", err)
	}

	return identityManager.CreateSigningIdentity(
		msp.WithPrivateKey([]byte(privateKeyPEM)),
		msp.WithCert([]byte(certPEM)),
	)
}

// CreateConfigSignature creates a signature for a config update using the organization's admin credentials
func (s *FabricOrg) CreateConfigSignature(ctx context.Context, channelID string, configUpdateBytes []byte) (*common.ConfigSignature, error) {
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

	// Get private key
	privateKeyPEM, err := s.keyMgmtService.GetDecryptedPrivateKey(int(org.AdminSignKeyID.Int64))
	if err != nil {
		return nil, fmt.Errorf("failed to get private key: %w", err)
	}

	// Generate network config for SDK initialization
	networkConfig, err := s.GenerateNetworkConfig(ctx, channelID, "", "") // Empty orderer details as they're not needed for signing
	if err != nil {
		return nil, fmt.Errorf("failed to generate network config: %w", err)
	}

	// Initialize SDK
	configBackend := config.FromRaw([]byte(networkConfig), "yaml")
	sdk, err := fabsdk.New(configBackend)
	if err != nil {
		return nil, fmt.Errorf("failed to create sdk: %w", err)
	}
	defer sdk.Close()

	// Create signing identity
	signingIdentity, err := s.createSigningIdentity(sdk, privateKeyPEM, *adminSignKey.Certificate)
	if err != nil {
		return nil, fmt.Errorf("failed to create signing identity: %w", err)
	}

	// Create SDK context with signing identity
	sdkContext := sdk.Context(
		fabsdk.WithIdentity(signingIdentity),
		fabsdk.WithOrg(s.mspID),
	)

	// Create resource management client
	resClient, err := resmgmt.New(sdkContext)
	if err != nil {
		return nil, fmt.Errorf("failed to create resmgmt client: %w", err)
	}

	// Create config signature from the config update bytes
	signature, err := resClient.CreateConfigSignatureFromReader(signingIdentity, bytes.NewReader(configUpdateBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create config signature: %w", err)
	}

	return signature, nil
}
