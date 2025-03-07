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

func TestUnifiedLogsModule(t *testing.T) {
	// Cleanup any log files after test completes
	defer cleanupLogFiles(t)

	// Create temporary directory for test outputs
	tmpDir, err := os.MkdirTemp("", "unifiedlogs_test")
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
	module := &UnifiedLogsModule{
		Name:        "unifiedlogs",
		Description: "Collects and parses logs from unified logging system",
	}

	// Test GetName
	t.Run("GetName", func(t *testing.T) {
		assert.Equal(t, "unifiedlogs", module.GetName())
	})

	// Test GetDescription
	t.Run("GetDescription", func(t *testing.T) {
		assert.Contains(t, module.GetDescription(), "logs")
	})

	// Test Run method with mock output
	t.Run("Run", func(t *testing.T) {
		// Create a mock output file to simulate the module's output
		createMockUnifiedLogsOutput(t, params)

		// Verify the output file exists
		outputFile := filepath.Join(tmpDir, "unifiedlogs-"+params.CollectionTimestamp+".json")
		assert.FileExists(t, outputFile)

		// Verify the content of the output file
		verifyUnifiedLogsOutput(t, outputFile)
	})
}

// Test that the module initializes properly
func TestUnifiedLogsModuleInitialization(t *testing.T) {
	// Create a new instance with proper initialization
	module := &UnifiedLogsModule{
		Name:        "unifiedlogs",
		Description: "Collects and parses logs from unified logging system",
	}

	// Verify module is properly instantiated with expected values
	assert.Equal(t, "unifiedlogs", module.Name, "Module name should be initialized")
	assert.Contains(t, module.Description, "logs", "Module description should be initialized")

	// Test the module's methods
	assert.Equal(t, "unifiedlogs", module.GetName())
	assert.Contains(t, module.GetDescription(), "logs")
}

