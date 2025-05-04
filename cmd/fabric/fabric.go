package fabric

import (
	"os"

	"github.com/chainlaunch/chainlaunch/cmd/fabric/install"
	"github.com/chainlaunch/chainlaunch/cmd/fabric/invoke"
	"github.com/chainlaunch/chainlaunch/cmd/fabric/nc"
	"github.com/chainlaunch/chainlaunch/cmd/fabric/org"
	"github.com/chainlaunch/chainlaunch/cmd/fabric/query"
	"github.com/chainlaunch/chainlaunch/pkg/logger"
	"github.com/spf13/cobra"
)

// RootCmd returns the root command
func NewFabricCmd(logger *logger.Logger) *cobra.Command {
	rootCmd := &cobra.Command{
		Use: "fabric",
	}
	rootCmd.AddCommand(
		install.NewInstallCmd(logger),
		query.NewQueryChaincodeCMD(os.Stdout, os.Stderr, logger),
		invoke.NewInvokeChaincodeCMD(os.Stdout, os.Stderr, logger),
		nc.NewNCCmd(logger),
		org.NewOrgCmd(logger),
	)
	return rootCmd
}
