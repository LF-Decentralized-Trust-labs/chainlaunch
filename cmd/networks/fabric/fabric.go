package fabric

import (
	"github.com/chainlaunch/chainlaunch/pkg/logger"
	"github.com/spf13/cobra"
)

// NewFabricCmd returns the fabric command
func NewFabricCmd(logger *logger.Logger) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "fabric",
		Short: "Manage Fabric networks",
		Long:  `Create, update, and manage Hyperledger Fabric networks.`,
	}

	rootCmd.AddCommand(
		newCreateCmd(logger),
		newUpdateCmd(logger),
	)

	return rootCmd
}
