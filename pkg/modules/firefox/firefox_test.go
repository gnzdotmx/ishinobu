package firefox

import (
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/gnzdotmx/ishinobu/pkg/mod"
	"github.com/gnzdotmx/ishinobu/pkg/modules/testutils"
	"github.com/gnzdotmx/ishinobu/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestCollectFirefoxHistory(t *testing.T) {
	// Setup test database
	testDir := t.TempDir()
	profileDir := filepath.Join(testDir, "profile123")
	err := os.MkdirAll(profileDir, 0755)
	assert.NoError(t, err)

	// Create test schema for places.sqlite - combine both tables in one schema
	schema := `
	CREATE TABLE moz_places (
		id INTEGER PRIMARY KEY,
		url TEXT,
		title TEXT,
		visit_count INTEGER,
		typed INTEGER,
		last_visit_date INTEGER,
		description TEXT
	);
	CREATE TABLE moz_historyvisits (
		id INTEGER PRIMARY KEY,
		place_id INTEGER,
		visit_date INTEGER,
		visit_type INTEGER
	);
	`

	placesDB := filepath.Join(profileDir, "places.sqlite")
	testutils.CreateSQLiteTestDB(t, placesDB, schema, nil, nil)

	// Now insert data into moz_places table
	placesRows := [][]interface{}{
		{1, "https://example.com", "Example Site", 5, 1, 13291740170497, "Example description"},
	}

	db, err := sql.Open("sqlite3", placesDB)
	assert.NoError(t, err)
	defer db.Close()

	// Insert places data
	_, err = db.Exec("INSERT INTO moz_places (id, url, title, visit_count, typed, last_visit_date, description) VALUES (?, ?, ?, ?, ?, ?, ?)",
		placesRows[0]...)
	assert.NoError(t, err)

	// Insert history visits data
	_, err = db.Exec("INSERT INTO moz_historyvisits (id, place_id, visit_date) VALUES (?, ?, ?)",
		1, 1, 13291740170497)
	assert.NoError(t, err)

	// Setup test params
	logger := testutils.NewTestLogger()

	params := mod.ModuleParams{
		Logger:              *logger,
		LogsDir:             testDir,
		OutputDir:           "./",
		ExportFormat:        "json",
		CollectionTimestamp: "2023-01-01T00:00:00Z",
	}

	// Run the function
	err = collectFirefoxHistory(profileDir, "firefox", params)
	assert.NoError(t, err)

	//find file which contains "firefox-history"
	files, err := filepath.Glob(filepath.Join(testDir, "firefox-history*"))
	assert.NoError(t, err)
	assert.NotEmpty(t, files, "Expected at least one file in output directory")

	// Load the first file
	content, err := os.ReadFile(files[0])
	assert.NoError(t, err)
	assert.NotEmpty(t, content, "Expected content in output file")

	// Unmarshal the content into a slice of records
	var records []map[string]interface{}
	err = json.Unmarshal(content, &records)
	if err != nil {
		// If not an array, try as a single JSON object
		var record map[string]interface{}
		err = json.Unmarshal(content, &record)
		assert.NoError(t, err, "Failed to parse JSON as either array or object")

		// Convert single record to slice for consistent handling
		records = []map[string]interface{}{record}
	}
	assert.NoError(t, err)
	assert.NotEmpty(t, records, "Expected records in output file")

	// Load the first row of the file as a record
	var record utils.Record
	err = json.Unmarshal(content, &record)
	assert.NoError(t, err)
	assert.NotEmpty(t, record, "Expected record in output file")

	// Verify results
	assert.GreaterOrEqual(t, len(records), 1)
	dataRecord := records[0]

	// Verify record fields
	assert.Equal(t, "profile123", dataRecord["profile"])
	assert.Equal(t, "Example Site", dataRecord["title"])
	assert.Equal(t, "https://example.com", dataRecord["url"])
}

