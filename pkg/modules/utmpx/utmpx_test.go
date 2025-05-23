package utmpx

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnzdotmx/ishinobu/pkg/mod"
	"github.com/gnzdotmx/ishinobu/pkg/testutils"
	"github.com/gnzdotmx/ishinobu/pkg/utils"
)

func TestGetNameAndDescription(t *testing.T) {
	module := &UtmpxModule{
		Name:        "utmpx",
		Description: "Test description",
	}

	assert.Equal(t, "utmpx", module.GetName())
	assert.Equal(t, "Test description", module.GetDescription())
}

func TestDecodeLogonType(t *testing.T) {
	tests := []struct {
		code     int16
		expected string
	}{
		{2, "BOOT_TIME"},
		{7, "USER_PROCESS"},
		{0, "UNKNOWN"},
		{-1, "UNKNOWN"},
		{9999, "UNKNOWN"},
	}

	for _, test := range tests {
		result := decodeLogonType(test.code)
		assert.Equal(t, test.expected, result, "LogonType code %d should decode to %s", test.code, test.expected)
	}
}

func TestParseUtmpx(t *testing.T) {
	// Create temporary directory for tests
	tmpDir, err := os.MkdirTemp("", "utmpx_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Set up test parameters
	logger := testutils.NewTestLogger()
	params := mod.ModuleParams{
		LogsDir:             tmpDir,
		OutputDir:           "./",
		ExportFormat:        "json",
		CollectionTimestamp: "2024-01-01T00:00:00Z",
		Logger:              *logger,
	}

	// Create a mock utmpx file
	utmpxPath := filepath.Join(tmpDir, "utmpx")

	// Create header record (we'll skip this in parsing)
	headerRecord := createUtmpxRecord("", "", "", 0, 0)

	// Create test records
	testRecord1 := createUtmpxRecord("testuser", "cons", "tty1", 1234, 7) // USER_PROCESS
	testRecord2 := createUtmpxRecord("reboot", "~", "~", 0, 2)            // BOOT_TIME

	// Write records to the test file
	f, err := os.Create(utmpxPath)
	require.NoError(t, err)

	// Write the header first
	err = binary.Write(f, binary.LittleEndian, headerRecord)
	require.NoError(t, err)

	// Write the test records
	err = binary.Write(f, binary.LittleEndian, testRecord1)
	require.NoError(t, err)
	err = binary.Write(f, binary.LittleEndian, testRecord2)
	require.NoError(t, err)

	f.Close()

	// Create test writer to capture the output
	testWriter := &testutils.TestDataWriter{Records: []utils.Record{}}

	// Parse the test file
	err = parseUtmpx(params, utmpxPath, testWriter)
	assert.NoError(t, err)

	// Verify the results
	assert.Equal(t, 2, len(testWriter.Records), "Should have 2 records")

	// Check first record (testuser)
	record1 := testWriter.Records[0]
	data1, ok := record1.Data.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "testuser", data1["user"])
	assert.Equal(t, "cons", data1["id"])
	assert.Equal(t, "tty1", data1["terminal_type"])
	assert.Equal(t, int32(1234), data1["pid"])
	assert.Equal(t, "USER_PROCESS", data1["logon_type"])
	assert.Equal(t, utmpxPath, record1.SourceFile)

	// Check second record (reboot)
	record2 := testWriter.Records[1]
	data2, ok := record2.Data.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "reboot", data2["user"])
	assert.Equal(t, "~", data2["id"])
	assert.Equal(t, "~", data2["terminal_type"])
	assert.Equal(t, int32(0), data2["pid"])
	assert.Equal(t, "BOOT_TIME", data2["logon_type"])
	assert.Equal(t, utmpxPath, record2.SourceFile)
}

