package status

import (
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "status",
		Short:                 "Show staged files",
		Long:                  "Show all files currently staged for export",
		RunE:                  CobraRunE,
		Args:                  cobra.NoArgs,
	}
	StagingDir string
)

func init() {
	baseCmd.PersistentFlags().StringVarP(&StagingDir, "staging-dir", "s", "", "Directory to use for staging (default: /var/lib/warewulf/staging)")
}

// GetCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
