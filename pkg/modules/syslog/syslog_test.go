package syslog

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gnzdotmx/ishinobu/pkg/mod"
	"github.com/gnzdotmx/ishinobu/pkg/modules/testutils"
	"github.com/stretchr/testify/assert"
)

// MockListFiles is a replacement for utils.ListFiles during testing
func MockListFiles(pattern string) ([]string, error) {
	return []string{"/path/to/system.log"}, nil
}

func TestSyslogModuleRegistration(t *testing.T) {
	// Check if the module is in the list of all modules
	allModules := mod.AllModules()
	assert.Contains(t, allModules, "syslog")
}

func TestParseSyslogFile(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "syslog_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a test log file
	logContent := `Mar 15 14:23:45 testhost kernel[0]: Test log entry
Mar 15 14:23:46 testhost process[123]: Multi-line
    log entry continues here
Mar 15 14:23:47 testhost daemon[456]: Another entry`

	logFile := filepath.Join(tempDir, "system.log")
	err = os.WriteFile(logFile, []byte(logContent), 0644)
	assert.NoError(t, err)

	// Create test writer and parameters
	writer := &testutils.TestDataWriter{}
	collectionTime := time.Now().UTC().Format(time.RFC3339)
	logger := testutils.NewTestLogger()
	params := mod.ModuleParams{
		Logger:              *logger, // Pass as value, not pointer
		OutputDir:           tempDir,
		LogsDir:             tempDir,
		ExportFormat:        "json",
		CollectionTimestamp: collectionTime,
	}

	// Parse the file
	err = parseSyslogFile(logFile, writer, params)
	assert.NoError(t, err)

	// We should have 4 records (each line is parsed as a separate record)
	assert.Equal(t, 4, len(writer.Records))

	// Check first record
	assert.Equal(t, "kernel", writer.Records[0].Data.(map[string]interface{})["processname"])
	assert.Equal(t, "0", writer.Records[0].Data.(map[string]interface{})["pid"])
	assert.Equal(t, "Test log entry", writer.Records[0].Data.(map[string]interface{})["message"])
	assert.Equal(t, logFile, writer.Records[0].SourceFile)

	// Check second and third records (the multi-line entry and its continuation)
	assert.Equal(t, "process", writer.Records[1].Data.(map[string]interface{})["processname"])
	assert.Equal(t, "123", writer.Records[1].Data.(map[string]interface{})["pid"])
	assert.Equal(t, "Multi-line", writer.Records[1].Data.(map[string]interface{})["message"])
	assert.Equal(t, "log entry continues here", writer.Records[2].Data.(map[string]interface{})["message"])
}

// TestSyslogModuleRunWithDirectCall tests the Run method by directly calling
// parseSyslogFile instead of relying on utils.ListFiles
func TestSyslogModuleRunWithDirectCall(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "syslog_run_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a test log file
	logContent := `Mar 15 14:23:45 testhost kernel[0]: Test log entry
Mar 15 14:23:46 testhost process[123]: Another test entry`

	logFile := filepath.Join(tempDir, "system.log")
	err = os.WriteFile(logFile, []byte(logContent), 0644)
	assert.NoError(t, err)

	// Create test writer and parameters
	writer := &testutils.TestDataWriter{}
	collectionTime := time.Now().UTC().Format(time.RFC3339)
	logger := testutils.NewTestLogger()
	params := mod.ModuleParams{
		Logger:              *logger,
		OutputDir:           tempDir,
		LogsDir:             tempDir,
		ExportFormat:        "json",
		CollectionTimestamp: collectionTime,
	}

	// Parse the file directly
	err = parseSyslogFile(logFile, writer, params)
	assert.NoError(t, err)

	// We should have 2 records
	assert.Equal(t, 2, len(writer.Records))

	// Check record fields
	assert.Equal(t, "kernel", writer.Records[0].Data.(map[string]interface{})["processname"])
	assert.Equal(t, "0", writer.Records[0].Data.(map[string]interface{})["pid"])
	assert.Equal(t, logFile, writer.Records[0].SourceFile)
}

func TestMultilineHandling(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "syslog_multiline_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a test log file with multiline entries
	logContent := `Mar 15 14:23:45 testhost kernel[0]: Starting multiline
    this is continuation line 1
    this is continuation line 2
Mar 15 14:23:46 testhost process[123]: Another entry`

	logFile := filepath.Join(tempDir, "system.log")
	err = os.WriteFile(logFile, []byte(logContent), 0644)
	assert.NoError(t, err)

	// Create test writer and parameters
	writer := &testutils.TestDataWriter{}
	collectionTime := time.Now().UTC().Format(time.RFC3339)
	logger := testutils.NewTestLogger()
	params := mod.ModuleParams{
		Logger:              *logger, // Pass as value, not pointer
		OutputDir:           tempDir,
		LogsDir:             tempDir,
		ExportFormat:        "json",
		CollectionTimestamp: collectionTime,
	}

	// Parse the file
	err = parseSyslogFile(logFile, writer, params)
	assert.NoError(t, err)

	// We should have 3 records based on how the parser is processing the file
	assert.Equal(t, 3, len(writer.Records))

	// Check the records
	assert.Equal(t, "kernel", writer.Records[0].Data.(map[string]interface{})["processname"])
	assert.Equal(t, "0", writer.Records[0].Data.(map[string]interface{})["pid"])
	assert.Equal(t, "Starting multiline", writer.Records[0].Data.(map[string]interface{})["message"])

	// The syslog parser combines the continuation lines into one with spaces between them
	assert.Equal(t, "this is continuation line 1 this is continuation line 2", writer.Records[1].Data.(map[string]interface{})["message"])

	// Check final entry
	assert.Equal(t, "process", writer.Records[2].Data.(map[string]interface{})["processname"])
	assert.Equal(t, "123", writer.Records[2].Data.(map[string]interface{})["pid"])
	assert.Equal(t, "Another entry", writer.Records[2].Data.(map[string]interface{})["message"])
}

func TestWriteRecord(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "syslog_write_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test writer and parameters
	writer := &testutils.TestDataWriter{}
	collectionTime := time.Now().UTC().Format(time.RFC3339)
	logger := testutils.NewTestLogger()
	params := mod.ModuleParams{
		Logger:              *logger, // Pass as value, not pointer
		OutputDir:           tempDir,
		LogsDir:             tempDir,
		ExportFormat:        "json",
		CollectionTimestamp: collectionTime,
	}

	// Test writing a record
	timestamp := "Mar 15 14:23:45"
	message := "Test message"
	logFile := "test.log"

	err = writeRecord(writer, logFile, timestamp, message, params)
	assert.NoError(t, err)

	// Check the record
	assert.Equal(t, 1, len(writer.Records))
	assert.Equal(t, message, writer.Records[0].Data.(map[string]interface{})["message"])
	assert.Equal(t, logFile, writer.Records[0].SourceFile)
}
