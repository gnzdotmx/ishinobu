// This module parses the QuickLook database for each user.
// It collects information about files that have been previewed using QuickLook, including:
// - Path and filename
// - Last hit date
// - Hit count
// - File last modified date
// - Generator
// - File size
package quicklook

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"

	"github.com/gnzdotmx/ishinobu/pkg/mod"
	"github.com/gnzdotmx/ishinobu/pkg/utils"
)

type QuickLookModule struct {
	Name        string
	Description string
}

func init() {
	module := &QuickLookModule{
		Name:        "quicklook",
		Description: "Collects QuickLook cache information"}
	mod.RegisterModule(module)
}

func (m *QuickLookModule) GetName() string {
	return m.Name
}

func (m *QuickLookModule) GetDescription() string {
	return m.Description
}

func (m *QuickLookModule) Run(params mod.ModuleParams) error {
	// Create a temporary folder to store database files
	ishinobuDir := "/tmp/ishinobu"
	if err := os.MkdirAll(ishinobuDir, os.ModePerm); err != nil {
		params.Logger.Debug("Failed to create directory /tmp/ishinobu", err)
		return err
	}

	// Find all QuickLook database files
	pattern := "/private/var/folders/*/*/C/com.apple.QuickLook.thumbnailcache/index.sqlite"
	files, err := filepath.Glob(pattern)
	if err != nil {
		params.Logger.Debug("Error listing QuickLook databases: %v", err)
		return err
	}

	if len(files) == 0 {
		params.Logger.Debug("No QuickLook databases found")
		return nil
	}

	outputFileName := utils.GetOutputFileName(m.GetName(), params.ExportFormat, params.OutputDir)
	writer, err := utils.NewDataWriter(params.LogsDir, outputFileName, params.ExportFormat)
	if err != nil {
		return err
	}

	for _, file := range files {
		err = processQuickLook(file, ishinobuDir, writer, params)
		if err != nil {
			params.Logger.Debug("Error processing QuickLook database: %v", err)
		}
	}

	// Clean up temporary directory
	err = os.RemoveAll(ishinobuDir)
	if err != nil {
		return fmt.Errorf("error removing directory /tmp/ishinobu: %w", err)
	}

	return nil
}

func processQuickLook(file string, ishinobuDir string, writer utils.DataWriter, params mod.ModuleParams) error {
	// Get user ID from file path
	uid := utils.GetUsernameFromPath(file)

	// Copy database to temp location to avoid locking issues
	dbCopy := filepath.Join(ishinobuDir, filepath.Base(file))
	err := utils.CopyFile(file, dbCopy)
	if err != nil {
		return fmt.Errorf("error copying database %s: %w", file, err)
	}

	// Open database connection
	db, err := sql.Open("sqlite3", dbCopy)
	if err != nil {
		return fmt.Errorf("error opening database %s: %w", dbCopy, err)
	}
	defer db.Close()

	// Query the database
	query := `
		SELECT DISTINCT 
			k.folder, 
			k.file_name, 
			t.hit_count, 
			t.last_hit_date, 
			k.version 
		FROM (
			SELECT rowid AS f_rowid, folder, file_name, version 
			FROM files
		) k 
		LEFT JOIN thumbnails t ON t.file_id = k.f_rowid 
		ORDER BY t.hit_count DESC`

	rows, err := db.Query(query)
	if err != nil {
		return fmt.Errorf("error querying database %s: %w", dbCopy, err)
	}
	defer rows.Close()

	for rows.Next() {
		var folder, fileName string
		var hitCount, lastHitDate sql.NullInt64
		var version []byte

		err = rows.Scan(&folder, &fileName, &hitCount, &lastHitDate, &version)
		if err != nil {
			return fmt.Errorf("error scanning row: %w", err)
		}

		recordData := make(map[string]interface{})
		recordData["uid"] = uid
		recordData["path"] = folder
		recordData["name"] = fileName

		lastHitDateStr := ""
		if lastHitDate.Valid {
			lastHitDateStr = utils.ParseChromeTimestamp(fmt.Sprintf("%d", lastHitDate.Int64))
			recordData["last_hit_date"] = lastHitDateStr
		}

		if hitCount.Valid {
			recordData["hit_count"] = hitCount.Int64
		}

		// Parse the binary plist in version field
		if len(version) > 0 {
			plistData, err := utils.ParseBiPList(string(version))
			if err == nil {
				if date, ok := plistData["date"].(float64); ok {
					recordData["file_last_modified"] = utils.ParseChromeTimestamp(fmt.Sprintf("%f", date))
				}
				if gen, ok := plistData["gen"].(string); ok {
					recordData["generator"] = gen
				}
				if size, ok := plistData["size"].(float64); ok {
					recordData["file_size"] = int64(size)
				}
			} else {
				params.Logger.Debug("Error parsing binary plist: %v", err)
			}
		}

		record := utils.Record{
			CollectionTimestamp: params.CollectionTimestamp,
			EventTimestamp:      lastHitDateStr,
			Data:                recordData,
			SourceFile:          file,
		}

		err = writer.WriteRecord(record)
		if err != nil {
			return fmt.Errorf("failed to write record: %w", err)
		}
	}

	return nil
}
