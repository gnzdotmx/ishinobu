package utils

import (
	"database/sql"
	"os"
	"testing"
)

func TestQuerySQLite(t *testing.T) {
	// Create a temporary test database
	tempDB := "test.db"
	db, err := sql.Open("sqlite3", tempDB)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer func() {
		db.Close()
		os.Remove(tempDB) // Clean up the test database file
	}()

	// Create a test table and insert some data
	_, err = db.Exec(`
		CREATE TABLE test_table (
			id INTEGER PRIMARY KEY,
			name TEXT
		);
		INSERT INTO test_table (id, name) VALUES (1, 'test1');
		INSERT INTO test_table (id, name) VALUES (2, 'test2');
	`)
	if err != nil {
		t.Fatalf("Failed to set up test data: %v", err)
	}

	// Test cases
	tests := []struct {
		name    string
		query   string
		wantErr bool
		rows    int
	}{
		{
			name:    "Valid query",
			query:   "SELECT * FROM test_table",
			wantErr: false,
			rows:    2,
		},
		{
			name:    "Invalid query",
			query:   "SELECT * FROM nonexistent_table",
			wantErr: true,
			rows:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows, err := QuerySQLite(tempDB, tt.query)

			if (err != nil) != tt.wantErr {
				t.Errorf("QuerySQLite() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				defer rows.Close()

				// Count the actual rows
				count := 0
				for rows.Next() {
					count++
				}

				if count != tt.rows {
					t.Errorf("QuerySQLite() got %d rows, want %d", count, tt.rows)
				}
			}
		})
	}
}
