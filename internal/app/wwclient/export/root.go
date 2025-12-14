package export

import (
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "export OVERLAY_NAME",
		Short:                 "Export staged files to a Warewulf overlay",
		Long:                  "Export all staged files to a Warewulf overlay",
		RunE:                  CobraRunE,
		Args:                  cobra.ExactArgs(1),
	}
	StagingDir string
	ClearAfter bool
)

func init() {
	baseCmd.PersistentFlags().StringVarP(&StagingDir, "staging-dir", "s", "", "Directory to use for staging (default: /var/lib/warewulf/staging)")
	baseCmd.PersistentFlags().BoolVarP(&ClearAfter, "clear", "c", false, "Clear staging area after export")
}

// GetCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
