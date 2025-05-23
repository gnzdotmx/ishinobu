// This module collects and parses Firefox browser history, downloads, and extensions.
// It also collects information about the Firefox profile.
// It collects the following information:
// - Firefox history
// - Firefox downloads
// - Firefox extensions
// - Firefox profile
package firefox

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gnzdotmx/ishinobu/pkg/mod"
	"github.com/gnzdotmx/ishinobu/pkg/utils"
)

type FirefoxModule struct {
	Name        string
	Description string
}

func init() {
	module := &FirefoxModule{
		Name:        "firefox",
		Description: "Collects and parses Firefox browser history, downloads, and extensions"}
	mod.RegisterModule(module)
}

func (m *FirefoxModule) GetName() string {
	return m.Name
}

func (m *FirefoxModule) GetDescription() string {
	return m.Description
}

func (m *FirefoxModule) Run(params mod.ModuleParams) error {
	// Firefox locations
	firefoxLocations, err := filepath.Glob("/Users/*/Library/Application Support/Firefox/Profiles/*")
	if err != nil {
		params.Logger.Debug("Error listing Firefox locations: %v", err)
		return err
	}

	for _, location := range firefoxLocations {
		// Parse history
		err = collectFirefoxHistory(location, m.GetName(), params)
		if err != nil {
			params.Logger.Debug("Error collecting Firefox history: %v", err)
		}

		// Parse downloads
		err = collectFirefoxDownloads(location, m.GetName(), params)
		if err != nil {
			params.Logger.Debug("Error collecting Firefox downloads: %v", err)
		}

		// Parse extensions
		err = collectFirefoxExtensions(location, m.GetName(), params)
		if err != nil {
			params.Logger.Debug("Error collecting Firefox extensions: %v", err)
		}
	}

	return nil
}

