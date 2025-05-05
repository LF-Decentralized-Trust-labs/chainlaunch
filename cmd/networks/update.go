package networks

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
		Short: "Update an existing blockchain network",
		Long:  `Update the configuration of an existing blockchain network.`,
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

			// First get the network to determine its platform
			resp, err := client.Get(fmt.Sprintf("/networks/%d", networkID))
			if err != nil {
				return fmt.Errorf("failed to get network: %w", err)
			}

			if err := common.CheckResponse(resp, 200); err != nil {
				return fmt.Errorf("failed to get network: %w", err)
			}

			var network http.NetworkResponse
			if err := json.NewDecoder(resp.Body).Decode(&network); err != nil {
				return fmt.Errorf("failed to decode response: %w", err)
			}

			var updateResp *http.NetworkResponse
			switch network.Platform {
			case "besu":
				var config http.CreateBesuNetworkRequest
				if err := json.Unmarshal(configData, &config); err != nil {
					return fmt.Errorf("failed to parse besu config: %w", err)
				}
				updateResp, err = updateBesuNetwork(client, networkID, &config)
			case "fabric":
				var config http.CreateFabricNetworkRequest
				if err := json.Unmarshal(configData, &config); err != nil {
					return fmt.Errorf("failed to parse fabric config: %w", err)
				}
				updateResp, err = updateFabricNetwork(client, networkID, &config)
			default:
				return fmt.Errorf("unsupported platform: %s", network.Platform)
			}

			if err != nil {
				return fmt.Errorf("failed to update network: %w", err)
			}

			// Print response
			fmt.Printf("Network updated successfully:\n")
			fmt.Printf("ID: %d\n", updateResp.ID)
			fmt.Printf("Name: %s\n", updateResp.Name)
			fmt.Printf("Platform: %s\n", updateResp.Platform)
			fmt.Printf("Status: %s\n", updateResp.Status)

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

func updateFabricNetwork(client *common.Client, networkID int64, req *http.CreateFabricNetworkRequest) (*http.NetworkResponse, error) {
	resp, err := client.Put(fmt.Sprintf("/networks/fabric/%d", networkID), req)
	if err != nil {
		return nil, fmt.Errorf("failed to update fabric network: %w", err)
	}

	if err := common.CheckResponse(resp, 200); err != nil {
		return nil, fmt.Errorf("failed to update fabric network: %w", err)
	}

	var network http.NetworkResponse
	if err := json.NewDecoder(resp.Body).Decode(&network); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &network, nil
}
