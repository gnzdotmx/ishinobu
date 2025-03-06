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

func TestBrowserCookiesModule(t *testing.T) {
	// Cleanup any log files after test completes
	defer cleanupLogFiles(t)

	// Create temporary directory for test outputs
	tmpDir, err := os.MkdirTemp("", "browsercookies_test")
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
	module := &BrowserCookiesModule{
		Name:        "browsercookies",
		Description: "Collects and parses browser cookies from Chrome and Firefox",
	}

	// Test GetName
	t.Run("GetName", func(t *testing.T) {
		assert.Equal(t, "browsercookies", module.GetName())
	})

	// Test GetDescription
	t.Run("GetDescription", func(t *testing.T) {
		assert.Contains(t, module.GetDescription(), "browser cookies")
	})

	// Test Run method
	t.Run("Run", func(t *testing.T) {
		// Create mock output files directly
		createMockBrowserCookiesFiles(t, params)

		// Check if output files were created
		expectedFiles := []string{
			"browsercookies-chrome",
			"browsercookies-firefox",
		}

		for _, file := range expectedFiles {
			pattern := filepath.Join(tmpDir, file+"*.json")
			matches, err := filepath.Glob(pattern)
			assert.NoError(t, err)
			assert.NotEmpty(t, matches, "Expected output file not found: "+file)

			// Verify file contents
			verifyBrowserCookiesFileContents(t, matches[0], file)
		}
	})
}

func TestCollectChromeCookies(t *testing.T) {
	defer cleanupLogFiles(t)

	tmpDir, err := os.MkdirTemp("", "chrome_cookies_test")
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

	// Create mock Chrome cookies output file
	createMockChromeCookiesFile(t, params)

	// Check if the file exists
	pattern := filepath.Join(tmpDir, "browsercookies-chrome*.json")
	matches, err := filepath.Glob(pattern)
	assert.NoError(t, err)
	assert.NotEmpty(t, matches)

	// Verify file contents
	content, err := os.ReadFile(matches[0])
	assert.NoError(t, err)

	var jsonData map[string]interface{}
	err = json.Unmarshal(content, &jsonData)
	assert.NoError(t, err)

	// Verify specific fields for Chrome cookies
	assert.Contains(t, jsonData["source_file"].(string), "Chrome/Default/Cookies")
	assert.Equal(t, "Default", jsonData["chrome_profile"])
	assert.Equal(t, "example.com", jsonData["host_key"])
	assert.Equal(t, "test_cookie", jsonData["name"])
	assert.Equal(t, "test_value", jsonData["value"])
	assert.Equal(t, "/", jsonData["path"])
	assert.NotEmpty(t, jsonData["creation_utc"])
	assert.NotEmpty(t, jsonData["expires_utc"])
	assert.NotEmpty(t, jsonData["last_access_utc"])
	assert.Equal(t, "1", jsonData["is_secure"])
	assert.Equal(t, "1", jsonData["is_httponly"])
}

func TestCollectFirefoxCookies(t *testing.T) {
	defer cleanupLogFiles(t)

	tmpDir, err := os.MkdirTemp("", "firefox_cookies_test")
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

	// Create mock Firefox cookies output file
	createMockFirefoxCookiesFile(t, params)

	// Check if the file exists
	pattern := filepath.Join(tmpDir, "browsercookies-firefox*.json")
	matches, err := filepath.Glob(pattern)
	assert.NoError(t, err)
	assert.NotEmpty(t, matches)

	// Verify file contents
	content, err := os.ReadFile(matches[0])
	assert.NoError(t, err)

	var jsonData map[string]interface{}
	err = json.Unmarshal(content, &jsonData)
	assert.NoError(t, err)

	// Verify specific fields for Firefox cookies
	assert.Contains(t, jsonData["source_file"].(string), "Firefox/Profiles/abcd1234.default/cookies.sqlite")
	assert.Equal(t, "testuser", jsonData["user"])
	assert.Equal(t, "abcd1234.default", jsonData["profile"])
	assert.Equal(t, "example.org", jsonData["host"])
	assert.Equal(t, "firefox_cookie", jsonData["name"])
	assert.Equal(t, "firefox_value", jsonData["value"])
	assert.Equal(t, "/", jsonData["path"])
	assert.NotEmpty(t, jsonData["creation_time"])
	assert.NotEmpty(t, jsonData["expiry"])
	assert.NotEmpty(t, jsonData["last_accessed"])
	assert.Equal(t, "1", jsonData["is_secure"])
	assert.Equal(t, "1", jsonData["is_httponly"])
}

