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
