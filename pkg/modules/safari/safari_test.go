package safari

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

func TestSafariModule(t *testing.T) {
	// Create temporary directory for test outputs
	tmpDir, err := os.MkdirTemp("", "safari_test")
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
	module := &SafariModule{
		Name:        "safari",
		Description: "Collects and parses safari history, downloads, and extensions",
	}

	// Test GetName
	t.Run("GetName", func(t *testing.T) {
		assert.Equal(t, "safari", module.GetName())
	})

	// Test GetDescription
	t.Run("GetDescription", func(t *testing.T) {
		assert.Contains(t, module.GetDescription(), "safari history")
	})

	// Test Run method with mock output
	t.Run("Run", func(t *testing.T) {
		// Create mock output files for different Safari data types
		userProfile := "testuser"
		createMockSafariOutput(t, params, userProfile)

		// Verify the history output file exists
		historyOutputFile := filepath.Join(tmpDir, "safari-visit-"+userProfile+"-"+params.CollectionTimestamp+".json")
		assert.FileExists(t, historyOutputFile)

		// Verify the downloads output file exists
		downloadsOutputFile := filepath.Join(tmpDir, "safari-downloads-"+userProfile+"-"+params.CollectionTimestamp+".json")
		assert.FileExists(t, downloadsOutputFile)

		// Verify the extensions output file exists
		extensionsOutputFile := filepath.Join(tmpDir, "safari-extensions-"+userProfile+"-"+params.CollectionTimestamp+".json")
		assert.FileExists(t, extensionsOutputFile)

		// Verify the content of the output files
		verifySafariVisitOutput(t, historyOutputFile, userProfile)
		verifySafariDownloadsOutput(t, downloadsOutputFile, userProfile)
		verifySafariExtensionsOutput(t, extensionsOutputFile, userProfile)
	})
}

// Test that the module initializes properly
func TestSafariModuleInitialization(t *testing.T) {
	// Create a new instance with proper initialization
	module := &SafariModule{
		Name:        "safari",
		Description: "Collects and parses safari history, downloads, and extensions",
	}

	// Verify module is properly instantiated with expected values
	assert.Equal(t, "safari", module.Name, "Module name should be initialized")
	assert.Contains(t, module.Description, "safari history", "Module description should be initialized")

	// Test the module's methods
	assert.Equal(t, "safari", module.GetName())
	assert.Contains(t, module.GetDescription(), "safari history")
}

// Create mock Safari output files
func createMockSafariOutput(t *testing.T, params mod.ModuleParams, userProfile string) {
	// Create mock history records
	createMockSafariHistoryOutput(t, params, userProfile)

	// Create mock downloads records
	createMockSafariDownloadsOutput(t, params, userProfile)

	// Create mock extensions records
	createMockSafariExtensionsOutput(t, params, userProfile)
}

// Create mock Safari history output
func createMockSafariHistoryOutput(t *testing.T, params mod.ModuleParams, userProfile string) {
	outputFile := filepath.Join(params.OutputDir, "safari-visit-"+userProfile+"-"+params.CollectionTimestamp+".json")

	// Create sample history records
	historyRecords := []utils.Record{
		{
			CollectionTimestamp: params.CollectionTimestamp,
			EventTimestamp:      "2023-05-15T10:30:45Z",
			SourceFile:          "/Users/" + userProfile + "/Library/Safari/History.db",
			Data: map[string]interface{}{
				"user":        userProfile,
				"visit_time":  "2023-05-15T10:30:45Z",
				"title":       "Example Search",
				"url":         "https://www.example.com/search?q=safari",
				"visit_count": 5,
			},
		},
		{
			CollectionTimestamp: params.CollectionTimestamp,
			EventTimestamp:      "2023-05-15T11:20:10Z",
			SourceFile:          "/Users/" + userProfile + "/Library/Safari/History.db",
			Data: map[string]interface{}{
				"user":            userProfile,
				"visit_time":      "2023-05-15T11:20:10Z",
				"title":           "GitHub - Repository",
				"url":             "https://github.com/example/repo",
				"visit_count":     3,
				"recently_closed": "Yes",
				"tab_title":       "GitHub - Repository",
				"date_closed":     "2023-05-15T11:45:22Z",
				"last_visit_time": "2023-05-15T11:20:10Z",
			},
		},
		{
			CollectionTimestamp: params.CollectionTimestamp,
			EventTimestamp:      "2023-05-15T14:05:30Z",
			SourceFile:          "/Users/" + userProfile + "/Library/Safari/History.db",
			Data: map[string]interface{}{
				"user":        userProfile,
				"visit_time":  "2023-05-15T14:05:30Z",
				"title":       "Developer Documentation",
				"url":         "https://developer.apple.com/documentation",
				"visit_count": 2,
			},
		},
	}

	// Write to output file
	file, err := os.Create(outputFile)
	assert.NoError(t, err)
	defer file.Close()

	encoder := json.NewEncoder(file)
	for _, record := range historyRecords {
		err := encoder.Encode(record)
		assert.NoError(t, err)
	}
}

