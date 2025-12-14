package add

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/staging"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	sourcePath := args[0]
	destPath := ""
	if len(args) > 1 {
		destPath = args[1]
	}

	// Create staging index
	stagingIndex := staging.NewStagingIndex(StagingDir)

	// Load existing index
	if err := stagingIndex.Load(); err != nil {
		return fmt.Errorf("failed to load staging index: %w", err)
	}

	// Add file to staging
	if err := stagingIndex.AddFile(sourcePath, destPath); err != nil {
		return fmt.Errorf("failed to add file to staging: %w", err)
	}

	// Save index
	if err := stagingIndex.Save(); err != nil {
		return fmt.Errorf("failed to save staging index: %w", err)
	}

	fmt.Printf("Added file to staging: %s\n", sourcePath)
	if destPath != "" {
		fmt.Printf("  Destination: %s\n", destPath)
	}

	return nil
}
