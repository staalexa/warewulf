package capture

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStartStopCommand_TableOutput(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	stateFile := filepath.Join(tmpDir, "capture.json")
	if !assert.NoError(t, os.MkdirAll(sourceDir, 0755)) {
		return
	}

	if !assert.NoError(t, os.WriteFile(filepath.Join(sourceDir, "file.txt"), []byte("old"), 0644)) {
		return
	}

	startCmd := GetStartCommand()
	startCmd.SetArgs([]string{"--source", sourceDir, "--state-file", stateFile})
	startOut := new(bytes.Buffer)
	startCmd.SetOut(startOut)
	startCmd.SetErr(new(bytes.Buffer))

	if !assert.NoError(t, startCmd.Execute()) {
		return
	}

	if !assert.NoError(t, os.WriteFile(filepath.Join(sourceDir, "file.txt"), []byte("new"), 0644)) {
		return
	}

	stopCmd := GetStopCommand()
	stopCmd.SetArgs([]string{"--source", sourceDir, "--state-file", stateFile})
	stopOut := new(bytes.Buffer)
	stopCmd.SetOut(stopOut)
	stopCmd.SetErr(new(bytes.Buffer))

	if !assert.NoError(t, stopCmd.Execute()) {
		return
	}

	assert.Contains(t, stopOut.String(), "CHANGE")
	assert.Contains(t, stopOut.String(), "modified")
	assert.Contains(t, stopOut.String(), "/file.txt")
	assert.Contains(t, stopOut.String(), "Decision summary:")

	_, err := os.Stat(stateFile)
	assert.NoError(t, err)
}

func TestStartStopCommand_JSONOutput(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	stateFile := filepath.Join(tmpDir, "capture.json")
	if !assert.NoError(t, os.MkdirAll(sourceDir, 0755)) {
		return
	}

	if !assert.NoError(t, os.WriteFile(filepath.Join(sourceDir, "new.txt"), []byte("x"), 0644)) {
		return
	}

	startCmd := GetStartCommand()
	startCmd.SetArgs([]string{"--source", sourceDir, "--state-file", stateFile})
	startCmd.SetOut(new(bytes.Buffer))
	startCmd.SetErr(new(bytes.Buffer))

	if !assert.NoError(t, startCmd.Execute()) {
		return
	}

	if !assert.NoError(t, os.WriteFile(filepath.Join(sourceDir, "new.txt"), []byte("y"), 0644)) {
		return
	}

	stopCmd := GetStopCommand()
	stopCmd.SetArgs([]string{"--source", sourceDir, "--state-file", stateFile, "--format", "json"})
	stopOut := new(bytes.Buffer)
	stopErr := new(bytes.Buffer)
	stopCmd.SetOut(stopOut)
	stopCmd.SetErr(stopErr)

	if !assert.NoError(t, stopCmd.Execute()) {
		return
	}

	assert.Contains(t, stopOut.String(), "\"change\": \"modified\"")
	assert.Contains(t, stopOut.String(), "\"path\": \"/new.txt\"")
	assert.NotContains(t, stopOut.String(), "Decision summary:")

	var payload []map[string]interface{}
	if !assert.NoError(t, json.Unmarshal(stopOut.Bytes(), &payload)) {
		return
	}
	assert.NotEmpty(t, payload)

	assert.Contains(t, stopErr.String(), "Decision summary:")
}

func TestStopCommand_InteractivePersistsDecision(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	stateFile := filepath.Join(tmpDir, "capture.json")
	if !assert.NoError(t, os.MkdirAll(sourceDir, 0o755)) {
		return
	}

	if !assert.NoError(t, os.WriteFile(filepath.Join(sourceDir, "new.txt"), []byte("old"), 0o644)) {
		return
	}

	startCmd := GetStartCommand()
	startCmd.SetArgs([]string{"--source", sourceDir, "--state-file", stateFile})
	startCmd.SetOut(new(bytes.Buffer))
	startCmd.SetErr(new(bytes.Buffer))
	if !assert.NoError(t, startCmd.Execute()) {
		return
	}

	if !assert.NoError(t, os.WriteFile(filepath.Join(sourceDir, "new.txt"), []byte("new"), 0o644)) {
		return
	}

	stopCmd := GetStopCommand()
	stopCmd.SetArgs([]string{"--source", sourceDir, "--state-file", stateFile, "--interactive"})
	stopCmd.SetIn(strings.NewReader("y\n"))
	stopCmd.SetOut(new(bytes.Buffer))
	stopCmd.SetErr(new(bytes.Buffer))

	if !assert.NoError(t, stopCmd.Execute()) {
		return
	}

	data, err := os.ReadFile(stateFile)
	if !assert.NoError(t, err) {
		return
	}
	assert.Contains(t, string(data), "\"/new.txt\": \"yes\"")
}

func TestStopCommand_FilterAndExportSelected(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	stateFile := filepath.Join(tmpDir, "capture.json")
	exportDir := filepath.Join(tmpDir, "export")

	if !assert.NoError(t, os.MkdirAll(sourceDir, 0o755)) {
		return
	}
	if !assert.NoError(t, os.WriteFile(filepath.Join(sourceDir, "a.txt"), []byte("a-old"), 0o644)) {
		return
	}
	if !assert.NoError(t, os.WriteFile(filepath.Join(sourceDir, "b.txt"), []byte("b-old"), 0o644)) {
		return
	}

	startCmd := GetStartCommand()
	startCmd.SetArgs([]string{"--source", sourceDir, "--state-file", stateFile})
	startCmd.SetOut(new(bytes.Buffer))
	startCmd.SetErr(new(bytes.Buffer))
	if !assert.NoError(t, startCmd.Execute()) {
		return
	}

	if !assert.NoError(t, os.WriteFile(filepath.Join(sourceDir, "a.txt"), []byte("a-new"), 0o644)) {
		return
	}
	if !assert.NoError(t, os.WriteFile(filepath.Join(sourceDir, "b.txt"), []byte("b-new"), 0o644)) {
		return
	}

	stopCmd := GetStopCommand()
	stopCmd.SetArgs([]string{"--source", sourceDir, "--state-file", stateFile, "--interactive", "--only", "modified", "--path-prefix", "/a", "--export", "--export-dir", exportDir})
	stopCmd.SetIn(strings.NewReader("y\n"))
	stopCmd.SetOut(new(bytes.Buffer))
	stopCmd.SetErr(new(bytes.Buffer))

	if !assert.NoError(t, stopCmd.Execute()) {
		return
	}

	data, err := os.ReadFile(filepath.Join(exportDir, "a.txt"))
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, "a-new", string(data))

	_, err = os.Stat(filepath.Join(exportDir, "b.txt"))
	assert.Error(t, err)
}

func TestStopCommand_MissingSnapshot(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	stateFile := filepath.Join(tmpDir, "missing.json")
	if !assert.NoError(t, os.MkdirAll(sourceDir, 0755)) {
		return
	}

	stopCmd := GetStopCommand()
	stopCmd.SetArgs([]string{"--source", sourceDir, "--state-file", stateFile})
	stopCmd.SetOut(new(bytes.Buffer))
	stopCmd.SetErr(new(bytes.Buffer))

	assert.Error(t, stopCmd.Execute())
}
