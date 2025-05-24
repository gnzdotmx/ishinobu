package cursor

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnzdotmx/ishinobu/pkg/mod"
	"github.com/gnzdotmx/ishinobu/pkg/testutils"
)

func TestCollectCursorExtensions(t *testing.T) {
	// Set up test environment
	tmpDir, err := os.MkdirTemp("", "cursor_extensions_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Convert to absolute path to avoid any path resolution issues
	tmpDir, err = filepath.Abs(tmpDir)
	require.NoError(t, err)

	// Create output and logs directories
	outputDir := filepath.Join(tmpDir, "output")
	logsDir := filepath.Join(tmpDir, "logs")
	err = os.MkdirAll(outputDir, 0755)
	require.NoError(t, err)
	err = os.MkdirAll(logsDir, 0755)
	require.NoError(t, err)

	// Create mock extensions directory structure
	username := "testuser"
	extensionsDir := filepath.Join(tmpDir, "Users", username, "extensions")
	err = os.MkdirAll(extensionsDir, 0755)
	require.NoError(t, err)

	// Create mock extension directories
	extensions := []struct {
		name         string
		packageJSON  map[string]interface{}
		expectedData map[string]interface{}
	}{
		{
			name: "publisher1.extension1",
			packageJSON: map[string]interface{}{
				"name":        "extension1",
				"displayName": "Extension One",
				"publisher":   "publisher1",
				"version":     "1.0.0",
				"description": "Test extension 1",
				"repository":  "https://github.com/publisher1/extension1",
			},
			expectedData: map[string]interface{}{
				"username":           username,
				"extension_location": filepath.Join(extensionsDir, "publisher1.extension1"),
				"publisher":          "publisher1",
				"extension_id":       "extension1",
			},
		},
		{
			name: "publisher2.extension2",
			packageJSON: map[string]interface{}{
				"name":        "extension2",
				"displayName": "Extension Two",
				"publisher":   "publisher2",
				"version":     "2.1.0",
				"description": "Test extension 2",
				"repository": map[string]interface{}{
					"url": "https://github.com/publisher2/extension2",
				},
			},
			expectedData: map[string]interface{}{
				"username":           username,
				"extension_location": filepath.Join(extensionsDir, "publisher2.extension2"),
				"publisher":          "publisher2",
				"extension_id":       "extension2",
			},
		},
		{
			name: "single-name-extension",
			packageJSON: map[string]interface{}{
				"name":        "singleext",
				"displayName": "Single Name Extension",
				"publisher":   "publisher3",
				"version":     "1.2.3",
				"description": "Test extension with single name",
			},
			expectedData: map[string]interface{}{
				"username":           username,
				"extension_location": filepath.Join(extensionsDir, "single-name-extension"),
				"publisher":          "publisher3",
				"extension_id":       "single-name-extension",
			},
		},
	}

	// Create the mock extension directories and package.json files
	for _, ext := range extensions {
		extDir := filepath.Join(extensionsDir, ext.name)
		err = os.MkdirAll(extDir, 0755)
		require.NoError(t, err)

		packageJSONBytes, err := json.Marshal(ext.packageJSON)
		require.NoError(t, err)

		err = os.WriteFile(filepath.Join(extDir, "package.json"), packageJSONBytes, 0600)
		require.NoError(t, err)
	}

	// Set up test parameters
	logger := testutils.NewTestLogger()
	params := mod.ModuleParams{
		OutputDir:           "./",
		LogsDir:             logsDir,
		ExportFormat:        "json",
		CollectionTimestamp: time.Now().Format(time.RFC3339),
		Logger:              *logger,
	}

	// Call the function under test
	err = collectCursorExtensions(extensionsDir, "cursor", params)
	require.NoError(t, err)

	// Read and verify the output file
	outputFile := filepath.Join(logsDir, "cursor-extensions-testuser.json")
	data, err := os.ReadFile(outputFile)
	require.NoError(t, err)

	// Split the file into lines and parse each line as a separate JSON object
	lines := strings.Split(string(data), "\n")
	var records []map[string]interface{}
	for _, line := range lines {
		if line == "" {
			continue
		}
		var record map[string]interface{}
		err = json.Unmarshal([]byte(line), &record)
		require.NoError(t, err)
		records = append(records, record)
	}

	assert.Equal(t, len(extensions), len(records), "Number of records should match number of extensions")

	// Verify each record contains expected data
	for _, ext := range extensions {
		found := false
		for _, record := range records {
			if record["extension_id"] == ext.expectedData["extension_id"] {
				found = true
				// Verify expected fields
				for key, expectedValue := range ext.expectedData {
					assert.Equal(t, expectedValue, record[key], "Field %s mismatch", key)
				}
				break
			}
		}
		assert.True(t, found, "Record for extension %s not found", ext.name)
	}
}

// TestCollectCursorExtensionsError tests error handling in collectCursorExtensions
func TestCollectCursorExtensionsError(t *testing.T) {
	// Set up test environment
	tmpDir, err := os.MkdirTemp("", "cursor_extensions_error_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create an extension directory that can't be read
	invalidDir := filepath.Join(tmpDir, "invalid-dir")
	// Instead of creating a file, let's make a directory that can't be read
	err = os.MkdirAll(invalidDir, 0755)
	require.NoError(t, err)
	// Create a file where we expect a directory to satisfy the pathExists check
	// but cause a different error when trying to read the directory
	dummyFile := filepath.Join(invalidDir, "extensions")
	err = os.WriteFile(dummyFile, []byte("not a directory"), 0600)
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

	// Call the function under test with invalid directory
	err = collectCursorExtensions(dummyFile, "cursor", params)
	assert.Error(t, err, "Should return error when extensions directory cannot be read")
	assert.Contains(t, err.Error(), errNoExtensionsDir.Error())
}

// TestCursorModuleRun tests the Run method of CursorModule
func TestCursorModuleRun(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir, err := os.MkdirTemp("", "cursor_module_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create output and logs directories
	outputDir := filepath.Join(tmpDir, "output")
	logsDir := filepath.Join(tmpDir, "logs")
	err = os.MkdirAll(outputDir, 0755)
	require.NoError(t, err)
	err = os.MkdirAll(logsDir, 0755)
	require.NoError(t, err)

	// Set up test parameters
	logger := testutils.NewTestLogger()
	params := mod.ModuleParams{
		OutputDir:           outputDir,
		LogsDir:             logsDir,
		ExportFormat:        "json",
		CollectionTimestamp: time.Now().Format(time.RFC3339),
		Logger:              *logger,
	}

	// Create and run the module
	module := &CursorModule{
		Name:        "cursor",
		Description: "Collects installed Cursor extensions",
	}

	// Run the module - it may log errors but shouldn't crash
	err = module.Run(params)
	assert.NoError(t, err)
}

// TestCursorModuleBasics tests the basic module methods
func TestCursorModuleBasics(t *testing.T) {
	// Makes sure to initialize the module with name and description
	module := &CursorModule{
		Name:        "cursor",
		Description: "Collects installed Cursor extensions",
	}

	// Verify module name and description
	assert.Equal(t, "cursor", module.GetName())
	assert.Equal(t, "Collects installed Cursor extensions", module.GetDescription())
}

// TestCursorExtractUsername tests the username extraction from path logic
func TestCursorExtractUsername(t *testing.T) {
	tests := []struct {
		path         string
		expectedUser string
		description  string
	}{
		{
			path:         "/Users/testuser/Library/Application Support/Cursor/User/extensions",
			expectedUser: "testuser",
			description:  "Standard macOS path",
		},
		{
			path:         "/Users/test.user/Library/Application Support/Cursor/User/extensions",
			expectedUser: "test.user",
			description:  "Username with period",
		},
		{
			path:         "/Users/test-user/.cursor/extensions",
			expectedUser: "test-user",
			description:  "Username with hyphen",
		},
		{
			path:         "/home/testuser/.cursor/extensions",
			expectedUser: "",
			description:  "Linux path without Users component",
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			// Parse the path manually using the same logic as in collectCursorExtensions
			pathParts := strings.Split(tt.path, "/")
			var username string
			for i, part := range pathParts {
				if part == "Users" && i+1 < len(pathParts) {
					username = pathParts[i+1]
					break
				}
			}

			assert.Equal(t, tt.expectedUser, username, "Username extraction mismatch")
		})
	}
}
