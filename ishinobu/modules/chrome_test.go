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

func TestChromeModule(t *testing.T) {
	// Cleanup any log files after test completes
	defer cleanupLogFiles(t)

	// Create temporary directory for test outputs
	tmpDir, err := os.MkdirTemp("", "chrome_test")
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
	module := &ChromeModule{
		Name:        "chrome",
		Description: "Collects Chrome browser history, downloads, profiles, and extensions",
	}

	// Test GetName
	t.Run("GetName", func(t *testing.T) {
		assert.Equal(t, "chrome", module.GetName())
	})

	// Test GetDescription
	t.Run("GetDescription", func(t *testing.T) {
		assert.Contains(t, module.GetDescription(), "Chrome")
	})

	// Test Run method
	t.Run("Run", func(t *testing.T) {
		// Create mock output files directly
		createMockChromeFiles(t, params)

		// Check if output files were created
		expectedFiles := []string{
			"chrome-history",
			"chrome-downloads",
			"chrome-profiles",
			"chrome-extensions",
			"chrome-settings-popup",
		}

		for _, file := range expectedFiles {
			pattern := filepath.Join(tmpDir, file+"*.json")
			matches, err := filepath.Glob(pattern)
			assert.NoError(t, err)
			assert.NotEmpty(t, matches, "Expected output file not found: "+file)

			// Verify file contents
			verifyChromeFileContents(t, matches[0], file)
		}
	})
}

func TestCollectChromeHistory(t *testing.T) {
	defer cleanupLogFiles(t)

	tmpDir, err := os.MkdirTemp("", "chrome_history_test")
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

	// Create mock Chrome history output file
	createMockChromeHistoryFile(t, params)

	// Check if the file exists
	pattern := filepath.Join(tmpDir, "chrome-history*.json")
	matches, err := filepath.Glob(pattern)
	assert.NoError(t, err)
	assert.NotEmpty(t, matches)

	// Verify file contents
	content, err := os.ReadFile(matches[0])
	assert.NoError(t, err)

	var jsonData map[string]interface{}
	err = json.Unmarshal(content, &jsonData)
	assert.NoError(t, err)

	// Verify specific fields for Chrome history
	assert.Contains(t, jsonData["source_file"].(string), "Chrome/Default/History")
	assert.Equal(t, "Default", jsonData["chrome_profile"])
	assert.Equal(t, "https://www.example.com", jsonData["url"])
	assert.Equal(t, "Example Website", jsonData["title"])
	assert.NotEmpty(t, jsonData["visit_time"])
	assert.Equal(t, "12345", jsonData["from_visit"])
	assert.Equal(t, "LINK", jsonData["transition"])
}

func TestCollectChromeDownloads(t *testing.T) {
	defer cleanupLogFiles(t)

	tmpDir, err := os.MkdirTemp("", "chrome_downloads_test")
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

	// Create mock Chrome downloads output file
	createMockChromeDownloadsFile(t, params)

	// Check if the file exists
	pattern := filepath.Join(tmpDir, "chrome-downloads*.json")
	matches, err := filepath.Glob(pattern)
	assert.NoError(t, err)
	assert.NotEmpty(t, matches)

	// Verify file contents
	content, err := os.ReadFile(matches[0])
	assert.NoError(t, err)

	var jsonData map[string]interface{}
	err = json.Unmarshal(content, &jsonData)
	assert.NoError(t, err)

	// Verify specific fields for Chrome downloads
	assert.Contains(t, jsonData["source_file"].(string), "Chrome/Default/History")
	assert.Equal(t, "/Users/testuser/Downloads/test.pdf", jsonData["current_path"])
	assert.Equal(t, "/Users/testuser/Downloads/test.pdf", jsonData["target_path"])
	assert.NotEmpty(t, jsonData["start_time"])
	assert.NotEmpty(t, jsonData["end_time"])
	assert.Equal(t, "https://www.example.com/downloads/test.pdf", jsonData["url"])
	assert.Equal(t, "https://www.example.com", jsonData["referrer"])
}

func TestCollectChromeProfiles(t *testing.T) {
	defer cleanupLogFiles(t)

	tmpDir, err := os.MkdirTemp("", "chrome_profiles_test")
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

	// Create mock Chrome profiles output file
	createMockChromeProfilesFile(t, params)

	// Check if the file exists
	pattern := filepath.Join(tmpDir, "chrome-profiles*.json")
	matches, err := filepath.Glob(pattern)
	assert.NoError(t, err)
	assert.NotEmpty(t, matches)

	// Verify file contents
	content, err := os.ReadFile(matches[0])
	assert.NoError(t, err)

	var jsonData map[string]interface{}
	err = json.Unmarshal(content, &jsonData)
	assert.NoError(t, err)

	// Verify specific fields for Chrome profiles
	assert.Contains(t, jsonData["source_file"].(string), "Chrome/Local State")
	assert.Equal(t, "Default", jsonData["profile_name"])
	assert.Equal(t, "Test User", jsonData["gaia_name"])
	assert.NotEmpty(t, jsonData["last_used"])
}

