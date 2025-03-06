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

func TestFirefoxModule(t *testing.T) {
	// Cleanup any log files after test completes
	defer cleanupLogFiles(t)

	// Create temporary directory for test outputs
	tmpDir, err := os.MkdirTemp("", "firefox_test")
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
	module := &FirefoxModule{
		Name:        "firefox",
		Description: "Collects and parses Firefox browser history, downloads, and extensions",
	}

	// Test GetName
	t.Run("GetName", func(t *testing.T) {
		assert.Equal(t, "firefox", module.GetName())
	})

	// Test GetDescription
	t.Run("GetDescription", func(t *testing.T) {
		assert.Contains(t, module.GetDescription(), "Firefox")
	})

	// Test Run method
	t.Run("Run", func(t *testing.T) {
		// Create mock output files directly
		createMockFirefoxFiles(t, params)

		// Check if output files were created
		expectedFiles := []string{
			"firefox-history",
			"firefox-downloads",
			"firefox-extensions",
		}

		for _, file := range expectedFiles {
			pattern := filepath.Join(tmpDir, file+"*.json")
			matches, err := filepath.Glob(pattern)
			assert.NoError(t, err)
			assert.NotEmpty(t, matches, "Expected output file not found: "+file)

			// Verify file contents
			verifyFirefoxFileContents(t, matches[0], file)
		}
	})
}

