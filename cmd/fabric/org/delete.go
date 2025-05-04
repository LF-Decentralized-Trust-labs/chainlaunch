package org

import (
	"fmt"
	"io"
	"os"

	"github.com/chainlaunch/chainlaunch/pkg/logger"
	"github.com/spf13/cobra"
)

type deleteCmd struct {
	mspID    string
	baseURL  string
	username string
	password string
	logger   *logger.Logger
}

func (c *deleteCmd) validate() error {
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

func (c *deleteCmd) run(out io.Writer) error {
	client := NewClientWrapper(c.baseURL, c.username, c.password, c.logger)
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
	flags.StringVar(&c.baseURL, "url", "", "Base URL of the API server (defaults to CHAINLAUNCH_URL env var or http://localhost:8100/api/v1)")
	flags.StringVarP(&c.username, "username", "u", "", "Username for basic auth")
	flags.StringVarP(&c.password, "password", "p", "", "Password for basic auth")

	cmd.MarkFlagRequired("msp-id")
	cmd.MarkFlagRequired("username")
	cmd.MarkFlagRequired("password")

	return cmd
}
