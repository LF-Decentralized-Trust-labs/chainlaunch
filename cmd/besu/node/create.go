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
	name       string
	p2pPort    int64
	rpcPort    int64
	p2pHost    string
	rpcHost    string
	externalIP string
	internalIP string
	envVars    map[string]string
	keyID      int64
	bootNodes  []string
	networkID  int64
	logger     *logger.Logger
}

func (c *createCmd) validate() error {
	if c.name == "" {
		return fmt.Errorf("Name is required")
	}
	if c.p2pPort == 0 {
		return fmt.Errorf("P2P port is required")
	}
	if c.rpcPort == 0 {
		return fmt.Errorf("RPC port is required")
	}
	if c.p2pHost == "" {
		return fmt.Errorf("P2P host is required")
	}
	if c.rpcHost == "" {
		return fmt.Errorf("RPC host is required")
	}
	if c.externalIP == "" {
		return fmt.Errorf("External IP is required")
	}
	if c.internalIP == "" {
		return fmt.Errorf("Internal IP is required")
	}
	if c.keyID == 0 {
		return fmt.Errorf("Key ID is required")
	}
	if len(c.bootNodes) == 0 {
		return fmt.Errorf("Boot nodes are required")
	}
	if c.networkID == 0 {
		return fmt.Errorf("Network ID is required")
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

	// Create node request with only the specified parameters
	req := &types.BesuNodeConfig{
		BaseNodeConfig: types.BaseNodeConfig{
			Type: "besu",
			Mode: "service",
		},
		NetworkID:  c.networkID,
		P2PPort:    uint(c.p2pPort),
		RPCPort:    uint(c.rpcPort),
		P2PHost:    c.p2pHost,
		RPCHost:    c.rpcHost,
		ExternalIP: c.externalIP,
		InternalIP: c.internalIP,
		Env:        envVars,
		KeyID:      c.keyID,
		BootNodes:  c.bootNodes,
	}

	// Create node
	node, err := client.CreateBesuNode(c.name, req)
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
	flags.Int64Var(&c.p2pPort, "p2p-port", 30303, "P2P port")
	flags.Int64Var(&c.rpcPort, "rpc-port", 8545, "RPC port")
	flags.StringVar(&c.p2pHost, "p2p-host", "0.0.0.0", "P2P host")
	flags.StringVar(&c.rpcHost, "rpc-host", "0.0.0.0", "RPC host")
	flags.StringVar(&c.externalIP, "external-ip", "127.0.0.1", "External IP")
	flags.StringVar(&c.internalIP, "internal-ip", "127.0.0.1", "Internal IP")
	flags.StringToStringVar(&c.envVars, "env", map[string]string{}, "Environment variables (format: KEY=VALUE)")
	flags.Int64Var(&c.keyID, "key-id", 0, "Key ID")
	flags.StringSliceVar(&c.bootNodes, "boot-nodes", []string{}, "Boot nodes")
	flags.Int64Var(&c.networkID, "network-id", 0, "Network ID")
	flags.StringVar(&c.name, "name", "", "Name")
	// Mark all flags as required
	cmd.MarkFlagRequired("p2p-port")
	cmd.MarkFlagRequired("rpc-port")
	cmd.MarkFlagRequired("p2p-host")
	cmd.MarkFlagRequired("rpc-host")
	cmd.MarkFlagRequired("external-ip")
	cmd.MarkFlagRequired("internal-ip")
	cmd.MarkFlagRequired("key-id")
	cmd.MarkFlagRequired("network-id")
	cmd.MarkFlagRequired("name")
	return cmd
}
