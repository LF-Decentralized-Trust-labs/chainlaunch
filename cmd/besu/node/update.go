package node

import (
	"fmt"
	"os"

	"github.com/chainlaunch/chainlaunch/cmd/common"
	"github.com/chainlaunch/chainlaunch/pkg/logger"
	"github.com/chainlaunch/chainlaunch/pkg/nodes/types"
	"github.com/spf13/cobra"
)

type updateCmd struct {
	id          int64
	networkID   int64
	envVars     []string
	domainNames []string
	addresses   []string
	logger      *logger.Logger
}

func (c *updateCmd) validate() error {
	if c.id == 0 {
		return fmt.Errorf("node ID is required")
	}
	if c.networkID == 0 {
		return fmt.Errorf("network ID is required")
	}
	return nil
}

func (c *updateCmd) run(out *os.File) error {
	client, err := common.NewClientFromEnv()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	// Parse environment variables
	envVars := make(map[string]string)
	for _, env := range c.envVars {
		key, value, err := parseEnvVar(env)
		if err != nil {
			return err
		}
		envVars[key] = value
	}

	// Parse domain names
	domainNames := make(map[string]string)
	for _, domain := range c.domainNames {
		key, value, err := parseDomainName(domain)
		if err != nil {
			return err
		}
		domainNames[key] = value
	}

	// Parse addresses
	addresses := make(map[string]string)
	for _, addr := range c.addresses {
		key, value, err := parseAddress(addr)
		if err != nil {
			return err
		}
		addresses[key] = value
	}

	// Create update request
	req := &types.BesuNodeConfig{
		BaseNodeConfig: types.BaseNodeConfig{
			Type: "besu",
			Mode: "service",
		},
		NetworkID: c.networkID,
		Env:       envVars,
	}

	// Update node
	node, err := client.UpdateBesuNode(c.id, req)
	if err != nil {
		return fmt.Errorf("failed to update Besu node: %w", err)
	}

	// Print success message
	fmt.Fprintf(out, "Successfully updated Besu node:\n")
	fmt.Fprintf(out, "ID: %d\n", node.ID)
	fmt.Fprintf(out, "Name: %s\n", node.Name)
	fmt.Fprintf(out, "Type: %s\n", node.NodeType)
	fmt.Fprintf(out, "Status: %s\n", node.Status)
	fmt.Fprintf(out, "Endpoint: %s\n", node.Endpoint)

	return nil
}

// NewUpdateCmd returns the update Besu node command
func NewUpdateCmd(logger *logger.Logger) *cobra.Command {
	c := &updateCmd{
		logger: logger,
	}

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a Besu node",
		Long:  `Update an existing Besu node with new configuration`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := c.validate(); err != nil {
				return err
			}
			return c.run(os.Stdout)
		},
	}

	flags := cmd.Flags()
	flags.Int64VarP(&c.id, "id", "i", 0, "ID of the node to update")
	flags.Int64VarP(&c.networkID, "network-id", "n", 0, "Network ID")
	flags.StringArrayVar(&c.envVars, "env", []string{}, "Environment variables (format: KEY=VALUE)")
	flags.StringArrayVar(&c.domainNames, "domain", []string{}, "Domain names (format: KEY=VALUE)")
	flags.StringArrayVar(&c.addresses, "address", []string{}, "Addresses (format: KEY=VALUE)")

	cmd.MarkFlagRequired("id")
	cmd.MarkFlagRequired("network-id")

	return cmd
}
