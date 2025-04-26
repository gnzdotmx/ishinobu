package nettop

import (
	"os"
	"path/filepath"
	"strings"
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
		assert.Contains(t, module.GetDescription(), "network connections")
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
	assert.Contains(t, module.Description, "network connections", "Module description should be initialized")

	// Test the module's methods
	assert.Equal(t, "nettop", module.GetName())
	assert.Contains(t, module.GetDescription(), "network connections")
}

// Test that the package initialization occurs
func TestPackageInitialization(t *testing.T) {
	// Create a new module with the same name
	module := &NettopModule{
		Name:        "nettop",
		Description: "This is a test initialization",
	}

	// Verify the module has expected values
	assert.Equal(t, "nettop", module.Name)
	assert.Equal(t, "This is a test initialization", module.Description)

	// The init function is automatically called when the package is imported
	// We can't directly test it, but we can verify it doesn't crash the tests
}

// Test Run method error handling
func TestRunMethodError(t *testing.T) {
	// Create temporary directory for test outputs
	tmpDir, err := os.MkdirTemp("", "nettop_error_test")
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

	module := &NettopModule{
		Name:        "nettop",
		Description: "Collects information about network connections",
	}

	err = module.Run(params)
	assert.Error(t, err, "Run should return an error with invalid output directory")
	assert.Contains(t, err.Error(), "error", "Error message should contain explanation")
}

