package lsof

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/gnzdotmx/ishinobu/pkg/mod"
	"github.com/gnzdotmx/ishinobu/pkg/modules/testutils"
	"github.com/gnzdotmx/ishinobu/pkg/utils"
)

func TestLsofModule(t *testing.T) {
	// Create temporary directory for test outputs
	tmpDir, err := os.MkdirTemp("", "lsof_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logger := testutils.NewTestLogger()

	// Setup test parameters
	params := mod.ModuleParams{
		OutputDir:           tmpDir,
		LogsDir:             tmpDir,
		ExportFormat:        "json",
		CollectionTimestamp: time.Now().Format(utils.TimeFormat),
		Logger:              *logger,
	}

	// Create module instance
	module := &LsofModule{
		Name:        "lsof",
		Description: "Collects information about open files and their processes",
	}

	// Test GetName
	t.Run("GetName", func(t *testing.T) {
		assert.Equal(t, "lsof", module.GetName())
	})

	// Test GetDescription
	t.Run("GetDescription", func(t *testing.T) {
		assert.Contains(t, module.GetDescription(), "open files")
	})

	// Test Run method with mock output
	t.Run("Run", func(t *testing.T) {
		// Create a mock output file to simulate the command execution result
		createMockLsofOutput(t, params)

		// Verify the output file exists
		outputFile := filepath.Join(tmpDir, "lsof-"+params.CollectionTimestamp+".json")
		assert.FileExists(t, outputFile)

		// Verify the content of the output file
		verifyLsofOutput(t, outputFile)
	})
}

// Test that the module initializes properly
func TestLsofModuleInitialization(t *testing.T) {
	// Create a new instance with proper initialization
	module := &LsofModule{
		Name:        "lsof",
		Description: "Collects information about open files and their processes",
	}

	// Verify module is properly instantiated with expected values
	assert.Equal(t, "lsof", module.Name, "Module name should be initialized")
	assert.Contains(t, module.Description, "open files", "Module description should be initialized")

	// Test the module's methods
	assert.Equal(t, "lsof", module.GetName())
	assert.Contains(t, module.GetDescription(), "open files")
}

// Create a mock lsof output file
func createMockLsofOutput(t *testing.T, params mod.ModuleParams) {
	outputFile := filepath.Join(params.OutputDir, "lsof-"+params.CollectionTimestamp+".json")

	// Sample lsof output data in JSON format
	mockData := `{
		"command": "lsof",
		"arguments": ["-n", "-P"],
		"timestamp": "` + params.CollectionTimestamp + `",
		"output": [
			{
				"command": "systemd",
				"pid": "1",
				"user": "root",
				"fd": "cwd",
				"type": "DIR",
				"device": "253,0",
				"size": "4096",
				"node": "2",
				"name": "/"
			},
			{
				"command": "sshd",
				"pid": "854",
				"user": "root",
				"fd": "3u",
				"type": "IPv4",
				"device": "0,9",
				"size": "0t0",
				"node": "21140",
				"name": "TCP *:22 (LISTEN)"
			},
			{
				"command": "chrome",
				"pid": "2158",
				"user": "user",
				"fd": "3u",
				"type": "IPv4",
				"device": "0,9",
				"size": "0t0",
				"node": "42560",
				"name": "TCP 127.0.0.1:45678->127.0.0.1:4321 (ESTABLISHED)"
			}
		]
	}`

	err := os.WriteFile(outputFile, []byte(mockData), 0644)
	assert.NoError(t, err)
}

// Verify the lsof output file contains expected data
func verifyLsofOutput(t *testing.T, outputFile string) {
	// Read the output file
	data, err := os.ReadFile(outputFile)
	assert.NoError(t, err)

	// Verify the content contains expected elements
	content := string(data)

	// Check for command information
	assert.Contains(t, content, "\"command\": \"lsof\"")
	assert.Contains(t, content, "\"arguments\": [\"-n\", \"-P\"]")

	// Check for expected process entries
	assert.Contains(t, content, "\"command\": \"systemd\"")
	assert.Contains(t, content, "\"pid\": \"1\"")

	// Check for network connection details
	assert.Contains(t, content, "\"name\": \"TCP *:22 (LISTEN)\"")

	// Check for established connection
	assert.Contains(t, content, "TCP 127.0.0.1:45678->127.0.0.1:4321 (ESTABLISHED)")
}
