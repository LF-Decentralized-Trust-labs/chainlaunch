package org

import (
	"io"
	"os"
	"time"

	"github.com/chainlaunch/chainlaunch/pkg/logger"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

type listCmd struct {
	logger *logger.Logger
}

func (c *listCmd) run(out io.Writer) error {
	client := NewClientWrapper(c.logger)

	// Get organizations data
	orgs, err := client.ListOrganizations()
	if err != nil {
		return err
	}

	// Create table writer
	table := tablewriter.NewWriter(out)
	table.SetHeader([]string{"MSP ID", "Created At", "Description"})

	// Configure table style
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("-")
	table.SetHeaderLine(true)
	table.SetBorder(false)
	table.SetTablePadding("\t")
	table.SetNoWhiteSpace(true)

	// Add data to table
	for _, org := range orgs.Items {
		table.Append([]string{
			org.MspID,
			org.CreatedAt.Format(time.RFC3339),
			org.Description,
		})
	}

	table.Render()
	return nil
}

// NewListCmd returns the list organizations command
func NewListCmd(logger *logger.Logger) *cobra.Command {
	c := &listCmd{
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

	return cmd
}
