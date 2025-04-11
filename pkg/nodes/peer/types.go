package peer

import "github.com/chainlaunch/chainlaunch/pkg/nodes/types"

// StartPeerOpts represents the options for starting a peer
type StartPeerOpts struct {
	ID                      string                  `json:"id"`
	ListenAddress           string                  `json:"listenAddress"`
	ChaincodeAddress        string                  `json:"chaincodeAddress"`
	EventsAddress           string                  `json:"eventsAddress"`
	OperationsListenAddress string                  `json:"operationsListenAddress"`
	ExternalEndpoint        string                  `json:"externalEndpoint"`
	DomainNames             []string                `json:"domainNames"`
	Env                     map[string]string       `json:"env"`
	Version                 string                  `json:"version"` // Fabric version to use
	AddressOverrides        []types.AddressOverride `json:"addressOverrides,omitempty"`
}

// PeerConfig represents the configuration for a peer node
type PeerConfig struct {
	Mode                    string `json:"mode"`
	ListenAddress           string `json:"listenAddress"`
	ChaincodeAddress        string `json:"chaincodeAddress"`
	EventsAddress           string `json:"eventsAddress"`
	OperationsListenAddress string `json:"operationsListenAddress"`
	ExternalEndpoint        string `json:"externalEndpoint"`
	SignCert                string `json:"signCert"`
	SignCACert              string `json:"signCACert"`
	SignKey                 string `json:"signKey"`
	PeerName                string `json:"peerName"`
	TLSCert                 string `json:"tlsCert"`
	TLSCACert               string `json:"tlsCACert"`
	TLSKey                  string `json:"tlsKey"`
}

// StartServiceResponse represents the response when starting a peer as a service
type StartServiceResponse struct {
	Mode        string `json:"mode"`
	Type        string `json:"type"`
	ServiceName string `json:"serviceName"`
}

// StartDockerResponse represents the response when starting a peer as a docker container
type StartDockerResponse struct {
	Mode          string `json:"mode"`
	ContainerName string `json:"containerName"`
}

type BlockInfo struct {
	Height            uint64 `json:"height"`
	CurrentBlockHash  string `json:"currentBlockHash"`
	PreviousBlockHash string `json:"previousBlockHash"`
}
