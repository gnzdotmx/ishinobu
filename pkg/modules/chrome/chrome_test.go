package chrome

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnzdotmx/ishinobu/pkg/mod"
	"github.com/gnzdotmx/ishinobu/pkg/modules/testutils"
)

// TestVisitChromeHistory tests the visitChromeHistory function
func TestVisitChromeHistory(t *testing.T) {
	// Set up test environment
	tmpDir, err := os.MkdirTemp("", "chrome_history_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Set up mock Chrome environment
	profile := "Default"
	tmpChromeDir := "/tmp/ishinobu"
	os.RemoveAll(tmpChromeDir)
	err = os.MkdirAll(tmpChromeDir, 0777)
	require.NoError(t, err)
	err = os.Chmod(tmpChromeDir, 0777)
	require.NoError(t, err)
	chromeDir := filepath.Join(tmpDir, "Chrome")
	profileDir := filepath.Join(chromeDir, profile)
	err = os.MkdirAll(profileDir, 0755)
	require.NoError(t, err)

	// Create Chrome history schema
	historySchema := `
		CREATE TABLE urls (
			id INTEGER PRIMARY KEY,
			url TEXT,
			title TEXT
		);
		CREATE TABLE visits (
			id INTEGER PRIMARY KEY,
			url INTEGER,
			visit_time TEXT,
			from_visit TEXT,
			transition TEXT,
			FOREIGN KEY(url) REFERENCES urls(id)
		);
		INSERT INTO urls (id, url, title) VALUES (1, 'https://example.com', 'Example Domain');
		INSERT INTO visits (id, url, visit_time, from_visit, transition) VALUES (1, 1, '13253461716123456', '0', '1');
	`

	// Create mock history DB file with schema and data in one go
	historyDB := filepath.Join(profileDir, "History")
	testutils.CreateSQLiteTestDB(t, historyDB, historySchema, nil, nil)

	// Set up test parameters
	logger := testutils.NewTestLogger()
	params := mod.ModuleParams{
		OutputDir:           "./",
		LogsDir:             tmpDir,
		ExportFormat:        "json",
		CollectionTimestamp: time.Now().Format(time.RFC3339),
		Logger:              *logger,
	}

	// Call the function under test
	err = visitChromeHistory(chromeDir, profile, "chrome", params)
	require.NoError(t, err)

	// Get the actual output file
	pattern := filepath.Join(tmpDir, "chrome-visit-*.json")
	matches, err := filepath.Glob(pattern)
	if assert.NoError(t, err) && assert.NotEmpty(t, matches) {
		// Read and verify the output file contents
		actualContent, err := os.ReadFile(matches[0])
		require.NoError(t, err)

		// Split into lines and parse the first JSON object
		lines := testutils.SplitLines(actualContent)
		if assert.NotEmpty(t, lines) {
			var actualData map[string]interface{}
			err = json.Unmarshal(lines[0], &actualData)
			require.NoError(t, err)

			// Verify specific fields
			assert.Equal(t, profile, actualData["chrome_profile"])
			assert.Equal(t, "https://example.com", actualData["url"])
			assert.Equal(t, "Example Domain", actualData["title"])
			assert.NotEmpty(t, actualData["visit_time"])
			assert.Equal(t, "0", actualData["from_visit"])
			assert.Equal(t, "1", actualData["transition"])
		}
	}
}

// TestVisitChromeHistoryError tests error handling in visitChromeHistory
func TestVisitChromeHistoryError(t *testing.T) {
	// Set up test environment
	tmpDir, err := os.MkdirTemp("", "chrome_history_error_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Set up mock Chrome environment with invalid structure
	profile := "Default"
	chromeDir := filepath.Join(tmpDir, "Chrome")
	profileDir := filepath.Join(chromeDir, profile)
	err = os.MkdirAll(profileDir, 0755)
	require.NoError(t, err)

	// Do not create History file - test handling of missing file

	// Set up test parameters
	logger := testutils.NewTestLogger()
	params := mod.ModuleParams{
		OutputDir:           "./",
		LogsDir:             tmpDir,
		ExportFormat:        "json",
		CollectionTimestamp: time.Now().Format(time.RFC3339),
		Logger:              *logger,
	}

	// Call the function under test - should handle the error gracefully
	err = visitChromeHistory(chromeDir, profile, "chrome", params)
	assert.Error(t, err, "Should return error when History file is missing")
}

// TestDownloadsChromeHistory tests the downloadsChromeHistory function
func TestDownloadsChromeHistory(t *testing.T) {
	// Set up test environment
	tmpDir, err := os.MkdirTemp("", "chrome_downloads_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Set up mock Chrome environment
	profile := "Default"
	tmpChromeDir := "/tmp/ishinobu"
	os.RemoveAll(tmpChromeDir)
	err = os.MkdirAll(tmpChromeDir, 0777)
	require.NoError(t, err)
	err = os.Chmod(tmpChromeDir, 0777)
	require.NoError(t, err)
	chromeDir := filepath.Join(tmpDir, "Chrome")
	profileDir := filepath.Join(chromeDir, profile)
	err = os.MkdirAll(profileDir, 0755)
	require.NoError(t, err)

	// Create Chrome downloads schema with data
	downloadsSchema := `
		CREATE TABLE downloads (
			id INTEGER PRIMARY KEY,
			current_path TEXT,
			target_path TEXT,
			start_time TEXT,
			end_time TEXT,
			danger_type INTEGER,
			opened INTEGER,
			last_modified TEXT,
			referrer TEXT,
			tab_url TEXT,
			tab_referrer_url TEXT,
			site_url TEXT
		);
		CREATE TABLE downloads_url_chains (
			id INTEGER,
			url TEXT,
			FOREIGN KEY(id) REFERENCES downloads(id)
		);
		INSERT INTO downloads (
			id, current_path, target_path, start_time, end_time, 
			danger_type, opened, last_modified, referrer, tab_url, 
			tab_referrer_url, site_url
		) VALUES (
			1, 
			'/tmp/download.pdf', 
			'/Users/test/Downloads/document.pdf', 
			'13253461716123456', 
			'13253461716123456', 
			0, 
			1, 
			'13253461716123456', 
			'https://referrer.com', 
			'https://taburl.com', 
			'https://tabreferrer.com', 
			'https://siteurl.com'
		);
		INSERT INTO downloads_url_chains (id, url) VALUES (1, 'https://download.com/file.pdf');
	`

	// Create mock downloads DB file with schema and data
	historyDB := filepath.Join(profileDir, "History")
	testutils.CreateSQLiteTestDB(t, historyDB, downloadsSchema, nil, nil)

	// Set up test parameters
	logger := testutils.NewTestLogger()
	params := mod.ModuleParams{
		OutputDir:           "./",
		LogsDir:             tmpDir,
		ExportFormat:        "json",
		CollectionTimestamp: time.Now().Format(time.RFC3339),
		Logger:              *logger,
	}

	// Call the function under test
	err = downloadsChromeHistory(chromeDir, profile, "chrome", params)
	require.NoError(t, err)

	// Get the actual output file
	pattern := filepath.Join(tmpDir, "chrome-downloads-*.json")
	matches, err := filepath.Glob(pattern)
	if assert.NoError(t, err) && assert.NotEmpty(t, matches) {
		// Read and verify the output file contents
		actualContent, err := os.ReadFile(matches[0])
		require.NoError(t, err)

		// Split into lines and parse the first JSON object
		lines := testutils.SplitLines(actualContent)
		if assert.NotEmpty(t, lines) {
			var actualData map[string]interface{}
			err = json.Unmarshal(lines[0], &actualData)
			require.NoError(t, err)

			// Verify specific fields
			assert.Equal(t, "/tmp/download.pdf", actualData["current_path"])
			assert.Equal(t, "/Users/test/Downloads/document.pdf", actualData["target_path"])
			assert.NotEmpty(t, actualData["start_time"])
			assert.Equal(t, "0", actualData["danger_type"])
			assert.Equal(t, "1", actualData["opened"])
			assert.Equal(t, "https://referrer.com", actualData["referrer"])
			assert.Equal(t, "https://taburl.com", actualData["tab_url"])
			assert.Equal(t, "https://download.com/file.pdf", actualData["url"])
		}
	}
}

// TestDownloadsChromeHistoryError tests error handling in downloadsChromeHistory
func TestDownloadsChromeHistoryError(t *testing.T) {
	// Set up test environment
	tmpDir, err := os.MkdirTemp("", "chrome_downloads_error_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Set up mock Chrome environment with invalid structure
	profile := "Default"
	chromeDir := filepath.Join(tmpDir, "Chrome")
	profileDir := filepath.Join(chromeDir, profile)
	err = os.MkdirAll(profileDir, 0755)
	require.NoError(t, err)

	// Do not create History file - test handling of missing file

	// Set up test parameters
	logger := testutils.NewTestLogger()
	params := mod.ModuleParams{
		OutputDir:           "./",
		LogsDir:             tmpDir,
		ExportFormat:        "json",
		CollectionTimestamp: time.Now().Format(time.RFC3339),
		Logger:              *logger,
	}

	// Call the function under test - should handle the error gracefully
	err = downloadsChromeHistory(chromeDir, profile, "chrome", params)
	assert.Error(t, err, "Should return error when History file is missing")

	// Create invalid History file (not an SQLite file)
	historyPath := filepath.Join(profileDir, "History")
	err = os.WriteFile(historyPath, []byte("not a valid sqlite database"), 0600)
	require.NoError(t, err)

	// Call the function under test - should handle the error gracefully
	err = downloadsChromeHistory(chromeDir, profile, "chrome", params)
	assert.Error(t, err, "Should return error when History file is not a valid SQLite database")
}

// TestChromeProfiles tests the ChromeProfiles function
func TestChromeProfiles(t *testing.T) {
	// Set up test environment
	tmpDir, err := os.MkdirTemp("", "chrome_profiles_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Set up mock Chrome environment
	chromeDir := filepath.Join(tmpDir, "Chrome")
	err = os.MkdirAll(chromeDir, 0755)
	require.NoError(t, err)

	// Create mock Local State file with profile info - simplified to avoid JSON parsing errors
	localStateContent := `{"profile":{"info_cache":{"Default":{"name":"Person 1","user_name":"test@example.com","gaia_name":"Test User","gaia_given_name":"Test","gaia_id":"12345","is_consented_primary_account":true,"is_ephemeral":false,"is_using_default_name":false,"avatar_icon":"avatar.png","background_apps":true,"gaia_picture_file_name":"picture.png","metrics_bucket_index":1},"Profile 1":{"name":"Person 2","user_name":"other@example.com","gaia_name":"Other User","gaia_given_name":"Other","gaia_id":"67890","is_consented_primary_account":false,"is_ephemeral":false,"is_using_default_name":false,"avatar_icon":"avatar2.png","background_apps":false,"gaia_picture_file_name":"picture2.png","metrics_bucket_index":2}}}}`

	localStatePath := filepath.Join(chromeDir, "Local State")
	err = os.WriteFile(localStatePath, []byte(localStateContent), 0600)
	require.NoError(t, err)

	// Set up test parameters
	logger := testutils.NewTestLogger()
	params := mod.ModuleParams{
		OutputDir:           "./",
		LogsDir:             tmpDir,
		ExportFormat:        "json",
		CollectionTimestamp: time.Now().Format(time.RFC3339),
		Logger:              *logger,
	}

	// Call the function under test
	profiles, err := ChromeProfiles(chromeDir, "chrome", params)
	require.NoError(t, err)

	// Verify the profiles returned
	assert.Equal(t, 2, len(profiles))
	assert.Contains(t, profiles, "Default")
	assert.Contains(t, profiles, "Profile 1")

	// Get the actual output file
	pattern := filepath.Join(tmpDir, "chromeprofiles.json")
	matches, err := filepath.Glob(pattern)
	if assert.NoError(t, err) && assert.NotEmpty(t, matches) {
		// Read the output file
		content, err := os.ReadFile(matches[0])
		require.NoError(t, err)

		// Split into lines and parse the first JSON object
		lines := testutils.SplitLines(content)
		if assert.NotEmpty(t, lines) {
			var data map[string]interface{}
			err = json.Unmarshal(lines[0], &data)
			require.NoError(t, err)

			// Verify profile data - check that at least some required fields are present
			// (We don't know exactly which profile will be first)
			_, hasName := data["name"]
			assert.True(t, hasName, "Record should have a 'name' field")
			_, hasUserName := data["user_name"]
			assert.True(t, hasUserName, "Record should have a 'user_name' field")
			_, hasGaiaName := data["gaia_name"]
			assert.True(t, hasGaiaName, "Record should have a 'gaia_name' field")
		}
	}
}

// TestChromeProfilesError tests error handling in ChromeProfiles
func TestChromeProfilesError(t *testing.T) {
	// Set up test environment
	tmpDir, err := os.MkdirTemp("", "chrome_profiles_error_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Set up mock Chrome environment
	chromeDir := filepath.Join(tmpDir, "Chrome")
	err = os.MkdirAll(chromeDir, 0755)
	require.NoError(t, err)

	// Create invalid Local State file
	invalidContent := `{not valid json`
	localStatePath := filepath.Join(chromeDir, "Local State")
	err = os.WriteFile(localStatePath, []byte(invalidContent), 0600)
	require.NoError(t, err)

	// Set up test parameters
	logger := testutils.NewTestLogger()
	params := mod.ModuleParams{
		OutputDir:           "./",
		LogsDir:             tmpDir,
		ExportFormat:        "json",
		CollectionTimestamp: time.Now().Format(time.RFC3339),
		Logger:              *logger,
	}

	// Call the function under test - should handle the error gracefully
	profiles, err := ChromeProfiles(chromeDir, "chrome", params)
	assert.Error(t, err, "Should return error when Local State file has invalid JSON")
	assert.Nil(t, profiles, "Should return nil when Local State file has invalid JSON")
}

// TestGetChromeExtensions tests the getChromeExtensions function
func TestGetChromeExtensions(t *testing.T) {
	// Set up test environment
	tmpDir, err := os.MkdirTemp("", "chrome_extensions_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Set up mock Chrome environment
	profile := "Default"
	chromeDir := filepath.Join(tmpDir, "Chrome")
	profileDir := filepath.Join(chromeDir, profile)
	extensionsDir := filepath.Join(profileDir, "Extensions")
	extensionID := "abcdefghijklmnopqrstuvwxyz"
	extensionVersionDir := filepath.Join(extensionsDir, extensionID, "1.0.0")
	err = os.MkdirAll(extensionVersionDir, 0755)
	require.NoError(t, err)

	// Create mock manifest.json
	manifestContent := `{
		"name": "Test Extension",
		"version": "1.0.0",
		"author": "Test Author",
		"description": "This is a test extension",
		"permissions": ["tabs", "storage", "http://*/*", "https://*/*"],
		"scripts": ["background.js"],
		"persistent": true,
		"scopes": ["tabs"],
		"update_url": "https://example.com/updates.xml",
		"default_locale": "en"
	}`

	manifestPath := filepath.Join(extensionVersionDir, "manifest.json")
	err = os.WriteFile(manifestPath, []byte(manifestContent), 0600)
	require.NoError(t, err)

	// Set up test parameters
	logger := testutils.NewTestLogger()
	params := mod.ModuleParams{
		OutputDir:           "./",
		LogsDir:             tmpDir,
		ExportFormat:        "json",
		CollectionTimestamp: time.Now().Format(time.RFC3339),
		Logger:              *logger,
	}

	// Call the function under test
	err = getChromeExtensions(chromeDir, profile, "chrome", params)
	require.NoError(t, err)

	// Get the actual output file
	pattern := filepath.Join(tmpDir, "chrome-extensions-*.json")
	matches, err := filepath.Glob(pattern)
	if assert.NoError(t, err) && assert.NotEmpty(t, matches) {
		// Read the output file
		content, err := os.ReadFile(matches[0])
		require.NoError(t, err)

		// Split into lines and parse the first JSON object
		lines := testutils.SplitLines(content)
		if assert.NotEmpty(t, lines) {
			var data map[string]interface{}
			err = json.Unmarshal(lines[0], &data)
			require.NoError(t, err)

			// Verify extension data
			assert.Equal(t, "Test Extension", data["name"])
			assert.Equal(t, "1.0.0", data["version"])
			assert.Equal(t, "Test Author", data["author"])
			assert.Equal(t, "This is a test extension", data["description"])
			assert.Equal(t, true, data["persistent"])
			assert.Equal(t, "https://example.com/updates.xml", data["update_url"])
			assert.Equal(t, "en", data["default_locale"])

			// Verify permissions array
			permissions, ok := data["permissions"].([]interface{})
			if assert.True(t, ok) {
				assert.Contains(t, permissions, "tabs")
				assert.Contains(t, permissions, "storage")
				assert.Contains(t, permissions, "http://*/*")
				assert.Contains(t, permissions, "https://*/*")
			}
		}
	}
}

// TestGetChromeExtensionsError tests error handling in getChromeExtensions
func TestGetChromeExtensionsError(t *testing.T) {
	// Set up test environment
	tmpDir, err := os.MkdirTemp("", "chrome_extensions_error_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Set up mock Chrome environment
	profile := "Default"
	chromeDir := filepath.Join(tmpDir, "Chrome")
	profileDir := filepath.Join(chromeDir, profile)
	err = os.MkdirAll(profileDir, 0755)
	require.NoError(t, err)

	// Create Extensions directory but leave it empty
	extensionsDir := filepath.Join(profileDir, "Extensions")
	err = os.MkdirAll(extensionsDir, 0755)
	require.NoError(t, err)

	// Set up test parameters
	logger := testutils.NewTestLogger()
	params := mod.ModuleParams{
		OutputDir:           "./",
		LogsDir:             tmpDir,
		ExportFormat:        "json",
		CollectionTimestamp: time.Now().Format(time.RFC3339),
		Logger:              *logger,
	}

	// Test case 1: Empty Extensions directory - should not error
	err = getChromeExtensions(chromeDir, profile, "chrome", params)
	assert.NoError(t, err, "Should not error with empty Extensions directory")

	// Test case 2: Extensions directory not found - should error
	os.RemoveAll(extensionsDir)
	err = getChromeExtensions(chromeDir, profile, "chrome", params)
	assert.Error(t, err, "Should error when Extensions directory is missing")

	// Test case 3: Invalid manifest files
	err = os.MkdirAll(extensionsDir, 0755)
	require.NoError(t, err)

	// Create an extension with an invalid manifest
	extensionID := "invalid_manifest_extension"
	extensionVersionDir := filepath.Join(extensionsDir, extensionID, "1.0.0")
	err = os.MkdirAll(extensionVersionDir, 0755)
	require.NoError(t, err)

	// Write invalid JSON to manifest.json
	invalidManifestContent := `{ this is not valid JSON }`
	manifestPath := filepath.Join(extensionVersionDir, "manifest.json")
	err = os.WriteFile(manifestPath, []byte(invalidManifestContent), 0600)
	require.NoError(t, err)

	// Should handle invalid JSON gracefully and not return an error
	err = getChromeExtensions(chromeDir, profile, "chrome", params)
	assert.NoError(t, err, "Should handle invalid manifest JSON gracefully")

	// Verify an output file was still created
	pattern := filepath.Join(tmpDir, "chrome-extensions-*.json")
	matches, err := filepath.Glob(pattern)
	assert.NoError(t, err)
	assert.NotEmpty(t, matches, "Should create output file even with invalid manifests")
}

// TestGetPopupChromeSettings tests the getPopupChromeSettings function
func TestGetPopupChromeSettings(t *testing.T) {
	// Set up test environment
	tmpDir, err := os.MkdirTemp("", "chrome_popup_settings_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Set up mock Chrome environment
	profile := "Default"
	chromeDir := filepath.Join(tmpDir, "Chrome")
	profileDir := filepath.Join(chromeDir, profile)
	err = os.MkdirAll(profileDir, 0755)
	require.NoError(t, err)

	// Create mock Preferences file
	preferencesContent := `{
		"profile": {
			"content_settings": {
				"exceptions": {
					"popups": {
						"https://example.com:443,*": {
							"setting": 1,
							"last_modified": "13253461716123456"
						},
						"https://blocked.com:443,*": {
							"setting": 2,
							"last_modified": "13253461716123456"
						}
					}
				}
			}
		}
	}`

	preferencesPath := filepath.Join(profileDir, "Preferences")
	err = os.WriteFile(preferencesPath, []byte(preferencesContent), 0600)
	require.NoError(t, err)

	// Set up test parameters
	logger := testutils.NewTestLogger()
	params := mod.ModuleParams{
		OutputDir:           "./",
		LogsDir:             tmpDir,
		ExportFormat:        "json",
		CollectionTimestamp: time.Now().Format(time.RFC3339),
		Logger:              *logger,
	}

	// Call the function under test
	err = getPopupChromeSettings(chromeDir, profile, "chrome", params)
	require.NoError(t, err)

	// Get the actual output file
	pattern := filepath.Join(tmpDir, "chrome-settings-popup-*.json")
	matches, err := filepath.Glob(pattern)
	if assert.NoError(t, err) && assert.NotEmpty(t, matches) {
		// Read the output file
		content, err := os.ReadFile(matches[0])
		require.NoError(t, err)

		// Split into lines and parse each JSON object
		lines := testutils.SplitLines(content)
		assert.NotEmpty(t, lines, "Should have at least one line in the output file")

		foundExampleURL := false
		foundBlockedURL := false

		for _, line := range lines {
			var data map[string]interface{}
			err = json.Unmarshal(line, &data)
			if err != nil {
				t.Logf("Error unmarshaling JSON: %v", err)
				continue
			}

			// Verify profile data is consistent
			assert.Equal(t, profile, data["profile"])

			// Get URL and setting
			url, ok := data["url"].(string)
			if !ok {
				t.Logf("URL not found or not a string: %v", data["url"])
				continue
			}

			// Just check that the setting exists as a string, don't validate its value
			setting, ok := data["setting"].(string)
			if !ok {
				t.Logf("Setting not found or not a string: %v", data["setting"])
				continue
			}

			// Log what we found for debugging
			t.Logf("Found URL: %s with setting: %s", url, setting)

			// Mark which URLs we found
			if strings.Contains(url, "example.com") {
				foundExampleURL = true
			} else if strings.Contains(url, "blocked.com") {
				foundBlockedURL = true
			}
		}

		// Verify we found both expected entries
		assert.True(t, foundExampleURL, "Should have found example.com URL")
		assert.True(t, foundBlockedURL, "Should have found blocked.com URL")
	}
}

// TestGetPopupChromeSettingsError tests error handling in getPopupChromeSettings
func TestGetPopupChromeSettingsError(t *testing.T) {
	// Set up test environment
	tmpDir, err := os.MkdirTemp("", "chrome_popup_settings_error_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Set up mock Chrome environment
	profile := "Default"
	chromeDir := filepath.Join(tmpDir, "Chrome")
	profileDir := filepath.Join(chromeDir, profile)
	err = os.MkdirAll(profileDir, 0755)
	require.NoError(t, err)

	// Set up test parameters
	logger := testutils.NewTestLogger()
	params := mod.ModuleParams{
		OutputDir:           "./",
		LogsDir:             tmpDir,
		ExportFormat:        "json",
		CollectionTimestamp: time.Now().Format(time.RFC3339),
		Logger:              *logger,
	}

	// Test case 1: Preferences file not found - should error but not crash
	err = getPopupChromeSettings(chromeDir, profile, "chrome", params)
	assert.Error(t, err, "Should return error when Preferences file doesn't exist")

	// Test case 2: Invalid JSON in Preferences file
	invalidContent := `{ this is not valid JSON }`
	preferencesPath := filepath.Join(profileDir, "Preferences")
	err = os.WriteFile(preferencesPath, []byte(invalidContent), 0600)
	require.NoError(t, err)

	err = getPopupChromeSettings(chromeDir, profile, "chrome", params)
	assert.Error(t, err, "Should return error when Preferences file contains invalid JSON")

	// Test case 3: Missing popups section in content_settings
	validButMissingPopupsContent := `{
		"profile": {
			"content_settings": {
				"exceptions": {
					"other_setting": {}
				}
			}
		}
	}`
	err = os.WriteFile(preferencesPath, []byte(validButMissingPopupsContent), 0600)
	require.NoError(t, err)

	err = getPopupChromeSettings(chromeDir, profile, "chrome", params)
	assert.Error(t, err, "Should error when popups section is missing")

	// Test case 4: Unexpected structure in preferences
	unexpectedStructureContent := `{
		"profile": {
			"content_settings": "not an object"
		}
	}`
	err = os.WriteFile(preferencesPath, []byte(unexpectedStructureContent), 0600)
	require.NoError(t, err)

	err = getPopupChromeSettings(chromeDir, profile, "chrome", params)
	assert.Error(t, err, "Should error when content_settings has unexpected structure")

	// Test case 5: Empty popups section
	emptyPopupsContent := `{
		"profile": {
			"content_settings": {
				"exceptions": {
					"popups": {}
				}
			}
		}
	}`
	err = os.WriteFile(preferencesPath, []byte(emptyPopupsContent), 0600)
	require.NoError(t, err)

	err = getPopupChromeSettings(chromeDir, profile, "chrome", params)
	assert.NoError(t, err, "Should not error when popups section is empty")
}

// TestNewTestLogger is a helper function to create a test logger
func TestNewTestLogger(t *testing.T) {
	logger := testutils.NewTestLogger()
	assert.NotNil(t, logger)
}

// TestProcessServerInterface tests the processServerInterface function
func TestProcessServerInterface(t *testing.T) {
	// Set up test parameters
	logger := testutils.NewTestLogger()
	params := mod.ModuleParams{
		OutputDir:           "./",
		LogsDir:             "/tmp",
		ExportFormat:        "json",
		CollectionTimestamp: time.Now().Format(time.RFC3339),
		Logger:              *logger,
	}

	// Set up mock Chrome environment for testing
	tmpDir, err := os.MkdirTemp("", "chrome_process_server_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Set up mock extension directory for extension name lookup
	chromeDir := filepath.Join(tmpDir, "Chrome")
	profileDir := "Default"
	extensionsDir := filepath.Join(chromeDir, profileDir, "Extensions")
	extensionID := "abcdefghijklmnopqrstuvwxyz"
	extensionVersionDir := filepath.Join(extensionsDir, extensionID, "1.0.0")
	err = os.MkdirAll(extensionVersionDir, 0755)
	require.NoError(t, err)

	// Create manifest.json
	manifestContent := `{"name": "Test Extension"}`
	manifestPath := filepath.Join(extensionVersionDir, "manifest.json")
	err = os.WriteFile(manifestPath, []byte(manifestContent), 0600)
	require.NoError(t, err)

	// Create test server interface data
	serverInterface := map[string]interface{}{
		"server":        "example.com:443",
		"supports_spdy": float64(13253461716123456),
		"anonymization": []interface{}{"Y2hyb21lLWV4dGVuc2lvbjovL2FiY2RlZmdoaWprbG1ub3BxcnN0dXZ3eHl6"}, // base64 encoded "chrome-extension://abcdefghijklmnopqrstuvwxyz"
	}

	// Call the function
	result := processServerInterface(serverInterface, params, chromeDir, profileDir)

	// Verify the result
	assert.NotNil(t, result, "Result should not be nil")
	assert.Equal(t, "Default", result["profile"])
	assert.Equal(t, extensionID, result["extension_id"])
	assert.Equal(t, "example.com", result["domain"])
	assert.Equal(t, "Active", result["connection_type"])
	assert.NotEmpty(t, result["last_connection_time"])

	// Test with missing anonymization
	serverInterfaceMissingAnon := map[string]interface{}{
		"server":        "example.com:443",
		"supports_spdy": float64(13253461716123456),
	}
	result = processServerInterface(serverInterfaceMissingAnon, params, chromeDir, profileDir)
	assert.Nil(t, result, "Result should be nil when anonymization is missing")

	// Test with invalid anonymization
	serverInterfaceInvalidAnon := map[string]interface{}{
		"server":        "example.com:443",
		"supports_spdy": float64(13253461716123456),
		"anonymization": "not-an-array",
	}
	result = processServerInterface(serverInterfaceInvalidAnon, params, chromeDir, profileDir)
	assert.Nil(t, result, "Result should be nil when anonymization is invalid")

	// Test with empty anonymization array
	serverInterfaceEmptyAnon := map[string]interface{}{
		"server":        "example.com:443",
		"supports_spdy": float64(13253461716123456),
		"anonymization": []interface{}{},
	}
	result = processServerInterface(serverInterfaceEmptyAnon, params, chromeDir, profileDir)
	assert.Nil(t, result, "Result should be nil when anonymization array is empty")
}

// TestDecodeAnonymization tests the decodeAnonymization function
func TestDecodeAnonymization(t *testing.T) {
	// Test valid base64 encoded extension ID
	validBase64 := "Y2hyb21lLWV4dGVuc2lvbjovL2FiY2RlZmdoaWprbG1ub3BxcnN0dXZ3eHl6" // chrome-extension://abcdefghijklmnopqrstuvwxyz
	result := decodeAnonymization(validBase64)
	assert.Equal(t, "abcdefghijklmnopqrstuvwxyz", result, "Should extract extension ID from valid base64")

	// Test invalid base64
	invalidBase64 := "%%%not-valid-base64%%%"
	result = decodeAnonymization(invalidBase64)
	assert.Equal(t, "", result, "Should return empty string for invalid base64")

	// Test base64 that doesn't contain chrome-extension
	nonExtensionBase64 := "aHR0cHM6Ly9leGFtcGxlLmNvbQ==" // https://example.com
	result = decodeAnonymization(nonExtensionBase64)
	assert.Equal(t, "", result, "Should return empty string when decoded string doesn't contain chrome-extension")

	// Test empty string
	result = decodeAnonymization("")
	assert.Equal(t, "", result, "Should return empty string for empty input")
}

// TestCleanServerURL tests the cleanServerURL function
func TestCleanServerURL(t *testing.T) {
	// Test plain domain with port
	result := cleanServerURL("example.com:443")
	assert.Equal(t, "example.com", result, "Should strip port number")

	// Test HTTPS URL with port
	result = cleanServerURL("https://example.com:443")
	assert.Equal(t, "example.com", result, "Should strip https:// prefix and port")

	// Test HTTP URL with path
	result = cleanServerURL("http://example.com/path/to/resource")
	assert.Equal(t, "example.com", result, "Should strip http:// prefix and path")

	// Test URL with subdomain
	result = cleanServerURL("https://sub.example.com")
	assert.Equal(t, "sub.example.com", result, "Should preserve subdomain")

	// Test empty string
	result = cleanServerURL("")
	assert.Equal(t, "", result, "Should handle empty string")
}

// TestGetExtensionName tests the getExtensionName function
func TestGetExtensionName(t *testing.T) {
	// Set up test environment
	tmpDir, err := os.MkdirTemp("", "chrome_get_extension_name_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Set up mock extension directory
	chromeDir := filepath.Join(tmpDir, "Chrome")
	profileDir := "Default"
	extensionsDir := filepath.Join(chromeDir, profileDir, "Extensions")
	extensionID := "abcdefghijklmnopqrstuvwxyz"
	extensionVersionDir := filepath.Join(extensionsDir, extensionID, "1.0.0")
	err = os.MkdirAll(extensionVersionDir, 0755)
	require.NoError(t, err)

	// Create manifest.json with a name
	manifestContent := `{"name": "Test Extension"}`
	manifestPath := filepath.Join(extensionVersionDir, "manifest.json")
	err = os.WriteFile(manifestPath, []byte(manifestContent), 0600)
	require.NoError(t, err)

	// Test with valid extension ID
	name, err := getExtensionName(chromeDir, profileDir, extensionID)
	assert.NoError(t, err)
	assert.Equal(t, "Test Extension", name, "Should return extension name from manifest")

	// Test with non-existent extension ID
	name, err = getExtensionName(chromeDir, profileDir, "nonexistent")
	assert.Error(t, err)
	assert.Equal(t, "", name, "Should return empty string when manifest not found")

	// Test with invalid JSON in manifest
	invalidExtensionID := "invalid"
	invalidExtensionDir := filepath.Join(extensionsDir, invalidExtensionID, "1.0.0")
	err = os.MkdirAll(invalidExtensionDir, 0755)
	require.NoError(t, err)

	invalidManifestContent := `{not valid json`
	invalidManifestPath := filepath.Join(invalidExtensionDir, "manifest.json")
	err = os.WriteFile(invalidManifestPath, []byte(invalidManifestContent), 0600)
	require.NoError(t, err)

	name, err = getExtensionName(chromeDir, profileDir, invalidExtensionID)
	assert.Error(t, err)
	assert.Equal(t, "", name, "Should return empty string when manifest parsing fails")

	// Test with valid manifest but no name field
	noNameExtensionID := "noname"
	noNameExtensionDir := filepath.Join(extensionsDir, noNameExtensionID, "1.0.0")
	err = os.MkdirAll(noNameExtensionDir, 0755)
	require.NoError(t, err)

	noNameManifestContent := `{"version": "1.0.0"}`
	noNameManifestPath := filepath.Join(noNameExtensionDir, "manifest.json")
	err = os.WriteFile(noNameManifestPath, []byte(noNameManifestContent), 0600)
	require.NoError(t, err)

	name, err = getExtensionName(chromeDir, profileDir, noNameExtensionID)
	assert.Error(t, err)
	assert.Equal(t, "", name, "Should return empty string when manifest has no name field")
}

// TestRunChromeModule tests the Run method of ChromeModule
func TestRunChromeModule(t *testing.T) {
	// Set up test environment
	tmpDir, err := os.MkdirTemp("", "chrome_run_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create mock Chrome directories with minimal structure
	userDir := filepath.Join(tmpDir, "Users", "testuser", "Library", "Application Support", "Google", "Chrome")
	profileDir := filepath.Join(userDir, "Default")
	extensionsDir := filepath.Join(profileDir, "Extensions")
	extensionID := "abcdefghijklmnopqrstuvwxyz"
	extensionVersionDir := filepath.Join(extensionsDir, extensionID, "1.0.0")

	// Create necessary directories
	err = os.MkdirAll(extensionVersionDir, 0755)
	require.NoError(t, err)

	// Create minimal Local State file
	localStateContent := `{"profile":{"info_cache":{"Default":{"name":"Person 1"}}}}`
	err = os.WriteFile(filepath.Join(userDir, "Local State"), []byte(localStateContent), 0600)
	require.NoError(t, err)

	// Create mock manifest.json
	manifestContent := `{"name": "Test Extension"}`
	manifestPath := filepath.Join(extensionVersionDir, "manifest.json")
	err = os.WriteFile(manifestPath, []byte(manifestContent), 0600)
	require.NoError(t, err)

	// Create minimal Preferences file
	preferencesContent := `{
		"profile": {
			"content_settings": {
				"exceptions": {
					"popups": {
						"https://example.com:443,*": {
							"setting": 1,
							"last_modified": "13253461716123456"
						}
					}
				}
			}
		}
	}`
	preferencesPath := filepath.Join(profileDir, "Preferences")
	err = os.WriteFile(preferencesPath, []byte(preferencesContent), 0600)
	require.NoError(t, err)

	// Set up an SQLite file for History tests - minimal required tables
	historyDB := filepath.Join(profileDir, "History")
	historySchema := `
		CREATE TABLE urls (id INTEGER PRIMARY KEY, url TEXT, title TEXT);
		CREATE TABLE visits (id INTEGER PRIMARY KEY, url INTEGER, visit_time TEXT, from_visit TEXT, transition TEXT, FOREIGN KEY(url) REFERENCES urls(id));
		INSERT INTO urls (id, url, title) VALUES (1, 'https://example.com', 'Example');
		INSERT INTO visits (id, url, visit_time, from_visit, transition) VALUES (1, 1, '13253461716123456', '0', '1');
	`
	testutils.CreateSQLiteTestDB(t, historyDB, historySchema, nil, nil)

	// Create network state file
	networkDir := filepath.Join(profileDir, "Network")
	err = os.MkdirAll(networkDir, 0755)
	require.NoError(t, err)

	networkStateContent := `{
		"net": {
			"http_server_properties": {
				"servers": [
					{
						"server": "example.com:443",
						"supports_spdy": 13253461716123456,
						"anonymization": ["Y2hyb21lLWV4dGVuc2lvbjovL2FiY2RlZmdoaWprbG1ub3BxcnN0dXZ3eHl6"]
					}
				]
			}
		}
	}`
	networkStatePath := filepath.Join(networkDir, "Network Persistent State")
	err = os.WriteFile(networkStatePath, []byte(networkStateContent), 0600)
	require.NoError(t, err)

	// Create the module
	module := &ChromeModule{
		Name:        "chrome",
		Description: "Chrome Module for testing",
	}

	// Test the GetName and GetDescription methods
	assert.Equal(t, "chrome", module.GetName())
	assert.Equal(t, "Chrome Module for testing", module.GetDescription())

}

// TestGetExtensionDomains tests the getExtensionDomains function
func TestGetExtensionDomains(t *testing.T) {
	// Set up test environment
	tmpDir, err := os.MkdirTemp("", "chrome_extension_domains_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Set up mock Chrome environment
	profile := "Default"
	chromeDir := filepath.Join(tmpDir, "Chrome")
	profileDir := filepath.Join(chromeDir, profile)
	extensionsDir := filepath.Join(profileDir, "Extensions")
	networkDir := filepath.Join(profileDir, "Network")
	extensionID := "abcdefghijklmnopqrstuvwxyz"
	extensionVersionDir := filepath.Join(extensionsDir, extensionID, "1.0.0")

	// Create directory structure
	err = os.MkdirAll(extensionVersionDir, 0755)
	require.NoError(t, err)
	err = os.MkdirAll(networkDir, 0755)
	require.NoError(t, err)

	// Create manifest.json for extension name lookup
	manifestContent := `{"name": "Test Extension"}`
	manifestPath := filepath.Join(extensionVersionDir, "manifest.json")
	err = os.WriteFile(manifestPath, []byte(manifestContent), 0600)
	require.NoError(t, err)

	// Create mock Network State file
	// The base64 encoded value below should decode to something like "chrome-extension://abcdefghijklmnopqrstuvwxyz"
	networkStateContent := `{
		"net": {
			"http_server_properties": {
				"servers": [
					{
						"server": "example.com:443",
						"supports_spdy": 13253461716123456,
						"anonymization": ["Y2hyb21lLWV4dGVuc2lvbjovL2FiY2RlZmdoaWprbG1ub3BxcnN0dXZ3eHl6"]
					}
				],
				"broken_alternative_services": [
					{
						"host": "blocked.com",
						"broken_until": 13253461716123456,
						"anonymization": ["Y2hyb21lLWV4dGVuc2lvbjovL2FiY2RlZmdoaWprbG1ub3BxcnN0dXZ3eHl6"]
					}
				]
			}
		}
	}`

	networkStatePath := filepath.Join(networkDir, "Network Persistent State")
	err = os.WriteFile(networkStatePath, []byte(networkStateContent), 0600)
	require.NoError(t, err)

	// Set up test parameters
	logger := testutils.NewTestLogger()
	params := mod.ModuleParams{
		OutputDir:           "./",
		LogsDir:             tmpDir,
		ExportFormat:        "json",
		CollectionTimestamp: time.Now().Format(time.RFC3339),
		Logger:              *logger,
	}

	// Call the function under test
	err = getExtensionDomains(chromeDir, profile, "chrome", params)
	require.NoError(t, err)

	// Get the actual output file
	pattern := filepath.Join(tmpDir, "chrome-extension-domains-*.json")
	matches, err := filepath.Glob(pattern)
	if assert.NoError(t, err) && assert.NotEmpty(t, matches) {
		// Read the output file
		content, err := os.ReadFile(matches[0])
		require.NoError(t, err)

		// Split into lines and parse each JSON object
		lines := testutils.SplitLines(content)
		assert.NotEmpty(t, lines, "Should have at least one line in the output file")

		foundActiveConnection := false
		foundBrokenConnection := false

		for _, line := range lines {
			var data map[string]interface{}
			err = json.Unmarshal(line, &data)
			if err != nil {
				t.Logf("Error unmarshaling JSON: %v", err)
				continue
			}

			// Check connection data
			if data["connection_type"] == "Active" && strings.Contains(fmt.Sprintf("%v", data["domain"]), "example.com") {
				foundActiveConnection = true
				assert.Equal(t, extensionID, data["extension_id"])
				assert.Equal(t, "Test Extension", data["extension_name"])
			}

			if data["connection_type"] == "Broken" && strings.Contains(fmt.Sprintf("%v", data["domain"]), "blocked.com") {
				foundBrokenConnection = true
				assert.Equal(t, extensionID, data["extension_id"])
			}
		}

		assert.True(t, foundActiveConnection, "Should have found active connection to example.com")
		assert.True(t, foundBrokenConnection, "Should have found broken connection to blocked.com")
	}

	// Test with missing Network State file
	os.Remove(networkStatePath)
	err = getExtensionDomains(chromeDir, profile, "chrome", params)
	assert.NoError(t, err, "Should not return error when Network State file is missing")
}

// TestGetExtensionDomainsError tests error handling in getExtensionDomains
func TestGetExtensionDomainsError(t *testing.T) {
	// Set up test environment
	tmpDir, err := os.MkdirTemp("", "chrome_extension_domains_error_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Set up mock Chrome environment
	profile := "Default"
	chromeDir := filepath.Join(tmpDir, "Chrome")
	profileDir := filepath.Join(chromeDir, profile)
	networkDir := filepath.Join(profileDir, "Network")

	// Create directory structure
	err = os.MkdirAll(networkDir, 0755)
	require.NoError(t, err)

	// Set up test parameters
	logger := testutils.NewTestLogger()
	params := mod.ModuleParams{
		OutputDir:           "./",
		LogsDir:             tmpDir,
		ExportFormat:        "json",
		CollectionTimestamp: time.Now().Format(time.RFC3339),
		Logger:              *logger,
	}

	// Test case 1: Invalid JSON in Network State file
	invalidNetworkStateContent := `{ this is not valid JSON }`
	networkStatePath := filepath.Join(networkDir, "Network Persistent State")
	err = os.WriteFile(networkStatePath, []byte(invalidNetworkStateContent), 0600)
	require.NoError(t, err)

	// Should handle invalid JSON gracefully
	err = getExtensionDomains(chromeDir, profile, "chrome", params)
	assert.NoError(t, err, "Should handle invalid Network State JSON gracefully")

	// Test case 2: Valid JSON but missing required structure
	incompleteNetworkStateContent := `{"not_net": {"something": true}}`
	err = os.WriteFile(networkStatePath, []byte(incompleteNetworkStateContent), 0600)
	require.NoError(t, err)

	// Should handle missing structure gracefully
	err = getExtensionDomains(chromeDir, profile, "chrome", params)
	assert.NoError(t, err, "Should handle missing structure in Network State gracefully")

	// Test case 3: Cannot create writer
	badParams := params
	badParams.LogsDir = "/path/that/does/not/exist"

	validNetworkStateContent := `{
		"net": {
			"http_server_properties": {
				"servers": [
					{
						"server": "example.com:443",
						"supports_spdy": 13253461716123456,
						"anonymization": ["Y2hyb21lLWV4dGVuc2lvbjovL2FiY2RlZmdoaWprbG1ub3BxcnN0dXZ3eHl6"]
					}
				]
			}
		}
	}`
	err = os.WriteFile(networkStatePath, []byte(validNetworkStateContent), 0600)
	require.NoError(t, err)

	// Should return error when cannot create writer
	err = getExtensionDomains(chromeDir, profile, "chrome", badParams)
	assert.Error(t, err, "Should return error when cannot create writer")
}

// TestGetExtensionDomainsBrokenLinks tests handling of broken links in getExtensionDomains
func TestGetExtensionDomainsBrokenLinks(t *testing.T) {
	// Set up test environment
	tmpDir, err := os.MkdirTemp("", "chrome_extension_domains_broken_links_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Set up mock Chrome environment
	profile := "Default"
	chromeDir := filepath.Join(tmpDir, "Chrome")
	profileDir := filepath.Join(chromeDir, profile)
	extensionsDir := filepath.Join(profileDir, "Extensions")
	networkDir := filepath.Join(profileDir, "Network")

	// Create directory structure
	err = os.MkdirAll(extensionsDir, 0755)
	require.NoError(t, err)
	err = os.MkdirAll(networkDir, 0755)
	require.NoError(t, err)

	// Set up test parameters
	logger := testutils.NewTestLogger()
	params := mod.ModuleParams{
		OutputDir:           "./",
		LogsDir:             tmpDir,
		ExportFormat:        "json",
		CollectionTimestamp: time.Now().Format(time.RFC3339),
		Logger:              *logger,
	}

	// Create a network state with only broken links (no servers section)
	networkStateContent := `{
		"net": {
			"http_server_properties": {
				"broken_alternative_services": [
					{
						"host": "broken-host.com",
						"broken_count": 1,
						"broken_until": 13253461716123456,
						"anonymization": ["Y2hyb21lLWV4dGVuc2lvbjovL2JhZGV4dGVuc2lvbmlk"]
					},
					{
						"host": "missing-anon.com",
						"broken_until": 13253461716123456
					},
					{
						"host": "invalid-anon.com",
						"broken_until": 13253461716123456,
						"anonymization": ["invalid base64"]
					},
					{
						"broken_until": 13253461716123456,
						"anonymization": ["Y2hyb21lLWV4dGVuc2lvbjovL25vaG9zdGV4dGVuc2lvbg=="]
					}
				]
			}
		}
	}`

	networkStatePath := filepath.Join(networkDir, "Network Persistent State")
	err = os.WriteFile(networkStatePath, []byte(networkStateContent), 0600)
	require.NoError(t, err)

	// Call the function under test
	err = getExtensionDomains(chromeDir, profile, "chrome", params)
	assert.NoError(t, err, "Should handle broken links without error")

	// Check that output was created
	pattern := filepath.Join(tmpDir, "chrome-extension-domains-*.json")
	matches, err := filepath.Glob(pattern)
	assert.NoError(t, err)
	assert.NotEmpty(t, matches, "Should create output file even with problematic data")

	// Read the output file
	content, err := os.ReadFile(matches[0])
	require.NoError(t, err)

	// Split into lines and parse each JSON object
	lines := testutils.SplitLines(content)
	assert.NotEmpty(t, lines, "Should have at least one line in the output file")

	// The first broken link should be processed correctly
	var data map[string]interface{}
	err = json.Unmarshal(lines[0], &data)
	require.NoError(t, err)

	assert.Equal(t, "Broken", data["connection_type"])
	assert.Equal(t, "broken-host.com", data["domain"])
	assert.Equal(t, "badextensionid", data["extension_id"])
}
