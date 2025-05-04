package org

import (
	"fmt"
	"io"
	"os"

	"github.com/chainlaunch/chainlaunch/pkg/logger"
	"github.com/spf13/cobra"
)

type updateCmd struct {
	mspID    string
	name     string
	domain   string
	baseURL  string
	username string
	password string
	logger   *logger.Logger
}

func (c *updateCmd) validate() error {
	if c.mspID == "" {
		return fmt.Errorf("MSP ID is required")
	}
	if c.username == "" {
		return fmt.Errorf("username is required")
	}
	if c.password == "" {
		return fmt.Errorf("password is required")
	}
	return nil
}

func (c *updateCmd) run(out io.Writer) error {
	client := NewClientWrapper(c.baseURL, c.username, c.password, c.logger)
	return client.UpdateOrganization(c.mspID, c.name, c.domain)
}

// NewUpdateCmd returns the update organization command
func NewUpdateCmd(logger *logger.Logger) *cobra.Command {
	c := &updateCmd{
		logger: logger,
	}

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update an organization",
		Long:  `Update a Hyperledger Fabric organization`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := c.validate(); err != nil {
				return err
			}
			return c.run(os.Stdout)
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&c.mspID, "msp-id", "m", "", "MSP ID of the organization to update")
	flags.StringVarP(&c.name, "name", "n", "", "New organization name")
	flags.StringVarP(&c.domain, "domain", "d", "", "New organization domain")
	flags.StringVar(&c.baseURL, "url", "", "Base URL of the API server (defaults to CHAINLAUNCH_URL env var or http://localhost:8100/api/v1)")
	flags.StringVarP(&c.username, "username", "u", "", "Username for basic auth")
	flags.StringVarP(&c.password, "password", "p", "", "Password for basic auth")

	cmd.MarkFlagRequired("msp-id")
	cmd.MarkFlagRequired("username")
	cmd.MarkFlagRequired("password")

	return cmd
}
