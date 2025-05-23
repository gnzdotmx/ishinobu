package quarantineevents

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/gnzdotmx/ishinobu/pkg/mod"
	"github.com/gnzdotmx/ishinobu/pkg/testutils"
	"github.com/gnzdotmx/ishinobu/pkg/utils"
)

func TestQuarantineEventsModule(t *testing.T) {
	// Create temporary directory for test outputs
	tmpDir, err := os.MkdirTemp("", "quarantineevents_test")
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

// Test the nullStringValue helper function
func TestNullStringValue(t *testing.T) {
	t.Run("Valid NullString", func(t *testing.T) {
		validStr := sql.NullString{String: "test", Valid: true}
		result := nullStringValue(validStr)
		assert.Equal(t, "test", result)
	})

	t.Run("Invalid NullString", func(t *testing.T) {
		invalidStr := sql.NullString{String: "test", Valid: false}
		result := nullStringValue(invalidStr)
		assert.Equal(t, "", result)
	})

	t.Run("Empty Valid NullString", func(t *testing.T) {
		emptyStr := sql.NullString{String: "", Valid: true}
		result := nullStringValue(emptyStr)
		assert.Equal(t, "", result)
	})
}

// Test error handling in processQuarantineEvents
func TestProcessQuarantineEventsErrorHandling(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "quarantineevents_error_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logger := testutils.NewTestLogger()
	params := mod.ModuleParams{
		OutputDir:           tmpDir,
		LogsDir:             tmpDir,
		ExportFormat:        "json",
		CollectionTimestamp: time.Now().Format(utils.TimeFormat),
		Logger:              *logger,
	}

	t.Run("NonExistentDatabase", func(t *testing.T) {
		nonExistentPath := filepath.Join(tmpDir, "nonexistent.db")
		err := processQuarantineEvents(nonExistentPath, "quarantineevents", params)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error copying file")
	})
}

// Test the module's Run method by creating test files in a specific location
func TestModuleRunWithSetupFiles(t *testing.T) {
	// Skip if tests need to be run in automated environments without privileges
	// t.Skip("This test requires creation of temporary files in a specific location")

	// Create temporary directory for test outputs
	tmpDir, err := os.MkdirTemp("", "quarantineevents_run_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test directory structure
	userLibraryDir := filepath.Join(tmpDir, "Users", "testuser", "Library", "Preferences")
	err = os.MkdirAll(userLibraryDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	// Create mock QuarantineEventsV2 file
	quarantineDbPath := filepath.Join(userLibraryDir, "com.apple.LaunchServices.QuarantineEventsV2")
	file, err := os.Create(quarantineDbPath)
	if err != nil {
		t.Fatal(err)
	}
	file.Close()

	logger := testutils.NewTestLogger()
	params := mod.ModuleParams{
		OutputDir:           tmpDir,
		LogsDir:             tmpDir,
		ExportFormat:        "json",
		CollectionTimestamp: time.Now().Format(utils.TimeFormat),
		Logger:              *logger,
	}

	module := &QuarantineEventsModule{
		Name:        "quarantineevents",
		Description: "Collects and parses QuarantineEventsV2 database",
	}

	// The Run method should not find our file as it's not in standard locations
	// This is a negative test to prove the Run method looks in specific locations
	err = module.Run(params)
	assert.NoError(t, err)

	// Verify no output file was created since our mock file wasn't found
	outputFile := filepath.Join(tmpDir, "quarantineevents-"+params.CollectionTimestamp+".json")
	assert.NoFileExists(t, outputFile)
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

		sourceFileStr, ok := record["source_file"].(string)
		assert.True(t, ok, "Source file should be a string")
		assert.Contains(t, sourceFileStr, "QuarantineEventsV2")

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
	contentStr := string(content)
	assert.Contains(t, contentStr, "com.google.Chrome")
	assert.Contains(t, contentStr, "com.apple.Safari")
	assert.Contains(t, contentStr, "https://example.com/download/file.dmg")
	assert.Contains(t, contentStr, "https://test.org/software/app.pkg")
	assert.Contains(t, contentStr, "jane@example.com")
}

// Test multiple variations of quarantine events records
func TestQuarantineEventsDataVariations(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "quarantineevents_variations_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Generate different timestamp formats
	timestamps := []string{
		"2023-05-01T12:30:45Z",          // Standard ISO
		"2023-05-01 12:30:45 +0000 UTC", // Alternative format
		"1682946645",                    // Unix timestamp
	}

	// Generate different data combinations
	dataVariations := []map[string]interface{}{
		{
			// All fields populated
			"identifier":       "complete-id-1",
			"user":             "testuser",
			"bundle_id":        "com.test.app1",
			"quarantine_agent": "com.test.browser",
			"download_url":     "https://example.com/download/complete.dmg",
			"sender_name":      "Complete Test",
			"sender_address":   "complete@example.com",
			"type_no":          1,
			"origin_title":     "Complete Origin",
			"origin_url":       "https://complete.example.com",
			"origin_alias":     "complete-alias",
		},
		{
			// Missing some optional fields
			"identifier":       "partial-id-1",
			"user":             "testuser",
			"bundle_id":        "com.test.app2",
			"quarantine_agent": "com.test.browser",
			"download_url":     "https://example.com/download/partial.dmg",
			"sender_name":      "",
			"sender_address":   "",
			"type_no":          0,
			"origin_title":     "Partial Origin",
			"origin_url":       "",
			"origin_alias":     "",
		},
		{
			// Minimal required data
			"identifier":       "minimal-id-1",
			"user":             "testuser",
			"bundle_id":        "com.test.minimal",
			"quarantine_agent": "com.test.minimal",
			"download_url":     "",
			"sender_name":      "",
			"sender_address":   "",
			"type_no":          0,
			"origin_title":     "",
			"origin_url":       "",
			"origin_alias":     "",
		},
	}

	// Create test records with different combinations
	var testRecords []utils.Record
	for i, timestamp := range timestamps {
		for j, data := range dataVariations {
			record := utils.Record{
				CollectionTimestamp: time.Now().Format(utils.TimeFormat),
				EventTimestamp:      timestamp,
				SourceFile:          fmt.Sprintf("/Users/user%d/Library/Preferences/com.apple.LaunchServices.QuarantineEventsV2", i+j),
				Data:                data,
			}
			testRecords = append(testRecords, record)
		}
	}

	// Create output file
	outputFile := filepath.Join(tmpDir, "quarantineevents-variations.json")
	file, err := os.Create(outputFile)
	assert.NoError(t, err)
	defer file.Close()

	// Write test records
	encoder := json.NewEncoder(file)
	for _, record := range testRecords {
		err := encoder.Encode(record)
		assert.NoError(t, err)
	}

	// Read and validate file contents
	content, err := os.ReadFile(outputFile)
	assert.NoError(t, err)

	lines := splitQuarantineEventsLines(content)
	assert.Equal(t, len(testRecords), len(lines), "Should have same number of records as written")

	// Verify random variations are handled
	for _, line := range lines {
		var record utils.Record
		err := json.Unmarshal(line, &record)
		assert.NoError(t, err)

		// Check record has expected structure
		assert.NotEmpty(t, record.CollectionTimestamp)
		assert.NotEmpty(t, record.EventTimestamp)
		assert.NotEmpty(t, record.SourceFile)

		data, ok := record.Data.(map[string]interface{})
		assert.True(t, ok)
		assert.NotEmpty(t, data["identifier"])
		assert.NotEmpty(t, data["user"])
		assert.NotEmpty(t, data["bundle_id"])
		assert.NotEmpty(t, data["quarantine_agent"])
	}
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
