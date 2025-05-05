package besu

import (
	"github.com/chainlaunch/chainlaunch/cmd/besu/node"
	"github.com/chainlaunch/chainlaunch/pkg/logger"
	"github.com/spf13/cobra"
)

// NewBesuCmd returns the root command for Besu operations
func NewBesuCmd(logger *logger.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "besu",
		Short: "Manage Besu nodes",
		Long:  `Create, list, update, and delete Besu nodes`,
	}

	// Add node management commands
	cmd.AddCommand(
		node.NewCreateCmd(logger),
		node.NewListCmd(logger),
		node.NewUpdateCmd(logger),
		node.NewDeleteCmd(logger),
	)

	return cmd
}