func TestCollectFirefoxHistoryError(t *testing.T) {
	// Setup test directory with no database
	testDir := t.TempDir()
	profileDir := filepath.Join(testDir, "profile123")
	err := os.MkdirAll(profileDir, 0755)
	assert.NoError(t, err)

	// Setup test params
	logger := testutils.NewTestLogger()
	params := mod.ModuleParams{
		Logger:              *logger,
		LogsDir:             testDir,
		OutputDir:           "./",
		ExportFormat:        "json",
		CollectionTimestamp: "2023-01-01T00:00:00Z",
	}

	// Run the function (should fail because the database file doesn't exist)
	err = collectFirefoxHistory(profileDir, "firefox", params)
	assert.Error(t, err)
}

func TestCollectFirefoxDownloads(t *testing.T) {
	// Setup test database
	testDir := t.TempDir()
	profileDir := filepath.Join(testDir, "profile123")
	err := os.MkdirAll(profileDir, 0755)
	assert.NoError(t, err)

	// Create test schema for places.sqlite with both tables in one schema
	schema := `
	CREATE TABLE moz_places (
		id INTEGER PRIMARY KEY,
		url TEXT
	);
	CREATE TABLE moz_annos (
		id INTEGER PRIMARY KEY,
		place_id INTEGER,
		content TEXT,
		dateAdded INTEGER
	);
	`

	placesDB := filepath.Join(profileDir, "places.sqlite")
	testutils.CreateSQLiteTestDB(t, placesDB, schema, nil, nil)

	// Now insert data directly using SQL
	db, err := sql.Open("sqlite3", placesDB)
	assert.NoError(t, err)
	defer db.Close()

	// Insert places data
	_, err = db.Exec("INSERT INTO moz_places (id, url) VALUES (?, ?)",
		1, "https://example.com/file.zip")
	assert.NoError(t, err)

	// Insert annotations data
	_, err = db.Exec("INSERT INTO moz_annos (id, place_id, content, dateAdded) VALUES (?, ?, ?, ?)",
		1, 1, "/path/to/file.zip,{, finished:12345678,totalBytes:1024}", 13291740170497)
	assert.NoError(t, err)

	// Setup test params
	logger := testutils.NewTestLogger()

	params := mod.ModuleParams{
		Logger:              *logger,
		LogsDir:             testDir,
		OutputDir:           "./",
		ExportFormat:        "json",
		CollectionTimestamp: "2023-01-01T00:00:00Z",
	}

	// Run the function
	err = collectFirefoxDownloads(profileDir, "firefox", params)
	assert.NoError(t, err)

	// Find the file which contains "firefox-downloads"
	files, err := filepath.Glob(filepath.Join(testDir, "firefox-downloads*"))
	assert.NoError(t, err)
	assert.NotEmpty(t, files, "Expected at least one file in output directory")

	// Load the first file
	content, err := os.ReadFile(files[0])
	assert.NoError(t, err)
	assert.NotEmpty(t, content, "Expected content in output file")

	// Unmarshal the content into a slice of records
	var records []map[string]interface{}
	err = json.Unmarshal(content, &records)
	if err != nil {
		// If not an array, try as a single JSON object
		var record map[string]interface{}
		err = json.Unmarshal(content, &record)
		assert.NoError(t, err, "Failed to parse JSON as either array or object")

		// Convert single record to slice for consistent handling
		records = []map[string]interface{}{record}
	}
	assert.NoError(t, err)
	// Verify results
	assert.GreaterOrEqual(t, len(records), 1)
	record := records[0]

	// Verify record fields
	assert.Equal(t, "profile123", record["profile"])
	assert.Equal(t, "https://example.com/file.zip", record["download_url"])
	assert.Equal(t, "/path/to/file.zip", record["download_path"])
}