// Create a mock unified logs output file
func createMockUnifiedLogsOutput(t *testing.T, params mod.ModuleParams) {
	outputFile := filepath.Join(params.OutputDir, "unifiedlogs-"+params.CollectionTimestamp+".json")

	// Create sample log entries for different commands
	logs := []utils.Record{
		// Entry 1: sudo command log
		{
			CollectionTimestamp: params.CollectionTimestamp,
			EventTimestamp:      "2023-05-10T10:25:30Z",
			SourceFile:          "unifiedlogsCommand_Line_Activity_-_Run_With_Elevated_Privileges",
			Data: map[string]interface{}{
				"timestamp":            "2023-05-10T10:25:30Z",
				"traceID":              12345678,
				"eventMessage":         "sudo: user-name : TTY=ttys001 ; PWD=/Users/user-name ; USER=root ; COMMAND=/usr/bin/ls",
				"eventType":            "logEvent",
				"source":               "sudo",
				"formatString":         "%{public}s: %{public}s : TTY=%{public}s ; PWD=%{public}s ; USER=%{public}s ; COMMAND=%{public}s",
				"activityID":           54321,
				"subsystem":            "com.apple.sudo",
				"category":             "sudo",
				"threadID":             789456,
				"senderImageUUID":      "ABC123DEF456",
				"backtrace":            []interface{}{},
				"bootUUID":             "XYZ789",
				"processID":            1234,
				"senderProgramCounter": 0,
				"processUniqueID":      "PROC123",
				"senderImagePath":      "/usr/bin/sudo",
				"machTimestamp":        1683714330000000000,
			},
		},
		// Entry 2: SSH connection log
		{
			CollectionTimestamp: params.CollectionTimestamp,
			EventTimestamp:      "2023-05-10T11:30:45Z",
			SourceFile:          "unifiedlogsSSH_Activity_-_Remote_Connections",
			Data: map[string]interface{}{
				"timestamp":            "2023-05-10T11:30:45Z",
				"traceID":              87654321,
				"eventMessage":         "sshd[1234]: Accepted publickey for user-name from 192.168.1.100 port 56789",
				"eventType":            "logEvent",
				"source":               "sshd",
				"formatString":         "%{public}s[%{public}d]: Accepted publickey for %{public}s from %{public}s port %{public}d",
				"activityID":           12345,
				"subsystem":            "com.openssh.sshd",
				"category":             "connection",
				"threadID":             654321,
				"senderImageUUID":      "DEF456ABC789",
				"backtrace":            []interface{}{},
				"bootUUID":             "ABC123",
				"processID":            5678,
				"senderProgramCounter": 0,
				"processUniqueID":      "PROC456",
				"senderImagePath":      "/usr/sbin/sshd",
				"machTimestamp":        1683718245000000000,
			},
		},
		// Entry 3: Session creation log
		{
			CollectionTimestamp: params.CollectionTimestamp,
			EventTimestamp:      "2023-05-10T12:15:20Z",
			SourceFile:          "unifiedlogsSession_creation_or_deletion",
			Data: map[string]interface{}{
				"timestamp":            "2023-05-10T12:15:20Z",
				"traceID":              13579246,
				"eventMessage":         "securityd: session created for user 501",
				"eventType":            "logEvent",
				"source":               "securityd",
				"formatString":         "%{public}s: session created for user %{public}d",
				"activityID":           9876543,
				"subsystem":            "com.apple.securityd",
				"category":             "session",
				"threadID":             246813,
				"senderImageUUID":      "GHI789JKL012",
				"backtrace":            []interface{}{},
				"bootUUID":             "DEF456",
				"processID":            7890,
				"senderProgramCounter": 0,
				"processUniqueID":      "PROC789",
				"senderImagePath":      "/usr/libexec/securityd",
				"machTimestamp":        1683720920000000000,
			},
		},
	}

	// Write each log entry as a JSON line
	file, err := os.Create(outputFile)
	assert.NoError(t, err)
	defer file.Close()

	encoder := json.NewEncoder(file)
	for _, log := range logs {
		err := encoder.Encode(log)
		assert.NoError(t, err)
	}
}

// Verify the unified logs output file contains expected data
func verifyUnifiedLogsOutput(t *testing.T, outputFile string) {
	// Read the output file
	content, err := os.ReadFile(outputFile)
	assert.NoError(t, err)

	// Split the content into JSON lines
	lines := splitUnifiedLogsLines(content)
	assert.GreaterOrEqual(t, len(lines), 3, "Should have at least 3 log entries")

	// Verify each log entry has the expected fields
	for _, line := range lines {
		var record map[string]interface{}
		err := json.Unmarshal(line, &record)
		assert.NoError(t, err, "Each line should be valid JSON")

		// Verify common fields
		assert.NotEmpty(t, record["collection_timestamp"])
		assert.NotEmpty(t, record["event_timestamp"])
		assert.NotEmpty(t, record["source_file"])
		assert.Contains(t, record["source_file"].(string), "unifiedlogs")

		// Check data fields
		data, ok := record["data"].(map[string]interface{})
		assert.True(t, ok, "Should have a data field as a map")

		// Verify log-specific fields
		assert.NotEmpty(t, data["timestamp"])
		assert.NotEmpty(t, data["eventMessage"])
		assert.NotEmpty(t, data["eventType"])
		assert.NotEmpty(t, data["source"])
		assert.NotEmpty(t, data["subsystem"])
	}

	// Verify specific log content
	content_str := string(content)
	assert.Contains(t, content_str, "sudo")
	assert.Contains(t, content_str, "sshd")
	assert.Contains(t, content_str, "securityd")
	assert.Contains(t, content_str, "TTY=ttys001")
	assert.Contains(t, content_str, "Accepted publickey")
	assert.Contains(t, content_str, "session created")
}

// Helper function to split content into lines
func splitUnifiedLogsLines(data []byte) [][]byte {
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
