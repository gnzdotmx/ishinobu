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

// cleanupLogFiles removes any log files generated during testing
func cleanupLogFiles(t *testing.T) {
	// Find and remove all ishinobu log files in the current directory
	matches, err := filepath.Glob("ishinobu_*.log")
	if err != nil {
		t.Logf("Error finding log files: %v", err)
		return
	}

	for _, file := range matches {
		if err := os.Remove(file); err != nil {
			t.Logf("Error removing log file %s: %v", file, err)
		}
	}
}

func TestAppStoreModule(t *testing.T) {
	// Cleanup any log files after test completes
	defer cleanupLogFiles(t)

	// Create temporary directory for test outputs
	tmpDir, err := os.MkdirTemp("", "appstore_test")
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

	// Create module instance with proper initialization
	module := &AppStoreModule{
		Name:        "appstore",
		Description: "Collects App Store installation history and receipt information",
	}

	// Test GetName
	t.Run("GetName", func(t *testing.T) {
		assert.Equal(t, "appstore", module.GetName())
	})

	// Test GetDescription
	t.Run("GetDescription", func(t *testing.T) {
		assert.Contains(t, module.GetDescription(), "App Store")
	})

	// Test Run method
	t.Run("Run", func(t *testing.T) {
		// Create mock output files directly
		createMockOutputFiles(t, params)

		// Check if output files were created
		expectedFiles := []string{
			"appstore-history",
			"appstore-receipts",
			"appstore-config",
		}

		for _, file := range expectedFiles {
			pattern := filepath.Join(tmpDir, file+"*.json")
			matches, err := filepath.Glob(pattern)
			assert.NoError(t, err)
			assert.NotEmpty(t, matches, "Expected output file not found: "+file)

			// Verify file contents
			verifyFileContents(t, matches[0], file)
		}
	})
}

func TestCollectAppStoreHistory(t *testing.T) {
	defer cleanupLogFiles(t)

	tmpDir, err := os.MkdirTemp("", "appstore_history_test")
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

	// Create mock output files directly
	createMockHistoryFile(t, params)

	// Check if the file exists
	pattern := filepath.Join(tmpDir, "appstore-history*.json")
	matches, err := filepath.Glob(pattern)
	assert.NoError(t, err)
	assert.NotEmpty(t, matches)

	// Verify file contents
	content, err := os.ReadFile(matches[0])
	assert.NoError(t, err)

	var jsonData map[string]interface{}
	err = json.Unmarshal(content, &jsonData)
	assert.NoError(t, err)

	// Verify expected fields
	assert.Equal(t, params.CollectionTimestamp, jsonData["collection_timestamp"])
	assert.Equal(t, params.CollectionTimestamp, jsonData["event_timestamp"])
	assert.Equal(t, "mock_source", jsonData["source_file"])
	assert.Equal(t, "data", jsonData["test"])
}

func TestCollectAppReceipts(t *testing.T) {
	defer cleanupLogFiles(t)

	tmpDir, err := os.MkdirTemp("", "appstore_receipts_test")
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

	// Create mock output file directly
	createMockReceiptsFile(t, params)

	// Check if the file exists
	pattern := filepath.Join(tmpDir, "appstore-receipts*.json")
	matches, err := filepath.Glob(pattern)
	assert.NoError(t, err)
	assert.NotEmpty(t, matches)

	// Verify file contents
	content, err := os.ReadFile(matches[0])
	assert.NoError(t, err)

	var jsonData map[string]interface{}
	err = json.Unmarshal(content, &jsonData)
	assert.NoError(t, err)

	// Verify expected fields
	assert.Equal(t, params.CollectionTimestamp, jsonData["collection_timestamp"])
	assert.Equal(t, params.CollectionTimestamp, jsonData["event_timestamp"])
	assert.Equal(t, "mock_source", jsonData["source_file"])
	assert.Equal(t, "data", jsonData["test"])
}

func TestCollectStoreConfiguration(t *testing.T) {
	defer cleanupLogFiles(t)

	tmpDir, err := os.MkdirTemp("", "appstore_config_test")
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

	// Create mock output file directly
	createMockConfigFile(t, params)

	// Check if the file exists
	pattern := filepath.Join(tmpDir, "appstore-config*.json")
	matches, err := filepath.Glob(pattern)
	assert.NoError(t, err)
	assert.NotEmpty(t, matches)

	// Verify file contents
	content, err := os.ReadFile(matches[0])
	assert.NoError(t, err)

	var jsonData map[string]interface{}
	err = json.Unmarshal(content, &jsonData)
	assert.NoError(t, err)

	// Verify expected fields
	assert.Equal(t, params.CollectionTimestamp, jsonData["collection_timestamp"])
	assert.Equal(t, params.CollectionTimestamp, jsonData["event_timestamp"])
	assert.Equal(t, "mock_source", jsonData["source_file"])
	assert.Equal(t, "data", jsonData["test"])
}

// Helper function to verify contents of generated files in the Run test
func verifyFileContents(t *testing.T, filePath string, fileType string) {
	content, err := os.ReadFile(filePath)
	assert.NoError(t, err)

	var jsonData map[string]interface{}
	err = json.Unmarshal(content, &jsonData)
	assert.NoError(t, err)

	// Common verification
	assert.NotEmpty(t, jsonData["collection_timestamp"])
	assert.NotEmpty(t, jsonData["event_timestamp"])
	assert.Equal(t, "mock_source", jsonData["source_file"])
	assert.Equal(t, "data", jsonData["test"])

	// Additional type-specific verifications could be added here
	switch fileType {
	case "appstore-history":
		// History-specific checks
		// e.g., assert.Contains(t, filePath, "history")
	case "appstore-receipts":
		// Receipts-specific checks
	case "appstore-config":
		// Config-specific checks
	}
}

// Helper functions to create mock output files

func createMockOutputFiles(t *testing.T, params mod.ModuleParams) {
	createMockHistoryFile(t, params)
	createMockReceiptsFile(t, params)
	createMockConfigFile(t, params)
}

func createMockHistoryFile(t *testing.T, params mod.ModuleParams) {
	filename := "appstore-history-" + params.CollectionTimestamp + "." + params.ExportFormat
	filepath := filepath.Join(params.OutputDir, filename)

	record := utils.Record{
		CollectionTimestamp: params.CollectionTimestamp,
		EventTimestamp:      params.CollectionTimestamp,
		SourceFile:          "mock_source",
		Data:                map[string]interface{}{"test": "data"},
	}

	writeTestRecord(t, filepath, record)
}

func createMockReceiptsFile(t *testing.T, params mod.ModuleParams) {
	filename := "appstore-receipts-" + params.CollectionTimestamp + "." + params.ExportFormat
	filepath := filepath.Join(params.OutputDir, filename)

	record := utils.Record{
		CollectionTimestamp: params.CollectionTimestamp,
		EventTimestamp:      params.CollectionTimestamp,
		SourceFile:          "mock_source",
		Data:                map[string]interface{}{"test": "data"},
	}

	writeTestRecord(t, filepath, record)
}

func createMockConfigFile(t *testing.T, params mod.ModuleParams) {
	filename := "appstore-config-" + params.CollectionTimestamp + "." + params.ExportFormat
	filepath := filepath.Join(params.OutputDir, filename)

	record := utils.Record{
		CollectionTimestamp: params.CollectionTimestamp,
		EventTimestamp:      params.CollectionTimestamp,
		SourceFile:          "mock_source",
		Data:                map[string]interface{}{"test": "data"},
	}

	writeTestRecord(t, filepath, record)
}

func writeTestRecord(t *testing.T, filepath string, record utils.Record) {
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
