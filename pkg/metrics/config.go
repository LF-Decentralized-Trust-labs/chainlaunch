package metrics

import (
	"time"
)

// Config represents the configuration for the metrics service
type Config struct {
	// PrometheusVersion is the version of Prometheus to deploy
	PrometheusVersion string
	// PrometheusPort is the port Prometheus will listen on
	PrometheusPort int
	// ScrapeInterval is the interval between scrapes
	ScrapeInterval time.Duration
	// DeploymentMode specifies how Prometheus is deployed (currently only supports "docker")
	DeploymentMode string
}

// DefaultConfig returns a Config with sensible default values
func DefaultConfig() *Config {
	return &Config{
		PrometheusVersion: "v3.3.1",
		PrometheusPort:    9090,
		ScrapeInterval:    15 * time.Second,
		DeploymentMode:    "docker",
	}
}
