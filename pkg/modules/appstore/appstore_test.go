package appstore

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/gnzdotmx/ishinobu/pkg/mod"
	"github.com/gnzdotmx/ishinobu/pkg/testutils"
	"github.com/gnzdotmx/ishinobu/pkg/utils"
)

func TestCollectAppStoreHistory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "appstore_history_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create necessary directory structure
	appStoreDir := filepath.Join(tmpDir, "Library", "Application Support", "App Store")
	err = os.MkdirAll(appStoreDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	// Create output directory
	outputDir := filepath.Join(tmpDir, "output")
	err = os.MkdirAll(outputDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	// Create logs directory
	logsDir := filepath.Join(tmpDir, "logs")
	err = os.MkdirAll(logsDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	// Create a mock SQLite database with test data
	dbPath := filepath.Join(appStoreDir, "storeagent.db")
	err = createMockSQLiteDB(dbPath)
	if err != nil {
		t.Fatal(err)
	}

	logger := testutils.NewTestLogger()
	params := mod.ModuleParams{
		OutputDir:           outputDir,
		LogsDir:             logsDir,
		ExportFormat:        "json",
		CollectionTimestamp: time.Now().Format(utils.TimeFormat),
		Logger:              *logger,
	}

	history_writer, err := utils.NewDataWriter(params.LogsDir, "appstore-history.json", params.ExportFormat)
	if err != nil {
		t.Fatal(err)
	}

	err = collectAppStoreHistory(params, []string{appStoreDir}, history_writer)
	assert.NoError(t, err)

	// Close the writer to ensure all data is written
	err = history_writer.Close()
	assert.NoError(t, err)

	// Give the system a moment to ensure the file is written
	time.Sleep(100 * time.Millisecond)

	// Verify output file contents
	pattern := filepath.Join(params.LogsDir, "appstore-history*.json")
	matches, err := filepath.Glob(pattern)
	assert.NoError(t, err)
	assert.NotEmpty(t, matches, "Expected output file not found")

	// Read the file to ensure it exists and is not empty
	content, err := os.ReadFile(matches[0])
	assert.NoError(t, err)
	assert.NotEmpty(t, content, "Expected content in output file")

	// Load the first row of the file as a record
	var record utils.Record
	err = json.Unmarshal(content, &record)
	assert.NoError(t, err)
	assert.NotEmpty(t, record, "Expected record in output file")

	// Verify the record contents
	assert.Equal(t, params.CollectionTimestamp, record.CollectionTimestamp)
	assert.NotEmpty(t, record.EventTimestamp)
	assert.Equal(t, dbPath, record.SourceFile)
}

func TestCollectAppReceipts(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "appstore_receipts_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create necessary directories
	outputDir := filepath.Join(tmpDir, "output")
	logsDir := filepath.Join(tmpDir, "logs")
	err = os.MkdirAll(outputDir, 0755)
	if err != nil {
		t.Fatal(err)
	}
	err = os.MkdirAll(logsDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	// Create mock app with receipt
	appPath := filepath.Join(tmpDir, "TestApp.app")
	receiptPath := filepath.Join(appPath, "Contents", "_MASReceipt", "receipt")
	err = os.MkdirAll(filepath.Dir(receiptPath), 0755)
	if err != nil {
		t.Fatal(err)
	}

	// Create mock Info.plist
	infoPlistPath := filepath.Join(appPath, "Contents", "Info.plist")
	err = os.MkdirAll(filepath.Dir(infoPlistPath), 0755)
	if err != nil {
		t.Fatal(err)
	}

	// Create mock receipt file
	err = os.WriteFile(receiptPath, []byte("mock receipt data"), 0600)
	if err != nil {
		t.Fatal(err)
	}

	// Create mock Info.plist
	infoPlistData := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleIdentifier</key>
    <string>com.example.testapp</string>
    <key>CFBundleShortVersionString</key>
    <string>1.0.0</string>
</dict>
</plist>`
	err = os.WriteFile(infoPlistPath, []byte(infoPlistData), 0600)
	if err != nil {
		t.Fatal(err)
	}

	logger := testutils.NewTestLogger()
	params := mod.ModuleParams{
		OutputDir:           outputDir,
		LogsDir:             logsDir,
		ExportFormat:        "json",
		CollectionTimestamp: time.Now().Format(utils.TimeFormat),
		Logger:              *logger,
	}

	receipts_writer, err := utils.NewDataWriter(params.LogsDir, "appstore-receipts.json", params.ExportFormat)
	if err != nil {
		t.Fatal(err)
	}

	err = collectAppReceipts(params, []string{appPath}, receipts_writer)
	assert.NoError(t, err)

	// Verify output file contents
	pattern := filepath.Join(params.LogsDir, "appstore-receipts*.json")
	matches, err := filepath.Glob(pattern)
	assert.NoError(t, err)
	assert.NotEmpty(t, matches, "Expected output file not found")

	// Read the file to ensure it exists and is not empty
	content, err := os.ReadFile(matches[0])
	assert.NoError(t, err)
	assert.NotEmpty(t, content, "Expected content in output file")

	// Load the first row of the file as a record
	var record utils.Record
	err = json.Unmarshal(content, &record)
	assert.NoError(t, err)
	assert.NotEmpty(t, record, "Expected record in output file")

	// Verify the record contents
	assert.Equal(t, params.CollectionTimestamp, record.CollectionTimestamp)
	assert.NotEmpty(t, record.EventTimestamp)
	assert.Contains(t, record.SourceFile, appPath)
}

func TestCollectStoreConfiguration(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "appstore_config_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create necessary directories
	outputDir := filepath.Join(tmpDir, "output")
	logsDir := filepath.Join(tmpDir, "logs")
	err = os.MkdirAll(outputDir, 0755)
	if err != nil {
		t.Fatal(err)
	}
	err = os.MkdirAll(logsDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	// Create mock plist file
	plistPath := filepath.Join(tmpDir, "com.apple.appstore.plist")
	plistData := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>AutomaticDownloadEnabled</key>
    <true/>
    <key>AutomaticUpdateEnabled</key>
    <true/>
    <key>FreeDownloadsRequirePassword</key>
    <true/>
    <key>LastUpdateCheck</key>
    <string>2024-03-20T10:00:00Z</string>
    <key>PasswordSetting</key>
    <string>always</string>
</dict>
</plist>`
	err = os.WriteFile(plistPath, []byte(plistData), 0600)
	if err != nil {
		t.Fatal(err)
	}

	logger := testutils.NewTestLogger()
	params := mod.ModuleParams{
		OutputDir:           outputDir,
		LogsDir:             logsDir,
		ExportFormat:        "json",
		CollectionTimestamp: time.Now().Format(utils.TimeFormat),
		Logger:              *logger,
	}

	store_writer, err := utils.NewDataWriter(params.LogsDir, "appstore-config.json", params.ExportFormat)
	if err != nil {
		t.Fatal(err)
	}

	err = collectStoreConfiguration(params, []string{plistPath}, store_writer)
	assert.NoError(t, err)

	// Verify output file contents
	pattern := filepath.Join(params.LogsDir, "appstore-config*.json")
	matches, err := filepath.Glob(pattern)
	assert.NoError(t, err)
	assert.NotEmpty(t, matches, "Expected output file not found")

	// Read the file to ensure it exists and is not empty
	content, err := os.ReadFile(matches[0])
	assert.NoError(t, err)
	assert.NotEmpty(t, content, "Expected content in output file")

	// Load the first row of the file as a record
	var record utils.Record
	err = json.Unmarshal(content, &record)
	assert.NoError(t, err)
	assert.NotEmpty(t, record, "Expected record in output file")

	// Verify the record contents
	assert.Equal(t, params.CollectionTimestamp, record.CollectionTimestamp)
	assert.NotEmpty(t, record.EventTimestamp)
	assert.Equal(t, plistPath, record.SourceFile)
}

func createMockSQLiteDB(dbPath string) error {
	// Create the database file
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Create the history table
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS history (
		item_id TEXT,
		bundle_id TEXT,
		title TEXT,
		version TEXT,
		download_size INTEGER,
		purchase_date TEXT,
		download_date TEXT,
		first_launch_date TEXT,
		last_launch_date TEXT
	);`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	// Insert test data
	insertDataSQL := `
	INSERT INTO history (
		item_id, bundle_id, title, version, download_size,
		purchase_date, download_date, first_launch_date, last_launch_date
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);`

	_, err = db.Exec(insertDataSQL,
		"123456789",
		"com.example.app",
		"Test App",
		"1.0.0",
		1024,
		"2024-03-20T10:00:00Z",
		"2024-03-20T10:01:00Z",
		"2024-03-20T10:02:00Z",
		"2024-03-20T10:03:00Z",
	)
	if err != nil {
		return fmt.Errorf("failed to insert test data: %w", err)
	}

	return nil
}
