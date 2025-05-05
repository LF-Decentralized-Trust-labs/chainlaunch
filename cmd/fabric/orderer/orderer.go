package orderer

import (
	"github.com/chainlaunch/chainlaunch/pkg/logger"
	"github.com/spf13/cobra"
)

// NewOrdererCmd returns the orderer command
func NewOrdererCmd(logger *logger.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "orderer",
		Short: "Manage Fabric orderer nodes",
		Long:  `Manage Hyperledger Fabric orderer nodes`,
	}

	cmd.AddCommand(NewCreateCmd(logger))
	cmd.AddCommand(NewDeleteCmd(logger))
	cmd.AddCommand(NewListCmd(logger))
	cmd.AddCommand(NewUpdateCmd(logger))

	return cmd
}
