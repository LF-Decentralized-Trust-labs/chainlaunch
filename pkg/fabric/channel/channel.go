package channel

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"

	"crypto/x509"
	"time"

	"github.com/hyperledger/fabric-config/configtx"
	"github.com/hyperledger/fabric-config/configtx/membership"
	"github.com/hyperledger/fabric-config/configtx/orderer"
	"github.com/hyperledger/fabric-config/protolator"
	cb "github.com/hyperledger/fabric-protos-go/common"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric/protoutil"
)

// ChannelService handles channel operations
type ChannelService struct {
	// Add any dependencies here
}

// NewChannelService creates a new channel service
func NewChannelService() *ChannelService {
	return &ChannelService{}
}

// HostPort represents a network host and port
type HostPort struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

// Organization represents a blockchain organization
type Organization struct {
	Name             string     `json:"name"`
	AnchorPeers      []HostPort `json:"anchorPeers"`
	OrdererEndpoints []string   `json:"ordererEndpoints"`
	SignCACert       string     `json:"signCACert"`
	TLSCACert        string     `json:"tlsCACert"`
}

// AddressWithCerts represents a network address with TLS certificates
type AddressWithCerts struct {
	Address       HostPort `json:"address"`
	ClientTLSCert string   `json:"clientTLSCert"`
	ServerTLSCert string   `json:"serverTLSCert"`
}

// CreateChannelInput represents the input for creating a new channel
type CreateChannelInput struct {
	Name        string             `json:"name"`
	PeerOrgs    []Organization     `json:"peerOrgs"`
	OrdererOrgs []Organization     `json:"ordererOrgs"`
	Consenters  []AddressWithCerts `json:"consenters"`
}

// SetAnchorPeersInput represents the input for setting anchor peers
type SetAnchorPeersInput struct {
	CurrentConfig *cb.Config
	AnchorPeers   []HostPort
	MSPID         string
	ChannelName   string
}

// CreateChannelResponse represents the response from creating a channel
type CreateChannelResponse struct {
	ChannelID  string `json:"channelId"`
	ConfigData string `json:"configData"`
}

// CreateChannel creates a new channel with the given configuration
func (s *ChannelService) CreateChannel(input CreateChannelInput) (*CreateChannelResponse, error) {
	channelConfig, err := s.parseAndCreateChannel(input)
	if err != nil {
		return nil, fmt.Errorf("failed to create channel: %w", err)
	}

	return &CreateChannelResponse{
		ChannelID:  input.Name,
		ConfigData: base64.StdEncoding.EncodeToString(channelConfig),
	}, nil
}

// SetAnchorPeers updates the anchor peers for an organization in a channel
func (s *ChannelService) SetAnchorPeers(input *SetAnchorPeersInput) ([]byte, error) {
	// Create config manager and update anchor peers
	cftxGen := configtx.New(input.CurrentConfig)
	app := cftxGen.Application().Organization(input.MSPID)

	// Remove existing anchor peers
	currentAnchorPeers, err := app.AnchorPeers()
	if err != nil {
		return nil, fmt.Errorf("failed to get current anchor peers: %w", err)
	}

	for _, ap := range currentAnchorPeers {
		if err := app.RemoveAnchorPeer(configtx.Address{
			Host: ap.Host,
			Port: ap.Port,
		}); err != nil {
			return nil, fmt.Errorf("failed to remove anchor peer: %w", err)
		}
	}

	// Add new anchor peers
	for _, ap := range input.AnchorPeers {
		if err := app.AddAnchorPeer(configtx.Address{
			Host: ap.Host,
			Port: ap.Port,
		}); err != nil {
			return nil, fmt.Errorf("failed to add anchor peer: %w", err)
		}
	}

	// Compute update
	configUpdateBytes, err := cftxGen.ComputeMarshaledUpdate(input.ChannelName)
	if err != nil {
		return nil, fmt.Errorf("failed to compute update: %w", err)
	}

	configUpdate := &cb.ConfigUpdate{}
	if err := proto.Unmarshal(configUpdateBytes, configUpdate); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config update: %w", err)
	}

	// Create envelope
	envelopeBytes, err := s.createConfigUpdateEnvelope(input.ChannelName, configUpdate)
	if err != nil {
		return nil, fmt.Errorf("failed to create config update envelope: %w", err)
	}

	return envelopeBytes, nil
}
func (s *ChannelService) createConfigUpdateEnvelope(channelID string, configUpdate *cb.ConfigUpdate) ([]byte, error) {
	configUpdate.ChannelId = channelID
	configUpdateData, err := proto.Marshal(configUpdate)
	if err != nil {
		return nil, err
	}
	configUpdateEnvelope := &cb.ConfigUpdateEnvelope{}
	configUpdateEnvelope.ConfigUpdate = configUpdateData
	envelope, err := protoutil.CreateSignedEnvelope(cb.HeaderType_CONFIG_UPDATE, channelID, nil, configUpdateEnvelope, 0, 0)
	if err != nil {
		return nil, err
	}
	envelopeData, err := proto.Marshal(envelope)
	if err != nil {
		return nil, err
	}
	return envelopeData, nil
}

