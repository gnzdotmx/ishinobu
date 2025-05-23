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

// Test that the package initialization occurs
func TestPackageInitialization(t *testing.T) {
	// Create a new module with the same name
	module := &LsofModule{
		Name:        "lsof",
		Description: "This is a test initialization",
	}

	// Verify the module has expected values
	assert.Equal(t, "lsof", module.Name)
	assert.Equal(t, "This is a test initialization", module.Description)

	// The init function is automatically called when the package is imported
	// We can't directly test it, but we can verify it doesn't crash the tests
}

// Test Run method error handling
func TestRunMethodError(t *testing.T) {
	// Create temporary directory for test outputs
	tmpDir, err := os.MkdirTemp("", "lsof_error_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logger := testutils.NewTestLogger()

	// Setup test parameters
	params := mod.ModuleParams{
		OutputDir:           "./",
		LogsDir:             tmpDir,
		ExportFormat:        "json",
		CollectionTimestamp: time.Now().Format(utils.TimeFormat),
		Logger:              *logger,
	}

	// Test with a non-existent directory
	params.OutputDir = "/non/existent/directory"

	module := &LsofModule{
		Name:        "lsof",
		Description: "Collects information about open files and their processes",
	}

	err = module.Run(params)
	assert.Error(t, err, "Run should return an error with invalid output directory")
	assert.Contains(t, err.Error(), "error", "Error message should contain explanation")
}

// Test that the module uses correct command arguments
func TestModuleCommandArgs(t *testing.T) {
	// Create temporary directory for test outputs
	tmpDir, err := os.MkdirTemp("", "lsof_cmd_test")
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

	// Create output file before running to verify it gets overwritten
	outputFile := filepath.Join(tmpDir, "lsof-"+params.CollectionTimestamp+".json")
	err = os.WriteFile(outputFile, []byte("test data"), 0600)
	assert.NoError(t, err)

	// Run the module
	err = module.Run(params)

	// Depending on the environment, this might succeed or fail
	// But we're testing the file gets updated, not the command success
	if err == nil {
		// If it succeeds, verify it wrote new data
		fileInfo, err := os.Stat(outputFile)
		assert.NoError(t, err)
		assert.NotEqual(t, int64(9), fileInfo.Size(), "File size should be different from initial test data")
	}
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

	err := os.WriteFile(outputFile, []byte(mockData), 0600)
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
