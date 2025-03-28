package monitoring

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/chainlaunch/chainlaunch/pkg/notifications"
)

// Service defines the interface for the monitoring service
type Service interface {
	// Start begins monitoring nodes
	Start(ctx context.Context) error
	// Stop stops monitoring nodes
	Stop() error
	// AddNode adds a node to be monitored
	AddNode(node *Node) error
	// RemoveNode removes a node from monitoring
	RemoveNode(nodeID int64) error
	// NodeExists checks if a node exists
	NodeExists(nodeID int64) bool

	// GetNodeStatus returns the current status of a node
	GetNodeStatus(nodeID int64) (*NodeCheck, error)
	// GetAllNodeStatuses returns the current status of all nodes
	GetAllNodeStatuses() []*NodeCheck
}

// service implements the Service interface
type service struct {
	config           *Config
	nodes            map[int64]*Node
	nodesMutex       sync.RWMutex
	httpClient       *http.Client
	notificationSvc  notifications.Service
	stopChan         chan struct{}
	workerWaitGroup  sync.WaitGroup
	lastCheckResults map[int64]*NodeCheck
	resultsMutex     sync.RWMutex
}

// NewService creates a new monitoring service
func NewService(config *Config, notificationSvc notifications.Service) Service {
	if config == nil {
		config = DefaultConfig()
	}

	return &service{
		config:           config,
		nodes:            make(map[int64]*Node),
		notificationSvc:  notificationSvc,
		stopChan:         make(chan struct{}),
		lastCheckResults: make(map[int64]*NodeCheck),
		httpClient: &http.Client{
			Timeout: config.DefaultTimeout,
		},
	}
}

// Start begins monitoring nodes
func (s *service) Start(ctx context.Context) error {
	// Create a worker pool to check nodes
	for i := 0; i < s.config.Workers; i++ {
		s.workerWaitGroup.Add(1)
		go s.worker(ctx, i)
	}
	return nil
}

// Stop stops monitoring nodes
func (s *service) Stop() error {
	close(s.stopChan)
	s.workerWaitGroup.Wait()
	return nil
}

// AddNode adds a node to be monitored
func (s *service) AddNode(node *Node) error {
	if node.ID == 0 {
		return fmt.Errorf("node ID cannot be zero")
	}
	if node.URL == "" {
		return fmt.Errorf("node URL cannot be empty")
	}

	// Set defaults if not provided
	if node.CheckInterval == 0 {
		node.CheckInterval = s.config.DefaultCheckInterval
	}
	if node.Timeout == 0 {
		node.Timeout = s.config.DefaultTimeout
	}
	if node.FailureThreshold == 0 {
		node.FailureThreshold = s.config.DefaultFailureThreshold
	}

	s.nodesMutex.Lock()
	s.nodes[node.ID] = node
	s.nodesMutex.Unlock()

	return nil
}

// RemoveNode removes a node from monitoring
func (s *service) RemoveNode(nodeID int64) error {
	s.nodesMutex.Lock()
	delete(s.nodes, nodeID)
	s.nodesMutex.Unlock()

	s.resultsMutex.Lock()
	delete(s.lastCheckResults, nodeID)
	s.resultsMutex.Unlock()

	return nil
}

func (s *service) NodeExists(nodeID int64) bool {
	s.nodesMutex.RLock()
	defer s.nodesMutex.RUnlock()
	_, exists := s.nodes[nodeID]
	return exists
}

// GetNodeStatus returns the current status of a node
func (s *service) GetNodeStatus(nodeID int64) (*NodeCheck, error) {
	s.resultsMutex.RLock()
	defer s.resultsMutex.RUnlock()

	result, exists := s.lastCheckResults[nodeID]
	if !exists {
		return nil, fmt.Errorf("node with ID %d not found", nodeID)
	}

	return result, nil
}

// GetAllNodeStatuses returns the current status of all nodes
func (s *service) GetAllNodeStatuses() []*NodeCheck {
	s.resultsMutex.RLock()
	defer s.resultsMutex.RUnlock()

	results := make([]*NodeCheck, 0, len(s.lastCheckResults))
	for _, result := range s.lastCheckResults {
		results = append(results, result)
	}

	return results
}

// worker is a background worker that checks nodes periodically
func (s *service) worker(ctx context.Context, workerID int) {
	defer s.workerWaitGroup.Done()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.checkNodes(ctx)
		}
	}
}

// checkNodes performs the actual node checks
func (s *service) checkNodes(ctx context.Context) {
	now := time.Now()

	// Get a copy of the nodes to check
	s.nodesMutex.RLock()
	nodesToCheck := make([]*Node, 0)
	for _, node := range s.nodes {
		if now.Sub(node.LastChecked) >= node.CheckInterval {
			nodesToCheck = append(nodesToCheck, node)
		}
	}
	s.nodesMutex.RUnlock()

	// Check each node
	for _, node := range nodesToCheck {
		s.checkNode(ctx, node)
	}
}

