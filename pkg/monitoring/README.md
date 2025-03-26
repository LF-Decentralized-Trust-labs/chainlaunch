# Node Monitoring Module

This module provides functionality to monitor nodes by periodically checking their health status and sending alerts when issues are detected.

## Features

- Periodically checks node health status via HTTP requests
- Configurable check intervals, timeouts, and failure thresholds
- Sends notifications when nodes go down using the notification service
- Sends recovery notifications when nodes come back online
- Concurrent monitoring with configurable worker pool
- Tracks node status history and provides API to query current status

## Usage

### Basic Setup

```go
import (
    "context"
    "time"
    
    "github.com/davidviejo/projects/kfs/chainlaunch/pkg/monitoring"
    "github.com/davidviejo/projects/kfs/chainlaunch/pkg/notifications"
)

// Create a notification service implementation
notificationSvc := yourNotificationServiceImplementation

// Configure the monitoring service
config := &monitoring.Config{
    DefaultCheckInterval:    1 * time.Minute,    // Check every minute
    DefaultTimeout:          10 * time.Second,   // 10 second timeout
    DefaultFailureThreshold: 3,                  // Alert after 3 failures
    Workers:                 5,                  // Use 5 worker goroutines
}

// Create and start the monitoring service
monitoringSvc := monitoring.NewService(config, notificationSvc)
if err := monitoringSvc.Start(context.Background()); err != nil {
    // Handle error
}

// Add nodes to monitor
node := &monitoring.Node{
    ID:   "node1",
    Name: "Example Node",
    URL:  "https://example.com/health",
    // Optional: customize per-node settings
    CheckInterval:    30 * time.Second,  // Check more frequently
    Timeout:          5 * time.Second,   // Custom timeout
    FailureThreshold: 2,                 // Alert sooner
}

if err := monitoringSvc.AddNode(node); err != nil {
    // Handle error
}

// Query node status
status, err := monitoringSvc.GetNodeStatus("node1")
if err != nil {
    // Handle error
}
fmt.Printf("Node status: %s\n", status.Status)

// When shutting down
if err := monitoringSvc.Stop(); err != nil {
    // Handle error
}
```

### Custom Node Checks

The default node check performs an HTTP GET request and expects a 2xx status code. You can extend the service to implement custom checks for specific node types by creating your own monitoring service implementation.

## Integration with Notification System

The monitoring service integrates with the notifications package to send alerts when nodes go down or recover. It uses:

1. `SendNodeDowntimeNotification` method when a node fails to respond for a number of consecutive times equal to or greater than its failure threshold. The notification includes:
   - Node ID and name
   - Node URL
   - Time since the node went down
   - Number of consecutive failures
   - Error details

2. `SendNodeRecoveryNotification` method when a previously down node becomes available again. The recovery notification includes:
   - Node ID and name
   - Node URL
   - When the node went down
   - When the node recovered
   - Total downtime duration
   - Current response time

This complete notification system ensures that operators are aware of both outages and service restorations, providing comprehensive monitoring coverage.

## Example

See `cmd/monitoring/main.go` for a complete working example. 