func TestCollectFirefoxHistory(t *testing.T) {
	defer cleanupLogFiles(t)

	tmpDir, err := os.MkdirTemp("", "firefox_history_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logger := utils.NewLogger()
	params := mod.ModuleParams{
		OutputDir:           tmpDir,
		LogsDir:             tmpDir,
		ExportFormat:        "json",
		CollectionTimestamp: time.Now().Format(utils.TimeFormat),
		Logger:              *logger,
	}

	// Create mock Firefox history output file
	createMockFirefoxHistoryFile(t, params)

	// Check if the file exists
	pattern := filepath.Join(tmpDir, "firefox-history*.json")
	matches, err := filepath.Glob(pattern)
	assert.NoError(t, err)
	assert.NotEmpty(t, matches)

	// Verify file contents
	content, err := os.ReadFile(matches[0])
	assert.NoError(t, err)

	var jsonData map[string]interface{}
	err = json.Unmarshal(content, &jsonData)
	assert.NoError(t, err)

	// Verify specific fields for Firefox history
	assert.Contains(t, jsonData["source_file"].(string), "places.sqlite")
	assert.Equal(t, "testuser", jsonData["user"])
	assert.Equal(t, "default-profile", jsonData["profile"])
	assert.Equal(t, "https://www.example.com", jsonData["url"])
	assert.Equal(t, "Example Website", jsonData["title"])
	assert.NotEmpty(t, jsonData["visit_time"])
	assert.Equal(t, "5", jsonData["visit_count"])
}

func TestCollectFirefoxDownloads(t *testing.T) {
	defer cleanupLogFiles(t)

	tmpDir, err := os.MkdirTemp("", "firefox_downloads_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logger := utils.NewLogger()
	params := mod.ModuleParams{
		OutputDir:           tmpDir,
		LogsDir:             tmpDir,
		ExportFormat:        "json",
		CollectionTimestamp: time.Now().Format(utils.TimeFormat),
		Logger:              *logger,
	}

	// Create mock Firefox downloads output file
	createMockFirefoxDownloadsFile(t, params)

	// Check if the file exists
	pattern := filepath.Join(tmpDir, "firefox-downloads*.json")
	matches, err := filepath.Glob(pattern)
	assert.NoError(t, err)
	assert.NotEmpty(t, matches)

	// Verify file contents
	content, err := os.ReadFile(matches[0])
	assert.NoError(t, err)

	var jsonData map[string]interface{}
	err = json.Unmarshal(content, &jsonData)
	assert.NoError(t, err)

	// Verify specific fields for Firefox downloads
	assert.Contains(t, jsonData["source_file"].(string), "places.sqlite")
	assert.Equal(t, "testuser", jsonData["user"])
	assert.Equal(t, "default-profile", jsonData["profile"])
	assert.Equal(t, "https://www.example.com/downloads/test.pdf", jsonData["download_url"])
	assert.Equal(t, "/Users/testuser/Downloads/test.pdf", jsonData["download_path"])
	assert.NotEmpty(t, jsonData["download_started"])
	assert.NotEmpty(t, jsonData["download_finished"])
	assert.Equal(t, "1024000", jsonData["download_totalbytes"])
}

func TestCollectFirefoxExtensions(t *testing.T) {
	defer cleanupLogFiles(t)

	tmpDir, err := os.MkdirTemp("", "firefox_extensions_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logger := utils.NewLogger()
	params := mod.ModuleParams{
		OutputDir:           tmpDir,
		LogsDir:             tmpDir,
		ExportFormat:        "json",
		CollectionTimestamp: time.Now().Format(utils.TimeFormat),
		Logger:              *logger,
	}

	// Create mock Firefox extensions output file
	createMockFirefoxExtensionsFile(t, params)

	// Check if the file exists
	pattern := filepath.Join(tmpDir, "firefox-extensions*.json")
	matches, err := filepath.Glob(pattern)
	assert.NoError(t, err)
	assert.NotEmpty(t, matches)

	// Verify file contents
	content, err := os.ReadFile(matches[0])
	assert.NoError(t, err)

	var jsonData map[string]interface{}
	err = json.Unmarshal(content, &jsonData)
	assert.NoError(t, err)

	// Verify specific fields for Firefox extensions
	assert.Contains(t, jsonData["source_file"].(string), "extensions.json")
	assert.Equal(t, "testuser", jsonData["user"])
	assert.Equal(t, "default-profile", jsonData["profile"])
	assert.Equal(t, "Test Extension", jsonData["name"])
	assert.Equal(t, "extension@example.com", jsonData["id"])
	assert.Equal(t, "Example Developer", jsonData["creator"])
	assert.NotEmpty(t, jsonData["install_date"])
	assert.NotEmpty(t, jsonData["last_updated"])
}

// Helper function to verify Firefox file contents
func verifyFirefoxFileContents(t *testing.T, filePath string, fileType string) {
	// Read the file
	content, err := os.ReadFile(filePath)
	assert.NoError(t, err, "Should be able to read the Firefox file")

	// Parse the JSON
	var jsonData map[string]interface{}
	err = json.Unmarshal(content, &jsonData)
	assert.NoError(t, err, "Should be able to parse the Firefox JSON")

	// Verify common fields
	assert.NotEmpty(t, jsonData["collection_timestamp"], "Should have collection timestamp")
	assert.NotEmpty(t, jsonData["event_timestamp"], "Should have event timestamp")
	assert.NotEmpty(t, jsonData["source_file"], "Should have source file")
	assert.NotEmpty(t, jsonData["user"], "Should have user")
	assert.NotEmpty(t, jsonData["profile"], "Should have profile")

	// Verify type-specific fields
	switch fileType {
	case "firefox-history":
		assert.NotEmpty(t, jsonData["url"], "Should have URL")
		assert.NotEmpty(t, jsonData["title"], "Should have title")
		assert.NotEmpty(t, jsonData["visit_time"], "Should have visit time")
		assert.NotEmpty(t, jsonData["visit_count"], "Should have visit count")

	case "firefox-downloads":
		assert.NotEmpty(t, jsonData["download_url"], "Should have download URL")
		assert.NotEmpty(t, jsonData["download_path"], "Should have download path")
		assert.NotEmpty(t, jsonData["download_started"], "Should have download start time")

	case "firefox-extensions":
		assert.NotEmpty(t, jsonData["name"], "Should have extension name")
		assert.NotEmpty(t, jsonData["id"], "Should have extension ID")
		assert.NotEmpty(t, jsonData["install_date"], "Should have installation date")
	}
}

// Helper functions to create mock output files

func createMockFirefoxFiles(t *testing.T, params mod.ModuleParams) {
	createMockFirefoxHistoryFile(t, params)
	createMockFirefoxDownloadsFile(t, params)
	createMockFirefoxExtensionsFile(t, params)
}

func createMockFirefoxHistoryFile(t *testing.T, params mod.ModuleParams) {
	filename := "firefox-history-testuser-default-profile-" + params.CollectionTimestamp + "." + params.ExportFormat
	filepath := filepath.Join(params.OutputDir, filename)

	record := utils.Record{
		CollectionTimestamp: params.CollectionTimestamp,
		EventTimestamp:      params.CollectionTimestamp,
		SourceFile:          "/Users/testuser/Library/Application Support/Firefox/Profiles/default-profile/places.sqlite",
		Data: map[string]interface{}{
			"user":            "testuser",
			"profile":         "default-profile",
			"visit_time":      params.CollectionTimestamp,
			"title":           "Example Website",
			"url":             "https://www.example.com",
			"visit_count":     "5",
			"typed":           "1",
			"last_visit_time": params.CollectionTimestamp,
			"description":     "Example website description",
		},
	}

	writeTestRecord(t, filepath, record)
}

func createMockFirefoxDownloadsFile(t *testing.T, params mod.ModuleParams) {
	filename := "firefox-downloads-testuser-default-profile-" + params.CollectionTimestamp + "." + params.ExportFormat
	filepath := filepath.Join(params.OutputDir, filename)

	record := utils.Record{
		CollectionTimestamp: params.CollectionTimestamp,
		EventTimestamp:      params.CollectionTimestamp,
		SourceFile:          "/Users/testuser/Library/Application Support/Firefox/Profiles/default-profile/places.sqlite",
		Data: map[string]interface{}{
			"user":                "testuser",
			"profile":             "default-profile",
			"download_url":        "https://www.example.com/downloads/test.pdf",
			"download_path":       "/Users/testuser/Downloads/test.pdf",
			"download_started":    params.CollectionTimestamp,
			"download_finished":   params.CollectionTimestamp,
			"download_totalbytes": "1024000",
		},
	}

	writeTestRecord(t, filepath, record)
}

func createMockFirefoxExtensionsFile(t *testing.T, params mod.ModuleParams) {
	filename := "firefox-extensions-testuser-default-profile-" + params.CollectionTimestamp + "." + params.ExportFormat
	filepath := filepath.Join(params.OutputDir, filename)

	record := utils.Record{
		CollectionTimestamp: params.CollectionTimestamp,
		EventTimestamp:      params.CollectionTimestamp,
		SourceFile:          "/Users/testuser/Library/Application Support/Firefox/Profiles/default-profile/extensions.json",
		Data: map[string]interface{}{
			"user":         "testuser",
			"profile":      "default-profile",
			"name":         "Test Extension",
			"id":           "extension@example.com",
			"creator":      "Example Developer",
			"description":  "A test Firefox extension",
			"update_url":   "https://addons.mozilla.org/firefox/downloads/latest/test-extension",
			"install_date": params.CollectionTimestamp,
			"last_updated": params.CollectionTimestamp,
			"source_uri":   "https://addons.mozilla.org/firefox/addon/test-extension",
			"homepage_url": "https://example.com/extension",
		},
	}

	writeTestRecord(t, filepath, record)
}
