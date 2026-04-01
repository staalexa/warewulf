package capture

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/overlaydiff"
)

var (
	stopSourcePath  string
	stopStateFile   string
	stopExcludes    []string
	stopFormat      string
	stopInteractive bool
	stopOnly        []string
	stopPathPrefix  []string
	stopExport      bool
	stopExportDir   string
)

func GetStopCommand() *cobra.Command {
	cmd := &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "stop",
		Short:                 "Stop capture and show changes",
		Long:                  "Stop capture, compare against the baseline, and show changes",
		RunE:                  runStop,
		Args:                  cobra.NoArgs,
	}

	cmd.Flags().StringVar(&stopSourcePath, "source", "", "Source directory to scan")
	cmd.Flags().StringVar(&stopStateFile, "state-file", "", "Snapshot file path")
	cmd.Flags().StringArrayVar(&stopExcludes, "exclude", nil, defaultExcludeHelp())
	cmd.Flags().StringVar(&stopFormat, "format", "table", "Output format: table|json")
	cmd.Flags().BoolVar(&stopInteractive, "interactive", false, "Prompt for per-path selection decisions")
	cmd.Flags().StringArrayVar(&stopOnly, "only", nil, "Only include change class (repeatable): added|modified|mode-changed")
	cmd.Flags().StringArrayVar(&stopPathPrefix, "path-prefix", nil, "Only include path prefix (repeatable, example: /etc)")
	cmd.Flags().BoolVar(&stopExport, "export", false, "Copy selected files to an export directory")
	cmd.Flags().StringVar(&stopExportDir, "export-dir", "", "Export destination directory (default: /tmp/wwclient-overlaydiff/<timestamp>)")

	_ = cmd.MarkFlagRequired("source")
	return cmd
}

func runStop(cmd *cobra.Command, args []string) error {
	format := strings.ToLower(stopFormat)
	if format != "table" && format != "json" {
		return fmt.Errorf("invalid format %q: expected table or json", stopFormat)
	}

	onlyFilters, err := overlaydiff.ParseChangeTypes(stopOnly)
	if err != nil {
		return err
	}

	stateFile := resolveStateFilePath(stopStateFile)
	snapshot, err := overlaydiff.LoadSnapshot(stateFile)
	if err != nil {
		return err
	}

	sourceAbs, err := filepath.Abs(stopSourcePath)
	if err != nil {
		return fmt.Errorf("failed to resolve source path: %w", err)
	}
	if snapshot.SourceRoot != "" && sourceAbs != snapshot.SourceRoot {
		return fmt.Errorf("source path %s does not match capture %s", sourceAbs, snapshot.SourceRoot)
	}

	var excludes []string
	if len(stopExcludes) > 0 {
		excludes = overlaydiff.ResolveExcludes(stopExcludes)
	} else if len(snapshot.Excludes) > 0 {
		excludes = overlaydiff.NormalizeExcludes(snapshot.Excludes)
	} else {
		excludes = overlaydiff.ResolveExcludes(nil)
	}

	entries, err := overlaydiff.ScanTreeWithOptions(sourceAbs, overlaydiff.ScanOptions{Excludes: excludes})
	if err != nil {
		return err
	}

	changes := overlaydiff.Compare(entries, snapshot.Entries)
	changes = overlaydiff.FilterChanges(changes, overlaydiff.FilterOptions{
		Only:       onlyFilters,
		PathPrefix: stopPathPrefix,
	})

	snapshot.Changes = overlaydiff.SummarizeChanges(changes)
	if err := overlaydiff.SaveSnapshot(stateFile, snapshot); err != nil {
		return err
	}

	if stopInteractive {
		if err := runInteractiveSelection(cmd.InOrStdin(), cmd.OutOrStdout(), changes, &snapshot, stateFile); err != nil {
			if errors.Is(err, errInteractiveCancelled) {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Cancelled. Decisions were saved and can be resumed later.")
				return err
			}
			return err
		}
	}

	switch format {
	case "json":
		data, err := overlaydiff.FormatJSON(changes)
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s\n", data)
	default:
		_, _ = fmt.Fprint(cmd.OutOrStdout(), overlaydiff.FormatTable(changes))
	}

	selected, skipped, templated, unset := summarizeDecisions(changes, snapshot.Decisions)
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Decision summary: selected=%d skipped=%d templated=%d unset=%d\n", selected, skipped, templated, unset)

	if stopExport || strings.TrimSpace(stopExportDir) != "" {
		exportDir := strings.TrimSpace(stopExportDir)
		if exportDir == "" {
			exportDir = defaultExportDir()
		}

		exported, err := exportSelected(sourceAbs, exportDir, changes, snapshot.Decisions)
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Exported %d selected entries to %s\n", exported, exportDir)
	}

	return nil
}

