package org

import (
	"github.com/chainlaunch/chainlaunch/pkg/logger"
	"github.com/spf13/cobra"
)

// NewOrgCmd returns the organization command
func NewOrgCmd(logger *logger.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "org",
		Short: "Organization management commands",
		Long:  `Commands for managing Hyperledger Fabric organizations`,
	}

	cmd.AddCommand(
		NewCreateCmd(logger),
		NewListCmd(logger),
		NewDeleteCmd(logger),
		NewUpdateCmd(logger),
	)

	return cmd
}
