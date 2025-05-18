package besu

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func NewBesuTestnetCmd() *cobra.Command {
	var name string
	var nodes int

	cmd := &cobra.Command{
		Use:   "besu",
		Short: "Create a Besu testnet",
		Run: func(cmd *cobra.Command, args []string) {
			if name == "" {
				fmt.Println("Error: --name is required")
				os.Exit(1)
			}
			fmt.Printf("Creating Besu testnet '%s' with %d node(s)\n", name, nodes)
			// TODO: Call Besu testnet creation logic here
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Name of the testnet (required)")
	cmd.Flags().IntVar(&nodes, "nodes", 1, "Number of nodes (default 1)")

	return cmd
}
