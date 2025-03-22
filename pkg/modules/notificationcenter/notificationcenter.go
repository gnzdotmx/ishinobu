// This module intends to collect and parse notifications from NotificationCenter.
// The notifications are stored in a SQLite database located at /private/var/folders/*/*/0/com.apple.notificationcenter/db2/db*.
// The notifications are stored temporarily until the user clears them from the NotificationCenter.
// It collects the following information:
// - Delivered date
// - Date
// - App
// - Category
// - URL
package notificationcenter

import (
	"fmt"
	"path/filepath"

	"github.com/gnzdotmx/ishinobu/pkg/mod"
	"github.com/gnzdotmx/ishinobu/pkg/utils"
)

type NotificationCenterModule struct {
	Name        string
	Description string
}

func init() {
	module := &NotificationCenterModule{
		Name:        "notificationcenter",
		Description: "Collects and parses notifications from NotificationCenter"}
	mod.RegisterModule(module)
}

func (m *NotificationCenterModule) GetName() string {
	return m.Name
}

func (m *NotificationCenterModule) GetDescription() string {
	return m.Description
}

func (m *NotificationCenterModule) Run(params mod.ModuleParams) error {
	notifications_db_path := "/private/var/folders/*/*/0/com.apple.notificationcenter/db2/db*"
	query := "SELECT data, delivered_date FROM record ORDER BY delivered_date DESC"

	notifications_db_paths, err := filepath.Glob(notifications_db_path)
	if err != nil {
		return err
	}

	outputFileName := utils.GetOutputFileName(m.GetName(), params.ExportFormat, params.OutputDir)
	writer, err := utils.NewDataWriter(params.LogsDir, outputFileName, params.ExportFormat)
	if err != nil {
		return err
	}
	defer writer.Close()

	for _, db_path := range notifications_db_paths {
		rows, err := utils.QuerySQLite(db_path, query)
		if err != nil {
			params.Logger.Debug("Error querying SQLite: %v", err)
			continue
		}

		for rows.Next() {
			var data string
			var delivered_date string
			err := rows.Scan(&data, &delivered_date)
			if err != nil {
				params.Logger.Debug("Error scanning row: %v", err)
				continue
			}
			plistData, err := utils.ParseBiPList(data)
			if err != nil {
				params.Logger.Debug("Error parsing plist: %v", err)
				continue
			}
			recordData := make(map[string]interface{})

			parsedDeliveredDate, err := utils.ConvertCFAbsoluteTimeToDate(delivered_date)
			if err == nil {
				delivered_date = parsedDeliveredDate
			}

			strDate := fmt.Sprintf("%v", plistData["date"])
			parsedDate, err := utils.ConvertCFAbsoluteTimeToDate(strDate)
			if err != nil {
				plistData["date"] = parsedDate
			}

			recordData["delivered_date"] = delivered_date
			recordData["date"] = parsedDate
			if app, ok := plistData["app"].(string); ok {
				recordData["app"] = app
			}
			if req, ok := plistData["req"].(map[string]interface{}); ok {
				if cate, ok := req["cate"].(string); ok {
					recordData["cate"] = cate
				}
				if durl, ok := req["durl"].(string); ok {
					recordData["durl"] = durl
				}
				if iden, ok := req["iden"].(string); ok {
					recordData["iden"] = iden
				}
				if titl, ok := req["titl"].(string); ok {
					recordData["title"] = titl
				}
				if subt, ok := req["subt"].(string); ok {
					recordData["subtitle"] = subt
				}
				if body, ok := req["body"].(string); ok {
					recordData["body"] = body
				}
			}

			record := utils.Record{
				CollectionTimestamp: params.CollectionTimestamp,
				EventTimestamp:      delivered_date,
				Data:                recordData,
				SourceFile:          db_path,
			}

			err = writer.WriteRecord(record)
			if err != nil {
				params.Logger.Debug("Failed to write record: %v", err)
			}
		}
	}
	return nil
}