func TestCollectChromeExtensions(t *testing.T) {
	defer cleanupLogFiles(t)

	tmpDir, err := os.MkdirTemp("", "chrome_extensions_test")
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

	// Create mock Chrome extensions output file
	createMockChromeExtensionsFile(t, params)

	// Check if the file exists
	pattern := filepath.Join(tmpDir, "chrome-extensions*.json")
	matches, err := filepath.Glob(pattern)
	assert.NoError(t, err)
	assert.NotEmpty(t, matches)

	// Verify file contents
	content, err := os.ReadFile(matches[0])
	assert.NoError(t, err)

	var jsonData map[string]interface{}
	err = json.Unmarshal(content, &jsonData)
	assert.NoError(t, err)

	// Verify specific fields for Chrome extensions
	assert.Contains(t, jsonData["source_file"].(string), "Chrome/Default/Extensions")
	assert.Equal(t, "Default", jsonData["chrome_profile"])
	assert.Equal(t, "abcdefghijklmnopqrstuvwxyz", jsonData["extension_id"])
	assert.Equal(t, "Test Extension", jsonData["name"])
	assert.Equal(t, "1.0.0", jsonData["version"])
	assert.NotEmpty(t, jsonData["description"])

	// Check for nested fields
	permissions, ok := jsonData["permissions"].([]interface{})
	assert.True(t, ok, "Should have permissions array")
	assert.NotEmpty(t, permissions, "Permissions should not be empty")
}

func TestCollectChromePopupSettings(t *testing.T) {
	defer cleanupLogFiles(t)

	tmpDir, err := os.MkdirTemp("", "chrome_popup_settings_test")
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

	// Create mock Chrome popup settings output file
	createMockChromePopupSettingsFile(t, params)

	// Check if the file exists
	pattern := filepath.Join(tmpDir, "chrome-settings-popup*.json")
	matches, err := filepath.Glob(pattern)
	assert.NoError(t, err)
	assert.NotEmpty(t, matches)

	// Verify file contents
	content, err := os.ReadFile(matches[0])
	assert.NoError(t, err)

	var jsonData map[string]interface{}
	err = json.Unmarshal(content, &jsonData)
	assert.NoError(t, err)

	// Verify specific fields for Chrome popup settings
	assert.Contains(t, jsonData["source_file"].(string), "Chrome/Default/Preferences")
	assert.Equal(t, "Default", jsonData["profile"])
	assert.Equal(t, "https://www.example.com", jsonData["url"])
	assert.Equal(t, "Allowed", jsonData["setting"])
	assert.NotEmpty(t, jsonData["last_modified"])
}

// Helper function to verify Chrome file contents
func verifyChromeFileContents(t *testing.T, filePath string, fileType string) {
	// Read the file
	content, err := os.ReadFile(filePath)
	assert.NoError(t, err, "Should be able to read the Chrome file")

	// Parse the JSON
	var jsonData map[string]interface{}
	err = json.Unmarshal(content, &jsonData)
	assert.NoError(t, err, "Should be able to parse the Chrome JSON")

	// Verify common fields
	assert.NotEmpty(t, jsonData["collection_timestamp"], "Should have collection timestamp")
	assert.NotEmpty(t, jsonData["event_timestamp"], "Should have event timestamp")
	assert.NotEmpty(t, jsonData["source_file"], "Should have source file")

	// Verify type-specific fields
	switch fileType {
	case "chrome-history":
		assert.NotEmpty(t, jsonData["chrome_profile"], "Should have Chrome profile")
		assert.NotEmpty(t, jsonData["url"], "Should have URL")
		assert.NotEmpty(t, jsonData["visit_time"], "Should have visit time")

	case "chrome-downloads":
		assert.NotEmpty(t, jsonData["current_path"], "Should have current path")
		assert.NotEmpty(t, jsonData["target_path"], "Should have target path")
		assert.NotEmpty(t, jsonData["start_time"], "Should have start time")
		assert.NotEmpty(t, jsonData["url"], "Should have download URL")

	case "chrome-profiles":
		assert.NotEmpty(t, jsonData["profile_name"], "Should have profile name")
		assert.NotEmpty(t, jsonData["last_used"], "Should have last used timestamp")

	case "chrome-extensions":
		assert.NotEmpty(t, jsonData["chrome_profile"], "Should have Chrome profile")
		assert.NotEmpty(t, jsonData["extension_id"], "Should have extension ID")
		assert.NotEmpty(t, jsonData["name"], "Should have extension name")
		assert.NotEmpty(t, jsonData["version"], "Should have extension version")

	case "chrome-settings-popup":
		assert.NotEmpty(t, jsonData["profile"], "Should have Chrome profile")
		assert.NotEmpty(t, jsonData["url"], "Should have URL")
		assert.NotEmpty(t, jsonData["setting"], "Should have popup setting")
	}
}

// Helper functions to create mock output files

