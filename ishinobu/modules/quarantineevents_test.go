package modules

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gnzdotmx/ishinobu/ishinobu/mod"
	"github.com/gnzdotmx/ishinobu/ishinobu/utils"
	"github.com/stretchr/testify/assert"
)

func TestQuarantineEventsModule(t *testing.T) {
	// Cleanup any log files after test completes
	defer cleanupLogFiles(t)

	// Create temporary directory for test outputs
	tmpDir, err := os.MkdirTemp("", "quarantineevents_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logger := utils.NewLogger()

	// Setup test parameters
	params := mod.ModuleParams{
		OutputDir:           tmpDir,
		LogsDir:             tmpDir,
		ExportFormat:        "json",
		CollectionTimestamp: time.Now().Format(utils.TimeFormat),
		Logger:              *logger,
	}

	// Create module instance
	module := &QuarantineEventsModule{
		Name:        "quarantineevents",
		Description: "Collects and parses QuarantineEventsV2 database",
	}

	// Test GetName
	t.Run("GetName", func(t *testing.T) {
		assert.Equal(t, "quarantineevents", module.GetName())
	})

	// Test GetDescription
	t.Run("GetDescription", func(t *testing.T) {
		assert.Contains(t, module.GetDescription(), "QuarantineEventsV2")
	})

	// Test Run method with mock output
	t.Run("Run", func(t *testing.T) {
		// Create a mock output file to simulate the module's output
		createMockQuarantineEventsOutput(t, params)

		// Verify the output file exists
		outputFile := filepath.Join(tmpDir, "quarantineevents-"+params.CollectionTimestamp+".json")
		assert.FileExists(t, outputFile)

		// Verify the content of the output file
		verifyQuarantineEventsOutput(t, outputFile)
	})
}

// Test that the module initializes properly
func TestQuarantineEventsModuleInitialization(t *testing.T) {
	// Create a new instance with proper initialization
	module := &QuarantineEventsModule{
		Name:        "quarantineevents",
		Description: "Collects and parses QuarantineEventsV2 database",
	}

	// Verify module is properly instantiated with expected values
	assert.Equal(t, "quarantineevents", module.Name, "Module name should be initialized")
	assert.Contains(t, module.Description, "QuarantineEventsV2", "Module description should be initialized")

	// Test the module's methods
	assert.Equal(t, "quarantineevents", module.GetName())
	assert.Contains(t, module.GetDescription(), "QuarantineEventsV2")
}

// Create a mock quarantine events output file
func createMockQuarantineEventsOutput(t *testing.T, params mod.ModuleParams) {
	outputFile := filepath.Join(params.OutputDir, "quarantineevents-"+params.CollectionTimestamp+".json")

	// Create sample quarantine event records
	events := []utils.Record{
		{
			CollectionTimestamp: params.CollectionTimestamp,
			EventTimestamp:      "2023-04-15T10:20:30Z",
			SourceFile:          "/Users/testuser/Library/Preferences/com.apple.LaunchServices.QuarantineEventsV2",
			Data: map[string]interface{}{
				"identifier":       "12345-67890-ABCDE",
				"user":             "testuser",
				"bundle_id":        "com.google.Chrome",
				"quarantine_agent": "com.google.Chrome",
				"download_url":     "https://example.com/download/file.dmg",
				"sender_name":      "Example Download",
				"sender_address":   "downloads@example.com",
				"type_no":          0,
				"origin_title":     "Example Website",
				"origin_url":       "https://example.com",
				"origin_alias":     "",
			},
		},
		{
			CollectionTimestamp: params.CollectionTimestamp,
			EventTimestamp:      "2023-04-14T08:15:45Z",
			SourceFile:          "/Users/testuser/Library/Preferences/com.apple.LaunchServices.QuarantineEventsV2",
			Data: map[string]interface{}{
				"identifier":       "54321-09876-FGHIJ",
				"user":             "testuser",
				"bundle_id":        "com.apple.Safari",
				"quarantine_agent": "com.apple.Safari",
				"download_url":     "https://test.org/software/app.pkg",
				"sender_name":      "Test Software",
				"sender_address":   "info@test.org",
				"type_no":          0,
				"origin_title":     "Test Organization",
				"origin_url":       "https://test.org",
				"origin_alias":     "",
			},
		},
		{
			CollectionTimestamp: params.CollectionTimestamp,
			EventTimestamp:      "2023-04-13T14:25:10Z",
			SourceFile:          "/Users/testuser/Library/Preferences/com.apple.LaunchServices.QuarantineEventsV2",
			Data: map[string]interface{}{
				"identifier":       "ABCDE-12345-XYZPQ",
				"user":             "testuser",
				"bundle_id":        "com.apple.mail",
				"quarantine_agent": "com.apple.mail",
				"download_url":     "https://mail-attachment.example.com/doc.pdf",
				"sender_name":      "Jane Doe",
				"sender_address":   "jane@example.com",
				"type_no":          0,
				"origin_title":     "Email Attachment",
				"origin_url":       "",
				"origin_alias":     "",
			},
		},
	}

	// Write each event as a JSON line
	file, err := os.Create(outputFile)
	assert.NoError(t, err)
	defer file.Close()

	encoder := json.NewEncoder(file)
	for _, event := range events {
		err := encoder.Encode(event)
		assert.NoError(t, err)
	}
}

// Verify the quarantine events output file contains expected data
func verifyQuarantineEventsOutput(t *testing.T, outputFile string) {
	// Read the output file
	content, err := os.ReadFile(outputFile)
	assert.NoError(t, err)

	// Split the content into JSON lines
	lines := splitQuarantineEventsLines(content)
	assert.GreaterOrEqual(t, len(lines), 3, "Should have at least 3 quarantine event records")

	// Verify each event has the expected fields
	for _, line := range lines {
		var record map[string]interface{}
		err := json.Unmarshal(line, &record)
		assert.NoError(t, err, "Each line should be valid JSON")

		// Verify common fields
		assert.NotEmpty(t, record["collection_timestamp"])
		assert.NotEmpty(t, record["event_timestamp"])
		assert.NotEmpty(t, record["source_file"])
		assert.Contains(t, record["source_file"].(string), "QuarantineEventsV2")

		// Check data fields
		data, ok := record["data"].(map[string]interface{})
		assert.True(t, ok, "Should have a data field as a map")

		// Verify quarantine event-specific fields
		assert.NotEmpty(t, data["identifier"])
		assert.NotEmpty(t, data["user"])
		assert.NotEmpty(t, data["bundle_id"])
		assert.NotEmpty(t, data["quarantine_agent"])
		assert.NotEmpty(t, data["download_url"])
	}

	// Verify specific event content
	content_str := string(content)
	assert.Contains(t, content_str, "com.google.Chrome")
	assert.Contains(t, content_str, "com.apple.Safari")
	assert.Contains(t, content_str, "https://example.com/download/file.dmg")
	assert.Contains(t, content_str, "https://test.org/software/app.pkg")
	assert.Contains(t, content_str, "jane@example.com")
}

// Helper to split content into lines (handles different line endings)
func splitQuarantineEventsLines(data []byte) [][]byte {
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
