package staging

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewStagingIndex(t *testing.T) {
	// Test with default directory
	si := NewStagingIndex("")
	assert.Equal(t, DefaultStagingDir, si.StagingDir)
	assert.NotNil(t, si.Files)
	assert.Equal(t, 0, len(si.Files))

	// Test with custom directory
	customDir := "/tmp/custom-staging"
	si = NewStagingIndex(customDir)
	assert.Equal(t, customDir, si.StagingDir)
}

func TestStagingIndex_AddFile(t *testing.T) {
	tmpDir := t.TempDir()
	stagingDir := filepath.Join(tmpDir, "staging")
	si := NewStagingIndex(stagingDir)

	// Create a test file
	testFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(testFile, []byte("test content"), 0644)
	assert.NoError(t, err)

	// Add file without destination
	err = si.AddFile(testFile, "")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(si.Files))
	assert.Contains(t, si.Files, "/test.txt")

	// Check that file was copied to staging
	stagedPath := si.GetStagedFilePath("/test.txt")
	assert.FileExists(t, stagedPath)

	// Add file with destination
	err = si.AddFile(testFile, "/etc/config.txt")
	assert.NoError(t, err)
	assert.Equal(t, 2, len(si.Files))
	assert.Contains(t, si.Files, "/etc/config.txt")
}

func TestStagingIndex_AddFile_Directory(t *testing.T) {
	tmpDir := t.TempDir()
	stagingDir := filepath.Join(tmpDir, "staging")
	si := NewStagingIndex(stagingDir)

	// Try to add a directory (should fail)
	testDir := filepath.Join(tmpDir, "testdir")
	err := os.MkdirAll(testDir, 0755)
	assert.NoError(t, err)

	err = si.AddFile(testDir, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "directory")
}

func TestStagingIndex_RemoveFile(t *testing.T) {
	tmpDir := t.TempDir()
	stagingDir := filepath.Join(tmpDir, "staging")
	si := NewStagingIndex(stagingDir)

	// Create and add a test file
	testFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(testFile, []byte("test content"), 0644)
	assert.NoError(t, err)

	err = si.AddFile(testFile, "/test.txt")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(si.Files))

	// Remove the file
	err = si.RemoveFile("/test.txt")
	assert.NoError(t, err)
	assert.Equal(t, 0, len(si.Files))

	// Try to remove non-existent file
	err = si.RemoveFile("/nonexistent.txt")
	assert.Error(t, err)
}

func TestStagingIndex_Clear(t *testing.T) {
	tmpDir := t.TempDir()
	stagingDir := filepath.Join(tmpDir, "staging")
	si := NewStagingIndex(stagingDir)

	// Create and add multiple test files
	for i := 0; i < 3; i++ {
		testFile := filepath.Join(tmpDir, fmt.Sprintf("test%d.txt", i))
		err := os.WriteFile(testFile, []byte("test content"), 0644)
		assert.NoError(t, err)
		err = si.AddFile(testFile, "")
		assert.NoError(t, err)
	}

	assert.Equal(t, 3, len(si.Files))

	// Clear all files
	err := si.Clear()
	assert.NoError(t, err)
	assert.Equal(t, 0, len(si.Files))
}

func TestStagingIndex_SaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	stagingDir := filepath.Join(tmpDir, "staging")

	// Create and save staging index
	si1 := NewStagingIndex(stagingDir)
	testFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(testFile, []byte("test content"), 0644)
	assert.NoError(t, err)

	err = si1.AddFile(testFile, "/test.txt")
	assert.NoError(t, err)

	err = si1.Save()
	assert.NoError(t, err)

	// Load staging index
	si2 := NewStagingIndex(stagingDir)
	err = si2.Load()
	assert.NoError(t, err)

	assert.Equal(t, len(si1.Files), len(si2.Files))
	assert.Contains(t, si2.Files, "/test.txt")
}

func TestStagingIndex_List(t *testing.T) {
	tmpDir := t.TempDir()
	stagingDir := filepath.Join(tmpDir, "staging")
	si := NewStagingIndex(stagingDir)

	// Empty list
	files := si.List()
	assert.Equal(t, 0, len(files))

	// Add files
	testFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(testFile, []byte("test content"), 0644)
	assert.NoError(t, err)

	err = si.AddFile(testFile, "/test.txt")
	assert.NoError(t, err)

	files = si.List()
	assert.Equal(t, 1, len(files))
	assert.Equal(t, "/test.txt", files[0].DestPath)
}
