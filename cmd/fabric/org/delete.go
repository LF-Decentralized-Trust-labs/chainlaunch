package org

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/chainlaunch/chainlaunch/pkg/logger"
	"github.com/spf13/cobra"
)

type deleteCmd struct {
	mspID  string
	yes    bool
	logger *logger.Logger
}

func (c *deleteCmd) validate() error {
	if c.mspID == "" {
		return fmt.Errorf("MSP ID is required")
	}
	return nil
}

func (c *deleteCmd) run(out io.Writer) error {
	client := NewClientWrapper(c.logger)

	// First check if the organization exists
	org, err := client.client.GetOrganizationByMspID(c.mspID)
	if err != nil {
		return fmt.Errorf("failed to get organization: %w", err)
	}

	// Skip confirmation if --yes flag is set
	if !c.yes {
		// Ask for confirmation
		fmt.Fprintf(out, "Are you sure you want to delete organization %s (MSP ID: %s)? [y/N]: ", org.Description, org.MspID)
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read confirmation: %w", err)
		}

		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Fprintln(out, "Deletion cancelled")
			return nil
		}
	}

	return client.DeleteOrganization(c.mspID)
}

// NewDeleteCmd returns the delete organization command
func NewDeleteCmd(logger *logger.Logger) *cobra.Command {
	c := &deleteCmd{
		logger: logger,
	}

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete an organization",
		Long:  `Delete a Hyperledger Fabric organization`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := c.validate(); err != nil {
				return err
			}
			return c.run(os.Stdout)
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&c.mspID, "msp-id", "m", "", "MSP ID of the organization to delete")
	flags.BoolVar(&c.yes, "yes", false, "Skip confirmation prompt")

	cmd.MarkFlagRequired("msp-id")

	return cmd
}
