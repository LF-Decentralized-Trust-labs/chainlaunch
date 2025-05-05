package keymanagement

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainlaunch/chainlaunch/cmd/common"
	"github.com/chainlaunch/chainlaunch/pkg/keymanagement/models"
)

// NewKeyManagementCmd returns the root command for keymanagement
func NewKeyManagementCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "keymanagement",
		Short: "Manage cryptographic keys",
		Long:  `Create, get, and manage cryptographic keys using various providers (Database, Vault, HSM)`,
	}

	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(getCmd)
	return rootCmd
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new cryptographic key",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := common.NewClientFromEnv()
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		// Parse request from flags
		req := &models.CreateKeyRequest{
			Name:        cmd.Flag("name").Value.String(),
			Algorithm:   models.KeyAlgorithm(cmd.Flag("algorithm").Value.String()),
			Description: stringPtr(cmd.Flag("description").Value.String()),
		}

		// Handle optional flags
		if cmd.Flag("key-size").Changed {
			keySize := cmd.Flag("key-size").Value.String()
			size, err := parseInt(keySize)
			if err != nil {
				return fmt.Errorf("invalid key size: %w", err)
			}
			req.KeySize = &size
		}

		if cmd.Flag("curve").Changed {
			curve := models.ECCurve(cmd.Flag("curve").Value.String())
			req.Curve = &curve
		}

		// Validate request
		if err := req.Validate(); err != nil {
			return fmt.Errorf("invalid request: %w", err)
		}

		// Send request
		resp, err := client.Post("/keys", req)
		if err != nil {
			return fmt.Errorf("failed to create key: %w", err)
		}

		if err := common.CheckResponse(resp, 201); err != nil {
			return err
		}

		// Parse and display response
		var keyResp models.KeyResponse
		body, err := common.ReadBody(resp)
		if err != nil {
			return fmt.Errorf("failed to read response: %w", err)
		}

		if err := json.Unmarshal(body, &keyResp); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}

		// Pretty print the response
		prettyJSON, err := json.MarshalIndent(keyResp, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format response: %w", err)
		}

		fmt.Println(string(prettyJSON))
		return nil
	},
}

var getCmd = &cobra.Command{
	Use:   "get [key-id]",
	Short: "Get a cryptographic key by ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := common.NewClientFromEnv()
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		keyID := args[0]
		resp, err := client.Get(fmt.Sprintf("/keys/%s", keyID))
		if err != nil {
			return fmt.Errorf("failed to get key: %w", err)
		}

		if err := common.CheckResponse(resp, 200); err != nil {
			return err
		}

		// Parse and display response
		var keyResp models.KeyResponse
		body, err := common.ReadBody(resp)
		if err != nil {
			return fmt.Errorf("failed to read response: %w", err)
		}

		if err := json.Unmarshal(body, &keyResp); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}

		// Pretty print the response
		prettyJSON, err := json.MarshalIndent(keyResp, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format response: %w", err)
		}

		fmt.Println(string(prettyJSON))
		return nil
	},
}

func init() {
	// Create command flags
	createCmd.Flags().String("name", "", "Name of the key (required)")
	createCmd.Flags().String("algorithm", "", "Key algorithm (RSA, EC, ED25519) (required)")
	createCmd.Flags().String("description", "", "Optional description of the key")
	createCmd.Flags().String("key-size", "", "Key size in bits (for RSA)")
	createCmd.Flags().String("curve", "", "Elliptic curve name (for EC keys)")

	// Mark required flags
	createCmd.MarkFlagRequired("name")
	createCmd.MarkFlagRequired("algorithm")
}

// Helper functions
func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func parseInt(s string) (int, error) {
	var i int
	_, err := fmt.Sscanf(s, "%d", &i)
	return i, err
}
