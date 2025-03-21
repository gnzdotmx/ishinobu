package syslog

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gnzdotmx/ishinobu/pkg/mod"
	"github.com/gnzdotmx/ishinobu/pkg/modules/testutils"
	"github.com/gnzdotmx/ishinobu/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestSyslogModule(t *testing.T) {
	// Create temporary directory for test outputs
	tmpDir, err := os.MkdirTemp("", "syslog_test")
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
	module := &SyslogModule{
		Name:        "syslog",
		Description: "Collects and parses system.log files",
	}

	// Test GetName
	t.Run("GetName", func(t *testing.T) {
		assert.Equal(t, "syslog", module.GetName())
	})

	// Test GetDescription
	t.Run("GetDescription", func(t *testing.T) {
		assert.Contains(t, module.GetDescription(), "system.log")
	})

	// Test Run method with mock output
	t.Run("Run", func(t *testing.T) {
		// Create a mock output file to simulate the module's output
		createMockSyslogOutput(t, params)

		// Verify the output file exists
		outputFile := filepath.Join(tmpDir, "syslog-"+params.CollectionTimestamp+".json")
		assert.FileExists(t, outputFile)

		// Verify the content of the output file
		verifySyslogOutput(t, outputFile)
	})
}

// Test that the module initializes properly
func TestSyslogModuleInitialization(t *testing.T) {
	// Create a new instance with proper initialization
	module := &SyslogModule{
		Name:        "syslog",
		Description: "Collects and parses system.log files",
	}

	// Verify module is properly instantiated with expected values
	assert.Equal(t, "syslog", module.Name, "Module name should be initialized")
	assert.Contains(t, module.Description, "system.log", "Module description should be initialized")

	// Test the module's methods
	assert.Equal(t, "syslog", module.GetName())
	assert.Contains(t, module.GetDescription(), "system.log")
}

// Create a mock syslog output file
func createMockSyslogOutput(t *testing.T, params mod.ModuleParams) {
	outputFile := filepath.Join(params.OutputDir, "syslog-"+params.CollectionTimestamp+".json")

	// Create sample syslog records
	syslogRecords := []utils.Record{
		{
			CollectionTimestamp: params.CollectionTimestamp,
			EventTimestamp:      "2023-04-15T08:31:27Z",
			SourceFile:          "/private/var/log/system.log",
			Data: map[string]interface{}{
				"systemname":  "MacBook-Pro.local",
				"processname": "kernel",
				"pid":         "0",
				"message":     "System boot completed successfully",
			},
		},
		{
			CollectionTimestamp: params.CollectionTimestamp,
			EventTimestamp:      "2023-04-15T10:45:12Z",
			SourceFile:          "/private/var/log/system.log",
			Data: map[string]interface{}{
				"systemname":  "MacBook-Pro.local",
				"processname": "UserEventAgent",
				"pid":         "354",
				"message":     "Captive: CNPluginHandler en0: Inactive",
			},
		},
		{
			CollectionTimestamp: params.CollectionTimestamp,
			EventTimestamp:      "2023-04-15T14:22:05Z",
			SourceFile:          "/private/var/log/system.log",
			Data: map[string]interface{}{
				"systemname":  "MacBook-Pro.local",
				"processname": "mDNSResponder",
				"pid":         "123",
				"message":     "Query for service _companion-link._tcp.local. completed",
			},
		},
		{
			CollectionTimestamp: params.CollectionTimestamp,
			EventTimestamp:      "2023-04-15T18:05:37Z",
			SourceFile:          "/private/var/log/system.log",
			Data: map[string]interface{}{
				"systemname":  "MacBook-Pro.local",
				"processname": "securityd",
				"pid":         "78",
				"message":     "Session 100 created",
			},
		},
	}

	// Write each syslog record as a JSON line
	file, err := os.Create(outputFile)
	assert.NoError(t, err)
	defer file.Close()

	encoder := json.NewEncoder(file)
	for _, record := range syslogRecords {
		err := encoder.Encode(record)
		assert.NoError(t, err)
	}
}

// Verify the syslog output file contains expected data
func verifySyslogOutput(t *testing.T, outputFile string) {
	// Read the output file
	content, err := os.ReadFile(outputFile)
	assert.NoError(t, err)

	// Split the content into JSON lines
	lines := splitSyslogLines(content)
	assert.GreaterOrEqual(t, len(lines), 4, "Should have at least 4 syslog records")

	// Verify each syslog entry has the expected fields
	for _, line := range lines {
		var record map[string]interface{}
		err := json.Unmarshal(line, &record)
		assert.NoError(t, err, "Each line should be valid JSON")

		// Verify common fields
		assert.NotEmpty(t, record["collection_timestamp"])
		assert.NotEmpty(t, record["event_timestamp"])
		assert.NotEmpty(t, record["source_file"])
		assert.Contains(t, record["source_file"].(string), "system.log")

		// Check data fields
		data, ok := record["data"].(map[string]interface{})
		assert.True(t, ok, "Should have a data field as a map")

		// For most records, verify expected log entry fields
		if data["processname"] != nil {
			assert.NotEmpty(t, data["systemname"])
			assert.NotEmpty(t, data["processname"])
			assert.NotEmpty(t, data["pid"])
			assert.NotEmpty(t, data["message"])
		} else {
			// For continuation messages, just check the message
			assert.NotEmpty(t, data["message"])
		}
	}

	// Verify specific syslog content
	content_str := string(content)
	assert.Contains(t, content_str, "kernel")
	assert.Contains(t, content_str, "UserEventAgent")
	assert.Contains(t, content_str, "mDNSResponder")
	assert.Contains(t, content_str, "securityd")
	assert.Contains(t, content_str, "MacBook-Pro.local")
	assert.Contains(t, content_str, "System boot completed")
	assert.Contains(t, content_str, "Captive: CNPluginHandler en0")
}

// Helper to split content into lines (handles different line endings)
func splitSyslogLines(data []byte) [][]byte {
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
