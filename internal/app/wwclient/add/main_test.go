package add

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddCommand(t *testing.T) {
	tests := map[string]struct {
		setupFiles  map[string]string
		args        []string
		errExpected bool
	}{
		"add file with destination": {
			setupFiles: map[string]string{"test.txt": "test content"},
			args:       []string{"test.txt", "/etc/test.txt"},
		},
		"add file without destination": {
			setupFiles: map[string]string{"test.txt": "test content"},
			args:       []string{"test.txt"},
		},
		"add missing file": {
			setupFiles:  map[string]string{},
			args:        []string{"missing.txt"},
			errExpected: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tmpDir := t.TempDir()
			stagingDir := filepath.Join(tmpDir, "staging")
			StagingDir = stagingDir

			// Change to temp directory
			wd, err := os.Getwd()
			assert.NoError(t, err)
			defer func() { assert.NoError(t, os.Chdir(wd)) }()
			assert.NoError(t, os.Chdir(tmpDir))

			// Create test files
			for file, content := range tt.setupFiles {
				err := os.WriteFile(file, []byte(content), 0644)
				assert.NoError(t, err)
			}

			cmd := GetCommand()
			cmd.SetArgs(tt.args)
			stdout := new(bytes.Buffer)
			cmd.SetOut(stdout)
			stderr := new(bytes.Buffer)
			cmd.SetErr(stderr)

			err = cmd.Execute()
			if tt.errExpected {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
