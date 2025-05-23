package browsercookies

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnzdotmx/ishinobu/pkg/mod"
	"github.com/gnzdotmx/ishinobu/pkg/testutils"
)

// TestCollectChromeCookies tests the collectChromeCookies function
func TestCollectChromeCookies(t *testing.T) {
	// Set up test environment
	tmpDir, err := os.MkdirTemp("", "chrome_cookies_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Set up mock Chrome environment
	profile := "Default"
	tmpChromeDir := "/tmp/ishinobu-Chrome-Cookies"
	os.RemoveAll(tmpChromeDir)            // Remove any existing directory that might have wrong permissions
	err = os.MkdirAll(tmpChromeDir, 0777) // Create with full permissions
	require.NoError(t, err)
	// Make sure it's writable by all users
	err = os.Chmod(tmpChromeDir, 0777)
	require.NoError(t, err)
	chromeDir := filepath.Join(tmpDir, "Chrome")
	profileDir := filepath.Join(chromeDir, profile)
	err = os.MkdirAll(profileDir, 0755)
	require.NoError(t, err)

	// Create Chrome cookies schema
	chromeSchema := `
		CREATE TABLE cookies (
			host_key TEXT, 
			name TEXT, 
			value TEXT, 
			path TEXT, 
			creation_utc TEXT, 
			expires_utc TEXT, 
			last_access_utc TEXT, 
			is_secure TEXT, 
			is_httponly TEXT, 
			has_expires TEXT, 
			is_persistent TEXT,
			priority TEXT, 
			encrypted_value TEXT, 
			samesite TEXT, 
			source_scheme TEXT
		)
	`

	// Create test data
	chromeData := [][]interface{}{
		{
			"example.com",
			"test_cookie",
			"test_value",
			"/",
			"13253461716123456",
			"13253461716123456",
			"13253461716123456",
			"1",
			"1",
			"1",
			"1",
			"1",
			"encrypted",
			"0",
			"1",
		},
	}

	chromeColumns := []string{
		"host_key", "name", "value", "path", "creation_utc", "expires_utc",
		"last_access_utc", "is_secure", "is_httponly", "has_expires",
		"is_persistent", "priority", "encrypted_value", "samesite", "source_scheme",
	}

	// Create mock cookies DB file with schema and data
	cookiesDB := filepath.Join(profileDir, "Cookies")
	testutils.CreateSQLiteTestDB(t, cookiesDB, chromeSchema, chromeData, chromeColumns)

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
	err = collectChromeCookies(chromeDir, profile, "browsercookies", params)
	require.NoError(t, err)

	// Get the actual output file
	pattern := filepath.Join(tmpDir, "browsercookies-chrome*.json")
	matches, err := filepath.Glob(pattern)
	if assert.NoError(t, err) && assert.NotEmpty(t, matches) {
		// Read and verify the output file contents
		actualContent, err := os.ReadFile(matches[0])
		require.NoError(t, err)

		var actualData map[string]interface{}
		err = json.Unmarshal(actualContent, &actualData)
		require.NoError(t, err)

		// Verify specific fields
		assert.Equal(t, profile, actualData["chrome_profile"])
		assert.Equal(t, "example.com", actualData["host_key"])
		assert.Equal(t, "test_cookie", actualData["name"])
		assert.Equal(t, "test_value", actualData["value"])
		assert.Equal(t, "/", actualData["path"])
		assert.NotEmpty(t, actualData["creation_utc"])
		assert.Equal(t, "1", actualData["is_secure"])
		assert.Equal(t, "1", actualData["is_httponly"])
		assert.Equal(t, true, actualData["encrypted_value"])
	}
}

// TestCollectFirefoxCookies tests the collectFirefoxCookies function
func TestCollectFirefoxCookies(t *testing.T) {
	// Set up test environment
	tmpDir, err := os.MkdirTemp("", "firefox_cookies_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create a temp directory for the test
	tmpFirefoxDir := "/tmp/ishinobu-Firefox-Cookies"
	os.RemoveAll(tmpFirefoxDir)            // Remove any existing directory that might have wrong permissions
	err = os.MkdirAll(tmpFirefoxDir, 0777) // Create with full permissions
	require.NoError(t, err)
	// Make sure it's writable by all users
	err = os.Chmod(tmpFirefoxDir, 0755)
	require.NoError(t, err)

	// Set up mock Firefox environment
	profile := "abcd1234.default"
	firefoxProfileDir := filepath.Join(tmpDir, "Firefox", "Profiles", profile)
	err = os.MkdirAll(firefoxProfileDir, 0755)
	require.NoError(t, err)

	// Create Firefox cookies schema
	firefoxSchema := `
		CREATE TABLE moz_cookies (
			host TEXT, 
			name TEXT, 
			value TEXT, 
			path TEXT, 
			creationTime TEXT, 
			expiry TEXT, 
			lastAccessed TEXT, 
			isSecure TEXT, 
			isHTTPOnly TEXT, 
			inBrowserElement TEXT, 
			sameSite TEXT
		)
	`

	// Create test data
	firefoxData := [][]interface{}{
		{
			"example.org",
			"firefox_cookie",
			"firefox_value",
			"/",
			"13253461716123456",
			"1735689600",
			"13253461716123456",
			"1",
			"1",
			"0",
			"1",
		},
	}

	firefoxColumns := []string{
		"host", "name", "value", "path", "creationTime", "expiry",
		"lastAccessed", "isSecure", "isHTTPOnly", "inBrowserElement", "sameSite",
	}

	// Create mock cookies DB file with schema and data
	cookiesDB := filepath.Join(firefoxProfileDir, "cookies.sqlite")
	testutils.CreateSQLiteTestDB(t, cookiesDB, firefoxSchema, firefoxData, firefoxColumns)

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
	err = collectFirefoxCookies(firefoxProfileDir, "browsercookies", params)
	require.NoError(t, err)

	// Get the actual output file
	pattern := filepath.Join(tmpDir, "browsercookies-firefox*.json")
	matches, err := filepath.Glob(pattern)
	if assert.NoError(t, err) && assert.NotEmpty(t, matches) {
		// Read and verify the output file contents
		actualContent, err := os.ReadFile(matches[0])
		require.NoError(t, err)

		var actualData map[string]interface{}
		err = json.Unmarshal(actualContent, &actualData)
		require.NoError(t, err)

		// Verify specific fields
		assert.Equal(t, profile, actualData["profile"])
		assert.Equal(t, "example.org", actualData["host"])
		assert.Equal(t, "firefox_cookie", actualData["name"])
		assert.Equal(t, "firefox_value", actualData["value"])
		assert.Equal(t, "/", actualData["path"])
		assert.NotEmpty(t, actualData["creation_time"])
		assert.Equal(t, "1735689600", actualData["expiry"])
		assert.Equal(t, "1", actualData["is_secure"])
		assert.Equal(t, "1", actualData["is_httponly"])
		assert.Equal(t, "0", actualData["in_browser_element"])
	}
}
