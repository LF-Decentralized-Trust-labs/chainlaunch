package keymanagement

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/chainlaunch/chainlaunch/cmd/common"
	"github.com/chainlaunch/chainlaunch/pkg/keymanagement/models"
)

type getCmd struct {
	keyID string
}

func (c *getCmd) validate() error {
	if c.keyID == "" {
		return fmt.Errorf("key ID is required")
	}
	return nil
}

func (c *getCmd) run(out *os.File) error {
	client, err := common.NewClientFromEnv()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	resp, err := client.Get(fmt.Sprintf("/keys/%s", c.keyID))
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

	fmt.Fprintln(out, string(prettyJSON))
	return nil
}

// NewGetCmd returns the get key command
func NewGetCmd() *cobra.Command {
	c := &getCmd{}

	cmd := &cobra.Command{
		Use:   "get [key-id]",
		Short: "Get a cryptographic key by ID",
		Long:  `Get a cryptographic key by its ID`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c.keyID = args[0]
			if err := c.validate(); err != nil {
				return err
			}
			return c.run(os.Stdout)
		},
	}

	return cmd
}