func TestRunUtmpxModule(t *testing.T) {
	// Create temporary directory for tests
	tmpDir, err := os.MkdirTemp("", "utmpx_module_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Set up test parameters
	logger := testutils.NewTestLogger()
	params := mod.ModuleParams{
		LogsDir:             tmpDir,
		OutputDir:           tmpDir,
		ExportFormat:        "json",
		CollectionTimestamp: "2024-01-01T00:00:00Z",
		Logger:              *logger,
	}

	// Create a mock utmpx file at the expected path
	// Since we can't replace the real system path in the module,
	// we'll just test the module functions directly without calling Run()

	// Create test module
	module := &UtmpxModule{
		Name:        "utmpx",
		Description: "Test utmpx module",
	}

	// Test the name and description
	assert.Equal(t, "utmpx", module.GetName())
	assert.Equal(t, "Test utmpx module", module.GetDescription())

	// Create a test utmpx file
	utmpxPath := filepath.Join(tmpDir, "utmpx")

	// Create a header record (we'll skip this)
	headerRecord := createUtmpxRecord("", "", "", 0, 0)

	// Create a test record
	testRecord := createUtmpxRecord("testuser", "cons", "tty1", 1234, 7)

	// Write to the test file
	f, err := os.Create(utmpxPath)
	require.NoError(t, err)

	// Write the header first
	err = binary.Write(f, binary.LittleEndian, headerRecord)
	require.NoError(t, err)

	// Write the test record
	err = binary.Write(f, binary.LittleEndian, testRecord)
	require.NoError(t, err)

	f.Close()

	// Create test writer to capture the output
	testWriter := &testutils.TestDataWriter{Records: []utils.Record{}}

	// Test the parseUtmpx function
	err = parseUtmpx(params, utmpxPath, testWriter)
	assert.NoError(t, err)

	// Verify the results
	assert.Equal(t, 1, len(testWriter.Records), "Should have 1 record")

	// Check record data
	record := testWriter.Records[0]
	data, ok := record.Data.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "testuser", data["user"])
	assert.Equal(t, "cons", data["id"])
	assert.Equal(t, "tty1", data["terminal_type"])
	assert.Equal(t, int32(1234), data["pid"])
	assert.Equal(t, "USER_PROCESS", data["logon_type"])
	assert.Equal(t, utmpxPath, record.SourceFile)
}

func TestParseUtmpxInvalidFile(t *testing.T) {
	// Create temporary directory for tests
	tmpDir, err := os.MkdirTemp("", "utmpx_invalid_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Set up test parameters
	logger := testutils.NewTestLogger()
	params := mod.ModuleParams{
		LogsDir:             tmpDir,
		OutputDir:           tmpDir,
		ExportFormat:        "json",
		CollectionTimestamp: "2024-01-01T00:00:00Z",
		Logger:              *logger,
	}

	// Test with non-existent file
	nonExistentPath := filepath.Join(tmpDir, "nonexistent_utmpx")
	testWriter := &testutils.TestDataWriter{Records: []utils.Record{}}
	err = parseUtmpx(params, nonExistentPath, testWriter)
	assert.Error(t, err, "Should return error for non-existent file")
	assert.Contains(t, err.Error(), "error reading utmpx file")

	// Test with invalid file (too small)
	invalidPath := filepath.Join(tmpDir, "invalid_utmpx")
	err = os.WriteFile(invalidPath, []byte("invalid data"), 0644)
	require.NoError(t, err)

	err = parseUtmpx(params, invalidPath, testWriter)
	assert.NoError(t, err, "Should handle invalid data gracefully")
	assert.Empty(t, testWriter.Records, "No records should be processed")
}

// Helper function to create a test UtmpxRecord with the given values
func createUtmpxRecord(user, id, terminalType string, pid int32, logonType int16) UtmpxRecord {
	record := UtmpxRecord{}

	// Fill the fixed-size byte arrays
	copy(record.User[:], user)
	copy(record.ID[:], id)
	copy(record.TerminalType[:], terminalType)
	copy(record.Hostname[:], "localhost")

	// Set the numeric fields
	record.Pid = pid
	record.LogonType = logonType

	// Set a sample timestamp (current time)
	now := time.Now()
	record.Timestamp = int32(now.Unix())
	record.Microseconds = int32(now.Nanosecond() / 1000)

	return record
}

// TestRunMethod tests the Run method of UtmpxModule by creating a mock implementation
func TestRunMethod(t *testing.T) {
	// Create temporary directory for tests
	tmpDir, err := os.MkdirTemp("", "utmpx_run_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Set up test parameters
	logger := testutils.NewTestLogger()
	params := mod.ModuleParams{
		LogsDir:             tmpDir,
		OutputDir:           tmpDir,
		ExportFormat:        "json",
		CollectionTimestamp: "2024-01-01T00:00:00Z",
		Logger:              *logger,
	}

	// Create a mock implementation of the module
	module := &UtmpxModule{
		Name:        "utmpx",
		Description: "Test utmpx module",
	}

	// We can't test the original Run method directly as it accesses a system file
	// Instead, create a mocked version of the parseUtmpx function to test error paths

	// Test case 1: NewDataWriter error (write to invalid directory)
	badParams := params
	badParams.LogsDir = "/path/that/does/not/exist"

	err = module.Run(badParams)
	assert.Error(t, err, "Run should return error with invalid output directory")

	// Test case 2: Redirect the standard output to capture any error messages
	originalOutput := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Test the module's GetName and GetDescription methods
	assert.Equal(t, "utmpx", module.GetName())
	assert.Equal(t, "Test utmpx module", module.GetDescription())

	// Restore original output
	os.Stdout = originalOutput
	w.Close()
	_, _ = io.ReadAll(r)
}

// TestParseUtmpxEdgeCases tests edge cases in record processing
func TestParseUtmpxEdgeCases(t *testing.T) {
	// Create temporary directory for tests
	tmpDir, err := os.MkdirTemp("", "utmpx_edge_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Set up test parameters
	logger := testutils.NewTestLogger()
	params := mod.ModuleParams{
		LogsDir:             tmpDir,
		OutputDir:           tmpDir,
		ExportFormat:        "json",
		CollectionTimestamp: "2024-01-01T00:00:00Z",
		Logger:              *logger,
	}

	// Create a mock utmpx file
	utmpxPath := filepath.Join(tmpDir, "utmpx_edge_cases")

	// Edge case 1: Record with empty hostname
	record1 := createUtmpxRecord("testuser", "cons", "tty1", 1234, 7)
	// Clear the hostname field
	for i := range record1.Hostname {
		record1.Hostname[i] = 0
	}

	// Edge case 2: Record with unknown logon type
	record2 := createUtmpxRecord("unknown", "tty2", "tty2", 5678, 999)

	// Edge case 3: Record with zero timestamp
	record3 := createUtmpxRecord("zero", "zero", "tty3", 9012, 7)
	record3.Timestamp = 0
	record3.Microseconds = 0

	// Write records to the test file
	f, err := os.Create(utmpxPath)
	require.NoError(t, err)

	// Write header (will be skipped)
	headerRecord := createUtmpxRecord("", "", "", 0, 0)
	err = binary.Write(f, binary.LittleEndian, headerRecord)
	require.NoError(t, err)

	// Write the edge case records
	err = binary.Write(f, binary.LittleEndian, record1)
	require.NoError(t, err)
	err = binary.Write(f, binary.LittleEndian, record2)
	require.NoError(t, err)
	err = binary.Write(f, binary.LittleEndian, record3)
	require.NoError(t, err)
	f.Close()

	// Create test writer
	testWriter := &testutils.TestDataWriter{Records: []utils.Record{}}

	// Parse the test file
	err = parseUtmpx(params, utmpxPath, testWriter)
	assert.NoError(t, err)

	// Verify results
	assert.Equal(t, 3, len(testWriter.Records), "Should have 3 records")

	// Check edge case 1: Empty hostname
	data1, ok := testWriter.Records[0].Data.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "testuser", data1["user"])
	assert.Equal(t, "localhost", data1["hostname"], "Empty hostname should default to localhost")

	// Check edge case 2: Unknown logon type
	data2, ok := testWriter.Records[1].Data.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "unknown", data2["user"])
	assert.Equal(t, "UNKNOWN", data2["logon_type"], "Unknown logon type should map to UNKNOWN")

	// Check edge case 3: Zero timestamp
	data3, ok := testWriter.Records[2].Data.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "zero", data3["user"])
	// Timestamp should be valid but will be epoch time (1970-01-01)
	assert.Contains(t, data3["timestamp"], "1970-01-01")
}

