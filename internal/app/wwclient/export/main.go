package export

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/overlay"
	"github.com/warewulf/warewulf/internal/pkg/staging"
	"github.com/warewulf/warewulf/internal/pkg/util"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	overlayName := args[0]

	// Create staging index
	stagingIndex := staging.NewStagingIndex(StagingDir)

	// Load existing index
	if err := stagingIndex.Load(); err != nil {
		return fmt.Errorf("failed to load staging index: %w", err)
	}

	// Check if there are any staged files
	if len(stagingIndex.Files) == 0 {
		return fmt.Errorf("no files staged for export")
	}

	// Get or create overlay
	ovl, err := overlay.Get(overlayName)
	if err != nil {
		// Overlay doesn't exist, create it
		ovl, err = overlay.Create(overlayName)
		if err != nil {
			return fmt.Errorf("failed to create overlay: %w", err)
		}
	}

	// Ensure it's a site overlay
	if !ovl.IsSiteOverlay() {
		siteOverlay, err := ovl.CloneToSite()
		if err != nil {
			return fmt.Errorf("failed to clone overlay to site: %w", err)
		}
		ovl = siteOverlay
	}

	// Export each staged file to the overlay
	for destPath, stagedFile := range stagingIndex.Files {
		stagingPath := stagingIndex.GetStagedFilePath(destPath)
		overlayPath := ovl.File(destPath)

		// Ensure parent directory exists
		if err := os.MkdirAll(filepath.Dir(overlayPath), 0755); err != nil {
			return fmt.Errorf("failed to create parent directory in overlay: %w", err)
		}

		// Copy file from staging to overlay
		if err := util.CopyFile(stagingPath, overlayPath); err != nil {
			return fmt.Errorf("failed to copy file %s to overlay: %w", destPath, err)
		}

		fmt.Printf("Exported: %s -> %s:%s\n", stagedFile.SourcePath, overlayName, destPath)
	}

	fmt.Printf("\nSuccessfully exported %d file(s) to overlay '%s'\n", len(stagingIndex.Files), overlayName)

	// Clear staging area if requested
	if ClearAfter {
		if err := stagingIndex.Clear(); err != nil {
			return fmt.Errorf("failed to clear staging area: %w", err)
		}
		if err := stagingIndex.Save(); err != nil {
			return fmt.Errorf("failed to save staging index: %w", err)
		}
		fmt.Println("Staging area cleared")
	}

	return nil
}