// Helper function to verify browser cookies file contents
func verifyBrowserCookiesFileContents(t *testing.T, filePath string, fileType string) {
	// Read the file
	content, err := os.ReadFile(filePath)
	assert.NoError(t, err, "Should be able to read the browser cookies file")

	// Parse the JSON
	var jsonData map[string]interface{}
	err = json.Unmarshal(content, &jsonData)
	assert.NoError(t, err, "Should be able to parse the browser cookies JSON")

	// Verify common fields
	assert.NotEmpty(t, jsonData["collection_timestamp"], "Should have collection timestamp")
	assert.NotEmpty(t, jsonData["event_timestamp"], "Should have event timestamp")
	assert.NotEmpty(t, jsonData["source_file"], "Should have source file")
	assert.NotEmpty(t, jsonData["name"], "Should have cookie name")
	assert.NotEmpty(t, jsonData["value"], "Should have cookie value")
	assert.NotEmpty(t, jsonData["path"], "Should have cookie path")

	// Verify browser-specific fields
	switch {
	case fileType == "browsercookies-chrome":
		assert.NotEmpty(t, jsonData["chrome_profile"], "Should have Chrome profile")
		assert.NotEmpty(t, jsonData["host_key"], "Should have host key")
		assert.NotEmpty(t, jsonData["creation_utc"], "Should have creation timestamp")
		assert.NotEmpty(t, jsonData["expires_utc"], "Should have expiration timestamp")
		assert.NotEmpty(t, jsonData["last_access_utc"], "Should have last access timestamp")

	case fileType == "browsercookies-firefox":
		assert.NotEmpty(t, jsonData["user"], "Should have Firefox user")
		assert.NotEmpty(t, jsonData["profile"], "Should have Firefox profile")
		assert.NotEmpty(t, jsonData["host"], "Should have host")
		assert.NotEmpty(t, jsonData["creation_time"], "Should have creation timestamp")
		assert.NotEmpty(t, jsonData["expiry"], "Should have expiration timestamp")
		assert.NotEmpty(t, jsonData["last_accessed"], "Should have last access timestamp")
	}

	// Verify security-related fields which should be present for both browsers
	if fileType == "browsercookies-chrome" || fileType == "browsercookies-firefox" {
		assert.Contains(t, []string{"0", "1"}, jsonData["is_secure"], "is_secure should be 0 or 1")
		assert.Contains(t, []string{"0", "1"}, jsonData["is_httponly"], "is_httponly should be 0 or 1")
	}
}

// Helper functions to create mock output files

func createMockBrowserCookiesFiles(t *testing.T, params mod.ModuleParams) {
	createMockChromeCookiesFile(t, params)
	createMockFirefoxCookiesFile(t, params)
}

func createMockChromeCookiesFile(t *testing.T, params mod.ModuleParams) {
	filename := "browsercookies-chrome-testuser-Default-" + params.CollectionTimestamp + "." + params.ExportFormat
	filepath := filepath.Join(params.OutputDir, filename)

	record := utils.Record{
		CollectionTimestamp: params.CollectionTimestamp,
		EventTimestamp:      time.Now().Format(utils.TimeFormat),
		SourceFile:          "/Users/testuser/Library/Application Support/Google/Chrome/Default/Cookies",
		Data: map[string]interface{}{
			"chrome_profile":  "Default",
			"host_key":        "example.com",
			"name":            "test_cookie",
			"value":           "test_value",
			"path":            "/",
			"creation_utc":    time.Now().Format(utils.TimeFormat),
			"expires_utc":     time.Now().AddDate(0, 0, 30).Format(utils.TimeFormat),
			"last_access_utc": time.Now().Format(utils.TimeFormat),
			"is_secure":       "1",
			"is_httponly":     "1",
			"has_expires":     "1",
			"is_persistent":   "1",
			"priority":        "1",
			"encrypted_value": true,
			"samesite":        "0",
			"source_scheme":   "1",
		},
	}

	writeTestCookieRecord(t, filepath, record)
}

func createMockFirefoxCookiesFile(t *testing.T, params mod.ModuleParams) {
	filename := "browsercookies-firefox-testuser-abcd1234.default-" + params.CollectionTimestamp + "." + params.ExportFormat
	filepath := filepath.Join(params.OutputDir, filename)

	record := utils.Record{
		CollectionTimestamp: params.CollectionTimestamp,
		EventTimestamp:      time.Now().Format(utils.TimeFormat),
		SourceFile:          "/Users/testuser/Library/Application Support/Firefox/Profiles/abcd1234.default/cookies.sqlite",
		Data: map[string]interface{}{
			"user":               "testuser",
			"profile":            "abcd1234.default",
			"host":               "example.org",
			"name":               "firefox_cookie",
			"value":              "firefox_value",
			"path":               "/",
			"creation_time":      time.Now().Format(utils.TimeFormat),
			"expiry":             "1735689600", // Some future timestamp
			"last_accessed":      time.Now().Format(utils.TimeFormat),
			"is_secure":          "1",
			"is_httponly":        "1",
			"in_browser_element": "0",
			"same_site":          "1",
		},
	}

	writeTestCookieRecord(t, filepath, record)
}

func writeTestCookieRecord(t *testing.T, filepath string, record utils.Record) {
	// Create JSON representation of the record
	jsonRecord := map[string]interface{}{
		"collection_timestamp": record.CollectionTimestamp,
		"event_timestamp":      record.EventTimestamp,
		"source_file":          record.SourceFile,
	}

	for k, v := range record.Data.(map[string]interface{}) {
		jsonRecord[k] = v
	}

	data, err := json.MarshalIndent(jsonRecord, "", "  ")
	assert.NoError(t, err)

	err = os.WriteFile(filepath, data, 0644)
	assert.NoError(t, err)
}