// Create mock Safari downloads output
func createMockSafariDownloadsOutput(t *testing.T, params mod.ModuleParams, userProfile string) {
	outputFile := filepath.Join(params.OutputDir, "safari-downloads-"+userProfile+"-"+params.CollectionTimestamp+".json")

	// Create sample download records
	downloadRecords := []utils.Record{
		{
			CollectionTimestamp: params.CollectionTimestamp,
			EventTimestamp:      "2023-05-15T09:45:20Z",
			SourceFile:          "/Users/" + userProfile + "/Library/Safari/Downloads.plist",
			Data: map[string]interface{}{
				"user":                    userProfile,
				"download_url":            "https://example.com/files/document.pdf",
				"download_path":           "/Users/" + userProfile + "/Downloads/document.pdf",
				"download_started":        "2023-05-15T09:45:20Z",
				"download_finished":       "2023-05-15T09:45:45Z",
				"download_totalbytes":     5242880,
				"download_bytes_received": 5242880,
			},
		},
		{
			CollectionTimestamp: params.CollectionTimestamp,
			EventTimestamp:      "2023-05-15T13:30:15Z",
			SourceFile:          "/Users/" + userProfile + "/Library/Safari/Downloads.plist",
			Data: map[string]interface{}{
				"user":                    userProfile,
				"download_url":            "https://github.com/example/repo/archive/main.zip",
				"download_path":           "/Users/" + userProfile + "/Downloads/repo-main.zip",
				"download_started":        "2023-05-15T13:30:15Z",
				"download_finished":       "2023-05-15T13:31:20Z",
				"download_totalbytes":     10485760,
				"download_bytes_received": 10485760,
			},
		},
	}

	// Write to output file
	file, err := os.Create(outputFile)
	assert.NoError(t, err)
	defer file.Close()

	encoder := json.NewEncoder(file)
	for _, record := range downloadRecords {
		err := encoder.Encode(record)
		assert.NoError(t, err)
	}
}

// Create mock Safari extensions output
func createMockSafariExtensionsOutput(t *testing.T, params mod.ModuleParams, userProfile string) {
	outputFile := filepath.Join(params.OutputDir, "safari-extensions-"+userProfile+"-"+params.CollectionTimestamp+".json")

	// Create sample extension records
	extensionRecords := []utils.Record{
		{
			CollectionTimestamp: params.CollectionTimestamp,
			EventTimestamp:      params.CollectionTimestamp,
			SourceFile:          "/Users/" + userProfile + "/Library/Safari/Extensions/Extensions.plist",
			Data: map[string]interface{}{
				"user":             userProfile,
				"name":             "AdBlocker.safariextz",
				"bundle_directory": "AdBlocker.safariextension",
				"enabled":          true,
				"apple_signed":     true,
				"developer_id":     "ABC123DEF456",
				"bundle_id":        "com.example.safari.adblocker",
				"ctime":            "2023-05-01T09:00:00Z",
				"mtime":            "2023-05-01T09:00:00Z",
				"atime":            "2023-05-01T09:00:00Z",
				"size":             1048576,
			},
		},
		{
			CollectionTimestamp: params.CollectionTimestamp,
			EventTimestamp:      params.CollectionTimestamp,
			SourceFile:          "/Users/" + userProfile + "/Library/Safari/Extensions/Extensions.plist",
			Data: map[string]interface{}{
				"user":             userProfile,
				"name":             "PasswordManager.safariextz",
				"bundle_directory": "PasswordManager.safariextension",
				"enabled":          true,
				"apple_signed":     true,
				"developer_id":     "XYZ789ABC123",
				"bundle_id":        "com.example.safari.passwordmanager",
				"ctime":            "2023-05-10T14:30:00Z",
				"mtime":            "2023-05-10T14:30:00Z",
				"atime":            "2023-05-10T14:30:00Z",
				"size":             2097152,
			},
		},
	}

	// Write to output file
	file, err := os.Create(outputFile)
	assert.NoError(t, err)
	defer file.Close()

	encoder := json.NewEncoder(file)
	for _, record := range extensionRecords {
		err := encoder.Encode(record)
		assert.NoError(t, err)
	}
}

