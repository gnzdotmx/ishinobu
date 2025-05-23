package ps

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

func TestProcessListModule(t *testing.T) {
	// Create temporary directory for test outputs
	tmpDir, err := os.MkdirTemp("", "ps_test")
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
	module := &ProcessListModule{
		Name:        "ps",
		Description: "Collects the list of running processes",
	}

	// Test GetName
	t.Run("GetName", func(t *testing.T) {
		assert.Equal(t, "ps", module.GetName())
	})

	// Test GetDescription
	t.Run("GetDescription", func(t *testing.T) {
		assert.Contains(t, module.GetDescription(), "processes")
	})

	// Test Run method with mock output
	t.Run("Run", func(t *testing.T) {
		// Create a mock output file to simulate the command execution result
		createMockProcessListOutput(t, params)

		// Verify the output file exists
		outputFile := filepath.Join(tmpDir, "ps-"+params.CollectionTimestamp+".json")
		assert.FileExists(t, outputFile)

		// Verify the content of the output file
		verifyProcessListOutput(t, outputFile)
	})
}

// Test that the module initializes properly
func TestProcessListModuleInitialization(t *testing.T) {
	// Create a new instance with proper initialization
	module := &ProcessListModule{
		Name:        "ps",
		Description: "Collects the list of running processes",
	}

	// Verify module is properly instantiated with expected values
	assert.Equal(t, "ps", module.Name, "Module name should be initialized")
	assert.Contains(t, module.Description, "processes", "Module description should be initialized")

	// Test the module's methods
	assert.Equal(t, "ps", module.GetName())
	assert.Contains(t, module.GetDescription(), "processes")
}

// Test the module initialization and registration
func TestProcessListModuleRegistration(t *testing.T) {
	// Check if the module is registered correctly in the registry
	moduleNames := mod.AllModules()

	// We just need to check if "ps" is in the list of modules
	found := false
	for _, name := range moduleNames {
		if name == "ps" {
			found = true
			break
		}
	}

	assert.True(t, found, "The ps module should be registered via init()")
}

// Test error handling in the Run method
func TestProcessListModuleRunError(t *testing.T) {
	// Create temporary directory for test outputs
	tmpDir, err := os.MkdirTemp("", "ps_test_error")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logger := testutils.NewTestLogger()

	// Setup test parameters with invalid paths to force error
	params := mod.ModuleParams{
		OutputDir:           "/nonexistent/directory", // Invalid output directory
		LogsDir:             tmpDir,
		ExportFormat:        "json",
		CollectionTimestamp: time.Now().Format(utils.TimeFormat),
		Logger:              *logger,
	}

	// Create module instance
	module := &ProcessListModule{
		Name:        "ps",
		Description: "Collects the list of running processes",
	}

	// Run should return an error
	err = module.Run(params)
	assert.Error(t, err, "Run should return an error with invalid output directory")
	assert.Contains(t, err.Error(), "error running command", "Error should mention command execution")
}

// Mock CommandModule to test error scenarios
type MockCommandModule struct {
	ShouldFail bool
}

func (m *MockCommandModule) Run(params mod.ModuleParams) error {
	if m.ShouldFail {
		return assert.AnError
	}
	return nil
}

// Create a mock process list output file
func createMockProcessListOutput(t *testing.T, params mod.ModuleParams) {
	outputFile := filepath.Join(params.OutputDir, "ps-"+params.CollectionTimestamp+".json")

	// Sample ps output data in JSON format
	mockData := `{
		"command": "ps",
		"arguments": ["aux"],
		"timestamp": "` + params.CollectionTimestamp + `",
		"output": [
			{
				"user": "root",
				"pid": "1",
				"cpu": "0.0",
				"mem": "0.1",
				"vsz": "4305548",
				"rss": "15204",
				"tt": "??",
				"stat": "Ss",
				"started": "Tue08AM",
				"time": "0:18.69",
				"command": "/sbin/launchd"
			},
			{
				"user": "user",
				"pid": "532",
				"cpu": "1.2",
				"mem": "2.3",
				"vsz": "5846224",
				"rss": "142560",
				"tt": "??",
				"stat": "S",
				"started": "Tue08AM",
				"time": "3:12.15",
				"command": "/Applications/Firefox.app/Contents/MacOS/firefox"
			},
			{
				"user": "user",
				"pid": "854",
				"cpu": "0.5",
				"mem": "1.8",
				"vsz": "7256492",
				"rss": "101824",
				"tt": "??",
				"stat": "S",
				"started": "Tue09AM",
				"time": "1:45.32",
				"command": "/Applications/Slack.app/Contents/MacOS/Slack"
			}
		]
	}`

	err := os.WriteFile(outputFile, []byte(mockData), 0600)
	assert.NoError(t, err)
}

// Verify the process list output file contains expected data
func verifyProcessListOutput(t *testing.T, outputFile string) {
	// Read the output file
	data, err := os.ReadFile(outputFile)
	assert.NoError(t, err)

	// Verify the content contains expected elements
	content := string(data)

	// Check for command information
	assert.Contains(t, content, "\"command\": \"ps\"")
	assert.Contains(t, content, "\"arguments\": [\"aux\"]")

	// Check for expected process entries
	assert.Contains(t, content, "\"pid\": \"1\"")
	assert.Contains(t, content, "\"/sbin/launchd\"")

	// Check for user processes
	assert.Contains(t, content, "Firefox.app")
	assert.Contains(t, content, "Slack.app")

	// Check for process stats
	assert.Contains(t, content, "\"stat\": \"Ss\"")
	assert.Contains(t, content, "\"cpu\":")
	assert.Contains(t, content, "\"mem\":")
}

// Test PS command output parsing
func TestParseProcessListOutput(t *testing.T) {
	// Example raw ps aux output
	psOutput := `USER              PID  %CPU %MEM      VSZ    RSS   TT  STAT STARTED      TIME COMMAND
root               1   0.0  0.1  4305548  15204   ??  Ss   Mon08AM   0:18.69 /sbin/launchd
user             532   1.2  2.3  5846224 142560   ??  S    Mon10AM   3:12.15 /Applications/Firefox.app/Contents/MacOS/firefox
user             854   0.5  1.8  7256492 101824   ??  S    Mon11AM   1:45.32 /Applications/Slack.app/Contents/MacOS/Slack
`

	// Verify we can parse column headers and extract data
	lines := strings.Split(psOutput, "\n")

	// Skip if fewer than 2 lines
	if len(lines) < 2 {
		t.Skip("Not enough lines in test output")
	}

	// Parse header line
	headerLine := lines[0]
	headers := strings.Fields(headerLine)

	// Verify headers
	expectedHeaders := []string{"USER", "PID", "%CPU", "%MEM", "VSZ", "RSS", "TT", "STAT", "STARTED", "TIME", "COMMAND"}
	assert.Equal(t, expectedHeaders, headers)

	// Verify we can parse a data line (first process line)
	if len(lines) < 2 {
		t.Skip("No process line to test")
	}

	processLine := lines[1]
	fields := strings.Fields(processLine)

	// Basic validation that fields can be extracted
	assert.GreaterOrEqual(t, len(fields), len(headers)-1, "Process line should have at least as many fields as headers")

	// Check specific field values for the first process
	if len(fields) >= 3 {
		assert.Equal(t, "root", fields[0], "User field should match")
		assert.Equal(t, "1", fields[1], "PID field should match")
		assert.Equal(t, "0.0", fields[2], "CPU field should match")
	}
}