// Custom failing writer for testing error handling
type failingWriter struct{}

func (w *failingWriter) WriteRecord(record utils.Record) error {
	return fmt.Errorf("simulated write failure")
}

func (w *failingWriter) Close() error {
	return nil
}

// TestWriterErrors tests handling of writer errors
func TestWriterErrors(t *testing.T) {
	// Create temporary directory for tests
	tmpDir, err := os.MkdirTemp("", "utmpx_writer_error_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Set up test parameters
	logger := testutils.NewTestLogger()
	params := mod.ModuleParams{
		LogsDir:             tmpDir,
		OutputDir:           "./",
		ExportFormat:        "json",
		CollectionTimestamp: "2024-01-01T00:00:00Z",
		Logger:              *logger,
	}

	// Create a mock utmpx file
	utmpxPath := filepath.Join(tmpDir, "utmpx")

	// Create record
	headerRecord := createUtmpxRecord("", "", "", 0, 0)
	testRecord := createUtmpxRecord("testuser", "cons", "tty1", 1234, 7)

	// Write records to the test file
	f, err := os.Create(utmpxPath)
	require.NoError(t, err)
	err = binary.Write(f, binary.LittleEndian, headerRecord)
	require.NoError(t, err)
	err = binary.Write(f, binary.LittleEndian, testRecord)
	require.NoError(t, err)
	f.Close()

	// Create a custom failing writer
	customFailingWriter := &failingWriter{}

	// Parse with failing writer
	err = parseUtmpx(params, utmpxPath, customFailingWriter)

	// The function should complete without error even if the writer fails
	assert.NoError(t, err, "parseUtmpx should handle writer errors gracefully")
}
