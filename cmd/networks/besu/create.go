package besu

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/chainlaunch/chainlaunch/cmd/common"
	"github.com/chainlaunch/chainlaunch/pkg/logger"
	"github.com/chainlaunch/chainlaunch/pkg/networks/http"
	"github.com/spf13/cobra"
)

func newCreateCmd(logger *logger.Logger) *cobra.Command {
	var (
		name        string
		description string
		configFile  string
	)

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new Besu network",
		Long:  `Create a new Besu network with the specified configuration.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := common.NewClientFromEnv()
			if err != nil {
				return fmt.Errorf("failed to create client: %w", err)
			}

			// Read config file
			configData, err := os.ReadFile(configFile)
			if err != nil {
				return fmt.Errorf("failed to read config file: %w", err)
			}

			var config http.CreateBesuNetworkRequest
			if err := json.Unmarshal(configData, &config); err != nil {
				return fmt.Errorf("failed to parse besu config: %w", err)
			}
			config.Name = name
			config.Description = description

			resp, err := createBesuNetwork(client, &config)
			if err != nil {
				return fmt.Errorf("failed to create besu network: %w", err)
			}

			// Print response
			fmt.Printf("Besu network created successfully:\n")
			fmt.Printf("ID: %d\n", resp.ID)
			fmt.Printf("Name: %s\n", resp.Name)
			fmt.Printf("Platform: %s\n", resp.Platform)
			fmt.Printf("Status: %s\n", resp.Status)

			return nil
		},
	}

	createCmd.Flags().StringVar(&name, "name", "", "Network name")
	createCmd.Flags().StringVar(&description, "description", "", "Network description")
	createCmd.Flags().StringVar(&configFile, "config", "", "Path to network configuration file")

	createCmd.MarkFlagRequired("name")
	createCmd.MarkFlagRequired("config")

	return createCmd
}

func createBesuNetwork(client *common.Client, req *http.CreateBesuNetworkRequest) (*http.NetworkResponse, error) {
	resp, err := client.Post("/networks/besu", req)
	if err != nil {
		return nil, fmt.Errorf("failed to create besu network: %w", err)
	}

	if err := common.CheckResponse(resp, 201); err != nil {
		return nil, fmt.Errorf("failed to create besu network: %w", err)
	}

	var network http.NetworkResponse
	if err := json.NewDecoder(resp.Body).Decode(&network); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &network, nil
}
