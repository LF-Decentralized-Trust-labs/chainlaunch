package peer

import (
	"github.com/chainlaunch/chainlaunch/pkg/logger"
	"github.com/spf13/cobra"
)

// NewPeerCmd returns the peer command
func NewPeerCmd(logger *logger.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "peer",
		Short: "Manage Fabric peer nodes",
		Long:  `Manage Hyperledger Fabric peer nodes`,
	}

	cmd.AddCommand(NewCreateCmd(logger))
	cmd.AddCommand(NewDeleteCmd(logger))
	cmd.AddCommand(NewListCmd(logger))
	cmd.AddCommand(NewUpdateCmd(logger))

	return cmd
}
