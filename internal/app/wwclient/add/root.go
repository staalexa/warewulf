package add

import (
	"github.com/spf13/cobra"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "add SOURCE [DEST]",
		Short:                 "Add a file to the staging area",
		Long:                  "Add a file to the staging area for later export to a Warewulf overlay",
		RunE:                  CobraRunE,
		Args:                  cobra.RangeArgs(1, 2),
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
