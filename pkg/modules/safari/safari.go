// This module collects and parses Safari history, downloads, and extensions.
// It collects the following information:
// - Safari history
// - Safari downloads
// - Safari extensions
package safari

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"howett.net/plist"

	"github.com/gnzdotmx/ishinobu/pkg/mod"
	"github.com/gnzdotmx/ishinobu/pkg/utils"
)

type SafariModule struct {
	Name        string
	Description string
}

func init() {
	module := &SafariModule{
		Name:        "safari",
		Description: "Collects and parses safari history, downloads, and extensions"}
	mod.RegisterModule(module)
}

func (m *SafariModule) GetName() string {
	return m.Name
}

func (m *SafariModule) GetDescription() string {
	return m.Description
}

func (m *SafariModule) Run(params mod.ModuleParams) error {
	locations, err := filepath.Glob("/Users/*/Library/Safari")
	if err != nil {
		params.Logger.Debug("Error listing Safari locations: %v", err)
		return err
	}

	// Create temporary directory
	ishinobuDir := "/tmp/ishinobu"
	if err := os.MkdirAll(ishinobuDir, os.ModePerm); err != nil {
		params.Logger.Debug("Failed to create directory /tmp/ishinobu", err)
		return err
	}

	for _, location := range locations {
		userProfile := strings.Split(location, "/")[2]
		outputFileNameVisit := utils.GetOutputFileName(m.GetName()+"-visit-"+userProfile, params.ExportFormat, params.OutputDir)
		writerVisit, err := utils.NewDataWriter(params.LogsDir, outputFileNameVisit, params.ExportFormat)
		if err != nil {
			return err
		}
		err = visitSafariHistory(location, ishinobuDir, writerVisit, params)
		if err != nil {
			params.Logger.Debug("Error when collecting Safari history: %v", err)
		}

		outputFileNameDownloads := utils.GetOutputFileName(m.GetName()+"-downloads-"+userProfile, params.ExportFormat, params.OutputDir)
		writerDownloads, err := utils.NewDataWriter(params.LogsDir, outputFileNameDownloads, params.ExportFormat)
		if err != nil {
			return err
		}
		err = downloadsSafariHistory(location, writerDownloads, params)
		if err != nil {
			params.Logger.Debug("Error when collecting Safari downloads: %v", err)
		}

		outputFileNameExtensions := utils.GetOutputFileName(m.GetName()+"-extensions-"+userProfile, params.ExportFormat, params.OutputDir)
		writerExtensions, err := utils.NewDataWriter(params.LogsDir, outputFileNameExtensions, params.ExportFormat)
		if err != nil {
			return err
		}
		err = getSafariExtensions(location, writerExtensions, params)
		if err != nil {
			params.Logger.Debug("Error when collecting Safari extensions: %v", err)
		}
	}

	// Remove temporary folder
	err = os.RemoveAll(ishinobuDir)
	if err != nil {
		return fmt.Errorf("error removing directory /tmp/ishinobu: %w", err)
	}

	return nil
}

func visitSafariHistory(location string, ishinobuDir string, writer utils.DataWriter, params mod.ModuleParams) error {

	userProfile := strings.Split(location, "/")[2]

	historyDB := filepath.Join(location, "History.db")
	dst := filepath.Join(ishinobuDir, "safari_history")
	err := utils.CopyFile(historyDB, dst)
	if err != nil {
		return fmt.Errorf("error copying file: %w", err)
	}

	query := `
		SELECT 
			history_visits.visit_time,
			history_items.title,
			history_items.url,
			history_items.visit_count
		FROM history_visits
		LEFT JOIN history_items ON history_items.id = history_visits.history_item
	`

	rows, err := utils.QuerySQLite(dst, query)
	if err != nil {
		return fmt.Errorf("error querying SQLite: %w", err)
	}

	// Parse recently closed tabs
	recentlyClosedFile := filepath.Join(location, "RecentlyClosedTabs.plist")
	recentlyClosed := make(map[string][]string)
	if data, err := os.ReadFile(recentlyClosedFile); err == nil {
		var plistData interface{}
		if _, err := plist.Unmarshal(data, &plistData); err == nil {
			if states, ok := plistData.(map[string]interface{})["ClosedTabOrWindowPersistentStates"].([]interface{}); ok {
				for _, state := range states {
					if persistentState, ok := state.(map[string]interface{})["PersistentState"].(map[string]interface{}); ok {
						if url, ok := persistentState["tabURL"].(string); ok {
							tabTitle, ok := persistentState["TabTitle"].(string)
							if !ok {
								params.Logger.Debug("TabTitle not found")
							}

							dateClosedStr, ok := persistentState["DateClosed"].(string)
							if !ok {
								params.Logger.Debug("DateClosed not found")
							}

							dateClosed, err := utils.ParseTimestamp(dateClosedStr)
							if err != nil {
								params.Logger.Debug("Error parsing timestamp: %v", err)
							}

							lastVisitTime := ""
							if lvt, ok := persistentState["LastVisitTime"]; ok {
								lastVisitTime = utils.ParseChromeTimestamp(fmt.Sprintf("%v", lvt))
							}

							recentlyClosed[url] = []string{tabTitle, dateClosed, lastVisitTime}
						}
					}
				}
			}
		}
	}

	// Iterate over each row and create a record
	recordData := make(map[string]interface{})
	for rows.Next() {
		var visitTime, title, url string
		var visitCount int
		err := rows.Scan(&visitTime, &title, &url, &visitCount)
		if err != nil {
			params.Logger.Debug("Error scanning row: %v", err)
			continue
		}

		visitTime = utils.ParseChromeTimestamp(visitTime)

		recordData["user"] = userProfile
		recordData["visit_time"] = visitTime
		recordData["title"] = title
		recordData["url"] = url
		recordData["visit_count"] = visitCount

		// Add recently closed data if available
		if closedData, ok := recentlyClosed[url]; ok {
			recordData["recently_closed"] = "Yes"
			recordData["tab_title"] = closedData[0]
			recordData["date_closed"] = closedData[1]
			recordData["last_visit_time"] = closedData[2]
		}

		record := utils.Record{
			CollectionTimestamp: params.CollectionTimestamp,
			EventTimestamp:      visitTime,
			Data:                recordData,
			SourceFile:          historyDB,
		}
		err = writer.WriteRecord(record)
		if err != nil {
			params.Logger.Debug("Failed to write record: %v", err)
		}
	}

	return nil
}

