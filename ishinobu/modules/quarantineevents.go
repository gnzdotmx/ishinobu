// This module collects and parses QuarantineEventsV2 database.
// It collects the following information:
// - Identifier
// - Timestamp
// - Bundle ID
// - Quarantine Agent
// - Download URL
package modules

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gnzdotmx/ishinobu/ishinobu/mod"
	"github.com/gnzdotmx/ishinobu/ishinobu/utils"
)

type QuarantineEventsModule struct {
	Name        string
	Description string
}

func init() {
	module := &QuarantineEventsModule{
		Name:        "quarantineevents",
		Description: "Collects and parses QuarantineEventsV2 database"}
	mod.RegisterModule(module)
}

func (m *QuarantineEventsModule) GetName() string {
	return m.Name
}

func (m *QuarantineEventsModule) GetDescription() string {
	return m.Description
}

func (m *QuarantineEventsModule) Run(params mod.ModuleParams) error {
	// Search for QuarantineEventsV2 databases in both user directories and private/var
	patterns := []string{
		"/Users/*/Library/Preferences/com.apple.LaunchServices.QuarantineEventsV2",
		"/private/var/*/Library/Preferences/com.apple.LaunchServices.QuarantineEventsV2",
	}

	for _, pattern := range patterns {
		files, err := filepath.Glob(pattern)
		if err != nil {
			params.Logger.Debug("Error listing QuarantineEvents locations: %v", err)
			continue
		}

		for _, dbPath := range files {
			err = processQuarantineEvents(dbPath, m.GetName(), params)
			if err != nil {
				params.Logger.Debug("Error processing QuarantineEvents database: %v", err)
			}
		}
	}

	return nil
}

func processQuarantineEvents(dbPath, moduleName string, params mod.ModuleParams) error {
	// Create a temporary folder to store database copy
	ishinobuDir := "/tmp/ishinobu"
	if err := os.MkdirAll(ishinobuDir, os.ModePerm); err != nil {
		params.Logger.Debug("Failed to create directory /tmp/ishinobu", err)
		return err
	}

	// Copy database to temporary location
	tempDB := filepath.Join(ishinobuDir, "quarantine_events")
	err := utils.CopyFile(dbPath, tempDB)
	if err != nil {
		return fmt.Errorf("error copying file: %v", err)
	}
	defer os.RemoveAll(ishinobuDir)

	// Setup output writer
	outputFileName := utils.GetOutputFileName(moduleName, params.ExportFormat, params.OutputDir)
	writer, err := utils.NewDataWriter(params.LogsDir, outputFileName, params.ExportFormat)
	if err != nil {
		return err
	}

	// Query the database
	query := `SELECT * FROM LSQuarantineEvent`
	rows, err := utils.QuerySQLite(tempDB, query)
	if err != nil {
		return fmt.Errorf("error querying SQLite: %v", err)
	}
	defer rows.Close()

	// Get username from path
	user := utils.GetUsernameFromPath(dbPath)

	// Process each row
	for rows.Next() {
		var (
			identifier      string
			timestamp       float64
			bundleID        sql.NullString
			quarantineAgent sql.NullString
			downloadURL     sql.NullString
			senderName      sql.NullString
			senderAddress   sql.NullString
			typeNo          int
			originTitle     sql.NullString
			originURL       sql.NullString
			originAlias     sql.NullString
		)

		err := rows.Scan(
			&identifier,
			&timestamp,
			&bundleID,
			&quarantineAgent,
			&downloadURL,
			&senderName,
			&senderAddress,
			&typeNo,
			&originTitle,
			&originURL,
			&originAlias,
		)
		if err != nil {
			params.Logger.Debug("Error scanning row: %v", err)
			continue
		}

		// Convert timestamp from Cocoa time to ISO format
		eventTime, err := utils.ConvertCFAbsoluteTimeToDate(fmt.Sprintf("%f", timestamp))
		if err != nil {
			params.Logger.Debug("Error converting timestamp: %v", err)
			eventTime = params.CollectionTimestamp // fallback to collection time
		}

		recordData := map[string]interface{}{
			"identifier":       identifier,
			"user":             user,
			"bundle_id":        nullStringValue(bundleID),
			"quarantine_agent": nullStringValue(quarantineAgent),
			"download_url":     nullStringValue(downloadURL),
			"sender_name":      nullStringValue(senderName),
			"sender_address":   nullStringValue(senderAddress),
			"type_no":          typeNo,
			"origin_title":     nullStringValue(originTitle),
			"origin_url":       nullStringValue(originURL),
			"origin_alias":     nullStringValue(originAlias),
		}

		record := utils.Record{
			CollectionTimestamp: params.CollectionTimestamp,
			EventTimestamp:      eventTime,
			Data:                recordData,
			SourceFile:          dbPath,
		}

		err = writer.WriteRecord(record)
		if err != nil {
			params.Logger.Debug("Failed to write record: %v", err)
		}
	}

	return nil
}

// Helper function to handle NULL values in database
func nullStringValue(s sql.NullString) string {
	if s.Valid {
		return s.String
	}
	return ""
}
