package orderer

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/chainlaunch/chainlaunch/cmd/common"
	"github.com/chainlaunch/chainlaunch/pkg/logger"
	"github.com/spf13/cobra"
)

type deleteCmd struct {
	id     int64
	yes    bool
	logger *logger.Logger
}

func (c *deleteCmd) validate() error {
	if c.id == 0 {
		return fmt.Errorf("node ID is required")
	}
	return nil
}

func (c *deleteCmd) run(out io.Writer) error {
	client, err := common.NewClientFromEnv()
	if err != nil {
		return err
	}

	// Confirm deletion if --yes flag is not set
	if !c.yes {
		fmt.Fprintf(out, "Are you sure you want to delete orderer node with ID %d? [y/N]: ", c.id)
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read user input: %w", err)
		}

		response = strings.ToLower(strings.TrimSpace(response))
		if response != "y" && response != "yes" {
			fmt.Fprintln(out, "Operation cancelled")
			return nil
		}
	}

	// Delete orderer node
	if err := client.DeleteNode(c.id); err != nil {
		return fmt.Errorf("failed to delete orderer node: %w", err)
	}

	fmt.Fprintf(out, "Orderer node with ID %d deleted successfully\n", c.id)
	return nil
}

// NewDeleteCmd returns the delete orderer command
func NewDeleteCmd(logger *logger.Logger) *cobra.Command {
	c := &deleteCmd{
		logger: logger,
	}

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete an orderer node",
		Long:  `Delete a Hyperledger Fabric orderer node`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := c.validate(); err != nil {
				return err
			}
			return c.run(os.Stdout)
		},
	}

	flags := cmd.Flags()
	flags.Int64VarP(&c.id, "id", "i", 0, "Node ID")
	flags.BoolVarP(&c.yes, "yes", "y", false, "Skip confirmation")

	cmd.MarkFlagRequired("id")

	return cmd
}
