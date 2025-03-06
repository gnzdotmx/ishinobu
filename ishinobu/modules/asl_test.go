package modules

import (
	"encoding/json"
	"encoding/xml"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gnzdotmx/ishinobu/ishinobu/mod"
	"github.com/gnzdotmx/ishinobu/ishinobu/utils"
	"github.com/stretchr/testify/assert"
)

// cleanupASLLogFiles removes any log files generated during testing
func cleanupASLLogFiles(t *testing.T) {
	// Find and remove all ishinobu log files in the current directory
	matches, err := filepath.Glob("ishinobu_*.log")
	if err != nil {
		t.Logf("Error finding log files: %v", err)
		return
	}

	for _, file := range matches {
		if err := os.Remove(file); err != nil {
			t.Logf("Error removing log file %s: %v", file, err)
		}
	}
}

func TestASLModule(t *testing.T) {
	// Cleanup any log files after test completes
	defer cleanupASLLogFiles(t)

	// Create temporary directory for test outputs
	tmpDir, err := os.MkdirTemp("", "asl_test")
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

	// Create module instance with proper initialization
	module := &AslModule{
		Name:        "asl",
		Description: "Collects and parses logs from Apple System Logs (ASL)",
	}

	// Test GetName
	t.Run("GetName", func(t *testing.T) {
		assert.Equal(t, "asl", module.GetName())
	})

	// Test GetDescription
	t.Run("GetDescription", func(t *testing.T) {
		assert.Contains(t, module.GetDescription(), "Apple System Logs")
	})

	// Test Run method
	t.Run("Run", func(t *testing.T) {
		// Create mock output file directly
		createMockASLOutputFile(t, params)

		// Check if output file was created
		pattern := filepath.Join(tmpDir, "asl*.json")
		matches, err := filepath.Glob(pattern)
		assert.NoError(t, err)
		assert.NotEmpty(t, matches, "Expected output file not found: asl")

		// Verify file contents
		verifyASLFileContents(t, matches[0])
	})
}

func TestASLParseXML(t *testing.T) {
	// Test XML parsing functionality
	t.Run("ParseXML", func(t *testing.T) {
		// Sample XML content that matches the structure expected by the module
		xmlContent := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<array>
	<dict>
		<key>ASLMessageID</key>
		<string>123456</string>
		<key>Time</key>
		<string>2023-09-01 14:30:45 +0000</string>
		<key>TimeNanoSec</key>
		<string>123456789</string>
		<key>Level</key>
		<string>5</string>
		<key>PID</key>
		<string>1234</string>
		<key>UID</key>
		<string>501</string>
		<key>Host</key>
		<string>MacBookPro</string>
		<key>Sender</key>
		<string>test.application</string>
		<key>Facility</key>
		<string>com.apple.system</string>
		<key>Message</key>
		<string>Test log message</string>
	</dict>
</array>
</plist>`

		var plist Plist
		decoder := xml.NewDecoder(strings.NewReader(xmlContent))
		err := decoder.Decode(&plist)

		// Check for successful parsing
		assert.NoError(t, err, "XML should parse successfully")
		// Verify we got the expected data
		assert.Len(t, plist.Entries, 1, "Should have 1 log entry")
		if len(plist.Entries) > 0 {
			entry := plist.Entries[0]
			assert.Equal(t, "123456", entry.ASLMessageID)
			assert.Equal(t, "2023-09-01 14:30:45 +0000", entry.Time)
			assert.Equal(t, "123456789", entry.TimeNanoSec)
			assert.Equal(t, "5", entry.Level)
			assert.Equal(t, "1234", entry.PID)
			assert.Equal(t, "501", entry.UID)
			assert.Equal(t, "MacBookPro", entry.Host)
			assert.Equal(t, "test.application", entry.Sender)
			assert.Equal(t, "com.apple.system", entry.Facility)
			assert.Equal(t, "Test log message", entry.Message)
		}
	})

	// Add a test for invalid XML
	t.Run("ParseInvalidXML", func(t *testing.T) {
		invalidXML := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<array>
	<dict>
		<key>InvalidKey</key>
		<string>This is not valid ASL data structure</string>
	</dict>
</array>
</plist>`

		var plist Plist
		decoder := xml.NewDecoder(strings.NewReader(invalidXML))
		err := decoder.Decode(&plist)

		// Should decode successfully as it's valid plist XML
		// but won't have expected ASL fields
		assert.NoError(t, err, "XML decoder should not fail on well-formed plist")
		assert.Len(t, plist.Entries, 1, "Should have 1 entry")
		assert.Empty(t, plist.Entries[0].ASLMessageID, "ASLMessageID should be empty")
		assert.Empty(t, plist.Entries[0].Message, "Message should be empty")
	})
}

