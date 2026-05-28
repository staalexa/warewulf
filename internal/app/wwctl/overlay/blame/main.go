package blame

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/overlay"
)

type blameLine struct {
	Path    string
	Overlay string
	Context string
}

func CobraRunE(vars *variables) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		nodeDB, err := node.New()
		if err != nil {
			return fmt.Errorf("could not open node configuration: %w", err)
		}

		nodeData, err := nodeDB.GetNode(args[0])
		if err != nil {
			return fmt.Errorf("could not get node %s: %w", args[0], err)
		}

		prefix := normalizePathPrefix(vars.PathPrefix)
		var lines []blameLine

		contextLines, err := collectBlameLines(nodeData.SystemOverlay, "system", vars.ShowModeChanges, prefix)
		if err != nil {
			return err
		}
		lines = append(lines, contextLines...)

		contextLines, err = collectBlameLines(nodeData.RuntimeOverlay, "runtime", vars.ShowModeChanges, prefix)
		if err != nil {
			return err
		}
		lines = append(lines, contextLines...)

		return printBlameLines(cmd, lines)
	}
}

func printBlameLines(cmd *cobra.Command, lines []blameLine) error {
	pathWidth := 0
	overlayWidth := 0
	for _, line := range lines {
		pathWidth = max(pathWidth, len(line.Path))
		overlayWidth = max(overlayWidth, len(line.Overlay))
	}

	for _, line := range lines {
		if _, err := fmt.Fprintf(cmd.OutOrStdout(), "%-*s  %-*s  [%s overlay]\n", pathWidth, line.Path, overlayWidth, line.Overlay, line.Context); err != nil {
			return err
		}
	}
	return nil
}

func collectBlameLines(overlayNames []string, context string, includeDirs bool, prefix string) ([]blameLine, error) {
	var lines []blameLine
	for _, overlayName := range overlayNames {
		overlayRoot, err := overlay.Get(overlayName)
		if err != nil {
			return nil, fmt.Errorf("could not get overlay %s: %w", overlayName, err)
		}

		overlayLines, err := collectOverlayLines(overlayRoot, overlayName, context, includeDirs, prefix)
		if err != nil {
			return nil, err
		}
		lines = append(lines, overlayLines...)
	}
	return lines, nil
}

func collectOverlayLines(overlayRoot overlay.Overlay, overlayName string, context string, includeDirs bool, prefix string) ([]blameLine, error) {
	var lines []blameLine
	rootfs := overlayRoot.Rootfs()
	err := filepath.Walk(rootfs, func(walkPath string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error walking overlay %s: %w", overlayName, err)
		}

		relPath, err := filepath.Rel(rootfs, walkPath)
		if err != nil {
			return fmt.Errorf("could not determine path for overlay %s: %w", overlayName, err)
		}
		if relPath == "." {
			return nil
		}

		if info.IsDir() {
			if !includeDirs {
				return nil
			}
		} else if !isBlameFile(info) {
			return nil
		}

		deployedPath := deployedOverlayPath(relPath)
		if !pathMatchesPrefix(deployedPath, prefix) {
			return nil
		}

		lines = append(lines, blameLine{
			Path:    deployedPath,
			Overlay: overlayName,
			Context: context,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.SliceStable(lines, func(i, j int) bool {
		return lines[i].Path < lines[j].Path
	})
	return lines, nil
}

func isBlameFile(info os.FileInfo) bool {
	return info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0
}

func deployedOverlayPath(relPath string) string {
	deployedPath := "/" + filepath.ToSlash(relPath)
	deployedPath = path.Clean(deployedPath)
	return strings.TrimSuffix(deployedPath, ".ww")
}

func normalizePathPrefix(prefix string) string {
	if prefix == "" {
		return ""
	}
	prefix = path.Clean("/" + strings.TrimPrefix(filepath.ToSlash(prefix), "/"))
	return prefix
}

func pathMatchesPrefix(filePath string, prefix string) bool {
	if prefix == "" || prefix == "/" {
		return true
	}
	return filePath == prefix || strings.HasPrefix(filePath, prefix+"/")
}
