package orderer

// StartOrdererOpts represents the options for starting an orderer
type StartOrdererOpts struct {
	ID                      string            `json:"id"`
	ListenAddress           string            `json:"listenAddress"`
	AdminAddress            string            `json:"adminAddress"`
	OperationsListenAddress string            `json:"operationsListenAddress"`
	ExternalEndpoint        string            `json:"externalEndpoint"`
	DomainNames             []string          `json:"domainNames"`
	Env                     map[string]string `json:"env"`
	Version                 string            `json:"version"` // Fabric version to use
}

// OrdererConfig represents the configuration for an orderer node
type OrdererConfig struct {
	Mode                    string `json:"mode"`
	ListenAddress           string `json:"listenAddress"`
	OperationsListenAddress string `json:"operationsListenAddress"`
	AdminAddress            string `json:"adminAddress"`
	ExternalEndpoint        string `json:"externalEndpoint"`
	SignCert                string `json:"signCert"`
	SignCACert              string `json:"signCACert"`
	SignKey                 string `json:"signKey"`
	OrdererName             string `json:"ordererName"`
	TLSCert                 string `json:"tlsCert"`
	TLSCACert               string `json:"tlsCACert"`
	TLSKey                  string `json:"tlsKey"`
}

// StartServiceResponse represents the response when starting an orderer as a service
type StartServiceResponse struct {
	Mode        string `json:"mode"`
	Type        string `json:"type"`
	ServiceName string `json:"serviceName"`
}

// StartDockerResponse represents the response when starting an orderer as a docker container
type StartDockerResponse struct {
	Mode          string `json:"mode"`
	ContainerName string `json:"containerName"`
}
