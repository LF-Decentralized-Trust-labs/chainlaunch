package service

import "time"

// Template represents a node configuration template
type Template struct {
	ID          int64                  `json:"id"`
	Type        string                 `json:"type"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Template    map[string]interface{} `json:"template"`
	CreatedAt   time.Time              `json:"createdAt"`
	UpdatedAt   time.Time              `json:"updatedAt"`
}

// CreateTemplateParams represents parameters for creating a template
type CreateTemplateParams struct {
	Type        string                 `json:"type" validate:"required"`
	Name        string                 `json:"name" validate:"required"`
	Description string                 `json:"description"`
	Template    map[string]interface{} `json:"template" validate:"required"`
}

// UpdateTemplateParams represents parameters for updating a template
type UpdateTemplateParams struct {
	Name        string                 `json:"name,omitempty"`
	Description string                 `json:"description,omitempty"`
	Template    map[string]interface{} `json:"template,omitempty"`
}

// Default templates for each node type
var (
	DefaultFabricPeerTemplate = map[string]interface{}{
		"command": "peer node start",
		"env": map[string]string{
			"CORE_PEER_ID":                      "{{.Name}}",
			"CORE_PEER_ADDRESS":                 "{{.ListenAddress}}",
			"CORE_PEER_LISTENADDRESS":           "{{.ListenAddress}}",
			"CORE_PEER_CHAINCODEADDRESS":        "{{.ChaincodeAddress}}",
			"CORE_PEER_CHAINCODELISTENADDRESS":  "{{.ChaincodeAddress}}",
			"CORE_PEER_GOSSIP_EXTERNALENDPOINT": "{{.ExternalEndpoint}}",
			"CORE_PEER_LOCALMSPID":              "{{.MSPID}}",
		},
	}

	DefaultFabricOrdererTemplate = map[string]interface{}{
		"command": "orderer",
		"env": map[string]string{
			"ORDERER_GENERAL_LISTENADDRESS":    "{{.ListenAddress}}",
			"ORDERER_GENERAL_LISTENPORT":       "{{.Port}}",
			"ORDERER_GENERAL_LOCALMSPID":       "{{.MSPID}}",
			"ORDERER_GENERAL_LOCALMSPDIR":      "{{.MSPDir}}",
			"ORDERER_OPERATIONS_LISTENADDRESS": "{{.OperationsListenAddress}}",
		},
	}

	DefaultBesuTemplate = map[string]interface{}{
		"command": "besu",
		"args": []string{
			"--data-path={{.DataDir}}",
			"--genesis-file={{.GenesisPath}}",
			"--rpc-http-enabled",
			"--rpc-http-host={{.RPCHost}}",
			"--rpc-http-port={{.RPCPort}}",
			"--p2p-host={{.P2PHost}}",
			"--p2p-port={{.P2PPort}}",
			"--network-id={{.NetworkID}}",
		},
	}
)
