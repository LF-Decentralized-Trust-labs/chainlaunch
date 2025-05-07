package keymanagement

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewKeyManagementCmd returns the root command for keymanagement
func NewKeyManagementCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "keys",
		Short: "Manage cryptographic keys",
		Long:  `Create, get, and manage cryptographic keys using various providers (Database, Vault, HSM)`,
	}

	rootCmd.AddCommand(NewCreateCmd())
	rootCmd.AddCommand(NewGetCmd())
	return rootCmd
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