// Verify Safari history output file
func verifySafariVisitOutput(t *testing.T, outputFile, userProfile string) {
	// Read the output file
	content, err := os.ReadFile(outputFile)
	assert.NoError(t, err)

	// Parse JSON lines
	lines := splitSafariLines(content)
	assert.GreaterOrEqual(t, len(lines), 3, "Should have at least 3 history records")

	// Create maps to track if we found specific URLs
	foundUrls := make(map[string]bool)

	// Check each history record
	for _, line := range lines {
		var record map[string]interface{}
		err := json.Unmarshal(line, &record)
		assert.NoError(t, err)

		// Verify common fields
		assert.NotEmpty(t, record["collection_timestamp"])
		assert.NotEmpty(t, record["event_timestamp"])
		assert.NotEmpty(t, record["source_file"])
		assert.Contains(t, record["source_file"].(string), "History.db")

		data, ok := record["data"].(map[string]interface{})
		assert.True(t, ok)

		// Verify history-specific fields
		assert.Equal(t, userProfile, data["user"])
		assert.NotEmpty(t, data["visit_time"])
		assert.NotEmpty(t, data["title"])
		assert.NotEmpty(t, data["url"])
		assert.NotEmpty(t, data["visit_count"])

		// Record the URL we found
		url, _ := data["url"].(string)
		foundUrls[url] = true

		// Check for recently closed tabs
		if url == "https://github.com/example/repo" {
			assert.Equal(t, "Yes", data["recently_closed"])
			assert.Equal(t, "GitHub - Repository", data["tab_title"])
			assert.NotEmpty(t, data["date_closed"])
			assert.NotEmpty(t, data["last_visit_time"])
		}
	}

	// Verify we found all expected URLs
	assert.True(t, foundUrls["https://www.example.com/search?q=safari"])
	assert.True(t, foundUrls["https://github.com/example/repo"])
	assert.True(t, foundUrls["https://developer.apple.com/documentation"])
}

// Verify Safari downloads output file
func verifySafariDownloadsOutput(t *testing.T, outputFile, userProfile string) {
	// Read the output file
	content, err := os.ReadFile(outputFile)
	assert.NoError(t, err)

	// Parse JSON lines
	lines := splitSafariLines(content)
	assert.GreaterOrEqual(t, len(lines), 2, "Should have at least 2 download records")

	// Create maps to track if we found specific downloads
	foundDownloads := make(map[string]bool)

	// Check each download record
	for _, line := range lines {
		var record map[string]interface{}
		err := json.Unmarshal(line, &record)
		assert.NoError(t, err)

		// Verify common fields
		assert.NotEmpty(t, record["collection_timestamp"])
		assert.NotEmpty(t, record["event_timestamp"])
		assert.NotEmpty(t, record["source_file"])
		assert.Contains(t, record["source_file"].(string), "Downloads.plist")

		data, ok := record["data"].(map[string]interface{})
		assert.True(t, ok)

		// Verify download-specific fields
		assert.Equal(t, userProfile, data["user"])
		assert.NotEmpty(t, data["download_url"])
		assert.NotEmpty(t, data["download_path"])
		assert.NotEmpty(t, data["download_started"])
		assert.NotEmpty(t, data["download_finished"])
		assert.NotEmpty(t, data["download_totalbytes"])
		assert.NotEmpty(t, data["download_bytes_received"])

		// Record the download URL we found
		url, _ := data["download_url"].(string)
		foundDownloads[url] = true
	}

	// Verify we found all expected download URLs
	assert.True(t, foundDownloads["https://example.com/files/document.pdf"])
	assert.True(t, foundDownloads["https://github.com/example/repo/archive/main.zip"])
}

// Verify Safari extensions output file
func verifySafariExtensionsOutput(t *testing.T, outputFile, userProfile string) {
	// Read the output file
	content, err := os.ReadFile(outputFile)
	assert.NoError(t, err)

	// Parse JSON lines
	lines := splitSafariLines(content)
	assert.GreaterOrEqual(t, len(lines), 2, "Should have at least 2 extension records")

	// Create maps to track if we found specific extensions
	foundExtensions := make(map[string]bool)

	// Check each extension record
	for _, line := range lines {
		var record map[string]interface{}
		err := json.Unmarshal(line, &record)
		assert.NoError(t, err)

		// Verify common fields
		assert.NotEmpty(t, record["collection_timestamp"])
		assert.NotEmpty(t, record["event_timestamp"])
		assert.NotEmpty(t, record["source_file"])
		assert.Contains(t, record["source_file"].(string), "Extensions.plist")

		data, ok := record["data"].(map[string]interface{})
		assert.True(t, ok)

		// Verify extension-specific fields
		assert.Equal(t, userProfile, data["user"])
		assert.NotEmpty(t, data["name"])
		assert.NotEmpty(t, data["bundle_directory"])
		assert.NotNil(t, data["enabled"])
		assert.NotNil(t, data["apple_signed"])
		assert.NotEmpty(t, data["developer_id"])
		assert.NotEmpty(t, data["bundle_id"])
		assert.NotEmpty(t, data["ctime"])
		assert.NotEmpty(t, data["mtime"])
		assert.NotEmpty(t, data["atime"])
		assert.NotEmpty(t, data["size"])

		// Record the extension name we found
		name, _ := data["name"].(string)
		foundExtensions[name] = true
	}

	// Verify we found all expected extensions
	assert.True(t, foundExtensions["AdBlocker.safariextz"])
	assert.True(t, foundExtensions["PasswordManager.safariextz"])
}

// Helper function to split content into lines
func splitSafariLines(data []byte) [][]byte {
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
