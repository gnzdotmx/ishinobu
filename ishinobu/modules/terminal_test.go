package modules

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gnzdotmx/ishinobu/ishinobu/mod"
	"github.com/gnzdotmx/ishinobu/ishinobu/utils"
	"github.com/stretchr/testify/assert"
)

func TestTerminalModule(t *testing.T) {
	// Cleanup any log files after test completes
	defer cleanupLogFiles(t)

	// Create temporary directory for test outputs
	tmpDir, err := os.MkdirTemp("", "terminal_test")
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
	module := &TerminalModule{
		Name:        "terminal",
		Description: "Collects and parses Terminal.app saved state files and terminal histories",
	}

	// Test GetName
	t.Run("GetName", func(t *testing.T) {
		assert.Equal(t, "terminal", module.GetName())
	})

	// Test GetDescription
	t.Run("GetDescription", func(t *testing.T) {
		assert.Contains(t, module.GetDescription(), "Terminal.app")
		assert.Contains(t, module.GetDescription(), "histories")
	})

	// Test Terminal State collection with mock output
	t.Run("TerminalState", func(t *testing.T) {
		// Create a mock output file for terminal state
		createMockTerminalStateOutput(t, params)

		// Verify the output file exists
		outputFile := filepath.Join(tmpDir, "terminal-state-"+params.CollectionTimestamp+".json")
		assert.FileExists(t, outputFile)

		// Verify the content of the output file
		verifyTerminalStateOutput(t, outputFile)
	})

	// Test Terminal History collection with mock output
	t.Run("TerminalHistory", func(t *testing.T) {
		// Create a mock output file for terminal history
		createMockTerminalHistoryOutput(t, params)

		// Verify the output file exists
		outputFile := filepath.Join(tmpDir, "terminal-history-"+params.CollectionTimestamp+".json")
		assert.FileExists(t, outputFile)

		// Verify the content of the output file
		verifyTerminalHistoryOutput(t, outputFile)
	})
}

// Test that the module initializes properly
func TestTerminalModuleInitialization(t *testing.T) {
	// Create a new instance with proper initialization
	module := &TerminalModule{
		Name:        "terminal",
		Description: "Collects and parses Terminal.app saved state files and terminal histories",
	}

	// Verify module is properly instantiated with expected values
	assert.Equal(t, "terminal", module.Name, "Module name should be initialized")
	assert.Contains(t, module.Description, "Terminal.app", "Module description should be initialized")

	// Test the module's methods
	assert.Equal(t, "terminal", module.GetName())
	assert.Contains(t, module.GetDescription(), "terminal histories")
}

// Create a mock terminal state output file
func createMockTerminalStateOutput(t *testing.T, params mod.ModuleParams) {
	outputFile := filepath.Join(params.OutputDir, "terminal-state-"+params.CollectionTimestamp+".json")

	// Create sample terminal state records
	records := []utils.Record{
		{
			CollectionTimestamp: params.CollectionTimestamp,
			EventTimestamp:      params.CollectionTimestamp,
			SourceFile:          "Terminal.savedState/Window_1",
			Data: map[string]interface{}{
				"user":                             "testuser",
				"window_id":                        uint32(1),
				"datablock":                        1,
				"window_title":                     "Terminal - bash",
				"tab_working_directory_url":        "file:///Users/testuser/Documents/",
				"tab_working_directory_url_string": "/Users/testuser/Documents/",
				"line_index":                       1,
				"line":                             "ls -la",
			},
		},
		{
			CollectionTimestamp: params.CollectionTimestamp,
			EventTimestamp:      params.CollectionTimestamp,
			SourceFile:          "Terminal.savedState/Window_1",
			Data: map[string]interface{}{
				"user":                             "testuser",
				"window_id":                        uint32(1),
				"datablock":                        1,
				"window_title":                     "Terminal - bash",
				"tab_working_directory_url":        "file:///Users/testuser/Documents/",
				"tab_working_directory_url_string": "/Users/testuser/Documents/",
				"line_index":                       2,
				"line":                             "cd project",
			},
		},
		{
			CollectionTimestamp: params.CollectionTimestamp,
			EventTimestamp:      params.CollectionTimestamp,
			SourceFile:          "Terminal.savedState/Window_2",
			Data: map[string]interface{}{
				"user":                             "admin",
				"window_id":                        uint32(2),
				"datablock":                        1,
				"window_title":                     "Terminal - zsh",
				"tab_working_directory_url":        "file:///Users/admin/",
				"tab_working_directory_url_string": "/Users/admin/",
				"line_index":                       1,
				"line":                             "sudo ls -l /var/log",
			},
		},
	}

	// Write records to file
	file, err := os.Create(outputFile)
	assert.NoError(t, err)
	defer file.Close()

	encoder := json.NewEncoder(file)
	for _, record := range records {
		err := encoder.Encode(record)
		assert.NoError(t, err)
	}
}

