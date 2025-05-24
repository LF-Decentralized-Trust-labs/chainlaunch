package types

import "time"

// DeployPrometheusRequest represents the request to deploy Prometheus
// Used for HTTP API data transfer
// See handler.go for usage
type DeployPrometheusRequest struct {
	PrometheusVersion string `json:"prometheus_version" binding:"required"`
	PrometheusPort    int    `json:"prometheus_port" binding:"required"`
	ScrapeInterval    int    `json:"scrape_interval" binding:"required"`
}

// RefreshNodesRequest represents the request to refresh nodes
// Used for HTTP API data transfer
// See handler.go for usage
type RefreshNodesRequest struct {
	Nodes []struct {
		ID      string `json:"id" binding:"required"`
		Address string `json:"address" binding:"required"`
		Port    int    `json:"port" binding:"required"`
	} `json:"nodes" binding:"required"`
}

// CustomQueryRequest represents the request body for custom Prometheus queries
type CustomQueryRequest struct {
	Query string     `json:"query" binding:"required"`
	Start *time.Time `json:"start,omitempty"`
	End   *time.Time `json:"end,omitempty"`
	Step  *string    `json:"step,omitempty"`
}

// MessageResponse is a generic response for endpoints that return a message
// Example: {"message": "Prometheus deployed successfully"}
type MessageResponse struct {
	Message string `json:"message"`
}

// LabelValuesResponse is the response for label values endpoints
// Example: {"status": "success", "data": ["val1", "val2"]}
type LabelValuesResponse struct {
	Status string   `json:"status"`
	Data   []string `json:"data"`
}

// MetricsDataResponse is the response for metrics data endpoints (range, etc.)
// Example: {"status": "success", "data": ...}
type MetricsDataResponse struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data"`
}
