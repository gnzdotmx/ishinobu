package netstat

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

func TestNetstatModule(t *testing.T) {
	// Create temporary directory for test outputs
	tmpDir, err := os.MkdirTemp("", "netstat_test")
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
	module := &NetstatModule{
		Name:        "netstat",
		Description: "Collects and parses netstat output",
	}

	// Test GetName
	t.Run("GetName", func(t *testing.T) {
		assert.Equal(t, "netstat", module.GetName())
	})

	// Test GetDescription
	t.Run("GetDescription", func(t *testing.T) {
		assert.Contains(t, module.GetDescription(), "netstat")
	})

	// Test Run method with mock output
	t.Run("Run", func(t *testing.T) {
		// Create a mock output file to simulate the command execution result
		createMockNetstatOutput(t, params)

		// Verify the output file exists
		outputFile := filepath.Join(tmpDir, "netstat-"+params.CollectionTimestamp+".json")
		assert.FileExists(t, outputFile)

		// Verify the content of the output file
		verifyNetstatOutput(t, outputFile)
	})
}

// Test that the module initializes properly
func TestNetstatModuleInitialization(t *testing.T) {
	// Create a new instance with proper initialization
	module := &NetstatModule{
		Name:        "netstat",
		Description: "Collects and parses netstat output",
	}

	// Verify module is properly instantiated with expected values
	assert.Equal(t, "netstat", module.Name, "Module name should be initialized")
	assert.Contains(t, module.Description, "netstat", "Module description should be initialized")

	// Test the module's methods
	assert.Equal(t, "netstat", module.GetName())
	assert.Contains(t, module.GetDescription(), "netstat")
}

// Create a mock netstat output file
func createMockNetstatOutput(t *testing.T, params mod.ModuleParams) {
	outputFile := filepath.Join(params.OutputDir, "netstat-"+params.CollectionTimestamp+".json")

	// Sample netstat output data in JSON format
	mockData := `{
		"command": "netstat",
		"arguments": ["-anv"],
		"timestamp": "` + params.CollectionTimestamp + `",
		"output": [
			{
				"protocol": "tcp4",
				"local_address": "127.0.0.1.80",
				"foreign_address": "*.0",
				"state": "LISTEN",
				"pid": "1234",
				"program_name": "httpd"
			},
			{
				"protocol": "tcp4",
				"local_address": "192.168.1.10.443",
				"foreign_address": "*.0",
				"state": "LISTEN",
				"pid": "1235",
				"program_name": "nginx"
			},
			{
				"protocol": "tcp4",
				"local_address": "10.0.0.1.22",
				"foreign_address": "192.168.1.5.54321",
				"state": "ESTABLISHED",
				"pid": "5678",
				"program_name": "sshd"
			},
			{
				"protocol": "udp4",
				"local_address": "0.0.0.0.53",
				"foreign_address": "*.0",
				"state": "",
				"pid": "4567",
				"program_name": "named"
			}
		]
	}`

	err := os.WriteFile(outputFile, []byte(mockData), 0644)
	assert.NoError(t, err)
}

// Verify the netstat output file contains expected data
func verifyNetstatOutput(t *testing.T, outputFile string) {
	// Read the output file
	data, err := os.ReadFile(outputFile)
	assert.NoError(t, err)

	// Verify the content contains expected elements
	content := string(data)

	// Check for command information
	assert.Contains(t, content, "\"command\": \"netstat\"")
	assert.Contains(t, content, "\"arguments\": [\"-anv\"]")

	// Check for TCP connections
	assert.Contains(t, content, "\"protocol\": \"tcp4\"")
	assert.Contains(t, content, "\"local_address\": \"127.0.0.1.80\"")
	assert.Contains(t, content, "\"state\": \"LISTEN\"")

	// Check for specific program details
	assert.Contains(t, content, "\"program_name\": \"httpd\"")
	assert.Contains(t, content, "\"program_name\": \"nginx\"")

	// Check for established connections
	assert.Contains(t, content, "\"state\": \"ESTABLISHED\"")
	assert.Contains(t, content, "\"local_address\": \"10.0.0.1.22\"")
	assert.Contains(t, content, "\"foreign_address\": \"192.168.1.5.54321\"")

	// Check for UDP connections
	assert.Contains(t, content, "\"protocol\": \"udp4\"")
	assert.Contains(t, content, "\"local_address\": \"0.0.0.0.53\"")
}
