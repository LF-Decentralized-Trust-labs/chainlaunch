package orderer

import (
	"fmt"
	"io"
	"os"

	"github.com/chainlaunch/chainlaunch/cmd/common"
	"github.com/chainlaunch/chainlaunch/pkg/logger"
	"github.com/chainlaunch/chainlaunch/pkg/nodes/types"
	"github.com/spf13/cobra"
)

type updateCmd struct {
	id             int64
	name           string
	listenAddr     string
	adminAddr      string
	operationsAddr string
	externalAddr   string
	version        string
	logger         *logger.Logger
}

func (c *updateCmd) validate() error {
	if c.id == 0 {
		return fmt.Errorf("node ID is required")
	}
	return nil
}

func (c *updateCmd) run(out io.Writer) error {
	client, err := common.NewClientFromEnv()
	if err != nil {
		return err
	}

	// Prepare request
	req := &types.FabricOrdererConfig{
		Name:                    c.name,
		ListenAddress:           c.listenAddr,
		AdminAddress:            c.adminAddr,
		OperationsListenAddress: c.operationsAddr,
		ExternalEndpoint:        c.externalAddr,
		Version:                 c.version,
	}

	// Update orderer node
	node, err := client.UpdateOrdererNode(c.id, req)
	if err != nil {
		return fmt.Errorf("failed to update orderer node: %w", err)
	}

	fmt.Fprintf(out, "Orderer node %s updated successfully\n", node.Name)
	return nil
}

// NewUpdateCmd returns the update orderer command
func NewUpdateCmd(logger *logger.Logger) *cobra.Command {
	c := &updateCmd{
		logger: logger,
	}

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update an orderer node",
		Long:  `Update a Hyperledger Fabric orderer node`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := c.validate(); err != nil {
				return err
			}
			return c.run(os.Stdout)
		},
	}

	flags := cmd.Flags()
	flags.Int64VarP(&c.id, "id", "i", 0, "Node ID")
	flags.StringVarP(&c.name, "name", "n", "", "Name of the orderer node")
	flags.StringVar(&c.listenAddr, "listen-addr", "", "Listen address for the orderer")
	flags.StringVar(&c.adminAddr, "admin-addr", "", "Admin listen address")
	flags.StringVar(&c.operationsAddr, "operations-addr", "", "Operations listen address")
	flags.StringVar(&c.externalAddr, "external-addr", "", "External endpoint for the orderer")
	flags.StringVar(&c.version, "version", "", "Fabric version to use")

	cmd.MarkFlagRequired("id")

	return cmd
}
