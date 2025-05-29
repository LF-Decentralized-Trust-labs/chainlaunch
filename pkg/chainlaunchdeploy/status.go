package chainlaunchdeploy

import (
	"sync"
)

// DeploymentStatus represents the current status of a deployment operation.
type DeploymentStatus string

const (
	StatusPending DeploymentStatus = "pending"
	StatusRunning DeploymentStatus = "running"
	StatusSuccess DeploymentStatus = "success"
	StatusFailed  DeploymentStatus = "failed"
)

// DeploymentStatusUpdate represents a status update for a deployment.
type DeploymentStatusUpdate struct {
	DeploymentID string           // Unique identifier for the deployment (e.g., tx hash, chaincode name+version)
	Status       DeploymentStatus // Current status
	Message      string           // Optional message or log
	Error        error            // Optional error
}

// DeploymentStatusReporter defines the interface for reporting and querying deployment status.
type DeploymentStatusReporter interface {
	ReportStatus(update DeploymentStatusUpdate)
	GetStatus(deploymentID string) DeploymentStatusUpdate
}

// InMemoryDeploymentStatusReporter is a simple in-memory implementation.
type InMemoryDeploymentStatusReporter struct {
	mu     sync.RWMutex
	status map[string]DeploymentStatusUpdate
}

func NewInMemoryDeploymentStatusReporter() *InMemoryDeploymentStatusReporter {
	return &InMemoryDeploymentStatusReporter{
		status: make(map[string]DeploymentStatusUpdate),
	}
}

func (r *InMemoryDeploymentStatusReporter) ReportStatus(update DeploymentStatusUpdate) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.status[update.DeploymentID] = update
}

func (r *InMemoryDeploymentStatusReporter) GetStatus(deploymentID string) DeploymentStatusUpdate {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.status[deploymentID]
}

// Usage:
// 1. Create a reporter: reporter := NewInMemoryDeploymentStatusReporter()
// 2. During deployment, call reporter.ReportStatus(...) at each stage.
// 3. Query status with reporter.GetStatus(deploymentID).
