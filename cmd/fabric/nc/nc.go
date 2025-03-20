package nc

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/chainlaunch/chainlaunch/pkg/fabric/client"
	"github.com/chainlaunch/chainlaunch/pkg/logger"
	"github.com/spf13/cobra"
)

type pullCmd struct {
	networkName string
	mspID       string
	output      string
	baseURL     string
	username    string
	password    string
	logger      *logger.Logger
}

func (c *pullCmd) validate() error {
	if c.networkName == "" {
		return fmt.Errorf("network name is required")
	}
	if c.mspID == "" {
		return fmt.Errorf("msp ID is required")
	}
	if c.username == "" {
		return fmt.Errorf("username is required")
	}
	if c.password == "" {
		return fmt.Errorf("password is required")
	}
	return nil
}

func (c *pullCmd) run(out io.Writer) error {
	// Create API client
	apiClient := client.NewClient(c.baseURL, c.username, c.password)

	network, err := apiClient.GetNetworkByName(c.networkName)
	if err != nil {
		return fmt.Errorf("failed to get network: %w", err)
	}

	// Get organization
	org, err := apiClient.GetOrganizationByMspID(c.mspID)
	if err != nil {
		return fmt.Errorf("failed to get organization: %w", err)
	}

	// Get network config
	configBytes, err := apiClient.GetNetworkConfig(network.ID, org.ID)
	if err != nil {
		return fmt.Errorf("failed to get network config: %w", err)
	}

	if c.output == "" {
		// Write to stdout if no output file specified
		_, err = out.Write(configBytes)
		return err
	}

	// Write to specified file
	err = ioutil.WriteFile(c.output, configBytes, 0644)
	if err != nil {
		return fmt.Errorf("failed to write config to file: %w", err)
	}

	c.logger.Infof("Network config written to %s", c.output)
	return nil
}

func NewNCCmd(logger *logger.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "network-config",
		Short: "Network configuration commands",
	}

	pullCmd := &pullCmd{
		baseURL: "http://localhost:8100/api/v1",
		logger:  logger,
	}

	pull := &cobra.Command{
		Use:   "pull",
		Short: "Pull network configuration from server",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := pullCmd.validate(); err != nil {
				return err
			}
			return pullCmd.run(os.Stdout)
		},
	}

	flags := pull.Flags()
	flags.StringVarP(&pullCmd.networkName, "network", "n", "", "Network name")
	flags.StringVarP(&pullCmd.mspID, "msp-id", "m", "", "MSP ID")
	flags.StringVarP(&pullCmd.output, "output", "f", "", "Output file (default: stdout)")
	flags.StringVar(&pullCmd.baseURL, "url", pullCmd.baseURL, "Base URL of the API server")
	flags.StringVarP(&pullCmd.username, "username", "u", "", "Username for basic auth")
	flags.StringVarP(&pullCmd.password, "password", "p", "", "Password for basic auth")

	pull.MarkFlagRequired("network-id")
	pull.MarkFlagRequired("org-id")
	pull.MarkFlagRequired("username")
	pull.MarkFlagRequired("password")

	cmd.AddCommand(pull)
	return cmd
}
