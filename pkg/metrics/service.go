package metrics

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/chainlaunch/chainlaunch/pkg/db"
	nodeservice "github.com/chainlaunch/chainlaunch/pkg/nodes/service"
)

// Service defines the interface for metrics operations
type Service interface {
	// Start starts the Prometheus instance with the given configuration
	Start(ctx context.Context, config *Config) error

	// Stop stops the Prometheus instance
	Stop(ctx context.Context) error

	// QueryMetrics retrieves metrics for a specific node
	QueryMetrics(ctx context.Context, nodeID int64, query string) (map[string]interface{}, error)

	// QueryMetricsRange retrieves metrics for a specific node within a time range
	QueryMetricsRange(ctx context.Context, nodeID int64, query string, start, end time.Time, step time.Duration) (map[string]interface{}, error)

	// GetLabelValues retrieves values for a specific label
	GetLabelValues(ctx context.Context, nodeID int64, labelName string, matches []string) ([]string, error)

	// Reload reloads the Prometheus configuration
	Reload(ctx context.Context) error
}

// service implements the Service interface
type service struct {
	manager     *PrometheusManager
	nodeService *nodeservice.NodeService
}

// NewService creates a new metrics service
func NewService(config *Config, db *db.Queries, nodeService *nodeservice.NodeService) (Service, error) {
	manager, err := NewPrometheusManager(config, db, nodeService)
	if err != nil {
		return nil, err
	}
	return &service{
		manager:     manager,
		nodeService: nodeService,
	}, nil
}

// Start starts the Prometheus instance
func (s *service) Start(ctx context.Context, config *Config) error {
	return s.manager.Start(ctx)
}

// Stop stops the Prometheus instance
func (s *service) Stop(ctx context.Context) error {
	return s.manager.Stop(ctx)
}

// QueryMetrics retrieves metrics for a specific node
func (s *service) QueryMetrics(ctx context.Context, nodeID int64, query string) (map[string]interface{}, error) {
	// Get node type and create job name
	node, err := s.nodeService.GetNodeByID(ctx, nodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get node: %w", err)
	}
	jobName := slugify(fmt.Sprintf("%d-%s", node.ID, node.Name))

	// If no query is provided, use default metrics
	if query == "" {
		query = fmt.Sprintf(`{job="%s"}`, jobName)
	} else {
		// If query is provided, it's just a label, so add job filter
		query = fmt.Sprintf(`%s{job="%s"}`, query, jobName)
	}

	// Query Prometheus for metrics
	result, err := s.manager.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query metrics: %w", err)
	}

	return map[string]interface{}{
		"node_id": nodeID,
		"job":     jobName,
		"query":   query,
		"result":  result,
	}, nil
}

// QueryMetricsRange retrieves metrics for a specific node within a time range
func (s *service) QueryMetricsRange(ctx context.Context, nodeID int64, query string, start, end time.Time, step time.Duration) (map[string]interface{}, error) {
	node, err := s.nodeService.GetNodeByID(ctx, nodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get node: %w", err)
	}
	jobName := slugify(fmt.Sprintf("%d-%s", node.ID, node.Name))

	// Add job filter to query
	if !strings.Contains(query, "job=") {
		query = fmt.Sprintf(`%s{job="%s"}`, query, jobName)
	}

	// Query Prometheus for metrics with time range
	result, err := s.manager.QueryRange(ctx, query, start, end, step)
	if err != nil {
		return nil, fmt.Errorf("failed to query metrics range: %w", err)
	}

	return map[string]interface{}{
		"node_id": nodeID,
		"job":     jobName,
		"query":   query,
		"result":  result,
	}, nil
}

// GetLabelValues retrieves values for a specific label
func (s *service) GetLabelValues(ctx context.Context, nodeID int64, labelName string, matches []string) ([]string, error) {
	node, err := s.nodeService.GetNodeByID(ctx, nodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get node: %w", err)
	}
	jobName := slugify(fmt.Sprintf("%s-%d", node.NodeType, node.ID))

	// Add job filter to matches
	jobMatch := fmt.Sprintf(`{job="%s"}`, jobName)
	_ = jobMatch
	// matches = append(matches, jobMatch)

	result, err := s.manager.GetLabelValues(ctx, labelName, matches)
	if err != nil {
		return nil, fmt.Errorf("failed to get label values: %w", err)
	}
	return result, nil
}

// Reload reloads the Prometheus configuration
func (s *service) Reload(ctx context.Context) error {
	return s.manager.Reload(ctx)
}
