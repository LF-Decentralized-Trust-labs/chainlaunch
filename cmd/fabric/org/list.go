package org

import (
	"fmt"
	"io"
	"os"

	"github.com/chainlaunch/chainlaunch/pkg/logger"
	"github.com/spf13/cobra"
)

type listCmd struct {
	baseURL  string
	username string
	password string
	logger   *logger.Logger
}

func (c *listCmd) validate() error {
	if c.username == "" {
		return fmt.Errorf("username is required")
	}
	if c.password == "" {
		return fmt.Errorf("password is required")
	}
	return nil
}

func (c *listCmd) run(out io.Writer) error {
	client := NewClientWrapper(c.baseURL, c.username, c.password, c.logger)
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
			if err := c.validate(); err != nil {
				return err
			}
			return c.run(os.Stdout)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&c.baseURL, "url", "", "Base URL of the API server (defaults to CHAINLAUNCH_URL env var or http://localhost:8100/api/v1)")
	flags.StringVarP(&c.username, "username", "u", "", "Username for basic auth")
	flags.StringVarP(&c.password, "password", "p", "", "Password for basic auth")

	cmd.MarkFlagRequired("username")
	cmd.MarkFlagRequired("password")

	return cmd
}
