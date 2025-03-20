package nodes

// NodeType represents the type of node (peer or orderer)
type NodeType string

const (
	PeerNode    NodeType = "peer"
	OrdererNode NodeType = "orderer"
)

// Node represents the interface that all nodes must implement
type Node interface {
	Start() (interface{}, error)
	Stop() error
	Init() (interface{}, error)
	RenewCertificates() error
}
