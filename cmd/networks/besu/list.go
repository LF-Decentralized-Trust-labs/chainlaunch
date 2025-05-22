package besu

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/chainlaunch/chainlaunch/cmd/common"
	"github.com/chainlaunch/chainlaunch/pkg/logger"
	"github.com/spf13/cobra"
)

// NewListCmd returns a command that lists all Besu networks
func NewListCmd(logger *logger.Logger) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all Besu networks",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := common.NewClientFromEnv()
			if err != nil {
				return err
			}
			result, err := client.ListBesuNetworks()
			if err != nil {
				return err
			}

			switch strings.ToLower(output) {
			case "json":
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			case "tsv":
				fmt.Println("ID\tName\tStatus\tCreatedAt")
				for _, n := range result.Networks {
					fmt.Printf("%d\t%s\t%s\t%s\n", n.ID, n.Name, n.Status, n.CreatedAt)
				}
				return nil
			default: // table
				for _, n := range result.Networks {
					fmt.Printf("ID: %d | Name: %s | Status: %s | CreatedAt: %s\n", n.ID, n.Name, n.Status, n.CreatedAt)
				}
				return nil
			}
		},
	}
	cmd.Flags().StringVarP(&output, "output", "o", "table", "Output format: table, json, or tsv")
	return cmd
}
