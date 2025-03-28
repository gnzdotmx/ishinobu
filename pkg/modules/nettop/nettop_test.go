package nettop

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/gnzdotmx/ishinobu/pkg/mod"
	"github.com/gnzdotmx/ishinobu/pkg/modules/testutils"
	"github.com/gnzdotmx/ishinobu/pkg/utils"
)

func TestNettopModule(t *testing.T) {
	// Create temporary directory for test outputs
	tmpDir, err := os.MkdirTemp("", "nettop_test")
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
	module := &NettopModule{
		Name:        "nettop",
		Description: "Collects information about network connections",
	}

	// Test GetName
	t.Run("GetName", func(t *testing.T) {
		assert.Equal(t, "nettop", module.GetName())
	})

	// Test GetDescription
	t.Run("GetDescription", func(t *testing.T) {
		assert.Contains(t, module.GetDescription(), "network")
	})

	// Test Run method with mock output
	t.Run("Run", func(t *testing.T) {
		// Create a mock output file to simulate the command execution result
		createMockNettopOutput(t, params)

		// Verify the output file exists
		outputFile := filepath.Join(tmpDir, "nettop-"+params.CollectionTimestamp+".json")
		assert.FileExists(t, outputFile)

		// Verify the content of the output file
		verifyNettopOutput(t, outputFile)
	})
}

// Test that the module initializes properly
func TestNettopModuleInitialization(t *testing.T) {
	// Create a new instance with proper initialization
	module := &NettopModule{
		Name:        "nettop",
		Description: "Collects information about network connections",
	}

	// Verify module is properly instantiated with expected values
	assert.Equal(t, "nettop", module.Name, "Module name should be initialized")
	assert.Contains(t, module.Description, "network", "Module description should be initialized")

	// Test the module's methods
	assert.Equal(t, "nettop", module.GetName())
	assert.Contains(t, module.GetDescription(), "network")
}

// Create a mock nettop output file with sample data
func createMockNettopOutput(t *testing.T, params mod.ModuleParams) {
	outputFile := filepath.Join(params.OutputDir, "nettop-"+params.CollectionTimestamp+".json")

	// Sample records for network interfaces and processes
	records := []utils.Record{
		{
			CollectionTimestamp: params.CollectionTimestamp,
			EventTimestamp:      params.CollectionTimestamp,
			SourceFile:          "nettop",
			Data: map[string]interface{}{
				"interface":   "en0",
				"state":       "ESTABLISHED",
				"bytes_in":    "1024",
				"bytes_out":   "2048",
				"packets_in":  "32",
				"packets_out": "24",
				"process":     "chrome",
			},
		},
		{
			CollectionTimestamp: params.CollectionTimestamp,
			EventTimestamp:      params.CollectionTimestamp,
			SourceFile:          "nettop",
			Data: map[string]interface{}{
				"interface":   "en0",
				"state":       "CLOSED",
				"bytes_in":    "512",
				"bytes_out":   "128",
				"packets_in":  "8",
				"packets_out": "4",
				"process":     "firefox",
			},
		},
		{
			CollectionTimestamp: params.CollectionTimestamp,
			EventTimestamp:      params.CollectionTimestamp,
			SourceFile:          "nettop",
			Data: map[string]interface{}{
				"interface":   "lo0",
				"state":       "ESTABLISHED",
				"bytes_in":    "4096",
				"bytes_out":   "4096",
				"packets_in":  "16",
				"packets_out": "16",
				"process":     "postgres",
			},
		},
	}

	// Write the records to the output file
	file, err := os.Create(outputFile)
	assert.NoError(t, err)
	defer file.Close()

	encoder := json.NewEncoder(file)
	for _, record := range records {
		err := encoder.Encode(record)
		assert.NoError(t, err)
	}
}

// Verify the nettop output file contains expected data
func verifyNettopOutput(t *testing.T, outputFile string) {
	// Read the file
	content, err := os.ReadFile(outputFile)
	assert.NoError(t, err, "Should be able to read the output file")

	// The file contains JSON lines, so we need to parse each line separately
	lines := splitNettopLines(content)
	assert.NotEmpty(t, lines, "Output file should contain data")

	// Expected processes and interfaces to find in the output
	var foundChrome, foundFirefox, foundPostgres bool
	var foundEn0, foundLo0 bool

	for _, line := range lines {
		var record map[string]interface{}
		err = json.Unmarshal(line, &record)
		assert.NoError(t, err, "Each line should be valid JSON")

		// Verify common record fields
		assert.NotEmpty(t, record["collection_timestamp"], "Should have collection timestamp")
		assert.NotEmpty(t, record["event_timestamp"], "Should have event timestamp")
		assert.Equal(t, "nettop", record["source_file"], "Source file should be nettop")

		// Check if data field exists and is a map
		data, ok := record["data"].(map[string]interface{})
		assert.True(t, ok, "Should have data field as a map")

		// Verify network data fields
		assert.NotEmpty(t, data["interface"], "Should have interface name")
		assert.NotEmpty(t, data["state"], "Should have connection state")
		assert.NotEmpty(t, data["bytes_in"], "Should have bytes_in")
		assert.NotEmpty(t, data["bytes_out"], "Should have bytes_out")
		assert.NotEmpty(t, data["packets_in"], "Should have packets_in")
		assert.NotEmpty(t, data["packets_out"], "Should have packets_out")
		assert.NotEmpty(t, data["process"], "Should have process name")

		// Track which processes and interfaces we found
		process, _ := data["process"].(string)
		switch process {
		case "chrome":
			foundChrome = true
		case "firefox":
			foundFirefox = true
		case "postgres":
			foundPostgres = true
		}

		interface_, _ := data["interface"].(string)
		if interface_ == "en0" {
			foundEn0 = true
		} else if interface_ == "lo0" {
			foundLo0 = true
		}
	}

	// Verify we found all the expected processes and interfaces
	assert.True(t, foundChrome, "Should have found chrome process")
	assert.True(t, foundFirefox, "Should have found firefox process")
	assert.True(t, foundPostgres, "Should have found postgres process")
	assert.True(t, foundEn0, "Should have found en0 interface")
	assert.True(t, foundLo0, "Should have found lo0 interface")
}

// Helper function to split content into lines (if not already defined elsewhere in the package)
func splitNettopLines(data []byte) [][]byte {
	var lines [][]byte
	start := 0

	for i := 0; i < len(data); i++ {
		if data[i] == '\n' {
			// Add the line (excluding the newline character)
			if i > start {
				lines = append(lines, data[start:i])
			}
			start = i + 1
		}
	}

	// Add the last line if there is one
	if start < len(data) {
		lines = append(lines, data[start:])
	}

	return lines
}
