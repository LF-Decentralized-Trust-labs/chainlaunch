package besu

import (
	"github.com/chainlaunch/chainlaunch/pkg/logger"
	"github.com/spf13/cobra"
)

// NewBesuCmd returns the besu command
func NewBesuCmd(logger *logger.Logger) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "besu",
		Short: "Manage Besu networks",
		Long:  `Create, update, and manage Hyperledger Besu networks.`,
	}

	rootCmd.AddCommand(
		newCreateCmd(logger),
		newUpdateCmd(logger),
	)

	return rootCmd
}
