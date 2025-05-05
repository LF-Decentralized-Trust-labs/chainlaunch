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

func newUpdateCmd(logger *logger.Logger) *cobra.Command {
	var (
		networkID  int64
		configFile string
	)

	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "Update an existing Besu network",
		Long:  `Update the configuration of an existing Besu network.`,
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

			resp, err := updateBesuNetwork(client, networkID, &config)
			if err != nil {
				return fmt.Errorf("failed to update besu network: %w", err)
			}

			// Print response
			fmt.Printf("Besu network updated successfully:\n")
			fmt.Printf("ID: %d\n", resp.ID)
			fmt.Printf("Name: %s\n", resp.Name)
			fmt.Printf("Platform: %s\n", resp.Platform)
			fmt.Printf("Status: %s\n", resp.Status)

			return nil
		},
	}

	updateCmd.Flags().Int64Var(&networkID, "id", 0, "Network ID")
	updateCmd.Flags().StringVar(&configFile, "config", "", "Path to network configuration file")

	updateCmd.MarkFlagRequired("id")
	updateCmd.MarkFlagRequired("config")

	return updateCmd
}

func updateBesuNetwork(client *common.Client, networkID int64, req *http.CreateBesuNetworkRequest) (*http.NetworkResponse, error) {
	resp, err := client.Put(fmt.Sprintf("/networks/besu/%d", networkID), req)
	if err != nil {
		return nil, fmt.Errorf("failed to update besu network: %w", err)
	}

	if err := common.CheckResponse(resp, 200); err != nil {
		return nil, fmt.Errorf("failed to update besu network: %w", err)
	}

	var network http.NetworkResponse
	if err := json.NewDecoder(resp.Body).Decode(&network); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &network, nil
}
