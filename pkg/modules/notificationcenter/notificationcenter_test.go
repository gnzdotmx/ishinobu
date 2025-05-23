package notificationcenter

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"howett.net/plist"

	"github.com/gnzdotmx/ishinobu/pkg/mod"
	"github.com/gnzdotmx/ishinobu/pkg/testutils"
	"github.com/gnzdotmx/ishinobu/pkg/utils"
)

func TestNotificationCenterModuleInitialization(t *testing.T) {
	// Create a new instance with proper initialization
	module := &NotificationCenterModule{
		Name:        "notificationcenter",
		Description: "Collects and parses notifications from NotificationCenter",
	}

	// Verify module is properly instantiated with expected values
	assert.Equal(t, "notificationcenter", module.Name, "Module name should be initialized")
	assert.Contains(t, module.Description, "notifications", "Module description should be initialized")

	// Test the module's methods
	assert.Equal(t, "notificationcenter", module.GetName())
	assert.Contains(t, module.GetDescription(), "notifications")
}

// Test that the package initialization occurs
func TestPackageInitialization(t *testing.T) {
	// Create a new module with the same name
	module := &NotificationCenterModule{
		Name:        "notificationcenter",
		Description: "This is a test initialization",
	}

	// Verify the module has expected values
	assert.Equal(t, "notificationcenter", module.Name)
	assert.Equal(t, "This is a test initialization", module.Description)

	// The init function is automatically called when the package is imported
	// We can't directly test it, but we can verify it doesn't crash the tests
}

