package version

import (
	"fmt"

	"github.com/chainlaunch/chainlaunch/pkg/version"
	"github.com/spf13/cobra"
)

// NewVersionCmd creates a new version command
func NewVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Long:  `Print detailed version information about the chainlaunch binary`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Version: %s\n", version.Version)
			fmt.Printf("Git Commit: %s\n", version.GitCommit)
			fmt.Printf("Build Time: %s\n", version.BuildTime)
		},
	}

	return cmd
}
