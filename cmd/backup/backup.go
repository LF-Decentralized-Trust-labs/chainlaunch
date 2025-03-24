package backup

import (
	"github.com/chainlaunch/chainlaunch/cmd/backup/restore"
	"github.com/spf13/cobra"
)

func NewBackupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backup",
		Short: "Backup the chainlaunch network",
	}
	cmd.AddCommand(restore.NewRestoreCmd())
	return cmd
}