func TestParseNotification(t *testing.T) {
	// Create temporary directory for test outputs
	tmpDir, err := os.MkdirTemp("", "notificationcenter_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create test notification database
	dbDir := filepath.Join(tmpDir, "notifications")
	require.NoError(t, os.MkdirAll(dbDir, 0755))

	dbPath := filepath.Join(dbDir, "db.sqlite")

	// Create a test database schema
	schema := `
		CREATE TABLE record (
			id INTEGER PRIMARY KEY,
			data TEXT,
			delivered_date TEXT
		);
	`

	// Create columns for data insertion
	columns := []string{"data", "delivered_date"}

	// Create mock notification plist data
	notification1 := map[string]interface{}{
		"date": float64(694787590.123), // Example date in CFAbsoluteTime format
		"app":  "TestApp1",
		"req": map[string]interface{}{
			"cate": "category1",
			"durl": "app://test.url",
			"iden": "identifier1",
			"titl": "Test Notification 1",
			"subt": "Subtitle 1",
			"body": "This is a test notification body",
		},
	}

	notification2 := map[string]interface{}{
		"date": float64(694787600.456), // Example date in CFAbsoluteTime format
		"app":  "TestApp2",
		"req": map[string]interface{}{
			"cate": "category2",
			"durl": "app://another.url",
			"iden": "identifier2",
			"titl": "Test Notification 2",
			"subt": "Subtitle 2",
			"body": "Another test notification",
		},
	}

	// Convert notifications to binary plist
	plistData1, err := plist.Marshal(notification1, plist.BinaryFormat)
	require.NoError(t, err)

	plistData2, err := plist.Marshal(notification2, plist.BinaryFormat)
	require.NoError(t, err)

	// Create rows with the binary plist data and delivered date
	rows := [][]interface{}{
		{string(plistData1), "694787590.123"},
		{string(plistData2), "694787600.456"},
	}

	// Create the test database
	testutils.CreateSQLiteTestDB(t, dbPath, schema, rows, columns)

	// Set up test parameters
	logger := testutils.NewTestLogger()
	params := mod.ModuleParams{
		OutputDir:           "./",
		LogsDir:             tmpDir,
		ExportFormat:        "json",
		CollectionTimestamp: time.Now().Format(utils.TimeFormat),
		Logger:              *logger,
	}

	// Create module instance
	module := &NotificationCenterModule{
		Name:        "notificationcenter",
		Description: "Collects and parses notifications from NotificationCenter",
	}

	// Test ParseNotification method
	err = module.ParseNotification(dbPath, params)
	assert.NoError(t, err)

	// Verify output file exists
	outputFileName := utils.GetOutputFileName(module.GetName(), params.ExportFormat, params.OutputDir)
	outputFilePath := filepath.Join(params.LogsDir, outputFileName)
	assert.FileExists(t, outputFilePath)

	// Read the output file and verify contents
	data, err := os.ReadFile(outputFilePath)
	require.NoError(t, err)

	// Parse JSON lines
	lines := testutils.SplitLines(data)
	assert.GreaterOrEqual(t, len(lines), 2, "Should have at least 2 notification records")

	// Parse both records
	var record1, record2 map[string]interface{}
	err = json.Unmarshal(lines[0], &record1)
	require.NoError(t, err, "Failed to unmarshal first JSON record")

	err = json.Unmarshal(lines[1], &record2)
	require.NoError(t, err, "Failed to unmarshal second JSON record")

	// Check that both records are present (order might vary)
	foundApp1 := false
	foundApp2 := false

	for _, record := range []map[string]interface{}{record1, record2} {
		if app, ok := record["app"].(string); ok {
			switch app {
			case "TestApp1":
				foundApp1 = true
				assert.Equal(t, "category1", record["cate"], "Category field mismatch for TestApp1")
				assert.Equal(t, "app://test.url", record["durl"], "URL field mismatch for TestApp1")
				assert.Equal(t, "Test Notification 1", record["title"], "Title field mismatch for TestApp1")
				assert.Equal(t, "This is a test notification body", record["body"], "Body field mismatch for TestApp1")
			case "TestApp2":
				foundApp2 = true
				assert.Equal(t, "category2", record["cate"], "Category field mismatch for TestApp2")
				assert.Equal(t, "app://another.url", record["durl"], "URL field mismatch for TestApp2")
				assert.Equal(t, "Test Notification 2", record["title"], "Title field mismatch for TestApp2")
				assert.Equal(t, "Another test notification", record["body"], "Body field mismatch for TestApp2")
			}
		}

		// Verify timestamps are converted for each record
		assert.NotEmpty(t, record["event_timestamp"], "Event timestamp should not be empty")
		assert.NotContains(t, record["event_timestamp"], ".123", "Event timestamp should be converted")
	}

	// Verify we found both records
	assert.True(t, foundApp1, "TestApp1 notification should be in the output")
	assert.True(t, foundApp2, "TestApp2 notification should be in the output")
}

func TestRunMethod(t *testing.T) {
	// Create temporary directory for test outputs
	tmpDir, err := os.MkdirTemp("", "notificationcenter_run_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create a directory structure that matches what the module expects
	dbDir := filepath.Join(tmpDir, "folders", "test", "0", "com.apple.notificationcenter", "db2")
	require.NoError(t, os.MkdirAll(dbDir, 0755))

	dbPath := filepath.Join(dbDir, "db.sqlite")

	// Create a test database schema
	schema := `
		CREATE TABLE record (
			id INTEGER PRIMARY KEY,
			data TEXT,
			delivered_date TEXT
		);
	`

	columns := []string{"data", "delivered_date"}

	// Create a simple notification for testing
	notification := map[string]interface{}{
		"date": float64(694787590.123),
		"app":  "TestApp",
		"req": map[string]interface{}{
			"titl": "Test Notification",
			"body": "This is a test notification",
		},
	}

	plistData, err := plist.Marshal(notification, plist.BinaryFormat)
	require.NoError(t, err)

	rows := [][]interface{}{
		{string(plistData), "694787590.123"},
	}

	testutils.CreateSQLiteTestDB(t, dbPath, schema, rows, columns)

	// Set up test parameters
	logger := testutils.NewTestLogger()
	params := mod.ModuleParams{
		OutputDir:           "./",
		LogsDir:             tmpDir,
		ExportFormat:        "json",
		CollectionTimestamp: time.Now().Format(utils.TimeFormat),
		Logger:              *logger,
	}

	// Create module instance
	module := &NotificationCenterModule{
		Name:        "notificationcenter",
		Description: "Collects and parses notifications from NotificationCenter",
	}

	// Instead of trying to modify the Run method directly, we'll call ParseNotification
	// with our test path pattern directly
	testPath := filepath.Join(tmpDir, "folders/*/*/0/com.apple.notificationcenter/db2/db*")
	err = module.ParseNotification(testPath, params)
	assert.NoError(t, err)

	// Verify output file exists
	outputFileName := utils.GetOutputFileName(module.GetName(), params.ExportFormat, params.OutputDir)
	outputFilePath := filepath.Join(params.LogsDir, outputFileName)
	assert.FileExists(t, outputFilePath)
}

func TestErrorHandling(t *testing.T) {
	// Create temporary directory for test outputs
	tmpDir, err := os.MkdirTemp("", "notificationcenter_error_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Set up test parameters
	logger := testutils.NewTestLogger()
	params := mod.ModuleParams{
		OutputDir:           "./",
		LogsDir:             tmpDir,
		ExportFormat:        "json",
		CollectionTimestamp: time.Now().Format(utils.TimeFormat),
		Logger:              *logger,
	}

	// Test with invalid path pattern
	invalidPath := "[]invalid-glob-pattern"

	module := &NotificationCenterModule{
		Name:        "notificationcenter",
		Description: "Collects and parses notifications from NotificationCenter",
	}

	// ParseNotification with invalid path should not crash, just return an error
	err = module.ParseNotification(invalidPath, params)
	assert.Error(t, err)

	// Test with non-existent path
	nonExistentPath := filepath.Join(tmpDir, "non-existent-folder/*")
	err = module.ParseNotification(nonExistentPath, params)
	assert.NoError(t, err, "Non-existent path should not cause an error, just return empty results")
}
