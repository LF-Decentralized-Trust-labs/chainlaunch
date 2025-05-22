package org

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"text/tabwriter"
	"time"

	"github.com/chainlaunch/chainlaunch/pkg/logger"
	"github.com/spf13/cobra"
)

type listCmd struct {
	output string // "tsv" or "json"
	logger *logger.Logger
}

func (c *listCmd) run(out io.Writer) error {
	client := NewClientWrapper(c.logger)

	// Get organizations data
	orgs, err := client.ListOrganizations()
	if err != nil {
		return err
	}

	switch c.output {
	case "json":
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		if err := enc.Encode(orgs.Items); err != nil {
			return fmt.Errorf("failed to encode organizations as JSON: %w", err)
		}
		return nil
	case "tsv":
		w := tabwriter.NewWriter(out, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "MSP ID\tCreated At\tDescription")
		fmt.Fprintln(w, "------\t----------\t-----------")
		for _, org := range orgs.Items {
			fmt.Fprintf(w, "%s\t%s\t%s\n",
				org.MspID,
				org.CreatedAt.Format(time.RFC3339),
				org.Description,
			)
		}
		w.Flush()
		return nil
	default:
		return fmt.Errorf("unsupported output type: %s (must be 'tsv' or 'json')", c.output)
	}
}

// NewListCmd returns the list organizations command
func NewListCmd(logger *logger.Logger) *cobra.Command {
	c := &listCmd{
		output: "tsv",
		logger: logger,
	}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List organizations",
		Long:  `List all Hyperledger Fabric organizations`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.run(os.Stdout)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&c.output, "output", "tsv", "Output type: tsv or json")

	return cmd
}