func downloadsSafariHistory(location string, writer utils.DataWriter, params mod.ModuleParams) error {
	userProfile := strings.Split(location, "/")[2]
	// Read Downloads.plist
	downloadsFile := filepath.Join(location, "Downloads.plist")
	data, err := os.ReadFile(downloadsFile)
	if err != nil {
		return fmt.Errorf("failed to read Downloads.plist: %w", err)
	}

	var plistData interface{}
	if _, err := plist.Unmarshal(data, &plistData); err != nil {
		return fmt.Errorf("failed to parse Downloads.plist: %w", err)
	}

	// Parse download history
	if downloads, ok := plistData.(map[string]interface{})["DownloadHistory"].([]interface{}); ok {
		for _, download := range downloads {
			if entry, ok := download.(map[string]interface{}); ok {
				recordData := make(map[string]interface{})
				recordData["user"] = userProfile
				recordData["download_url"] = entry["DownloadEntryURL"]
				recordData["download_path"] = entry["DownloadEntryPath"]

				downloadStarted, err := utils.ParseTimestamp(fmt.Sprintf("%v", entry["DownloadEntryDateAddedKey"]))
				if err != nil {
					params.Logger.Debug("Error parsing timestamp: %v", err)
				}
				recordData["download_started"] = downloadStarted

				recordData["download_finished"], err = utils.ParseTimestamp(fmt.Sprintf("%v", entry["DownloadEntryDateFinishedKey"]))
				if err != nil {
					params.Logger.Debug("Error parsing timestamp: %v", err)
				}

				recordData["download_totalbytes"] = entry["DownloadEntryProgressTotalToLoad"]
				recordData["download_bytes_received"] = entry["DownloadEntryProgressBytesSoFar"]

				record := utils.Record{
					CollectionTimestamp: params.CollectionTimestamp,
					EventTimestamp:      downloadStarted,
					Data:                recordData,
					SourceFile:          downloadsFile,
				}
				err = writer.WriteRecord(record)
				if err != nil {
					params.Logger.Debug("Failed to write record: %v", err)
				}
			}
		}
	}

	return nil
}

func getSafariExtensions(location string, writer utils.DataWriter, params mod.ModuleParams) error {
	userProfile := strings.Split(location, "/")[2]

	// Read Extensions.plist
	extensionsDir := filepath.Join(location, "Extensions")
	extensionsFile := filepath.Join(extensionsDir, "Extensions.plist")
	data, err := os.ReadFile(extensionsFile)
	if err != nil {
		return fmt.Errorf("failed to read Extensions.plist: %w", err)
	}

	var plistData interface{}
	if _, err := plist.Unmarshal(data, &plistData); err != nil {
		return fmt.Errorf("failed to parse Extensions.plist: %w", err)
	}

	// Parse installed extensions
	if extensions, ok := plistData.(map[string]interface{})["Installed Extensions"].([]interface{}); ok {
		for _, ext := range extensions {
			if extension, ok := ext.(map[string]interface{}); ok {
				recordData := make(map[string]interface{})
				recordData["user"] = userProfile
				recordData["name"] = extension["Archive File Name"]
				recordData["bundle_directory"] = extension["Bundle Directory Name"]
				recordData["enabled"] = extension["Enabled"]
				recordData["apple_signed"] = extension["Apple-signed"]
				recordData["developer_id"] = extension["Developer Identifier"]
				recordData["bundle_id"] = extension["Bundle Identifier"]

				// Get extension file metadata
				archiveFileName, ok := extension["Archive File Name"].(string)
				if !ok {
					params.Logger.Debug("Failed to get archive file name")
				}
				extensionFile := filepath.Join(extensionsDir, archiveFileName)
				if fileInfo, err := os.Stat(extensionFile); err == nil {
					recordData["ctime"] = fileInfo.ModTime().Format(time.RFC3339)
					recordData["mtime"] = fileInfo.ModTime().Format(time.RFC3339)
					recordData["atime"] = fileInfo.ModTime().Format(time.RFC3339)
					recordData["size"] = fileInfo.Size()
				}

				record := utils.Record{
					CollectionTimestamp: params.CollectionTimestamp,
					EventTimestamp:      params.CollectionTimestamp,
					Data:                recordData,
					SourceFile:          extensionsFile,
				}
				err = writer.WriteRecord(record)
				if err != nil {
					params.Logger.Debug("Failed to write record: %v", err)
				}
			}
		}
	}

	return nil
}