// DecodeBlock decodes a base64 encoded block into JSON
func (s *ChannelService) DecodeBlock(blockB64 string) (map[string]interface{}, error) {
	blockBytes, err := base64.StdEncoding.DecodeString(blockB64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode block: %w", err)
	}

	block := &cb.Block{}
	if err := proto.Unmarshal(blockBytes, block); err != nil {
		return nil, fmt.Errorf("failed to unmarshal block: %w", err)
	}

	var buf bytes.Buffer
	if err := protolator.DeepMarshalJSON(&buf, block); err != nil {
		return nil, fmt.Errorf("failed to marshal block to JSON: %w", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return data, nil
}

// Helper functions below...

func (s *ChannelService) parseAndCreateChannel(input CreateChannelInput) ([]byte, error) {
	// Parse organizations
	peerOrgs := []configtx.Organization{}
	for _, org := range input.PeerOrgs {
		// Parse certificates
		signCACert, err := parseCertificate(org.SignCACert)
		if err != nil {
			return nil, fmt.Errorf("failed to parse signing CA cert for org %s: %w", org.Name, err)
		}

		tlsCACert, err := parseCertificate(org.TLSCACert)
		if err != nil {
			return nil, fmt.Errorf("failed to parse TLS CA cert for org %s: %w", org.Name, err)
		}

		// Convert anchor peers
		anchorPeers := make([]configtx.Address, len(org.AnchorPeers))
		for i, ap := range org.AnchorPeers {
			anchorPeers[i] = configtx.Address{
				Host: ap.Host,
				Port: ap.Port,
			}
		}

		// Create organization config
		peerOrg := configtx.Organization{
			Name: org.Name,
			MSP: configtx.MSP{
				Name:         org.Name,
				RootCerts:    []*x509.Certificate{signCACert},
				TLSRootCerts: []*x509.Certificate{tlsCACert},
				NodeOUs: membership.NodeOUs{
					Enable: true,
					ClientOUIdentifier: membership.OUIdentifier{
						Certificate:                  signCACert,
						OrganizationalUnitIdentifier: "client",
					},
					PeerOUIdentifier: membership.OUIdentifier{
						Certificate:                  signCACert,
						OrganizationalUnitIdentifier: "peer",
					},
					AdminOUIdentifier: membership.OUIdentifier{
						Certificate:                  signCACert,
						OrganizationalUnitIdentifier: "admin",
					},
					OrdererOUIdentifier: membership.OUIdentifier{
						Certificate:                  signCACert,
						OrganizationalUnitIdentifier: "orderer",
					},
				},
			},
			Policies: map[string]configtx.Policy{
				"Admins": {
					Type: "Signature",
					Rule: fmt.Sprintf("OR('%s.admin')", org.Name),
				},
				"Readers": {
					Type: "Signature",
					Rule: fmt.Sprintf("OR('%s.member')", org.Name),
				},
				"Writers": {
					Type: "Signature",
					Rule: fmt.Sprintf("OR('%s.member')", org.Name),
				},
				"Endorsement": {
					Type: "Signature",
					Rule: fmt.Sprintf("OR('%s.member')", org.Name),
				},
			},
			AnchorPeers:      anchorPeers,
			OrdererEndpoints: org.OrdererEndpoints,
		}

		peerOrgs = append(peerOrgs, peerOrg)
	}

	// Parse orderer organizations
	ordererOrgs := []configtx.Organization{}
	for _, org := range input.OrdererOrgs {
		signCACert, err := parseCertificate(org.SignCACert)
		if err != nil {
			return nil, fmt.Errorf("failed to parse signing CA cert for orderer org %s: %w", org.Name, err)
		}

		tlsCACert, err := parseCertificate(org.TLSCACert)
		if err != nil {
			return nil, fmt.Errorf("failed to parse TLS CA cert for orderer org %s: %w", org.Name, err)
		}

		ordererOrg := configtx.Organization{
			Name: org.Name,
			MSP: configtx.MSP{
				Name:         org.Name,
				RootCerts:    []*x509.Certificate{signCACert},
				TLSRootCerts: []*x509.Certificate{tlsCACert},
				NodeOUs: membership.NodeOUs{
					Enable: true,
					ClientOUIdentifier: membership.OUIdentifier{
						Certificate:                  signCACert,
						OrganizationalUnitIdentifier: "client",
					},
					OrdererOUIdentifier: membership.OUIdentifier{
						Certificate:                  signCACert,
						OrganizationalUnitIdentifier: "orderer",
					},
					AdminOUIdentifier: membership.OUIdentifier{
						Certificate:                  signCACert,
						OrganizationalUnitIdentifier: "admin",
					},
					PeerOUIdentifier: membership.OUIdentifier{
						Certificate:                  signCACert,
						OrganizationalUnitIdentifier: "peer",
					},
				},
			},
			Policies: map[string]configtx.Policy{
				"Admins": {
					Type: "Signature",
					Rule: fmt.Sprintf("OR('%s.admin')", org.Name),
				},
				"Readers": {
					Type: "Signature",
					Rule: fmt.Sprintf("OR('%s.member')", org.Name),
				},
				"Writers": {
					Type: "Signature",
					Rule: fmt.Sprintf("OR('%s.member')", org.Name),
				},
			},
			OrdererEndpoints: org.OrdererEndpoints,
		}

		ordererOrgs = append(ordererOrgs, ordererOrg)
	}

	// Parse consenters
	consenters := []orderer.Consenter{}
	for _, cons := range input.Consenters {
		clientTLSCert, err := parseCertificate(cons.ClientTLSCert)
		if err != nil {
			return nil, fmt.Errorf("failed to parse client TLS cert for consenter %s: %w", cons.Address.Host, err)
		}

		serverTLSCert, err := parseCertificate(cons.ServerTLSCert)
		if err != nil {
			return nil, fmt.Errorf("failed to parse server TLS cert for consenter %s: %w", cons.Address.Host, err)
		}

		consenters = append(consenters, orderer.Consenter{
			Address: orderer.EtcdAddress{
				Host: cons.Address.Host,
				Port: cons.Address.Port,
			},
			ClientTLSCert: clientTLSCert,
			ServerTLSCert: serverTLSCert,
		})
	}

	// Create channel configuration
	channelConfig := configtx.Channel{
		Consortiums: nil, // Not needed for application channels
		Application: configtx.Application{
			Organizations: peerOrgs,
			Capabilities:  []string{"V2_0"},
			ACLs:          defaultACLs(),
			Policies: map[string]configtx.Policy{
				"Readers": {
					Type: "ImplicitMeta",
					Rule: "ANY Readers",
				},
				"Writers": {
					Type: "ImplicitMeta",
					Rule: "ANY Writers",
				},
				"Admins": {
					Type: "ImplicitMeta",
					Rule: "MAJORITY Admins",
				},
				"LifecycleEndorsement": {
					Type: "ImplicitMeta",
					Rule: "MAJORITY Endorsement",
				},
				"Endorsement": {
					Type: "ImplicitMeta",
					Rule: "MAJORITY Endorsement",
				},
			},
		},
		Orderer: configtx.Orderer{
			OrdererType:  orderer.ConsensusTypeEtcdRaft,
			BatchTimeout: 2 * time.Second,
			State:        orderer.ConsensusStateNormal,
			BatchSize: orderer.BatchSize{
				MaxMessageCount:   500,
				AbsoluteMaxBytes:  10 * 1024 * 1024,
				PreferredMaxBytes: 2 * 1024 * 1024,
			},
			EtcdRaft: orderer.EtcdRaft{
				Consenters: consenters,
				Options: orderer.EtcdRaftOptions{
					TickInterval:         "500ms",
					ElectionTick:         10,
					HeartbeatTick:        1,
					MaxInflightBlocks:    5,
					SnapshotIntervalSize: 16 * 1024 * 1024, // 16 MB
				},
			},
			Organizations: ordererOrgs,
			Capabilities:  []string{"V2_0"},
			Policies: map[string]configtx.Policy{
				"Readers": {
					Type: "ImplicitMeta",
					Rule: "ANY Readers",
				},
				"Writers": {
					Type: "ImplicitMeta",
					Rule: "ANY Writers",
				},
				"Admins": {
					Type: "ImplicitMeta",
					Rule: "MAJORITY Admins",
				},
				"BlockValidation": {
					Type: "ImplicitMeta",
					Rule: "ANY Writers",
				},
			},
		},
		Capabilities: []string{"V2_0"},
		Policies: map[string]configtx.Policy{
			"Readers": {
				Type: "ImplicitMeta",
				Rule: "ANY Readers",
			},
			"Writers": {
				Type: "ImplicitMeta",
				Rule: "ANY Writers",
			},
			"Admins": {
				Type: "ImplicitMeta",
				Rule: "MAJORITY Admins",
			},
		},
	}

	// Create genesis block
	block, err := configtx.NewApplicationChannelGenesisBlock(channelConfig, input.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to create genesis block: %w", err)
	}

	// Marshal the block
	blockBytes, err := proto.Marshal(block)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal genesis block: %w", err)
	}

	return blockBytes, nil
}

// Helper function to parse PEM certificates
func parseCertificate(certPEM string) (*x509.Certificate, error) {
	block, _ := pem.Decode([]byte(certPEM))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	return cert, nil
}

func defaultACLs() map[string]string {
	return map[string]string{
		"_lifecycle/CheckCommitReadiness": "/Channel/Application/Writers",

		//  ACL policy for _lifecycle's "CommitChaincodeDefinition" function
		"_lifecycle/CommitChaincodeDefinition": "/Channel/Application/Writers",

		//  ACL policy for _lifecycle's "QueryChaincodeDefinition" function
		"_lifecycle/QueryChaincodeDefinition": "/Channel/Application/Writers",

		//  ACL policy for _lifecycle's "QueryChaincodeDefinitions" function
		"_lifecycle/QueryChaincodeDefinitions": "/Channel/Application/Writers",

		// ---Lifecycle System Chaincode (lscc) function to policy mapping for access control---//

		//  ACL policy for lscc's "getid" function
		"lscc/ChaincodeExists": "/Channel/Application/Readers",

		//  ACL policy for lscc's "getdepspec" function
		"lscc/GetDeploymentSpec": "/Channel/Application/Readers",

		//  ACL policy for lscc's "getccdata" function
		"lscc/GetChaincodeData": "/Channel/Application/Readers",

		//  ACL Policy for lscc's "getchaincodes" function
		"lscc/GetInstantiatedChaincodes": "/Channel/Application/Readers",

		// ---Query System Chaincode (qscc) function to policy mapping for access control---//

		//  ACL policy for qscc's "GetChainInfo" function
		"qscc/GetChainInfo": "/Channel/Application/Readers",

		//  ACL policy for qscc's "GetBlockByNumber" function
		"qscc/GetBlockByNumber": "/Channel/Application/Readers",

		//  ACL policy for qscc's  "GetBlockByHash" function
		"qscc/GetBlockByHash": "/Channel/Application/Readers",

		//  ACL policy for qscc's "GetTransactionByID" function
		"qscc/GetTransactionByID": "/Channel/Application/Readers",

		//  ACL policy for qscc's "GetBlockByTxID" function
		"qscc/GetBlockByTxID": "/Channel/Application/Readers",

		// ---Configuration System Chaincode (cscc) function to policy mapping for access control---//

		//  ACL policy for cscc's "GetConfigBlock" function
		"cscc/GetConfigBlock": "/Channel/Application/Readers",

		//  ACL policy for cscc's "GetChannelConfig" function
		"cscc/GetChannelConfig": "/Channel/Application/Readers",

		// ---Miscellaneous peer function to policy mapping for access control---//

		//  ACL policy for invoking chaincodes on peer
		"peer/Propose": "/Channel/Application/Writers",

		//  ACL policy for chaincode to chaincode invocation
		"peer/ChaincodeToChaincode": "/Channel/Application/Writers",

		// ---Events resource to policy mapping for access control// // // ---//

		//  ACL policy for sending block events
		"event/Block": "/Channel/Application/Readers",

		//  ACL policy for sending filtered block events
		"event/FilteredBlock": "/Channel/Application/Readers",
	}
}
