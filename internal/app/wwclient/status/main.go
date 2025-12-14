package status

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/staging"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	// Create staging index
	stagingIndex := staging.NewStagingIndex(StagingDir)

	// Load existing index
	if err := stagingIndex.Load(); err != nil {
		return fmt.Errorf("failed to load staging index: %w", err)
	}

	// List staged files
	files := stagingIndex.List()
	if len(files) == 0 {
		fmt.Println("No files staged")
		return nil
	}

	fmt.Printf("Staged files (%d):\n\n", len(files))
	for _, file := range files {
		fmt.Printf("  %s\n", file.DestPath)
		fmt.Printf("    Source: %s\n", file.SourcePath)
		fmt.Printf("    Size: %d bytes\n", file.Size)
		fmt.Printf("    Added: %s\n\n", file.AddedAt.Format("2006-01-02 15:04:05"))
	}

	return nil
}
