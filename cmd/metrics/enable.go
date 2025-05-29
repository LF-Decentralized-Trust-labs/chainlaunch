// Package metrics provides the 'metrics enable' command to deploy Prometheus for metrics collection.
package metrics

import (
	"fmt"
	"time"

	"github.com/chainlaunch/chainlaunch/cmd/common"
	metricscommon "github.com/chainlaunch/chainlaunch/pkg/metrics/common"
	"github.com/spf13/cobra"
)

// EnableMetricsConfig holds the parameters for enabling metrics
type EnableMetricsConfig struct {
	PrometheusVersion string
	PrometheusPort    int
	ScrapeInterval    int
}

// EnableMetricsRunner encapsulates the config and logic for running and validating the enable metrics command
type EnableMetricsRunner struct {
	Config EnableMetricsConfig
}

// Validate checks the configuration for required fields
func (r *EnableMetricsRunner) Validate() error {
	if r.Config.PrometheusVersion == "" {
		return fmt.Errorf("--version is required")
	}
	if r.Config.PrometheusPort <= 0 {
		return fmt.Errorf("--port must be greater than 0")
	}
	if r.Config.ScrapeInterval <= 0 {
		return fmt.Errorf("--scrape-interval must be greater than 0")
	}
	return nil
}

// Run executes the enable metrics logic
func (r *EnableMetricsRunner) Run() error {
	if err := r.Validate(); err != nil {
		return err
	}
	client, err := common.NewClientFromEnv()
	if err != nil {
		return err
	}
	status, err := client.GetMetricsStatus()
	if err == nil && status.Status == "running" {
		return fmt.Errorf("Prometheus is already deployed (status: running)")
	}
	cfg := &metricscommon.Config{
		PrometheusVersion: r.Config.PrometheusVersion,
		PrometheusPort:    r.Config.PrometheusPort,
		ScrapeInterval:    time.Duration(r.Config.ScrapeInterval) * time.Second,
	}
	resp, err := client.EnableMetrics(cfg)
	if err != nil {
		return err
	}
	fmt.Printf("Prometheus deployed: %v\n", resp)
	return nil
}

func NewEnableCmd() *cobra.Command {
	runner := &EnableMetricsRunner{
		Config: EnableMetricsConfig{
			PrometheusVersion: "v3.4.0",
			PrometheusPort:    9090,
			ScrapeInterval:    15,
		},
	}

	cmd := &cobra.Command{
		Use:   "enable",
		Short: "Enable metrics by deploying Prometheus",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runner.Run()
		},
	}

	cmd.Flags().StringVar(&runner.Config.PrometheusVersion, "version", "v3.3.1", "Prometheus version")
	cmd.Flags().IntVar(&runner.Config.PrometheusPort, "port", 9090, "Prometheus port")
	cmd.Flags().IntVar(&runner.Config.ScrapeInterval, "scrape-interval", 15, "Scrape interval in seconds")

	return cmd
}