// checkNode checks a single node and updates its status
func (s *service) checkNode(ctx context.Context, node *Node) {
	start := time.Now()

	// Create a context with the node's timeout
	checkCtx, cancel := context.WithTimeout(ctx, node.Timeout)
	defer cancel()

	// Create a request to check the node
	req, err := http.NewRequestWithContext(checkCtx, http.MethodGet, node.URL, nil)
	if err != nil {
		s.handleNodeCheckResult(node, NodeStatusDown, 0, err)
		return
	}

	// Perform the request
	resp, err := s.httpClient.Do(req)
	responseTime := time.Since(start)

	if err != nil {
		s.handleNodeCheckResult(node, NodeStatusDown, responseTime, err)
		return
	}
	defer resp.Body.Close()

	// Check if the response is successful (2xx status code)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		err = fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		s.handleNodeCheckResult(node, NodeStatusDown, responseTime, err)
		return
	}

	// Node is up
	s.handleNodeCheckResult(node, NodeStatusUp, responseTime, nil)
}

// handleNodeCheckResult processes the result of a node check
func (s *service) handleNodeCheckResult(node *Node, status NodeStatus, responseTime time.Duration, err error) {
	now := time.Now()

	// Create the check result
	checkResult := &NodeCheck{
		Node:         node,
		Status:       status,
		ResponseTime: responseTime,
		Error:        err,
		Timestamp:    now,
	}

	// Update the node's status
	s.nodesMutex.Lock()

	// Capture the previous status for status change detection
	previousStatus := node.Status
	wasDown := previousStatus == NodeStatusDown

	// Update last checked time
	node.LastChecked = now

	// Handle status changes
	statusChanged := previousStatus != status
	if statusChanged {
		node.LastStatusChange = now
	}

	// Update failure count
	if status == NodeStatusDown {
		node.FailureCount++
	} else {
		node.FailureCount = 0
	}

	// Update node status
	node.Status = status

	// If node was down but is now up, record the recovery time
	var recoveryTime time.Time
	var downtimeDuration time.Duration
	if wasDown && status == NodeStatusUp {
		recoveryTime = now
		downtimeDuration = now.Sub(node.LastStatusChange)
	}

	s.nodesMutex.Unlock()

	// Store the check result
	s.resultsMutex.Lock()
	s.lastCheckResults[node.ID] = checkResult
	s.resultsMutex.Unlock()

	// Send notifications if needed
	ctx := context.Background()
	if status == NodeStatusDown && node.FailureCount >= node.FailureThreshold {
		s.sendNodeDownNotification(ctx, node, err)
	} else if statusChanged && status == NodeStatusUp && wasDown {
		// Node has recovered - send recovery notification
		s.sendNodeRecoveryNotification(ctx, node, responseTime, recoveryTime, downtimeDuration)
	}
}

// formatDuration formats a duration into a human-readable string
func formatDuration(d time.Duration) string {
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	parts := []string{}
	if days > 0 {
		parts = append(parts, fmt.Sprintf("%dd", days))
	}
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%dh", hours))
	}
	if minutes > 0 {
		parts = append(parts, fmt.Sprintf("%dm", minutes))
	}
	if seconds > 0 || len(parts) == 0 {
		parts = append(parts, fmt.Sprintf("%ds", seconds))
	}

	return strings.Join(parts, " ")
}

// sendNodeDownNotification sends a notification that a node is down
func (s *service) sendNodeDownNotification(ctx context.Context, node *Node, err error) {
	data := notifications.NodeDowntimeData{
		NodeID:       node.ID,
		NodeName:     node.Name,
		NodeURL:      node.URL,
		DownSince:    node.LastStatusChange,
		FailureCount: node.FailureCount,
		Error:        err.Error(),
	}

	// Send the notification
	if err := s.notificationSvc.SendNodeDowntimeNotification(ctx, data); err != nil {
		// Just log the error; we don't want to create a notification loop
		fmt.Printf("Failed to send node downtime notification: %v\n", err)
	}
}

// sendNodeRecoveryNotification sends a notification that a node has recovered
func (s *service) sendNodeRecoveryNotification(ctx context.Context, node *Node, responseTime time.Duration, recoveryTime time.Time, downtimeDuration time.Duration) {
	data := notifications.NodeUpData{
		NodeID:           node.ID,
		NodeName:         node.Name,
		NodeURL:          node.URL,
		DownSince:        node.LastStatusChange,
		RecoveredAt:      recoveryTime,
		DowntimeDuration: downtimeDuration,
		ResponseTime:     responseTime,
	}

	// Send the notification
	if err := s.notificationSvc.SendNodeRecoveryNotification(ctx, data); err != nil {
		// Just log the error; we don't want to create a notification loop
		fmt.Printf("Failed to send node recovery notification: %v\n", err)
	}
}
