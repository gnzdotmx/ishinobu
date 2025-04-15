package testutils

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gnzdotmx/ishinobu/pkg/utils"
	_ "github.com/mattn/go-sqlite3"
)

// WriteTestRecord writes a test record to a file
func WriteTestRecord(t *testing.T, filepath string, record utils.Record) {
	// Create JSON representation of the record
	jsonRecord := map[string]interface{}{
		"collection_timestamp": record.CollectionTimestamp,
		"event_timestamp":      record.EventTimestamp,
		"source_file":          record.SourceFile,
	}

	recordMap, ok := record.Data.(map[string]interface{})
	assert.True(t, ok, "Record data should be a map")

	for k, v := range recordMap {
		jsonRecord[k] = v
	}

	data, err := json.MarshalIndent(jsonRecord, "", "  ")
	assert.NoError(t, err)

	err = os.WriteFile(filepath, data, 0600)
	assert.NoError(t, err)
}

// TestDataWriter implements utils.DataWriter for testing
type TestDataWriter struct {
	Records []utils.Record
}

func (w *TestDataWriter) WriteRecord(record utils.Record) error {
	w.Records = append(w.Records, record)
	return nil
}

func (w *TestDataWriter) Close() error {
	return nil
}

// Helper to split content into lines (handles different line endings)
func SplitLines(data []byte) [][]byte {
	var lines [][]byte
	start := 0

	for i := 0; i < len(data); i++ {
		if data[i] == '\n' {
			// Add the line (excluding the newline character)
			if i > start {
				lines = append(lines, data[start:i])
			}
			start = i + 1
		}
	}

	// Add the last line if there is one
	if start < len(data) {
		lines = append(lines, data[start:])
	}

	return lines
}

// CreateSQLiteTestDB creates a test SQLite database with schema and test data
func CreateSQLiteTestDB(t *testing.T, dbPath string, schema string, rows [][]interface{}, columns []string) {
	dir := filepath.Dir(dbPath)
	err := os.MkdirAll(dir, 0755)
	assert.NoError(t, err, "Failed to create directory for test database")

	// Remove any existing database file
	_ = os.Remove(dbPath)

	// Open/create database
	db, err := sql.Open("sqlite3", dbPath)
	assert.NoError(t, err, "Failed to open SQLite database")
	defer db.Close()

	// Execute schema
	_, err = db.Exec(schema)
	assert.NoError(t, err, "Failed to create schema")

	// If we have data to insert
	if len(rows) > 0 && len(columns) > 0 {
		// Extract table name from schema - simple approach
		tableName := ""
		schemaParts := strings.Fields(schema)
		for i, part := range schemaParts {
			if strings.ToLower(part) == "table" && i > 0 && i+1 < len(schemaParts) {
				tableName = strings.Trim(schemaParts[i+1], " \t\n()")
				break
			}
		}

		// Prepare placeholders and column names for INSERT
		placeholders := strings.Repeat("?, ", len(columns))
		placeholders = placeholders[:len(placeholders)-2] // Remove trailing ", "
		columnsList := strings.Join(columns, ", ")

		// Create INSERT statement
		insertStmt := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
			tableName, columnsList, placeholders)

		// Insert each row
		for _, row := range rows {
			_, err = db.Exec(insertStmt, row...)
			assert.NoError(t, err, "Failed to insert test data")
		}
	}
}
