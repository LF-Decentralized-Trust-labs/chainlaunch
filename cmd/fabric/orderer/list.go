package orderer

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/chainlaunch/chainlaunch/cmd/common"
	"github.com/chainlaunch/chainlaunch/pkg/logger"
	"github.com/spf13/cobra"
)

type listCmd struct {
	page   int
	limit  int
	output string // "tsv" or "json"
	logger *logger.Logger
}

func (c *listCmd) run(out io.Writer) error {
	client, err := common.NewClientFromEnv()
	if err != nil {
		return err
	}

	// List orderer nodes
	nodes, err := client.ListOrdererNodes(c.page, c.limit)
	if err != nil {
		return fmt.Errorf("failed to list orderer nodes: %w", err)
	}

	switch c.output {
	case "json":
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		if err := enc.Encode(nodes.Items); err != nil {
			return fmt.Errorf("failed to encode orderer nodes as JSON: %w", err)
		}
		return nil
	case "tsv":
		w := tabwriter.NewWriter(out, 0, 0, 3, ' ', tabwriter.TabIndent)
		fmt.Fprintln(w, "ID\tNAME\tSTATUS\tENDPOINT")
		for _, node := range nodes.Items {
			fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", node.ID, node.Name, node.Status, node.Endpoint)
		}
		w.Flush()
		return nil
	default:
		return fmt.Errorf("unsupported output type: %s (must be 'tsv' or 'json')", c.output)
	}
}

// NewListCmd returns the list orderers command
func NewListCmd(logger *logger.Logger) *cobra.Command {
	c := &listCmd{
		output: "tsv",
		logger: logger,
	}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all orderer nodes",
		Long:  `List all Hyperledger Fabric orderer nodes`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.run(os.Stdout)
		},
	}

	flags := cmd.Flags()
	flags.IntVar(&c.page, "page", 1, "Page number")
	flags.IntVar(&c.limit, "limit", 10, "Number of items per page")
	flags.StringVar(&c.output, "output", "tsv", "Output type: tsv or json")

	return cmd
}
