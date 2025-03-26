package monitoring

import (
	"time"
)

// NodeStatus represents the current status of a node
type NodeStatus string

const (
	// NodeStatusUp indicates the node is operational
	NodeStatusUp NodeStatus = "up"
	// NodeStatusDown indicates the node is not responding
	NodeStatusDown NodeStatus = "down"
)

// Node represents a node to be monitored
type Node struct {
	// ID is a unique identifier for the node
	ID int64
	// Name is a human-readable name for the node
	Name string
	// URL is the endpoint to check for node status
	URL string
	// CheckInterval is how often this node should be checked
	CheckInterval time.Duration
	// Timeout is the maximum time to wait for a response
	Timeout time.Duration
	// Status is the current status of the node
	Status NodeStatus
	// LastChecked is when the node was last checked
	LastChecked time.Time
	// LastStatusChange is when the node status last changed
	LastStatusChange time.Time
	// FailureCount tracks consecutive failures
	FailureCount int
	// FailureThreshold is how many consecutive failures before alerting
	FailureThreshold int
}

// NodeCheck represents the result of a node check
type NodeCheck struct {
	// Node is the node that was checked
	Node *Node
	// Status is the status determined by the check
	Status NodeStatus
	// ResponseTime is how long the check took
	ResponseTime time.Duration
	// Error is any error that occurred during the check
	Error error
	// Timestamp is when the check was performed
	Timestamp time.Time
}
