package keymanagement

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/chainlaunch/chainlaunch/cmd/common"
	"github.com/chainlaunch/chainlaunch/pkg/keymanagement/models"
)

type createCmd struct {
	name        string
	algorithm   string
	description string
	keySize     string
	curve       string
}

func (c *createCmd) validate() error {
	if c.name == "" {
		return fmt.Errorf("name is required")
	}
	if c.algorithm == "" {
		return fmt.Errorf("algorithm is required")
	}
	return nil
}

func (c *createCmd) run(out *os.File) error {
	client, err := common.NewClientFromEnv()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	// Parse request from flags
	req := &models.CreateKeyRequest{
		Name:        c.name,
		Algorithm:   models.KeyAlgorithm(c.algorithm),
		Description: stringPtr(c.description),
	}

	// Handle optional flags
	if c.keySize != "" {
		size, err := parseInt(c.keySize)
		if err != nil {
			return fmt.Errorf("invalid key size: %w", err)
		}
		req.KeySize = &size
	}

	if c.curve != "" {
		curve := models.ECCurve(c.curve)
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

	fmt.Fprintln(out, string(prettyJSON))
	return nil
}

// NewCreateCmd returns the create key command
func NewCreateCmd() *cobra.Command {
	c := &createCmd{}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new cryptographic key",
		Long:  `Create a new cryptographic key with the specified algorithm and parameters`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := c.validate(); err != nil {
				return err
			}
			return c.run(os.Stdout)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&c.name, "name", "", "Name of the key (required)")
	flags.StringVar(&c.algorithm, "algorithm", "", "Key algorithm (RSA, EC, ED25519) (required)")
	flags.StringVar(&c.description, "description", "", "Optional description of the key")
	flags.StringVar(&c.keySize, "key-size", "", "Key size in bits (for RSA)")
	flags.StringVar(&c.curve, "curve", "", "Elliptic curve name (for EC keys)")

	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("algorithm")

	return cmd
}
