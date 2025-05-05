package org

import (
	"fmt"
	"io"
	"os"

	"github.com/chainlaunch/chainlaunch/pkg/logger"
	"github.com/spf13/cobra"
)

type createCmd struct {
	name       string
	mspID      string
	providerID int64
	logger     *logger.Logger
}

func (c *createCmd) validate() error {
	if c.name == "" {
		return fmt.Errorf("organization name is required")
	}
	if c.mspID == "" {
		return fmt.Errorf("MSP ID is required")
	}
	return nil
}

func (c *createCmd) run(out io.Writer) error {
	client := NewClientWrapper(c.logger)
	return client.CreateOrganization(c.name, c.mspID, c.providerID)
}

// NewCreateCmd returns the create organization command
func NewCreateCmd(logger *logger.Logger) *cobra.Command {
	c := &createCmd{
		logger: logger,
	}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new organization",
		Long:  `Create a new Hyperledger Fabric organization`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := c.validate(); err != nil {
				return err
			}
			return c.run(os.Stdout)
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&c.name, "name", "n", "", "Organization name")
	flags.StringVarP(&c.mspID, "msp-id", "m", "", "MSP ID")
	flags.Int64VarP(&c.providerID, "provider-id", "p", 0, "Key management provider ID")

	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("msp-id")
	cmd.MarkFlagRequired("provider-id")

	return cmd
}
