// This module reads and parses:
// - Chrome history database for each user on disk.
// - Chrome downloads database for each user on disk.
// - Chrome profiles from the Local State file.
// Relevant fields:
// - visit_time: Timestamp of the visit.
// - from_visit: ID of the previous visit (useful for tracing navigation paths).
// - transition: Type of transition (e.g., link click, typed URL).
// - target_path: Path to the downloaded file.
// - start_time: Timestamp of the download start.
// - end_time: Timestamp of the download end.
// - danger_type: Type of danger (e.g., safe, dangerous).
// - opened: Whether the file was opened after download.
// - last_modified: Timestamp of the last modification.
// - referrer: Referrer URL.
// - tab_url: URL of the tab.
// - tab_referrer_url: Referrer URL of the tab.
// - site_url: URL of the site.
// - url: URL of the download.
package modules

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/gnzdotmx/ishinobu/ishinobu/mod"
	"github.com/gnzdotmx/ishinobu/ishinobu/utils"
)

type ChromeModule struct {
	Name        string
	Description string
}

func init() {
	module := &ChromeModule{
		Name:        "chrome",
		Description: "Collects and parses chrome history, downloads, and profiles"}
	mod.RegisterModule(module)
}

func (m *ChromeModule) GetName() string {
	return m.Name
}

func (m *ChromeModule) GetDescription() string {
	return m.Description
}

func (m *ChromeModule) Run(params mod.ModuleParams) error {
	locations, err := filepath.Glob("/Users/*/Library/Application Support/Google/Chrome")
	if err != nil {
		params.Logger.Debug("Error listing Chrome locations: %v", err)
		return err
	}

	for _, location := range locations {
		err := visitChromeHistory(location, m.GetName(), params)
		if err != nil {
			params.Logger.Debug("Error when collecting visiting Chrome history: %v", err)
		}

		err = downloadsChromeHistory(location, m.GetName(), params)
		if err != nil {
			params.Logger.Debug("Error when collecting downloads Chrome history: %v", err)
		}

		err = chromeProfiles(location, m.GetName(), params)
		if err != nil {
			params.Logger.Debug("Error when collecting Chrome profiles: %v", err)
		}
	}
	return nil
}

func visitChromeHistory(location string, moduleName string, params mod.ModuleParams) error {
	profiles, err := utils.ListFiles(filepath.Join(location, "Default", "History"))
	if err != nil {
		params.Logger.Debug("Error listing Chrome history files: %v", err)
		return err
	}

	// Create a temporary folder to store history files
	ishinobuDir := "/tmp/ishinobu"
	if err := os.MkdirAll(ishinobuDir, os.ModePerm); err != nil {
		params.Logger.Debug("Failed to create directory /tmp/ishinobu", err)
		return err
	}

	outputFileName := utils.GetOutputFileName(moduleName+"visit", params.ExportFormat, params.OutputDir)
	writer, err := utils.NewDataWriter(params.LogsDir, outputFileName, params.ExportFormat)
	if err != nil {
		return err
	}
	for _, profile := range profiles {
		userProfile := strings.Split(profile, "/")[len(strings.Split(profile, "/"))-1]
		dst := "/tmp/ishinobu/" + userProfile + "_chrome_history"
		err := utils.CopyFile(profile, dst)
		if err != nil {
			params.Logger.Debug("Failed to copy file: %v", err)
			continue
		}

		query := "SELECT urls.url, urls.title, visits.visit_time, visits.from_visit, visits.transition FROM urls INNER JOIN visits ON urls.id = visits.url ORDER BY visits.visit_time DESC;"
		rows, err := utils.QuerySQLite(dst, query)
		if err != nil {
			params.Logger.Debug("Error querying SQLite: %v", err)
			continue
		}

		// Iterate over each row and create a record
		recordData := make(map[string]interface{})
		for rows.Next() {
			var url, title, visitTime, fromVisit, transition string
			err := rows.Scan(&url, &title, &visitTime, &fromVisit, &transition)
			if err != nil {
				params.Logger.Debug("Error scanning row: %v", err)
				continue
			}
			recordData["url"] = url
			recordData["title"] = title
			recordData["visit_time"] = utils.ParseChromeTimestamp(visitTime)
			recordData["from_visit"] = fromVisit
			recordData["transition"] = transition

			record := utils.Record{
				CollectionTimestamp: params.CollectionTimestamp,
				EventTimestamp:      recordData["visit_time"].(string),
				Data:                recordData,
				SourceFile:          profile,
			}
			err = writer.WriteRecord(record)
			if err != nil {
				params.Logger.Debug("Failed to write record: %v", err)
			}
		}
	}

	// Remove temporary folder to store collected logs
	err = os.RemoveAll(ishinobuDir)
	if err != nil {
		params.Logger.Debug("Failed to remove directory /tmp/ishinobu", err)
	}
	return nil
}