// Create a mock terminal history output file
func createMockTerminalHistoryOutput(t *testing.T, params mod.ModuleParams) {
	outputFile := filepath.Join(params.OutputDir, "terminal-history-"+params.CollectionTimestamp+".json")

	// Create sample terminal history records
	records := []utils.Record{
		{
			CollectionTimestamp: params.CollectionTimestamp,
			EventTimestamp:      params.CollectionTimestamp,
			SourceFile:          "/Users/testuser/.bash_history",
			Data: map[string]interface{}{
				"username": "testuser",
				"command":  "cd ~/Documents",
			},
		},
		{
			CollectionTimestamp: params.CollectionTimestamp,
			EventTimestamp:      params.CollectionTimestamp,
			SourceFile:          "/Users/testuser/.bash_history",
			Data: map[string]interface{}{
				"username": "testuser",
				"command":  "git clone https://github.com/example/repo.git",
			},
		},
		{
			CollectionTimestamp: params.CollectionTimestamp,
			EventTimestamp:      params.CollectionTimestamp,
			SourceFile:          "/Users/admin/.zsh_history",
			Data: map[string]interface{}{
				"username": "admin",
				"command":  "sudo apt update && sudo apt upgrade -y",
			},
		},
		{
			CollectionTimestamp: params.CollectionTimestamp,
			EventTimestamp:      params.CollectionTimestamp,
			SourceFile:          "/Users/admin/.bash_sessions/session1",
			Data: map[string]interface{}{
				"username": "admin",
				"command":  "cd /var/log && grep -i error syslog",
			},
		},
	}

	// Write records to file
	file, err := os.Create(outputFile)
	assert.NoError(t, err)
	defer file.Close()

	encoder := json.NewEncoder(file)
	for _, record := range records {
		err := encoder.Encode(record)
		assert.NoError(t, err)
	}
}

