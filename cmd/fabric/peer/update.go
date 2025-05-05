package peer

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
	chaincodeAddr  string
	eventsAddr     string
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
	req := &types.FabricPeerConfig{
		Name:                    c.name,
		ListenAddress:           c.listenAddr,
		ChaincodeAddress:        c.chaincodeAddr,
		EventsAddress:           c.eventsAddr,
		OperationsListenAddress: c.operationsAddr,
		ExternalEndpoint:        c.externalAddr,
		Version:                 c.version,
	}

	// Update peer node
	node, err := client.UpdatePeerNode(c.id, req)
	if err != nil {
		return fmt.Errorf("failed to update peer node: %w", err)
	}

	fmt.Fprintf(out, "Peer node %s updated successfully\n", node.Name)
	return nil
}

// NewUpdateCmd returns the update peer command
func NewUpdateCmd(logger *logger.Logger) *cobra.Command {
	c := &updateCmd{
		logger: logger,
	}

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a peer node",
		Long:  `Update a Hyperledger Fabric peer node`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := c.validate(); err != nil {
				return err
			}
			return c.run(os.Stdout)
		},
	}

	flags := cmd.Flags()
	flags.Int64VarP(&c.id, "id", "i", 0, "Node ID")
	flags.StringVarP(&c.name, "name", "n", "", "Name of the peer node")
	flags.StringVar(&c.listenAddr, "listen-addr", "", "Listen address for the peer")
	flags.StringVar(&c.chaincodeAddr, "chaincode-addr", "", "Chaincode listen address")
	flags.StringVar(&c.eventsAddr, "events-addr", "", "Events listen address")
	flags.StringVar(&c.operationsAddr, "operations-addr", "", "Operations listen address")
	flags.StringVar(&c.externalAddr, "external-addr", "", "External endpoint for the peer")
	flags.StringVar(&c.version, "version", "", "Fabric version to use")

	cmd.MarkFlagRequired("id")

	return cmd
}
