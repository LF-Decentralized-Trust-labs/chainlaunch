package testnet

import (
	"github.com/spf13/cobra"
	// Import subcommands for each network type
	"github.com/chainlaunch/chainlaunch/cmd/testnet/besu"
	"github.com/chainlaunch/chainlaunch/cmd/testnet/fabric"
)

// NewTestnetCmd returns the root testnet command
func NewTestnetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "testnet",
		Short: "Manage testnets for different blockchain networks",
	}

	// Add subcommands for each network type
	cmd.AddCommand(fabric.NewFabricTestnetCmd())
	cmd.AddCommand(besu.NewBesuTestnetCmd())

	return cmd
}
