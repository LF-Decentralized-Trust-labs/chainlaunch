package networks

import (
	"github.com/chainlaunch/chainlaunch/cmd/networks/besu"
	"github.com/chainlaunch/chainlaunch/cmd/networks/fabric"
	"github.com/chainlaunch/chainlaunch/pkg/logger"
	"github.com/spf13/cobra"
)

// NetworksCmd represents the networks command
func NewNetworksCmd(logger *logger.Logger) *cobra.Command {
	networksCmd := &cobra.Command{
		Use:   "networks",
		Short: "Manage blockchain networks",
		Long:  `Create and manage blockchain networks for different platforms (Besu, Fabric).`,
	}

	// Add subcommands
	networksCmd.AddCommand(fabric.NewFabricCmd(logger))
	networksCmd.AddCommand(besu.NewBesuCmd(logger))

	return networksCmd
}
