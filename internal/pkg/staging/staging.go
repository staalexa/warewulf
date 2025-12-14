package staging

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

const (
	// DefaultStagingDir is the default directory for staging files
	DefaultStagingDir = "/var/lib/warewulf/staging"
	// StagingIndexFile is the file that tracks staged files
	StagingIndexFile = ".staging.json"
)

// StagedFile represents a file that has been staged
type StagedFile struct {
	SourcePath string    `json:"source_path"`
	DestPath   string    `json:"dest_path"`
	AddedAt    time.Time `json:"added_at"`
	Size       int64     `json:"size"`
}

// StagingIndex maintains the list of staged files
type StagingIndex struct {
	Files      map[string]StagedFile `json:"files"`
	StagingDir string                `json:"staging_dir"`
}

// NewStagingIndex creates a new staging index
func NewStagingIndex(stagingDir string) *StagingIndex {
	if stagingDir == "" {
		stagingDir = DefaultStagingDir
	}
	return &StagingIndex{
		Files:      make(map[string]StagedFile),
		StagingDir: stagingDir,
	}
}

// Load loads the staging index from disk
func (si *StagingIndex) Load() error {
	indexPath := filepath.Join(si.StagingDir, StagingIndexFile)
	data, err := os.ReadFile(indexPath)
	if os.IsNotExist(err) {
		// No index file exists yet, return empty index
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to read staging index: %w", err)
	}

	err = json.Unmarshal(data, si)
	if err != nil {
		return fmt.Errorf("failed to parse staging index: %w", err)
	}

	return nil
}

// Save saves the staging index to disk
func (si *StagingIndex) Save() error {
	// Ensure staging directory exists
	if err := os.MkdirAll(si.StagingDir, 0755); err != nil {
		return fmt.Errorf("failed to create staging directory: %w", err)
	}

	indexPath := filepath.Join(si.StagingDir, StagingIndexFile)
	data, err := json.MarshalIndent(si, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal staging index: %w", err)
	}

	err = os.WriteFile(indexPath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write staging index: %w", err)
	}

	return nil
}

// AddFile adds a file to the staging area
func (si *StagingIndex) AddFile(sourcePath, destPath string) error {
	// Validate source file exists
	srcInfo, err := os.Stat(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to stat source file: %w", err)
	}

	if srcInfo.IsDir() {
		return fmt.Errorf("source path is a directory, only files are supported")
	}

	// Get absolute path
	absSourcePath, err := filepath.Abs(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// If destPath is not specified, use the source file name
	if destPath == "" {
		destPath = "/" + filepath.Base(absSourcePath)
	}

	// Ensure destPath is absolute
	if !filepath.IsAbs(destPath) {
		destPath = "/" + destPath
	}

	// Copy file to staging area
	stagingPath := filepath.Join(si.StagingDir, "files", destPath)
	if err := os.MkdirAll(filepath.Dir(stagingPath), 0755); err != nil {
		return fmt.Errorf("failed to create staging directory: %w", err)
	}

	srcFile, err := os.Open(absSourcePath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(stagingPath)
	if err != nil {
		return fmt.Errorf("failed to create staging file: %w", err)
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return fmt.Errorf("failed to copy file to staging: %w", err)
	}

	// Add to index
	si.Files[destPath] = StagedFile{
		SourcePath: absSourcePath,
		DestPath:   destPath,
		AddedAt:    time.Now(),
		Size:       srcInfo.Size(),
	}

	wwlog.Info("Staged file: %s -> %s", absSourcePath, destPath)

	return nil
}

// RemoveFile removes a file from the staging area
func (si *StagingIndex) RemoveFile(destPath string) error {
	// Ensure destPath is absolute
	if !filepath.IsAbs(destPath) {
		destPath = "/" + destPath
	}

	if _, exists := si.Files[destPath]; !exists {
		return fmt.Errorf("file not found in staging: %s", destPath)
	}

	// Remove from staging directory
	stagingPath := filepath.Join(si.StagingDir, "files", destPath)
	if err := os.Remove(stagingPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove staged file: %w", err)
	}

	// Remove from index
	delete(si.Files, destPath)

	wwlog.Info("Removed staged file: %s", destPath)

	return nil
}

// Clear removes all files from the staging area
func (si *StagingIndex) Clear() error {
	// Remove all staged files
	filesDir := filepath.Join(si.StagingDir, "files")
	if err := os.RemoveAll(filesDir); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove staged files: %w", err)
	}

	// Clear index
	si.Files = make(map[string]StagedFile)

	wwlog.Info("Cleared all staged files")

	return nil
}

// List returns all staged files
func (si *StagingIndex) List() []StagedFile {
	files := make([]StagedFile, 0, len(si.Files))
	for _, file := range si.Files {
		files = append(files, file)
	}
	return files
}

// GetStagedFilePath returns the path to a staged file in the staging directory
func (si *StagingIndex) GetStagedFilePath(destPath string) string {
	return filepath.Join(si.StagingDir, "files", destPath)
}
