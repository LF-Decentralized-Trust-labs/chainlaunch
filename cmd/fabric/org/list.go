package org

import (
	"io"
	"os"

	"github.com/chainlaunch/chainlaunch/pkg/logger"
	"github.com/spf13/cobra"
)

type listCmd struct {
	logger *logger.Logger
}

func (c *listCmd) run(out io.Writer) error {
	client := NewClientWrapper(c.logger)
	return client.ListOrganizations(out)
}

// NewListCmd returns the list organizations command
func NewListCmd(logger *logger.Logger) *cobra.Command {
	c := &listCmd{
		logger: logger,
	}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List organizations",
		Long:  `List all Hyperledger Fabric organizations`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.run(os.Stdout)
		},
	}

	return cmd
}
