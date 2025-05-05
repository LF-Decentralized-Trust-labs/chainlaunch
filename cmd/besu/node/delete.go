package node

import (
	"fmt"
	"os"

	"github.com/chainlaunch/chainlaunch/cmd/common"
	"github.com/chainlaunch/chainlaunch/pkg/logger"
	"github.com/spf13/cobra"
)

type deleteCmd struct {
	id     int64
	logger *logger.Logger
}

func (c *deleteCmd) validate() error {
	if c.id == 0 {
		return fmt.Errorf("node ID is required")
	}
	return nil
}

func (c *deleteCmd) run(out *os.File) error {
	client, err := common.NewClientFromEnv()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	if err := client.DeleteNode(c.id); err != nil {
		return fmt.Errorf("failed to delete Besu node: %w", err)
	}

	fmt.Fprintf(out, "Successfully deleted Besu node with ID: %d\n", c.id)
	return nil
}

// NewDeleteCmd returns the delete Besu node command
func NewDeleteCmd(logger *logger.Logger) *cobra.Command {
	c := &deleteCmd{
		logger: logger,
	}

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a Besu node",
		Long:  `Delete a Besu node by its ID`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := c.validate(); err != nil {
				return err
			}
			return c.run(os.Stdout)
		},
	}

	flags := cmd.Flags()
	flags.Int64VarP(&c.id, "id", "i", 0, "ID of the node to delete")
	cmd.MarkFlagRequired("id")

	return cmd
}