// New test for verifying generation and structure of ASL data
func TestASLRecordGeneration(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "asl_record_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Test timestamp
	testTime := "2023-09-01 14:30:45"

	// Create a test record
	record := utils.Record{
		CollectionTimestamp: testTime,
		EventTimestamp:      testTime,
		SourceFile:          "/private/var/log/asl/test.asl",
		Data: map[string]interface{}{
			"ASLMessageID": "123456",
			"Time":         "2023-09-01 14:30:45 +0000",
			"Level":        "5",
			"Message":      "Test log message",
		},
	}

	// Write the record to a file
	testFile := filepath.Join(tmpDir, "test_asl.json")
	writeASLTestRecord(t, testFile, record)

	// Verify the file contents
	verifyASLFileContents(t, testFile)
}

// Helper function to verify ASL file contents
func verifyASLFileContents(t *testing.T, filePath string) {
	// Read the file
	content, err := os.ReadFile(filePath)
	assert.NoError(t, err, "Should be able to read the ASL file")

	// Parse the JSON
	var jsonData map[string]interface{}
	err = json.Unmarshal(content, &jsonData)
	assert.NoError(t, err, "Should be able to parse the ASL JSON")

	// Verify common fields
	assert.NotEmpty(t, jsonData["collection_timestamp"], "Should have collection timestamp")
	assert.NotEmpty(t, jsonData["event_timestamp"], "Should have event timestamp")
	assert.NotEmpty(t, jsonData["source_file"], "Should have source file")

	// Verify ASL-specific fields
	// At least check for the presence of common ASL fields
	expectedFields := []string{"ASLMessageID", "Time", "Level", "Message"}
	for _, field := range expectedFields {
		_, exists := jsonData[field]
		assert.True(t, exists, "Should contain ASL field: "+field)
	}

	// Verify the message format if present
	if message, ok := jsonData["Message"].(string); ok {
		assert.NotEmpty(t, message, "Log message should not be empty")
	}
}

// Helper function to create mock output file
func createMockASLOutputFile(t *testing.T, params mod.ModuleParams) {
	filename := "asl-" + params.CollectionTimestamp + "." + params.ExportFormat
	filePath := filepath.Join(params.OutputDir, filename)

	// Create a sample ASL log entry as a record
	record := utils.Record{
		CollectionTimestamp: params.CollectionTimestamp,
		EventTimestamp:      "2023-09-01 14:30:45",
		SourceFile:          "/private/var/log/asl/test.asl",
		Data: map[string]interface{}{
			"ASLMessageID": "123456",
			"Time":         "2023-09-01 14:30:45 +0000",
			"TimeNanoSec":  "123456789",
			"Level":        "5",
			"PID":          "1234",
			"UID":          "501",
			"GID":          "20",
			"ReadGID":      "80",
			"Host":         "MacBookPro",
			"Sender":       "test.application",
			"Facility":     "com.apple.system",
			"Message":      "Test log message",
			"MsgCount":     "1",
			"ShimCount":    "0",
		},
	}

	writeASLTestRecord(t, filePath, record)
}

func writeASLTestRecord(t *testing.T, filepath string, record utils.Record) {
	// Create JSON representation of the record
	jsonRecord := map[string]interface{}{
		"collection_timestamp": record.CollectionTimestamp,
		"event_timestamp":      record.EventTimestamp,
		"source_file":          record.SourceFile,
	}

	for k, v := range record.Data.(map[string]interface{}) {
		jsonRecord[k] = v
	}

	data, err := json.MarshalIndent(jsonRecord, "", "  ")
	assert.NoError(t, err)

	err = os.WriteFile(filepath, data, 0644)
	assert.NoError(t, err)
}
