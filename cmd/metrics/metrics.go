package metrics

import (
	"github.com/spf13/cobra"
)

// NewMetricsCmd creates the 'metrics' parent command
func NewMetricsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "metrics",
		Short: "Manage metrics and Prometheus integration",
	}
	cmd.AddCommand(NewEnableCmd())
	cmd.AddCommand(NewDisableCmd())
	return cmd
}