func downloadsChromeHistory(location string, moduleName string, params mod.ModuleParams) error {
	profiles, err := utils.ListFiles(filepath.Join(location, "Default", "History"))
	if err != nil {
		params.Logger.Debug("Error listing Chrome history files: %v", err)
		return err
	}

	// Create a temporary folder to store history files
	ishinobuDir := "/tmp/ishinobu"
	if err := os.MkdirAll(ishinobuDir, os.ModePerm); err != nil {
		params.Logger.Debug("Failed to create directory /tmp/ishinobu", err)
		return err
	}

	outputFileName := utils.GetOutputFileName(moduleName+"downloads", params.ExportFormat, params.OutputDir)
	writer, err := utils.NewDataWriter(params.LogsDir, outputFileName, params.ExportFormat)
	if err != nil {
		return err
	}
	for _, profile := range profiles {
		userProfile := strings.Split(profile, "/")[len(strings.Split(profile, "/"))-1]
		dst := "/tmp/ishinobu/" + userProfile + "_download_chrome_history"
		err := utils.CopyFile(profile, dst)
		if err != nil {
			params.Logger.Debug("Failed to copy file: %v", err)
			continue
		}

		query := `
		SELECT 
			current_path, 
			target_path, 
			start_time, 
			end_time, 
			danger_type, 
			opened,
			last_modified, 
			referrer, 
			tab_url, 
			tab_referrer_url, 
			site_url, 
			url 
		FROM downloads
    		LEFT JOIN downloads_url_chains on downloads_url_chains.id = downloads.id
		`
		rows, err := utils.QuerySQLite(dst, query)
		if err != nil {
			params.Logger.Debug("Error querying SQLite: %v", err)
			continue
		}

		// Iterate over each row and create a record
		recordData := make(map[string]interface{})
		for rows.Next() {
			var current_path, target_path, start_time, end_time, danger_type, opened, last_modified, referrer, tab_url, tab_referrer_url, site_url, url string
			err := rows.Scan(&current_path, &target_path, &start_time, &end_time, &danger_type, &opened, &last_modified, &referrer, &tab_url, &tab_referrer_url, &site_url, &url)
			if err != nil {
				params.Logger.Debug("Error scanning row: %v", err)
				continue
			}
			recordData["current_path"] = current_path
			recordData["target_path"] = target_path
			recordData["start_time"] = utils.ParseChromeTimestamp(start_time)
			recordData["end_time"] = utils.ParseChromeTimestamp(end_time)
			recordData["danger_type"] = danger_type
			recordData["opened"] = opened
			recordData["last_modified"] = utils.ParseChromeTimestamp(last_modified)
			recordData["referrer"] = referrer
			recordData["tab_url"] = tab_url
			recordData["tab_referrer_url"] = tab_referrer_url
			recordData["site_url"] = site_url
			recordData["url"] = url

			record := utils.Record{
				CollectionTimestamp: params.CollectionTimestamp,
				EventTimestamp:      recordData["start_time"].(string),
				Data:                recordData,
				SourceFile:          profile,
			}
			err = writer.WriteRecord(record)
			if err != nil {
				params.Logger.Debug("Failed to write record: %v", err)
			}
		}
	}

	// Remove temporary folder to store collected logs
	err = os.RemoveAll(ishinobuDir)
	if err != nil {
		params.Logger.Debug("Failed to remove directory /tmp/ishinobu", err)
	}
	return nil
}

func chromeProfiles(location string, moduleName string, params mod.ModuleParams) error {
	userProfile := strings.Split(location, "/")[2]

	// Define the path to the Local State file
	localStatePath := filepath.Join(location, "Local State")

	// Read the contents of the Local State file
	data, err := ioutil.ReadFile(localStatePath)
	if err != nil {
		params.Logger.Debug("Failed to read Local State file: %v", err)
		return err
	}

	// Unmarshal the JSON data
	var localState map[string]interface{}
	if err := json.Unmarshal(data, &localState); err != nil {
		params.Logger.Debug("Failed to parse JSON: %v", err)
		return err
	}

	// Navigate to the "profile" -> "info_cache" section
	profileSection, ok := localState["profile"].(map[string]interface{})
	if !ok {
		params.Logger.Debug("No profile data found")
		return err
	}

	infoCache, ok := profileSection["info_cache"].(map[string]interface{})
	if !ok {
		params.Logger.Debug("No info_cache data found")
		return err
	}

	outputFileName := utils.GetOutputFileName(moduleName+"profiles", params.ExportFormat, params.OutputDir)
	writer, err := utils.NewDataWriter(params.LogsDir, outputFileName, params.ExportFormat)
	if err != nil {
		return err
	}

	// Collect and display profile information
	for profileDir, info := range infoCache {
		profileInfo, ok := info.(map[string]interface{})
		if !ok {
			continue
		}

		recordData := make(map[string]interface{})
		recordData["os_user_name"] = userProfile
		recordData["profile_directory"] = profileDir
		recordData["name"] = profileInfo["name"].(string)
		recordData["user_name"] = profileInfo["user_name"].(string)
		recordData["gaia_name"] = profileInfo["gaia_name"].(string)
		recordData["gaia_given_name"] = profileInfo["gaia_given_name"].(string)
		recordData["gaia_id"] = profileInfo["gaia_id"].(string)
		recordData["is_consented_primary_account"] = fmt.Sprintf("%t", profileInfo["is_consented_primary_account"].(bool))
		recordData["is_ephemeral"] = fmt.Sprintf("%t", profileInfo["is_ephemeral"].(bool))
		recordData["is_using_default_name"] = fmt.Sprintf("%t", profileInfo["is_using_default_name"].(bool))
		recordData["avatar_icon"] = profileInfo["avatar_icon"].(string)
		recordData["background_apps_enabled"] = fmt.Sprintf("%t", profileInfo["background_apps"].(bool))
		recordData["gaia_picture_file_name"] = profileInfo["gaia_picture_file_name"].(string)
		recordData["metrics_bucket_index"] = fmt.Sprintf("%v", profileInfo["metrics_bucket_index"].(float64))

		record := utils.Record{
			CollectionTimestamp: params.CollectionTimestamp,
			EventTimestamp:      params.CollectionTimestamp,
			Data:                recordData,
			SourceFile:          localStatePath,
		}

		err = writer.WriteRecord(record)
		if err != nil {
			params.Logger.Debug("Failed to write record: %v", err)
		}
	}

	return nil
}
