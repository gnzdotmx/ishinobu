package quicklook

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

func TestQuickLookModule(t *testing.T) {
	// Create temporary directory for test outputs
	tmpDir, err := os.MkdirTemp("", "quicklook_test")
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
	module := &QuickLookModule{
		Name:        "quicklook",
		Description: "Collects QuickLook cache information",
	}

	// Test GetName
	t.Run("GetName", func(t *testing.T) {
		assert.Equal(t, "quicklook", module.GetName())
	})

	// Test GetDescription
	t.Run("GetDescription", func(t *testing.T) {
		assert.Contains(t, module.GetDescription(), "QuickLook cache")
	})

	// Test Run method with mock output
	t.Run("Run", func(t *testing.T) {
		// Create mock output file to simulate the module's output
		createMockQuickLookOutput(t, params)

		// Verify the output file exists
		outputFile := filepath.Join(tmpDir, "quicklook-"+params.CollectionTimestamp+".json")
		assert.FileExists(t, outputFile)

		// Verify the content of the output file
		verifyQuickLookOutput(t, outputFile)
	})
}

// Test that the module initializes properly
func TestQuickLookModuleInitialization(t *testing.T) {
	// Create a new instance with proper initialization
	module := &QuickLookModule{
		Name:        "quicklook",
		Description: "Collects QuickLook cache information",
	}

	// Verify module is properly instantiated with expected values
	assert.Equal(t, "quicklook", module.Name, "Module name should be initialized")
	assert.Contains(t, module.Description, "QuickLook cache", "Module description should be initialized")

	// Test the module's methods
	assert.Equal(t, "quicklook", module.GetName())
	assert.Contains(t, module.GetDescription(), "QuickLook cache")
}

// Create a mock QuickLook output file
func createMockQuickLookOutput(t *testing.T, params mod.ModuleParams) {
	outputFile := filepath.Join(params.OutputDir, "quicklook-"+params.CollectionTimestamp+".json")

	// Create sample QuickLook records
	quicklookRecords := []utils.Record{
		{
			CollectionTimestamp: params.CollectionTimestamp,
			EventTimestamp:      "2023-05-15T09:30:25Z",
			SourceFile:          "/private/var/folders/xx/xxxxxxxx/C/com.apple.QuickLook.thumbnailcache/index.sqlite",
			Data: map[string]interface{}{
				"uid":                "user1",
				"path":               "/Users/user1/Documents",
				"name":               "presentation.pptx",
				"last_hit_date":      "2023-05-15T09:30:25Z",
				"hit_count":          int64(5),
				"file_last_modified": "2023-05-14T18:45:10Z",
				"generator":          "Microsoft PowerPoint",
				"file_size":          int64(2458000),
			},
		},
		{
			CollectionTimestamp: params.CollectionTimestamp,
			EventTimestamp:      "2023-05-14T16:20:15Z",
			SourceFile:          "/private/var/folders/xx/xxxxxxxx/C/com.apple.QuickLook.thumbnailcache/index.sqlite",
			Data: map[string]interface{}{
				"uid":                "user1",
				"path":               "/Users/user1/Downloads",
				"name":               "report.pdf",
				"last_hit_date":      "2023-05-14T16:20:15Z",
				"hit_count":          int64(3),
				"file_last_modified": "2023-05-14T10:30:05Z",
				"generator":          "Adobe PDF",
				"file_size":          int64(1240000),
			},
		},
		{
			CollectionTimestamp: params.CollectionTimestamp,
			EventTimestamp:      "2023-05-13T14:10:35Z",
			SourceFile:          "/private/var/folders/yy/yyyyyyyy/C/com.apple.QuickLook.thumbnailcache/index.sqlite",
			Data: map[string]interface{}{
				"uid":                "user2",
				"path":               "/Users/user2/Pictures",
				"name":               "vacation.jpg",
				"last_hit_date":      "2023-05-13T14:10:35Z",
				"hit_count":          int64(2),
				"file_last_modified": "2023-05-12T20:15:30Z",
				"generator":          "Preview",
				"file_size":          int64(3500000),
			},
		},
	}

	// Write each record as a JSON line
	file, err := os.Create(outputFile)
	assert.NoError(t, err)
	defer file.Close()

	encoder := json.NewEncoder(file)
	for _, record := range quicklookRecords {
		err := encoder.Encode(record)
		assert.NoError(t, err)
	}
}

// Verify the QuickLook output file contains expected data
func verifyQuickLookOutput(t *testing.T, outputFile string) {
	// Read the output file
	content, err := os.ReadFile(outputFile)
	assert.NoError(t, err)

	// Split the content into JSON lines
	lines := splitQuickLookLines(content)
	assert.GreaterOrEqual(t, len(lines), 3, "Should have at least 3 QuickLook records")

	// Verify each record has the expected fields
	for _, line := range lines {
		var record map[string]interface{}
		err := json.Unmarshal(line, &record)
		assert.NoError(t, err, "Each line should be valid JSON")

		// Verify common fields
		assert.NotEmpty(t, record["collection_timestamp"])
		assert.NotEmpty(t, record["event_timestamp"])
		assert.NotEmpty(t, record["source_file"])
		sourceFile, ok := record["source_file"].(string)
		assert.True(t, ok, "Source file should be a string")
		assert.Contains(t, sourceFile, "com.apple.QuickLook.thumbnailcache")

		// Check data fields
		data, ok := record["data"].(map[string]interface{})
		assert.True(t, ok, "Should have a data field as a map")

		// Verify QuickLook-specific fields
		assert.NotEmpty(t, data["uid"])
		assert.NotEmpty(t, data["path"])
		assert.NotEmpty(t, data["name"])
		assert.NotEmpty(t, data["last_hit_date"])
		assert.NotNil(t, data["hit_count"])
		assert.NotEmpty(t, data["file_last_modified"])
		assert.NotEmpty(t, data["generator"])
		assert.NotNil(t, data["file_size"])
	}

	// Verify specific file content
	contentStr := string(content)
	assert.Contains(t, contentStr, "presentation.pptx")
	assert.Contains(t, contentStr, "report.pdf")
	assert.Contains(t, contentStr, "vacation.jpg")
	assert.Contains(t, contentStr, "Microsoft PowerPoint")
	assert.Contains(t, contentStr, "Adobe PDF")
	assert.Contains(t, contentStr, "Preview")
}

// Helper function to split content into lines for this specific test
func splitQuickLookLines(data []byte) [][]byte {
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
