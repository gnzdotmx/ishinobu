package vscode

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

// TestCollectVSCodeExtensions tests the collectVSCodeExtensions function
func TestCollectVSCodeExtensions(t *testing.T) {
	// Set up test environment
	tmpDir, err := os.MkdirTemp("", "vscode_extensions_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create a directory to store output files
	outputDir := filepath.Join(tmpDir, "output")
	err = os.MkdirAll(outputDir, 0755)
	require.NoError(t, err)

	// Create mock extensions directory structure
	username := "testuser"
	extensionsDir := filepath.Join(tmpDir, "extensions")
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
		LogsDir:             tmpDir,
		ExportFormat:        "json",
		CollectionTimestamp: time.Now().Format(time.RFC3339),
		Logger:              *logger,
	}

	// Call the function under test
	err = collectVSCodeExtensions(extensionsDir, "vscode", params)
	require.NoError(t, err)

	// Verify that output files were created
	outputFiles, err := filepath.Glob(filepath.Join(tmpDir, "vscode-extensions-*.json"))
	require.NoError(t, err)
	assert.NotEmpty(t, outputFiles, "Should have created output files")
}

// TestCollectVSCodeExtensionsError tests error handling in collectVSCodeExtensions
func TestCollectVSCodeExtensionsError(t *testing.T) {
	// Set up test environment
	tmpDir, err := os.MkdirTemp("", "vscode_extensions_error_test")
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
	err = collectVSCodeExtensions(dummyFile, "vscode", params)
	assert.Error(t, err, "Should return error when extensions directory cannot be read")
	assert.Contains(t, err.Error(), errNoExtensionsDir.Error())
}

// TestVSCodeModuleRun tests the Run method of VSCodeModule
func TestVSCodeModuleRun(t *testing.T) {
	// Since the Run method makes glob calls to the system paths,
	// we'll just test that it doesn't crash and returns no error

	// Set up test parameters
	logger := testutils.NewTestLogger()
	params := mod.ModuleParams{
		OutputDir:           os.TempDir(), // Use system temp dir
		LogsDir:             os.TempDir(),
		ExportFormat:        "json",
		CollectionTimestamp: time.Now().Format(time.RFC3339),
		Logger:              *logger,
	}

	// Create and run the module
	module := &VSCodeModule{
		Name:        "vscode",
		Description: "Collects installed VSCode extensions",
	}

	// Run the module - it may log errors but shouldn't crash
	err := module.Run(params)
	assert.NoError(t, err)
}

// TestVSCodeModuleBasics tests the basic module methods
func TestVSCodeModuleBasics(t *testing.T) {
	// Makes sure to initialize the module with name and description
	module := &VSCodeModule{
		Name:        "vscode",
		Description: "Collects installed VSCode extensions",
	}

	// Verify module name and description
	assert.Equal(t, "vscode", module.GetName())
	assert.Equal(t, "Collects installed VSCode extensions", module.GetDescription())
}

// TestVSCodeExtractUsername tests the username extraction from path logic
func TestVSCodeExtractUsername(t *testing.T) {
	tests := []struct {
		path         string
		expectedUser string
		description  string
	}{
		{
			path:         "/Users/testuser/Library/Application Support/Code/User/extensions",
			expectedUser: "testuser",
			description:  "Standard macOS path",
		},
		{
			path:         "/Users/test.user/Library/Application Support/Code/User/extensions",
			expectedUser: "test.user",
			description:  "Username with period",
		},
		{
			path:         "/Users/test-user/.vscode/extensions",
			expectedUser: "test-user",
			description:  "Username with hyphen",
		},
		{
			path:         "/home/testuser/.vscode/extensions",
			expectedUser: "",
			description:  "Linux path without Users component",
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			// Parse the path manually using the same logic as in collectVSCodeExtensions
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
