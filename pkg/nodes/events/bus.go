package events

import (
	"sync"

	"github.com/chainlaunch/chainlaunch/pkg/nodes/eventtypes"
)

// EventBus implements the Bus interface
type EventBus struct {
	handlers []eventtypes.Handler
	mu       sync.RWMutex
}

// NewEventBus creates a new EventBus
func NewEventBus() *EventBus {
	return &EventBus{
		handlers: make([]eventtypes.Handler, 0),
	}
}

// Subscribe subscribes to node events
func (b *EventBus) Subscribe(handler eventtypes.Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers = append(b.handlers, handler)
}

// Publish publishes a node event
func (b *EventBus) Publish(event eventtypes.Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, handler := range b.handlers {
		handler(event)
	}
}
