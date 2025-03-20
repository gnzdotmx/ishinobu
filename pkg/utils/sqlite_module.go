package utils

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

func QuerySQLite(dbPath string, query string) (*sql.Rows, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}

	return rows, nil
}
