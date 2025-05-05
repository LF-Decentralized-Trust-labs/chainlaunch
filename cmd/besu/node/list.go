package node

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/chainlaunch/chainlaunch/cmd/common"
	"github.com/chainlaunch/chainlaunch/pkg/logger"
	"github.com/spf13/cobra"
)

type listCmd struct {
	page   int
	limit  int
	logger *logger.Logger
}

func (c *listCmd) run(out *os.File) error {
	client, err := common.NewClientFromEnv()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	nodes, err := client.ListBesuNodes(c.page, c.limit)
	if err != nil {
		return fmt.Errorf("failed to list Besu nodes: %w", err)
	}

	// Create tab writer
	w := tabwriter.NewWriter(out, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tName\tType\tStatus\tEndpoint")
	fmt.Fprintln(w, "--\t----\t----\t------\t--------")

	// Print nodes
	for _, node := range nodes.Items {
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\n",
			node.ID,
			node.Name,
			node.NodeType,
			node.Status,
			node.Endpoint,
		)
	}

	w.Flush()

	// Print pagination info
	fmt.Printf("\nPage %d of %d (Total: %d)\n", nodes.Page, nodes.PageCount, nodes.Total)
	if nodes.HasNextPage {
		fmt.Println("Use --page to view more results")
	}

	return nil
}

// NewListCmd returns the list Besu nodes command
func NewListCmd(logger *logger.Logger) *cobra.Command {
	c := &listCmd{
		page:   1,
		limit:  10,
		logger: logger,
	}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Besu nodes",
		Long:  `List all Besu nodes with pagination support`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.run(os.Stdout)
		},
	}

	flags := cmd.Flags()
	flags.IntVar(&c.page, "page", 1, "Page number")
	flags.IntVar(&c.limit, "limit", 10, "Number of items per page")

	return cmd
}
