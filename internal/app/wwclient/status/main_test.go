package status

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/warewulf/warewulf/internal/pkg/staging"
)

func TestStatusCommand(t *testing.T) {
	tests := map[string]struct {
		setupFiles  map[string]string
		errExpected bool
	}{
		"empty staging": {
			setupFiles: map[string]string{},
		},
		"with staged files": {
			setupFiles: map[string]string{
				"test1.txt": "content1",
				"test2.txt": "content2",
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tmpDir := t.TempDir()
			stagingDir := filepath.Join(tmpDir, "staging")
			StagingDir = stagingDir

			// Setup staging area
			si := staging.NewStagingIndex(stagingDir)
			for file, content := range tt.setupFiles {
				srcPath := filepath.Join(tmpDir, file)
				err := os.WriteFile(srcPath, []byte(content), 0644)
				assert.NoError(t, err)
				err = si.AddFile(srcPath, "/"+file)
				assert.NoError(t, err)
			}
			if len(tt.setupFiles) > 0 {
				err := si.Save()
				assert.NoError(t, err)
			}

			cmd := GetCommand()
			cmd.SetArgs([]string{})
			stdout := new(bytes.Buffer)
			cmd.SetOut(stdout)
			stderr := new(bytes.Buffer)
			cmd.SetErr(stderr)

			err := cmd.Execute()
			if tt.errExpected {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