var errInteractiveCancelled = errors.New("interactive selection cancelled")

func runInteractiveSelection(in io.Reader, out io.Writer, changes []overlaydiff.Change, snapshot *overlaydiff.Snapshot, stateFile string) error {
	reader := bufio.NewReader(in)

	for idx, change := range changes {
		currentDecision := snapshot.Decisions[change.Path]
		if currentDecision == "" {
			currentDecision = overlaydiff.DecisionUnset
		}

		if currentDecision != overlaydiff.DecisionUnset {
			continue
		}

		for {
			_, _ = fmt.Fprintf(out, "[%d/%d] %s %s (%s) -> (y)es, (n)o, (t)emplated, (e)xit: ", idx+1, len(changes), change.Change, change.Path, change.Type)
			answer, err := reader.ReadString('\n')
			if err != nil {
				if errors.Is(err, io.EOF) {
					return errInteractiveCancelled
				}
				return err
			}

			switch strings.ToLower(strings.TrimSpace(answer)) {
			case "y", "yes":
				snapshot.Decisions[change.Path] = overlaydiff.DecisionYes
			case "n", "no":
				snapshot.Decisions[change.Path] = overlaydiff.DecisionNo
			case "t", "templated":
				snapshot.Decisions[change.Path] = overlaydiff.DecisionTemplated
			case "e", "exit":
				if err := overlaydiff.SaveSnapshot(stateFile, *snapshot); err != nil {
					return err
				}
				return errInteractiveCancelled
			default:
				_, _ = fmt.Fprintln(out, "Invalid answer, use y/n/t/e")
				continue
			}

			if err := overlaydiff.SaveSnapshot(stateFile, *snapshot); err != nil {
				return err
			}
			break
		}
	}

	return nil
}

func summarizeDecisions(changes []overlaydiff.Change, decisions map[string]overlaydiff.Decision) (selected int, skipped int, templated int, unset int) {
	for _, change := range changes {
		switch decisions[change.Path] {
		case overlaydiff.DecisionYes:
			selected++
		case overlaydiff.DecisionNo:
			skipped++
		case overlaydiff.DecisionTemplated:
			templated++
		default:
			unset++
		}
	}
	return
}

func defaultExportDir() string {
	return filepath.Join("/tmp", "wwclient-overlaydiff", time.Now().UTC().Format("20060102-150405"))
}

func exportSelected(sourceRoot, exportDir string, changes []overlaydiff.Change, decisions map[string]overlaydiff.Decision) (int, error) {
	exported := 0
	for _, change := range changes {
		if decisions[change.Path] != overlaydiff.DecisionYes {
			continue
		}
		if change.Source == nil {
			continue
		}

		relPath := strings.TrimPrefix(change.Path, "/")
		sourcePath := filepath.Join(sourceRoot, relPath)
		destPath := filepath.Join(exportDir, relPath)

		if err := osMkdirAll(filepath.Dir(destPath), 0o755); err != nil {
			return exported, err
		}

		switch change.Source.Type {
		case overlaydiff.EntryDir:
			if err := osMkdirAll(destPath, fs.FileMode(change.Source.Mode)); err != nil {
				return exported, err
			}
		case overlaydiff.EntrySymlink:
			_ = osRemoveAll(destPath)
			if err := osSymlink(change.Source.LinkTarget, destPath); err != nil {
				return exported, err
			}
		default:
			if err := copyFile(sourcePath, destPath, fs.FileMode(change.Source.Mode)); err != nil {
				return exported, err
			}
		}
		exported++
	}

	return exported, nil
}

var (
	osMkdirAll  = os.MkdirAll
	osRemoveAll = os.RemoveAll
	osSymlink   = os.Symlink
)

func copyFile(sourcePath, destPath string, mode fs.FileMode) error {
	input, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to open source file %s: %w", sourcePath, err)
	}
	defer input.Close()

	output, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create export file %s: %w", destPath, err)
	}

	if _, err = io.Copy(output, input); err != nil {
		_ = output.Close()
		return fmt.Errorf("failed to copy file %s to %s: %w", sourcePath, destPath, err)
	}

	if err = output.Close(); err != nil {
		return fmt.Errorf("failed to close export file %s: %w", destPath, err)
	}

	if err = os.Chmod(destPath, mode); err != nil {
		return fmt.Errorf("failed to set mode for %s: %w", destPath, err)
	}

	return nil
}