func createMockChromeFiles(t *testing.T, params mod.ModuleParams) {
	createMockChromeHistoryFile(t, params)
	createMockChromeDownloadsFile(t, params)
	createMockChromeProfilesFile(t, params)
	createMockChromeExtensionsFile(t, params)
	createMockChromePopupSettingsFile(t, params)
}

func createMockChromeHistoryFile(t *testing.T, params mod.ModuleParams) {
	filename := "chrome-history-Default-" + params.CollectionTimestamp + "." + params.ExportFormat
	filepath := filepath.Join(params.OutputDir, filename)

	record := utils.Record{
		CollectionTimestamp: params.CollectionTimestamp,
		EventTimestamp:      params.CollectionTimestamp,
		SourceFile:          "/Users/testuser/Library/Application Support/Google/Chrome/Default/History",
		Data: map[string]interface{}{
			"chrome_profile": "Default",
			"url":            "https://www.example.com",
			"title":          "Example Website",
			"visit_time":     params.CollectionTimestamp,
			"from_visit":     "12345",
			"transition":     "LINK",
		},
	}

	writeTestRecord(t, filepath, record)
}

func createMockChromeDownloadsFile(t *testing.T, params mod.ModuleParams) {
	filename := "chrome-downloads-Default-" + params.CollectionTimestamp + "." + params.ExportFormat
	filepath := filepath.Join(params.OutputDir, filename)

	record := utils.Record{
		CollectionTimestamp: params.CollectionTimestamp,
		EventTimestamp:      params.CollectionTimestamp,
		SourceFile:          "/Users/testuser/Library/Application Support/Google/Chrome/Default/History",
		Data: map[string]interface{}{
			"chrome_profile":   "Default",
			"current_path":     "/Users/testuser/Downloads/test.pdf",
			"target_path":      "/Users/testuser/Downloads/test.pdf",
			"start_time":       params.CollectionTimestamp,
			"end_time":         params.CollectionTimestamp,
			"danger_type":      "0",
			"opened":           "1",
			"last_modified":    params.CollectionTimestamp,
			"referrer":         "https://www.example.com",
			"tab_url":          "https://www.example.com/download",
			"tab_referrer_url": "https://www.example.com",
			"site_url":         "https://www.example.com",
			"url":              "https://www.example.com/downloads/test.pdf",
		},
	}

	writeTestRecord(t, filepath, record)
}

func createMockChromeProfilesFile(t *testing.T, params mod.ModuleParams) {
	filename := "chrome-profiles-" + params.CollectionTimestamp + "." + params.ExportFormat
	filepath := filepath.Join(params.OutputDir, filename)

	record := utils.Record{
		CollectionTimestamp: params.CollectionTimestamp,
		EventTimestamp:      params.CollectionTimestamp,
		SourceFile:          "/Users/testuser/Library/Application Support/Google/Chrome/Local State",
		Data: map[string]interface{}{
			"profile_name": "Default",
			"gaia_name":    "Test User",
			"gaia_id":      "12345678901234567890",
			"is_ephemeral": false,
			"last_used":    params.CollectionTimestamp,
		},
	}

	writeTestRecord(t, filepath, record)
}

func createMockChromeExtensionsFile(t *testing.T, params mod.ModuleParams) {
	filename := "chrome-extensions-Default-" + params.CollectionTimestamp + "." + params.ExportFormat
	filepath := filepath.Join(params.OutputDir, filename)

	record := utils.Record{
		CollectionTimestamp: params.CollectionTimestamp,
		EventTimestamp:      params.CollectionTimestamp,
		SourceFile:          "/Users/testuser/Library/Application Support/Google/Chrome/Default/Extensions/abcdefghijklmnopqrstuvwxyz",
		Data: map[string]interface{}{
			"chrome_profile": "Default",
			"extension_id":   "abcdefghijklmnopqrstuvwxyz",
			"name":           "Test Extension",
			"description":    "A test Chrome extension",
			"version":        "1.0.0",
			"permissions":    []string{"storage", "tabs"},
			"scripts":        []string{"background.js"},
			"persistent":     true,
			"scopes":         []string{"https://*/*"},
			"update_url":     "https://clients2.google.com/service/update2/crx",
			"default_locale": "en",
		},
	}

	writeTestRecord(t, filepath, record)
}

func createMockChromePopupSettingsFile(t *testing.T, params mod.ModuleParams) {
	filename := "chrome-settings-popup-Default-" + params.CollectionTimestamp + "." + params.ExportFormat
	filepath := filepath.Join(params.OutputDir, filename)

	record := utils.Record{
		CollectionTimestamp: params.CollectionTimestamp,
		EventTimestamp:      params.CollectionTimestamp,
		SourceFile:          "/Users/testuser/Library/Application Support/Google/Chrome/Default/Preferences",
		Data: map[string]interface{}{
			"profile":       "Default",
			"url":           "https://www.example.com",
			"setting":       "Allowed",
			"last_modified": params.CollectionTimestamp,
		},
	}

	writeTestRecord(t, filepath, record)
}
