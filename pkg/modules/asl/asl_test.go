package asl

import (
	"encoding/xml"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gnzdotmx/ishinobu/pkg/mod"
	"github.com/gnzdotmx/ishinobu/pkg/modules/testutils"
	"github.com/gnzdotmx/ishinobu/pkg/utils"
)

func TestUnmarshalXML(t *testing.T) {
	xmlData := `
	<dict>
		<key>ASLMessageID</key>
		<string>12345</string>
		<key>Time</key>
		<string>2024-03-20T10:00:00Z</string>
		<key>TimeNanoSec</key>
		<string>123456789</string>
		<key>Level</key>
		<string>5</string>
		<key>PID</key>
		<string>1234</string>
		<key>UID</key>
		<string>501</string>
		<key>GID</key>
		<string>20</string>
		<key>ReadGID</key>
		<string>80</string>
		<key>Host</key>
		<string>localhost</string>
		<key>Sender</key>
		<string>test.app</string>
		<key>Facility</key>
		<string>com.apple.system</string>
		<key>Message</key>
		<string>Test message</string>
		<key>MsgCount</key>
		<string>1</string>
		<key>ShimCount</key>
		<string>0</string>
		<key>SenderMachUUID</key>
		<string>abcdef-123456</string>
	</dict>`

	var entry LogEntry
	decoder := xml.NewDecoder(strings.NewReader(xmlData))
	start := xml.StartElement{Name: xml.Name{Local: "dict"}}
	err := entry.UnmarshalXML(decoder, start)

	assert.NoError(t, err)
	assert.Equal(t, "12345", entry.ASLMessageID)
	assert.Equal(t, "2024-03-20T10:00:00Z", entry.Time)
	assert.Equal(t, "123456789", entry.TimeNanoSec)
	assert.Equal(t, "5", entry.Level)
	assert.Equal(t, "1234", entry.PID)
	assert.Equal(t, "501", entry.UID)
	assert.Equal(t, "20", entry.GID)
	assert.Equal(t, "80", entry.ReadGID)
	assert.Equal(t, "localhost", entry.Host)
	assert.Equal(t, "test.app", entry.Sender)
	assert.Equal(t, "com.apple.system", entry.Facility)
	assert.Equal(t, "Test message", entry.Message)
	assert.Equal(t, "1", entry.MsgCount)
	assert.Equal(t, "0", entry.ShimCount)
	assert.Equal(t, "abcdef-123456", entry.SenderMachUUID)
}

func TestParseASLFileWithRealSyslog(t *testing.T) {
	// Check if syslog command is available
	_, err := exec.LookPath("syslog")
	if err != nil {
		t.Skip("syslog command not available, skipping test")
	}

	// Find real ASL files in the system
	aslFiles, err := filepath.Glob("/private/var/log/asl/*.asl")
	if err != nil || len(aslFiles) == 0 {
		t.Skip("No ASL files found, skipping test")
	}

	// Set up test directories
	tmpDir, err := os.MkdirTemp("", "asl_real_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create test parameters
	logger := testutils.NewTestLogger()
	params := mod.ModuleParams{
		OutputDir:           tmpDir,
		LogsDir:             tmpDir,
		ExportFormat:        "json",
		CollectionTimestamp: "2024-01-01T00:00:00Z",
		Logger:              *logger,
	}

	// Create our TestDataWriter
	testWriter := &testutils.TestDataWriter{Records: []utils.Record{}}

	// Choose a recent ASL file to test with
	testAslFile := aslFiles[0]

	// Run parseASLFile with real syslog command
	err = parseASLFile(params, []string{testAslFile}, testWriter)
	assert.NoError(t, err)

	// Check that records were processed
	assert.NotEmpty(t, testWriter.Records, "No records were processed from real ASL file")

	// Verify a sample of records if any were found
	if len(testWriter.Records) > 0 {
		record := testWriter.Records[0]

		// Check basic structure but not specific values since real data varies
		assert.NotEmpty(t, record.EventTimestamp, "Event timestamp should not be empty")
		assert.Equal(t, testAslFile, record.SourceFile)
		assert.Contains(t, record.Data, "ASLMessageID")
		assert.Contains(t, record.Data, "Time")
		assert.Contains(t, record.Data, "Message")
		assert.Contains(t, record.Data, "Level")
		assert.Contains(t, record.Data, "PID")
		assert.Contains(t, record.Data, "UID")
		assert.Contains(t, record.Data, "GID")
		assert.Contains(t, record.Data, "ReadGID")

		// Verify data is a map with expected fields
		_, ok := record.Data.(map[string]interface{})
		assert.True(t, ok, "Record data should be a map")
	}
}
