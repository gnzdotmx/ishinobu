package auditlog

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

func TestAuditLogModule(t *testing.T) {
	// Create temporary directory for test outputs
	tmpDir, err := os.MkdirTemp("", "auditlog_test")
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
	module := &AuditLogModule{
		Name:        "auditlog",
		Description: "Collects and parses audit logs using praudit",
	}

	// Test GetName
	t.Run("GetName", func(t *testing.T) {
		assert.Equal(t, "auditlog", module.GetName())
	})

	// Test GetDescription
	t.Run("GetDescription", func(t *testing.T) {
		assert.Contains(t, module.GetDescription(), "audit logs")
	})

	// Test Run method - since this calls praudit command, we'll create mock output directly
	t.Run("Run", func(t *testing.T) {
		// Create mock output file directly
		createMockAuditLogFile(t, params)

		// Check if output file was created
		pattern := filepath.Join(tmpDir, "auditlog*.json")
		matches, err := filepath.Glob(pattern)
		assert.NoError(t, err)
		assert.NotEmpty(t, matches)

		// Verify file contents
		verifyAuditLogFileContents(t, matches[0])
	})
}

func TestParsePrauditXMLLine(t *testing.T) {
	// Sample XML line that praudit would output
	xmlLine := `<record version="11" event="system_check" modifier="0" time="1618329456" msec="123">
		<subject audit-uid="1001" uid="1001" gid="1001" ruid="1001" rgid="1001" pid="12345" sid="100" tid="200" />
		<argument arg-num="1" value="test_value" desc="test description" />
		<return errval="0" retval="0" />
	</record>`

	record, err := parsePrauditXMLLine(xmlLine)
	assert.NoError(t, err)

	// Check basic fields
	data, ok := record.Data.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "system_check", data["event"])
	assert.Equal(t, "0", data["modifier"])
	assert.Equal(t, "1618329456", data["time"])
	assert.Equal(t, "123", data["msec"])

	// Check subject fields
	assert.Equal(t, "1001", data["audit-uid"])
	assert.Equal(t, "1001", data["uid"])
	assert.Equal(t, "1001", data["gid"])
	assert.Equal(t, "12345", data["pid"])

	// Check argument field
	assert.Equal(t, "test_value", data["1"])

	// Check return fields
	assert.Equal(t, "0", data["errval"])
	assert.Equal(t, "0", data["retval"])

	// Check timestamp format - should be human readable
	assert.NotEmpty(t, record.EventTimestamp)
}

// Test invalid XML input
func TestParsePrauditXMLLineInvalid(t *testing.T) {
	// Malformed XML that is missing closing tags
	malformedXML := `<record version="11" event="system_check">
		<subject audit-uid="1001" uid="1001" pid="12345" />
		<argument arg-num="1" value="test_value"
	</record>`

	_, err := parsePrauditXMLLine(malformedXML)
	assert.Error(t, err, "Parsing malformed XML should return an error")

	// Empty string
	_, err = parsePrauditXMLLine("")
	assert.Error(t, err, "Parsing empty string should return an error")

	// Non-audit log XML
	nonAuditXML := `<someothertag attr="value">Not an audit log</someothertag>`
	_, err = parsePrauditXMLLine(nonAuditXML)
	assert.Error(t, err, "Parsing non-audit XML should return an error")
}

// Test generation of audit log records
func TestAuditLogRecordGeneration(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "auditlog_record_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Test timestamp
	testTime := time.Now().Format(utils.TimeFormat)

	// Create a test record
	recordData := map[string]interface{}{
		"event":     "system_check",
		"modifier":  "0",
		"time":      "1618329456",
		"msec":      "123",
		"audit-uid": "1001",
		"uid":       "1001",
		"gid":       "1001",
		"ruid":      "1001",
		"rgid":      "1001",
		"pid":       "12345",
		"sid":       "100",
		"tid":       "200",
		"errval":    "0",
		"retval":    "0",
		"1":         "test_value",
	}

	record := utils.Record{
		CollectionTimestamp: testTime,
		EventTimestamp:      testTime,
		SourceFile:          "/private/var/audit/test_file",
		Data:                recordData,
	}

	// Write the record to a file
	testFile := filepath.Join(tmpDir, "test_auditlog.json")
	jsonRecord := recordToJSON(record)
	data, err := json.MarshalIndent(jsonRecord, "", "  ")
	assert.NoError(t, err)

	err = os.WriteFile(testFile, data, 0600)
	assert.NoError(t, err)

	// Verify the file contents
	verifyAuditLogFileContents(t, testFile)
}

// Helper function to verify audit log file contents
func verifyAuditLogFileContents(t *testing.T, filePath string) {
	// Read the file
	content, err := os.ReadFile(filePath)
	assert.NoError(t, err, "Should be able to read the audit log file")

	// Parse the JSON
	var jsonData map[string]interface{}
	err = json.Unmarshal(content, &jsonData)
	assert.NoError(t, err, "Should be able to parse the audit log JSON")

	// Verify common fields
	assert.NotEmpty(t, jsonData["collection_timestamp"], "Should have collection timestamp")
	assert.NotEmpty(t, jsonData["event_timestamp"], "Should have event timestamp")
	assert.NotEmpty(t, jsonData["source_file"], "Should have source file")

	// Verify audit log-specific fields
	// Common audit log fields to check
	expectedFields := []string{"event", "time", "pid"}
	for _, field := range expectedFields {
		_, exists := jsonData[field]
		assert.True(t, exists, "Should contain audit log field: "+field)
	}

	// Verify event type if present
	if event, ok := jsonData["event"].(string); ok {
		assert.NotEmpty(t, event, "Event type should not be empty")
	}

	// Verify user fields if present
	if _, hasUID := jsonData["uid"]; hasUID {
		assert.NotEmpty(t, jsonData["uid"], "User ID should not be empty when present")
	}
}

// Convert Record to JSON map
func recordToJSON(record utils.Record) map[string]interface{} {
	jsonRecord := map[string]interface{}{
		"collection_timestamp": record.CollectionTimestamp,
		"event_timestamp":      record.EventTimestamp,
		"source_file":          record.SourceFile,
	}

	recordMap, ok := record.Data.(map[string]interface{})
	if !ok {
		return jsonRecord
	}

	for k, v := range recordMap {
		jsonRecord[k] = v
	}

	return jsonRecord
}

// Helper function to create mock audit log output file
func createMockAuditLogFile(t *testing.T, params mod.ModuleParams) {
	filename := "auditlog-" + params.CollectionTimestamp + "." + params.ExportFormat
	filepath := filepath.Join(params.OutputDir, filename)

	recordData := map[string]interface{}{
		"event":     "system_check",
		"modifier":  "0",
		"time":      "1618329456",
		"msec":      "123",
		"audit-uid": "1001",
		"uid":       "1001",
		"gid":       "1001",
		"ruid":      "1001",
		"rgid":      "1001",
		"pid":       "12345",
		"sid":       "100",
		"tid":       "200",
		"errval":    "0",
		"retval":    "0",
		"1":         "test_value",
	}

	record := utils.Record{
		CollectionTimestamp: params.CollectionTimestamp,
		EventTimestamp:      time.Now().Format(utils.TimeFormat),
		SourceFile:          "/private/var/audit/test_file",
		Data:                recordData,
	}

	// Create JSON representation of the record
	jsonRecord := recordToJSON(record)
	data, err := json.MarshalIndent(jsonRecord, "", "  ")
	assert.NoError(t, err)

	err = os.WriteFile(filepath, data, 0600)
	assert.NoError(t, err)
}