func collectFirefoxHistory(location, moduleName string, params mod.ModuleParams) error {
	// Create temporary directory
	ishinobuDir := "/tmp/ishinobu-Firefox-History"
	if err := os.MkdirAll(ishinobuDir, os.ModePerm); err != nil {
		params.Logger.Debug("Failed to create directory /tmp/ishinobu-Firefox-History", err)
		return err
	}
	// Use defer to ensure cleanup happens regardless of how the function exits
	defer func() {
		os.RemoveAll(ishinobuDir)
	}()

	profile := filepath.Base(location)
	user := strings.Split(location, "/")[2]

	outputFileName := utils.GetOutputFileName(moduleName+"-history-"+user+"-"+profile, params.ExportFormat, params.OutputDir)
	writer, err := utils.NewDataWriter(params.LogsDir, outputFileName, params.ExportFormat)
	if err != nil {
		return err
	}

	historyDB := filepath.Join(location, "places.sqlite")
	dst := "/tmp/ishinobu-Firefox-History/" + profile + "_firefox_history"
	err = utils.CopyFile(historyDB, dst)
	if err != nil {
		return fmt.Errorf("error copying file: %w", err)
	}

	query := `
		SELECT moz_historyvisits.visit_date, moz_places.title, moz_places.url,
		moz_places.visit_count, moz_places.typed, moz_places.last_visit_date,
		moz_places.description
		FROM moz_historyvisits 
		LEFT JOIN moz_places ON moz_places.id = moz_historyvisits.place_id
	`
	rows, err := utils.QuerySQLite(dst, query)
	if err != nil {
		return fmt.Errorf("error querying SQLite: %w", err)
	}
	defer rows.Close()

	recordData := make(map[string]interface{})
	for rows.Next() {
		var visitDate, title, url, visitCount, typed, lastVisitDate, description string
		err := rows.Scan(&visitDate, &title, &url, &visitCount, &typed, &lastVisitDate, &description)
		if err != nil {
			params.Logger.Debug("Error scanning row: %v", err)
			continue
		}

		vistTime := utils.ParseChromeTimestamp(visitDate)

		recordData["user"] = user
		recordData["profile"] = profile
		recordData["visit_time"] = vistTime
		recordData["title"] = title
		recordData["url"] = url
		recordData["visit_count"] = visitCount
		recordData["typed"] = typed
		recordData["last_visit_time"] = utils.ParseChromeTimestamp(lastVisitDate)
		recordData["description"] = description

		record := utils.Record{
			CollectionTimestamp: params.CollectionTimestamp,
			EventTimestamp:      vistTime,
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

func collectFirefoxDownloads(location, moduleName string, params mod.ModuleParams) error {
	// Create temporary directory
	ishinobuDir := "/tmp/ishinobu-Firefox-Downloads"
	if err := os.MkdirAll(ishinobuDir, os.ModePerm); err != nil {
		params.Logger.Debug("Failed to create directory /tmp/ishinobu-Firefox-Downloads", err)
		return err
	}
	// Use defer to ensure cleanup happens regardless of how the function exits
	defer func() {
		os.RemoveAll(ishinobuDir)
	}()

	profile := filepath.Base(location)
	user := strings.Split(location, "/")[2]

	outputFileName := utils.GetOutputFileName(moduleName+"-downloads-"+user+"-"+profile, params.ExportFormat, params.OutputDir)
	writer, err := utils.NewDataWriter(params.LogsDir, outputFileName, params.ExportFormat)
	if err != nil {
		return err
	}

	downloadsDB := filepath.Join(location, "places.sqlite")
	dst := "/tmp/ishinobu-Firefox-Downloads/" + profile + "_firefox_downloads"
	err = utils.CopyFile(downloadsDB, dst)
	if err != nil {
		return fmt.Errorf("error copying file: %w", err)
	}

	query := `
		SELECT url, group_concat(content), dateAdded 
		FROM moz_annos 
		LEFT JOIN moz_places ON moz_places.id = moz_annos.place_id 
		GROUP BY place_id
	`
	rows, err := utils.QuerySQLite(dst, query)
	if err != nil {
		return fmt.Errorf("error querying SQLite: %w", err)
	}
	defer rows.Close()

	recordData := make(map[string]interface{})
	for rows.Next() {
		var url, content, dateAdded string
		err := rows.Scan(&url, &content, &dateAdded)
		if err != nil {
			params.Logger.Debug("Error scanning row: %v", err)
			continue
		}

		contentParts := strings.Split(content, ",")
		if len(contentParts) < 4 {
			continue
		}

		downloadStarted := utils.ParseChromeTimestamp(dateAdded)

		recordData["user"] = user
		recordData["profile"] = profile
		recordData["download_url"] = url
		recordData["download_path"] = contentParts[0]
		recordData["download_started"] = downloadStarted

		// Parse finish time from content
		finishTimeParts := strings.Split(contentParts[2], ":")
		if len(finishTimeParts) > 1 {
			recordData["download_finished"] = utils.ParseChromeTimestamp(finishTimeParts[1])
		}

		// Parse total bytes from content
		totalBytesParts := strings.Split(strings.TrimRight(contentParts[3], "}"), ":")
		if len(totalBytesParts) > 1 {
			recordData["download_totalbytes"] = totalBytesParts[1]
		}

		record := utils.Record{
			CollectionTimestamp: params.CollectionTimestamp,
			EventTimestamp:      downloadStarted,
			Data:                recordData,
			SourceFile:          downloadsDB,
		}

		err = writer.WriteRecord(record)
		if err != nil {
			params.Logger.Debug("Failed to write record: %v", err)
		}
	}

	return nil
}

type Extension struct {
	DefaultLocale struct {
		Name        string `json:"name"`
		Creator     string `json:"creator"`
		Description string `json:"description"`
		HomepageURL string `json:"homepageURL"`
	} `json:"defaultLocale"`
	ID          string `json:"id"`
	UpdateURL   string `json:"updateURL"`
	InstallDate int64  `json:"installDate"`
	UpdateDate  int64  `json:"updateDate"`
	SourceURI   string `json:"sourceURI"`
}

type ExtensionsData struct {
	Addons []Extension `json:"addons"`
}

func collectFirefoxExtensions(location, moduleName string, params mod.ModuleParams) error {
	profile := filepath.Base(location)
	user := strings.Split(location, "/")[2]

	outputFileName := utils.GetOutputFileName(moduleName+"-extensions-"+user+"-"+profile, params.ExportFormat, params.OutputDir)
	writer, err := utils.NewDataWriter(params.LogsDir, outputFileName, params.ExportFormat)
	if err != nil {
		return err
	}

	extensionsFile := filepath.Join(location, "extensions.json")
	if _, err := os.Stat(extensionsFile); os.IsNotExist(err) {
		return fmt.Errorf("extensions.json not found: %w", err)
	}

	data, err := os.ReadFile(extensionsFile)
	if err != nil {
		return fmt.Errorf("error reading extensions.json: %w", err)
	}

	var extensions ExtensionsData
	if err := json.Unmarshal(data, &extensions); err != nil {
		return fmt.Errorf("error parsing extensions.json: %w", err)
	}

	recordData := make(map[string]interface{})
	for _, ext := range extensions.Addons {

		installDate := utils.ParseChromeTimestamp(fmt.Sprintf("%d", ext.InstallDate))

		recordData["user"] = user
		recordData["profile"] = profile
		recordData["name"] = ext.DefaultLocale.Name
		recordData["id"] = ext.ID
		recordData["creator"] = ext.DefaultLocale.Creator
		recordData["description"] = ext.DefaultLocale.Description
		recordData["update_url"] = ext.UpdateURL
		recordData["install_date"] = installDate
		recordData["last_updated"] = utils.ParseChromeTimestamp(fmt.Sprintf("%d", ext.UpdateDate))
		recordData["source_uri"] = ext.SourceURI
		recordData["homepage_url"] = ext.DefaultLocale.HomepageURL

		record := utils.Record{
			CollectionTimestamp: params.CollectionTimestamp,
			EventTimestamp:      installDate,
			Data:                recordData,
			SourceFile:          extensionsFile,
		}

		err = writer.WriteRecord(record)
		if err != nil {
			params.Logger.Debug("Failed to write record: %v", err)
		}
	}

	return nil
}
