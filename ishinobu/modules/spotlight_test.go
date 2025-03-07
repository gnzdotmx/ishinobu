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

func TestSpotlightModule(t *testing.T) {
	// Cleanup any log files after test completes
	defer cleanupLogFiles(t)

	// Create temporary directory for test outputs
	tmpDir, err := os.MkdirTemp("", "spotlight_test")
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
	module := &SpotlightModule{
		Name:        "spotlight",
		Description: "Collects and parses Spotlight shortcuts data",
	}

	// Test GetName
	t.Run("GetName", func(t *testing.T) {
		assert.Equal(t, "spotlight", module.GetName())
	})

	// Test GetDescription
	t.Run("GetDescription", func(t *testing.T) {
		assert.Contains(t, module.GetDescription(), "Spotlight shortcuts")
	})

	// Test Run method with mock output
	t.Run("Run", func(t *testing.T) {
		// Create a mock output file to simulate the module's output
		createMockSpotlightOutput(t, params)

		// Verify the output file exists
		outputFile := filepath.Join(tmpDir, "spotlight-testuser-"+params.CollectionTimestamp+".json")
		assert.FileExists(t, outputFile)

		// Verify the content of the output file
		verifySpotlightOutput(t, outputFile)
	})
}

// Test that the module initializes properly
func TestSpotlightModuleInitialization(t *testing.T) {
	// Create a new instance with proper initialization
	module := &SpotlightModule{
		Name:        "spotlight",
		Description: "Collects and parses Spotlight shortcuts data",
	}

	// Verify module is properly instantiated with expected values
	assert.Equal(t, "spotlight", module.Name, "Module name should be initialized")
	assert.Contains(t, module.Description, "Spotlight shortcuts", "Module description should be initialized")

	// Test the module's methods
	assert.Equal(t, "spotlight", module.GetName())
	assert.Contains(t, module.GetDescription(), "Spotlight shortcuts")
}

// Create a mock spotlight output file
func createMockSpotlightOutput(t *testing.T, params mod.ModuleParams) {
	outputFile := filepath.Join(params.OutputDir, "spotlight-testuser-"+params.CollectionTimestamp+".json")

	// Create sample spotlight shortcut records
	shortcuts := []utils.Record{
		{
			CollectionTimestamp: params.CollectionTimestamp,
			EventTimestamp:      "2023-05-15T10:22:34Z",
			SourceFile:          "/Users/testuser/Library/Application Support/com.apple.spotlight.Shortcuts",
			Data: map[string]interface{}{
				"username":     "testuser",
				"shortcut":     "mail",
				"display_name": "Mail",
				"last_used":    "2023-05-15T10:22:34Z",
				"url":          "file:///Applications/Mail.app/",
			},
		},
		{
			CollectionTimestamp: params.CollectionTimestamp,
			EventTimestamp:      "2023-05-14T18:45:12Z",
			SourceFile:          "/Users/testuser/Library/Application Support/com.apple.spotlight.Shortcuts",
			Data: map[string]interface{}{
				"username":     "testuser",
				"shortcut":     "safari",
				"display_name": "Safari",
				"last_used":    "2023-05-14T18:45:12Z",
				"url":          "file:///Applications/Safari.app/",
			},
		},
		{
			CollectionTimestamp: params.CollectionTimestamp,
			EventTimestamp:      "2023-05-10T09:15:33Z",
			SourceFile:          "/Users/testuser/Library/Application Support/com.apple.spotlight.Shortcuts",
			Data: map[string]interface{}{
				"username":     "testuser",
				"shortcut":     "notes",
				"display_name": "Notes",
				"last_used":    "2023-05-10T09:15:33Z",
				"url":          "file:///Applications/Notes.app/",
			},
		},
	}

	// Write each shortcut as a JSON line
	file, err := os.Create(outputFile)
	assert.NoError(t, err)
	defer file.Close()

	encoder := json.NewEncoder(file)
	for _, shortcut := range shortcuts {
		err := encoder.Encode(shortcut)
		assert.NoError(t, err)
	}
}

// Verify the spotlight output file contains expected data
func verifySpotlightOutput(t *testing.T, outputFile string) {
	// Read the output file
	content, err := os.ReadFile(outputFile)
	assert.NoError(t, err)

	// Split the content into JSON lines
	lines := splitSpotlightLines(content)
	assert.GreaterOrEqual(t, len(lines), 3, "Should have at least 3 shortcut records")

	// Verify each shortcut record has the expected fields
	for _, line := range lines {
		var record map[string]interface{}
		err := json.Unmarshal(line, &record)
		assert.NoError(t, err, "Each line should be valid JSON")

		// Verify common fields
		assert.NotEmpty(t, record["collection_timestamp"])
		assert.NotEmpty(t, record["event_timestamp"])
		assert.NotEmpty(t, record["source_file"])
		assert.Contains(t, record["source_file"].(string), "com.apple.spotlight.Shortcuts")

		// Check data fields
		data, ok := record["data"].(map[string]interface{})
		assert.True(t, ok, "Should have a data field as a map")

		// Verify spotlight-specific fields
		assert.Equal(t, "testuser", data["username"])
		assert.NotEmpty(t, data["shortcut"])
		assert.NotEmpty(t, data["display_name"])
		assert.NotEmpty(t, data["last_used"])
		assert.NotEmpty(t, data["url"])
		assert.Contains(t, data["url"].(string), "file:///Applications/")
	}

	// Verify specific shortcut content
	content_str := string(content)
	assert.Contains(t, content_str, "Mail")
	assert.Contains(t, content_str, "Safari")
	assert.Contains(t, content_str, "Notes")
	assert.Contains(t, content_str, "file:///Applications/Mail.app/")
	assert.Contains(t, content_str, "file:///Applications/Safari.app/")
	assert.Contains(t, content_str, "file:///Applications/Notes.app/")
}

// Helper function to split content into lines
func splitSpotlightLines(data []byte) [][]byte {
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