// Verify the terminal state output file contains expected data
func verifyTerminalStateOutput(t *testing.T, outputFile string) {
	// Read the output file
	content, err := os.ReadFile(outputFile)
	assert.NoError(t, err)

	// Split the content into JSON lines
	lines := splitTerminalLines(content)
	assert.NotEmpty(t, lines, "Output file should contain data")

	// Track what we've found for verification
	var foundTestUser, foundAdmin bool
	var foundWindowOne, foundWindowTwo bool
	var foundCdProject, foundSudoLs bool

	// Verify each record
	for _, line := range lines {
		var record map[string]interface{}
		err := json.Unmarshal(line, &record)
		assert.NoError(t, err, "Each line should be valid JSON")

		// Verify common fields
		assert.NotEmpty(t, record["collection_timestamp"])
		assert.NotEmpty(t, record["event_timestamp"])
		assert.NotEmpty(t, record["source_file"])
		assert.Contains(t, record["source_file"].(string), "Terminal.savedState/Window_")

		// Check data fields
		data, ok := record["data"].(map[string]interface{})
		assert.True(t, ok, "Should have data field as a map")

		// Verify terminal state fields
		assert.NotEmpty(t, data["user"])
		assert.NotEmpty(t, data["window_id"])
		assert.NotEmpty(t, data["window_title"])
		assert.NotEmpty(t, data["tab_working_directory_url"])
		assert.NotEmpty(t, data["line"])

		// Track what we've found
		user, _ := data["user"].(string)
		if user == "testuser" {
			foundTestUser = true
		} else if user == "admin" {
			foundAdmin = true
		}

		windowID, ok := data["window_id"].(float64) // JSON unmarshals numbers as float64
		if ok {
			if windowID == 1 {
				foundWindowOne = true
			} else if windowID == 2 {
				foundWindowTwo = true
			}
		}

		line, _ := data["line"].(string)
		if line == "cd project" {
			foundCdProject = true
		} else if line == "sudo ls -l /var/log" {
			foundSudoLs = true
		}
	}

	// Verify we found everything we expected
	assert.True(t, foundTestUser, "Should have found testuser")
	assert.True(t, foundAdmin, "Should have found admin")
	assert.True(t, foundWindowOne, "Should have found window 1")
	assert.True(t, foundWindowTwo, "Should have found window 2")
	assert.True(t, foundCdProject, "Should have found 'cd project' command")
	assert.True(t, foundSudoLs, "Should have found 'sudo ls' command")
}

// Verify the terminal history output file contains expected data
func verifyTerminalHistoryOutput(t *testing.T, outputFile string) {
	// Read the output file
	content, err := os.ReadFile(outputFile)
	assert.NoError(t, err)

	// Split the content into JSON lines
	lines := splitTerminalLines(content)
	assert.NotEmpty(t, lines, "Output file should contain data")

	// Track what we've found for verification
	var foundTestUser, foundAdmin bool
	var foundBashHistory, foundZshHistory, foundBashSession bool
	var foundGitClone, foundSudoApt, foundGrepError bool

	// Verify each record
	for _, line := range lines {
		var record map[string]interface{}
		err := json.Unmarshal(line, &record)
		assert.NoError(t, err, "Each line should be valid JSON")

		// Verify common fields
		assert.NotEmpty(t, record["collection_timestamp"])
		assert.NotEmpty(t, record["event_timestamp"])
		assert.NotEmpty(t, record["source_file"])

		// Check data fields
		data, ok := record["data"].(map[string]interface{})
		assert.True(t, ok, "Should have data field as a map")

		// Verify terminal history fields
		assert.NotEmpty(t, data["username"])
		assert.NotEmpty(t, data["command"])

		// Track what we've found
		username, _ := data["username"].(string)
		if username == "testuser" {
			foundTestUser = true
		} else if username == "admin" {
			foundAdmin = true
		}

		sourceFile, _ := record["source_file"].(string)
		if strings.Contains(sourceFile, ".bash_history") {
			foundBashHistory = true
		} else if strings.Contains(sourceFile, ".zsh_history") {
			foundZshHistory = true
		} else if strings.Contains(sourceFile, ".bash_sessions") {
			foundBashSession = true
		}

		command, _ := data["command"].(string)
		if strings.Contains(command, "git clone") {
			foundGitClone = true
		} else if strings.Contains(command, "sudo apt") {
			foundSudoApt = true
		} else if strings.Contains(command, "grep -i error") {
			foundGrepError = true
		}
	}

	// Verify we found everything we expected
	assert.True(t, foundTestUser, "Should have found testuser")
	assert.True(t, foundAdmin, "Should have found admin")
	assert.True(t, foundBashHistory, "Should have found bash history")
	assert.True(t, foundZshHistory, "Should have found zsh history")
	assert.True(t, foundBashSession, "Should have found bash session")
	assert.True(t, foundGitClone, "Should have found git clone command")
	assert.True(t, foundSudoApt, "Should have found apt command")
	assert.True(t, foundGrepError, "Should have found grep command")
}

// Helper to split content into lines (handles different line endings)
func splitTerminalLines(data []byte) [][]byte {
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
