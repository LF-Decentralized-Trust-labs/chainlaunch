package node

import (
	"fmt"
	"os"
	"strings"

	"github.com/chainlaunch/chainlaunch/cmd/common"
	"github.com/chainlaunch/chainlaunch/pkg/logger"
	"github.com/chainlaunch/chainlaunch/pkg/nodes/types"
	"github.com/spf13/cobra"
)

type createCmd struct {
	name           string
	networkID      int64
	networkName    string
	networkVersion string
	networkType    string
	consensus      string
	genesisFile    string
	configFile     string
	envVars        []string
	domainNames    []string
	addresses      []string
	logger         *logger.Logger
}

func (c *createCmd) validate() error {
	if c.name == "" {
		return fmt.Errorf("node name is required")
	}
	if c.networkID == 0 {
		return fmt.Errorf("network ID is required")
	}
	if c.networkName == "" {
		return fmt.Errorf("network name is required")
	}
	if c.networkVersion == "" {
		return fmt.Errorf("network version is required")
	}
	if c.networkType == "" {
		return fmt.Errorf("network type is required")
	}
	if c.consensus == "" {
		return fmt.Errorf("consensus type is required")
	}
	if c.genesisFile == "" {
		return fmt.Errorf("genesis file is required")
	}
	if c.configFile == "" {
		return fmt.Errorf("config file is required")
	}
	return nil
}

func parseEnvVar(env string) (string, string, error) {
	parts := strings.SplitN(env, "=", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid environment variable format: %s", env)
	}
	return parts[0], parts[1], nil
}

func parseDomainName(domain string) (string, string, error) {
	parts := strings.SplitN(domain, "=", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid domain name format: %s", domain)
	}
	return parts[0], parts[1], nil
}

func parseAddress(addr string) (string, string, error) {
	parts := strings.SplitN(addr, "=", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid address format: %s", addr)
	}
	return parts[0], parts[1], nil
}

func (c *createCmd) run(out *os.File) error {
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

	// Create node request
	req := &types.BesuNodeConfig{
		BaseNodeConfig: types.BaseNodeConfig{
			Type: "besu",
			Mode: "service",
		},
		NetworkID:  c.networkID,
		P2PPort:    30303,
		RPCPort:    8545,
		P2PHost:    "0.0.0.0",
		RPCHost:    "0.0.0.0",
		ExternalIP: "127.0.0.1",
		InternalIP: "127.0.0.1",
		Env:        envVars,
	}

	// Create node
	node, err := client.CreateBesuNode(req)
	if err != nil {
		return fmt.Errorf("failed to create Besu node: %w", err)
	}

	// Print success message
	fmt.Fprintf(out, "Successfully created Besu node:\n")
	fmt.Fprintf(out, "ID: %d\n", node.ID)
	fmt.Fprintf(out, "Name: %s\n", node.Name)
	fmt.Fprintf(out, "Type: %s\n", node.NodeType)
	fmt.Fprintf(out, "Status: %s\n", node.Status)
	fmt.Fprintf(out, "Endpoint: %s\n", node.Endpoint)

	return nil
}

// NewCreateCmd returns the create Besu node command
func NewCreateCmd(logger *logger.Logger) *cobra.Command {
	c := &createCmd{
		logger: logger,
	}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new Besu node",
		Long:  `Create a new Besu node with the specified configuration`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := c.validate(); err != nil {
				return err
			}
			return c.run(os.Stdout)
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&c.name, "name", "n", "", "Name of the node")
	flags.Int64VarP(&c.networkID, "network-id", "i", 0, "Network ID")
	flags.StringVar(&c.networkName, "network-name", "", "Network name")
	flags.StringVar(&c.networkVersion, "network-version", "", "Network version")
	flags.StringVar(&c.networkType, "network-type", "", "Network type")
	flags.StringVar(&c.consensus, "consensus", "", "Consensus type")
	flags.StringVar(&c.genesisFile, "genesis-file", "", "Path to genesis file")
	flags.StringVar(&c.configFile, "config-file", "", "Path to config file")
	flags.StringArrayVar(&c.envVars, "env", []string{}, "Environment variables (format: KEY=VALUE)")
	flags.StringArrayVar(&c.domainNames, "domain", []string{}, "Domain names (format: KEY=VALUE)")
	flags.StringArrayVar(&c.addresses, "address", []string{}, "Addresses (format: KEY=VALUE)")

	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("network-id")
	cmd.MarkFlagRequired("network-name")
	cmd.MarkFlagRequired("network-version")
	cmd.MarkFlagRequired("network-type")
	cmd.MarkFlagRequired("consensus")
	cmd.MarkFlagRequired("genesis-file")
	cmd.MarkFlagRequired("config-file")

	return cmd
}