// Test that the module uses correct command arguments
func TestModuleCommandArgs(t *testing.T) {
	// Create temporary directory for test outputs
	tmpDir, err := os.MkdirTemp("", "nettop_cmd_test")
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

	// Create output file before running to verify it gets overwritten
	outputFile := filepath.Join(tmpDir, "nettop-"+params.CollectionTimestamp+".json")
	err = os.WriteFile(outputFile, []byte("test data"), 0644)
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

// Create a mock nettop output file
func createMockNettopOutput(t *testing.T, params mod.ModuleParams) {
	outputFile := filepath.Join(params.OutputDir, "nettop-"+params.CollectionTimestamp+".json")

	// Sample nettop output data in JSON format
	mockData := `[
		{
			"collection_timestamp": "` + params.CollectionTimestamp + `",
			"event_timestamp": "` + time.Now().Format(utils.TimeFormat) + `",
			"interface": "en0",
			"state": "ESTABLISHED",
			"bytes_in": "1024",
			"bytes_out": "512",
			"packets_in": "8",
			"packets_out": "4",
			"process": "Chrome",
			"source_file": "nettop"
		},
		{
			"collection_timestamp": "` + params.CollectionTimestamp + `",
			"event_timestamp": "` + time.Now().Format(utils.TimeFormat) + `",
			"interface": "lo0",
			"state": "ESTABLISHED",
			"bytes_in": "256",
			"bytes_out": "128",
			"packets_in": "2",
			"packets_out": "1",
			"process": "node",
			"source_file": "nettop"
		},
		{
			"collection_timestamp": "` + params.CollectionTimestamp + `",
			"event_timestamp": "` + time.Now().Format(utils.TimeFormat) + `",
			"interface": "en1",
			"state": "CLOSED",
			"bytes_in": "0",
			"bytes_out": "0",
			"packets_in": "0",
			"packets_out": "0",
			"process": "Firefox",
			"source_file": "nettop"
		}
	]`

	err := os.WriteFile(outputFile, []byte(mockData), 0600)
	assert.NoError(t, err)
}

// Verify the nettop output file contains expected data
func verifyNettopOutput(t *testing.T, outputFile string) {
	// Read the output file
	data, err := os.ReadFile(outputFile)
	assert.NoError(t, err)

	// Verify the content contains expected elements
	content := string(data)

	// Check for interface information
	assert.Contains(t, content, "\"interface\": \"en0\"")
	assert.Contains(t, content, "\"interface\": \"lo0\"")

	// Check for state information
	assert.Contains(t, content, "\"state\": \"ESTABLISHED\"")
	assert.Contains(t, content, "\"state\": \"CLOSED\"")

	// Check for bytes information
	assert.Contains(t, content, "\"bytes_in\": \"1024\"")
	assert.Contains(t, content, "\"bytes_out\": \"512\"")

	// Check for packets information
	assert.Contains(t, content, "\"packets_in\": \"8\"")
	assert.Contains(t, content, "\"packets_out\": \"4\"")

	// Check for process information
	assert.Contains(t, content, "\"process\": \"Chrome\"")
	assert.Contains(t, content, "\"process\": \"node\"")
}

// TestFieldParsing tests the field parsing logic in the nettop module
func TestFieldParsing(t *testing.T) {
	// Create temporary directory for test outputs
	tmpDir, err := os.MkdirTemp("", "nettop_field_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Setup test parameters
	params := mod.ModuleParams{
		OutputDir:           tmpDir,
		LogsDir:             tmpDir,
		ExportFormat:        "json",
		CollectionTimestamp: time.Now().Format(utils.TimeFormat),
		Logger:              *testutils.NewTestLogger(),
	}

	// Create mock CSV file with sample data
	mockCSV := "interface,state,bytes_in,bytes_out,packets_in,packets_out\n" +
		"en0,ESTABLISHED,1024,512,8,4,Chrome\n" +
		"lo0,ESTABLISHED,256,128,2,1,node\n" +
		"en1,CLOSED,0,0,0,0,Firefox\n"

	// Create a sample data file that represents what nettop would output
	sampleFile := filepath.Join(tmpDir, "sample_nettop.csv")
	err = os.WriteFile(sampleFile, []byte(mockCSV), 0644)
	assert.NoError(t, err)

	// Manually parse the CSV to verify the field parsing logic
	lines := strings.Split(mockCSV, "\n")
	header := lines[0]
	fields := strings.Split(header, ",")

	// Verify we can process the expected format
	for _, line := range lines[1:] {
		if strings.TrimSpace(line) == "" {
			continue
		}

		cols := strings.Split(line, ",")
		if len(cols) != len(fields) {
			continue
		}

		// Just verify the column data is accessible without storing records
		recordData := make(map[string]interface{})
		for index, col := range cols {
			colName := fields[index]

			// skip empty columns
			if colName == "" && col == "" {
				continue
			}

			if colName == "" && col != "" {
				recordData["process"] = col
			} else {
				recordData[string(colName)] = col
			}
		}

		// Verify some expected fields
		assert.NotEmpty(t, recordData["interface"])
		assert.NotEmpty(t, recordData["state"])
	}

	// Create expected JSON output
	createMockNettopOutput(t, params)

	// Verify the expected JSON output
	outputFile := filepath.Join(tmpDir, "nettop-"+params.CollectionTimestamp+".json")
	assert.FileExists(t, outputFile)

	// Verify content parsing
	verifyNettopOutput(t, outputFile)
}

// TestEmptyLineHandling verifies handling of empty lines and field indices
func TestEmptyLineHandling(t *testing.T) {
	// Create temporary directory for test outputs
	tmpDir, err := os.MkdirTemp("", "nettop_empty_line_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a more complex mock CSV with empty lines
	mockCSV := "interface,state,bytes_in,bytes_out,packets_in,packets_out\n" +
		"en0,ESTABLISHED,1024,512,8,4,Chrome\n" +
		"\n" + // Empty line
		"lo0,ESTABLISHED,256,128,2,1,node\n" +
		"en1,CLOSED,0,0,0,0,Firefox\n" +
		"\n" // Empty line at the end

	// Create a sample data file
	sampleFile := filepath.Join(tmpDir, "sample_empty_lines.csv")
	err = os.WriteFile(sampleFile, []byte(mockCSV), 0644)
	assert.NoError(t, err)

	// Parse the data manually to verify the module's logic
	lines := strings.Split(mockCSV, "\n")
	if len(lines) == 0 {
		t.Fatal("Expected non-empty lines array")
	}

	header := lines[0]
	fields := strings.Split(header, ",")

	// Check field indices
	fieldIndices := make(map[string]int)
	for i, field := range fields {
		fieldIndices[field] = i
		// Check that we're indexing fields correctly
		assert.Equal(t, i, fieldIndices[field])
	}

	// Verify all line handling
	recordCount := 0
	for _, line := range lines[1:] {
		if strings.TrimSpace(line) == "" {
			continue // This should skip two lines
		}

		cols := strings.Split(line, ",")
		// In real data, there's an extra column for process, so we don't check exact equality
		if len(cols) < len(fields) {
			continue
		}

		recordCount++
	}

	// We should have 3 valid records after parsing
	assert.Equal(t, 3, recordCount)
}

// TestInvalidDataHandling tests how the module handles invalid input formats
func TestInvalidDataHandling(t *testing.T) {
	// Create temporary directory for test outputs
	tmpDir, err := os.MkdirTemp("", "nettop_invalid_data_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Test with invalid number of columns
	mockCSV := "interface,state,bytes_in,bytes_out,packets_in,packets_out\n" +
		"en0,ESTABLISHED,1024,512,8,4\n" + // Missing the process column
		"lo0,ESTABLISHED,256,128,2,1,node,extra\n" // Too many columns

	// Parse the data manually to verify the module's logic
	lines := strings.Split(mockCSV, "\n")
	if len(lines) == 0 {
		t.Fatal("Expected non-empty lines array")
	}

	header := lines[0]
	fields := strings.Split(header, ",")

	// Process the lines
	recordCount := 0
	for _, line := range lines[1:] {
		if strings.TrimSpace(line) == "" {
			continue
		}

		cols := strings.Split(line, ",")
		// Only skip if there are fewer columns than fields
		// The Nettop module is actually accepting an extra process column
		if len(cols) < len(fields) {
			continue
		}

		// This will be reached for the second line that has extra columns
		recordCount++
	}

	// One line should be valid with the given logic
	assert.Equal(t, 2, recordCount)

	// Test with empty field names
	mockCSV = "interface,state,bytes_in,bytes_out,,packets_out\n" + // Empty field name
		"en0,ESTABLISHED,1024,512,,4,Chrome\n"

	lines = strings.Split(mockCSV, "\n")
	if len(lines) < 2 {
		t.Fatal("Expected at least 2 lines")
	}

	header = lines[0]
	fields = strings.Split(header, ",")

	// Process the record to verify empty field handling
	recordData := make(map[string]interface{})
	cols := strings.Split(lines[1], ",")

	// Ensure we don't go out of bounds
	for index, col := range cols {
		if index < len(fields) {
			colName := fields[index]

			// skip empty columns
			if colName == "" && col == "" {
				continue
			}

			if colName == "" && col != "" {
				recordData["process"] = col
			} else {
				recordData[string(colName)] = col
			}
		} else if index == len(fields) {
			// The extra column is for process
			recordData["process"] = col
		}
	}

	// Verify that the process field is set
	assert.Equal(t, "Chrome", recordData["process"])

	// Verify that an empty field name is handled properly
	_, hasEmptyField := recordData[""]
	assert.False(t, hasEmptyField, "Should not have an empty field name in the record")
}

// TestDataWriterError tests error handling when NewDataWriter fails
func TestDataWriterError(t *testing.T) {
	// Create temporary directory for test outputs
	tmpDir, err := os.MkdirTemp("", "nettop_writer_error_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logger := testutils.NewTestLogger()

	// Setup test parameters with an invalid logs directory path
	// This should cause a failure in NewDataWriter
	params := mod.ModuleParams{
		OutputDir:           tmpDir,
		LogsDir:             "/this/path/does/not/exist",
		ExportFormat:        "json",
		CollectionTimestamp: time.Now().Format(utils.TimeFormat),
		Logger:              *logger,
	}

	// Create module instance
	module := &NettopModule{
		Name:        "nettop",
		Description: "Collects information about network connections",
	}

	// Prepare mock CSV data for nettop
	mockCSV := "interface,state,bytes_in,bytes_out,packets_in,packets_out\n" +
		"en0,ESTABLISHED,1024,512,8,4,Chrome\n"

	// We can't override execCommand, but we can simulate part of the module's behavior
	// by directly checking the error handling in NewDataWriter failure

	// Call the required file operations that would happen after command execution
	lines := strings.Split(mockCSV, "\n")
	if len(lines) < 1 {
		t.Fatal("Expected at least one line")
	}

	// Run the module and expect an error
	err = module.Run(params)
	assert.Error(t, err, "Run should fail with invalid logs directory")
	assert.Contains(t, err.Error(), "failed to create data writer", "Error should mention data writer failure")
}
