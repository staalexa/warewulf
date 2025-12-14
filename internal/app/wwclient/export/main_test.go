package export

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/staging"
)

func TestExportCommand(t *testing.T) {
	tmpDir := t.TempDir()
	stagingDir := filepath.Join(tmpDir, "staging")
	overlayDir := filepath.Join(tmpDir, "overlays")

	// Setup Warewulf configuration
	conf := warewulfconf.Get()
	conf.Paths.WWOverlaydir = overlayDir

	// Create test files and stage them
	testFile1 := filepath.Join(tmpDir, "test1.txt")
	err := os.WriteFile(testFile1, []byte("content1"), 0644)
	assert.NoError(t, err)

	testFile2 := filepath.Join(tmpDir, "test2.txt")
	err = os.WriteFile(testFile2, []byte("content2"), 0644)
	assert.NoError(t, err)

	si := staging.NewStagingIndex(stagingDir)
	err = si.AddFile(testFile1, "/etc/test1.txt")
	assert.NoError(t, err)
	err = si.AddFile(testFile2, "/test2.txt")
	assert.NoError(t, err)
	err = si.Save()
	assert.NoError(t, err)

	// Test export
	StagingDir = stagingDir
	ClearAfter = false

	cmd := GetCommand()
	cmd.SetArgs([]string{"test-overlay"})
	stdout := new(bytes.Buffer)
	cmd.SetOut(stdout)
	stderr := new(bytes.Buffer)
	cmd.SetErr(stderr)

	err = cmd.Execute()
	assert.NoError(t, err)

	// Verify files were exported
	exportedFile1 := filepath.Join(overlayDir, "test-overlay", "rootfs", "etc", "test1.txt")
	assert.FileExists(t, exportedFile1)
	content, err := os.ReadFile(exportedFile1)
	assert.NoError(t, err)
	assert.Equal(t, "content1", string(content))

	exportedFile2 := filepath.Join(overlayDir, "test-overlay", "rootfs", "test2.txt")
	assert.FileExists(t, exportedFile2)
	content, err = os.ReadFile(exportedFile2)
	assert.NoError(t, err)
	assert.Equal(t, "content2", string(content))

	// Verify staging area not cleared
	si2 := staging.NewStagingIndex(stagingDir)
	err = si2.Load()
	assert.NoError(t, err)
	assert.Equal(t, 2, len(si2.Files))
}

func TestExportCommand_WithClear(t *testing.T) {
	tmpDir := t.TempDir()
	stagingDir := filepath.Join(tmpDir, "staging")
	overlayDir := filepath.Join(tmpDir, "overlays")

	// Setup Warewulf configuration
	conf := warewulfconf.Get()
	conf.Paths.WWOverlaydir = overlayDir

	// Create test file and stage it
	testFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(testFile, []byte("content"), 0644)
	assert.NoError(t, err)

	si := staging.NewStagingIndex(stagingDir)
	err = si.AddFile(testFile, "/test.txt")
	assert.NoError(t, err)
	err = si.Save()
	assert.NoError(t, err)

	// Test export with clear
	StagingDir = stagingDir
	ClearAfter = true

	cmd := GetCommand()
	cmd.SetArgs([]string{"test-overlay"})
	stdout := new(bytes.Buffer)
	cmd.SetOut(stdout)
	stderr := new(bytes.Buffer)
	cmd.SetErr(stderr)

	err = cmd.Execute()
	assert.NoError(t, err)

	// Verify staging area was cleared
	si2 := staging.NewStagingIndex(stagingDir)
	err = si2.Load()
	assert.NoError(t, err)
	assert.Equal(t, 0, len(si2.Files))
}

func TestExportCommand_EmptyStaging(t *testing.T) {
	tmpDir := t.TempDir()
	stagingDir := filepath.Join(tmpDir, "staging")
	overlayDir := filepath.Join(tmpDir, "overlays")

	// Setup Warewulf configuration
	conf := warewulfconf.Get()
	conf.Paths.WWOverlaydir = overlayDir

	// Test export with empty staging
	StagingDir = stagingDir
	ClearAfter = false

	cmd := GetCommand()
	cmd.SetArgs([]string{"test-overlay"})
	stdout := new(bytes.Buffer)
	cmd.SetOut(stdout)
	stderr := new(bytes.Buffer)
	cmd.SetErr(stderr)

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no files staged")
}
