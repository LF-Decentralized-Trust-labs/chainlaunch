package metrics

import (
	"fmt"

	"github.com/chainlaunch/chainlaunch/cmd/common"
	"github.com/spf13/cobra"
)

// DisableMetricsRunner encapsulates the logic for disabling metrics
type DisableMetricsRunner struct{}

func (r *DisableMetricsRunner) Run() error {
	client, err := common.NewClientFromEnv()
	if err != nil {
		return err
	}
	status, err := client.GetMetricsStatus()
	if err != nil {
		return fmt.Errorf("could not get metrics status: %w", err)
	}
	if status.Status != "running" {
		return fmt.Errorf("Prometheus is not running (status: %s)", status.Status)
	}
	resp, err := client.DisableMetrics()
	if err != nil {
		return err
	}
	fmt.Printf("Prometheus undeployed: %v\n", resp)
	return nil
}

func NewDisableCmd() *cobra.Command {
	runner := &DisableMetricsRunner{}
	cmd := &cobra.Command{
		Use:   "disable",
		Short: "Disable metrics by undeploying Prometheus",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runner.Run()
		},
	}
	return cmd
}