func TestCollectFirefoxDownloadsError(t *testing.T) {
	// Setup test directory with no database
	testDir := t.TempDir()
	profileDir := filepath.Join(testDir, "profile123")
	err := os.MkdirAll(profileDir, 0755)
	assert.NoError(t, err)

	// Setup test params
	logger := testutils.NewTestLogger()
	params := mod.ModuleParams{
		Logger:              *logger,
		LogsDir:             testDir,
		OutputDir:           "./",
		ExportFormat:        "json",
		CollectionTimestamp: "2023-01-01T00:00:00Z",
	}

	// Run the function (should fail because the database file doesn't exist)
	err = collectFirefoxDownloads(profileDir, "firefox", params)
	assert.Error(t, err)
}

func TestCollectFirefoxExtensions(t *testing.T) {
	// Setup test directory and extensions.json
	testDir := t.TempDir()
	profileDir := filepath.Join(testDir, "profile123")
	err := os.MkdirAll(profileDir, 0755)
	assert.NoError(t, err)

	// Create test extensions.json
	extensionsData := ExtensionsData{
		Addons: []Extension{
			{
				DefaultLocale: struct {
					Name        string "json:\"name\""
					Creator     string "json:\"creator\""
					Description string "json:\"description\""
					HomepageURL string "json:\"homepageURL\""
				}{
					Name:        "Test Extension",
					Creator:     "Test Creator",
					Description: "Test Description",
					HomepageURL: "https://example.com/extension",
				},
				ID:          "test-extension@example.com",
				UpdateURL:   "https://example.com/updates",
				InstallDate: 1640995200000,
				UpdateDate:  1645995200000,
				SourceURI:   "https://example.com/source",
			},
		},
	}

	extensionsJSON, err := json.Marshal(extensionsData)
	assert.NoError(t, err)

	extensionsFile := filepath.Join(profileDir, "extensions.json")
	err = os.WriteFile(extensionsFile, extensionsJSON, 0644)
	assert.NoError(t, err)

	// Setup test params
	logger := testutils.NewTestLogger()

	params := mod.ModuleParams{
		Logger:              *logger,
		LogsDir:             testDir,
		OutputDir:           "./",
		ExportFormat:        "json",
		CollectionTimestamp: "2023-01-01T00:00:00Z",
	}

	// Run the function
	err = collectFirefoxExtensions(profileDir, "firefox", params)
	assert.NoError(t, err)

	// Find the file which contains "firefox-extensions"
	files, err := filepath.Glob(filepath.Join(testDir, "firefox-extensions*"))
	assert.NoError(t, err)
	assert.NotEmpty(t, files, "Expected at least one file in output directory")

	// Load the first file
	content, err := os.ReadFile(files[0])
	assert.NoError(t, err)
	assert.NotEmpty(t, content, "Expected content in output file")

	// Unmarshal the content into a slice of records
	var records []map[string]interface{}
	err = json.Unmarshal(content, &records)
	if err != nil {
		// If not an array, try as a single JSON object
		var record map[string]interface{}
		err = json.Unmarshal(content, &record)
		assert.NoError(t, err, "Failed to parse JSON as either array or object")

		// Convert single record to slice for consistent handling
		records = []map[string]interface{}{record}
	}

	// Verify record fields
	assert.GreaterOrEqual(t, len(records), 1)
	record := records[0]

	// Verify record fields
	assert.Equal(t, "profile123", record["profile"])
	assert.Equal(t, "Test Extension", record["name"])
	assert.Equal(t, "test-extension@example.com", record["id"])
	assert.Equal(t, "Test Creator", record["creator"])
}

func TestCollectFirefoxExtensionsError(t *testing.T) {
	// Setup test directory with no extensions.json
	testDir := t.TempDir()
	profileDir := filepath.Join(testDir, "profile123")
	err := os.MkdirAll(profileDir, 0755)
	assert.NoError(t, err)

	// Setup test params
	logger := testutils.NewTestLogger()
	params := mod.ModuleParams{
		Logger:              *logger,
		LogsDir:             testDir,
		OutputDir:           "./",
		ExportFormat:        "json",
		CollectionTimestamp: "2023-01-01T00:00:00Z",
	}

	// Run the function (should fail because extensions.json doesn't exist)
	err = collectFirefoxExtensions(profileDir, "firefox", params)
	assert.Error(t, err)
}
