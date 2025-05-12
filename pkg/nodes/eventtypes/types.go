package eventtypes

import (
	"time"

	"github.com/chainlaunch/chainlaunch/pkg/nodes/nodetypes"
)

// EventType represents the type of node event
type EventType string

const (
	// EventTypeNodeCreated is emitted when a node is created
	EventTypeNodeCreated EventType = "node.created"
	// EventTypeNodeUpdated is emitted when a node is updated
	EventTypeNodeUpdated EventType = "node.updated"
	// EventTypeNodeDeleted is emitted when a node is deleted
	EventTypeNodeDeleted EventType = "node.deleted"
)

// Event represents a node event
type Event struct {
	Type      EventType
	Node      *nodetypes.Node
	NodeID    int64
	Timestamp time.Time
}

// Handler is a function that handles node events
type Handler func(Event)
