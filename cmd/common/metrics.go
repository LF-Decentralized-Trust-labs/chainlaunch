package common

import (
	"encoding/json"
	"fmt"
	"net/http"

	metricscommon "github.com/chainlaunch/chainlaunch/pkg/metrics/common"
	metricstypes "github.com/chainlaunch/chainlaunch/pkg/metrics/types"
)

// EnableMetrics deploys Prometheus using the provided parameters
func (c *Client) EnableMetrics(req *metricscommon.Config) (map[string]string, error) {
	// Convert common.Config to the HTTP request struct expected by the API
	httpReq := metricstypes.DeployPrometheusRequest{
		PrometheusVersion: req.PrometheusVersion,
		PrometheusPort:    req.PrometheusPort,
		ScrapeInterval:    int(req.ScrapeInterval.Seconds()),
	}
	resp, err := c.Post("/metrics/deploy", httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call /metrics/deploy: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("unexpected status %d: %v", resp.StatusCode, errResp)
	}
	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return result, nil
}

// DisableMetrics undeploys Prometheus
func (c *Client) DisableMetrics() (map[string]string, error) {
	resp, err := c.Post("/metrics/undeploy", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to call /metrics/undeploy: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("unexpected status %d: %v", resp.StatusCode, errResp)
	}
	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return result, nil
}

// GetMetricsStatus returns the current status of the Prometheus instance
func (c *Client) GetMetricsStatus() (*metricscommon.Status, error) {
	resp, err := c.Get("/metrics/status")
	if err != nil {
		return nil, fmt.Errorf("failed to call /metrics/status: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("unexpected status %d: %v", resp.StatusCode, errResp)
	}
	var status metricscommon.Status
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &status, nil
}
