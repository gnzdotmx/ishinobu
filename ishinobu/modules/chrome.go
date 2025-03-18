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
	"encoding/base64"
	"encoding/json"
	"fmt"
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
		params.Logger.Error("Error listing Chrome locations: %v", err)
		return err
	}

	for _, location := range locations {
		profilesDir, err := chromeProfiles(location, m.GetName(), params)
		if err != nil {
			params.Logger.Error("Error when collecting Chrome profiles: %v", err)
		}

		for _, profile := range profilesDir {
			err = visitChromeHistory(location, profile, m.GetName(), params)
			if err != nil {
				params.Logger.Error("Error when collecting visiting Chrome history: %v", err)
			}

			err = downloadsChromeHistory(location, profile, m.GetName(), params)
			if err != nil {
				params.Logger.Error("Error when collecting downloads Chrome history: %v", err)
			}

			err = getChromeExtensions(location, profile, m.GetName(), params)
			if err != nil {
				params.Logger.Error("Error when collecting Chrome extensions %v", err)
			}

			err = getPopupChromeSettings(location, profile, m.GetName(), params)
			if err != nil {
				params.Logger.Error("Error when collecting Chrome popup settings %v", err)
			}

			err = getExtensionDomains(location, profile, m.GetName(), params)
			if err != nil {
				params.Logger.Error("Error when collecting Chrome extension domains %v", err)
			}
		}
	}
	return nil
}

func visitChromeHistory(location string, profileUsr string, moduleName string, params mod.ModuleParams) error {
	// Create a temporary folder to store history files
	ishinobuDir := "/tmp/ishinobu"
	if err := os.MkdirAll(ishinobuDir, os.ModePerm); err != nil {
		params.Logger.Debug("Failed to create directory /tmp/ishinobu", err)
		return err
	}

	outputFileName := utils.GetOutputFileName(moduleName+"-visit-"+profileUsr, params.ExportFormat, params.OutputDir)
	writer, err := utils.NewDataWriter(params.LogsDir, outputFileName, params.ExportFormat)
	if err != nil {
		return err
	}

	profile := filepath.Join(location, profileUsr, "History")
	userProfile := strings.Split(profile, "/")[len(strings.Split(profile, "/"))-1]
	dst := "/tmp/ishinobu/" + userProfile + "_chrome_history"
	err = utils.CopyFile(profile, dst)
	if err != nil {
		return fmt.Errorf("error copying file: %v", err)
	}

	query := "SELECT urls.url, urls.title, visits.visit_time, visits.from_visit, visits.transition FROM urls INNER JOIN visits ON urls.id = visits.url ORDER BY visits.visit_time DESC;"
	rows, err := utils.QuerySQLite(dst, query)
	if err != nil {
		return fmt.Errorf("error querying SQLite: %v", err)
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

		recordData["chrome_profile"] = profileUsr
		recordData["url"] = url
		recordData["title"] = title
		recordData["visit_time"] = utils.ParseChromeTimestamp(visitTime)
		recordData["from_visit"] = fromVisit
		recordData["transition"] = transition

		record := utils.Record{
			CollectionTimestamp: params.CollectionTimestamp,
			EventTimestamp:      recordData["visit_time"].(string),
			Data:                recordData,
			SourceFile:          location + "/" + profileUsr + "/History",
		}
		err = writer.WriteRecord(record)
		if err != nil {
			params.Logger.Debug("Failed to write record: %v", err)
		}
	}

	// Remove temporary folder to store collected logs
	err = os.RemoveAll(ishinobuDir)
	if err != nil {
		return fmt.Errorf("error removing directory /tmp/ishinobu: %v", err)
	}
	return nil
}

func downloadsChromeHistory(location string, profileUsr string, moduleName string, params mod.ModuleParams) error {
	profile := filepath.Join(location, profileUsr, "History")

	// Create a temporary folder to store history files
	ishinobuDir := "/tmp/ishinobu"
	if err := os.MkdirAll(ishinobuDir, os.ModePerm); err != nil {
		params.Logger.Debug("Failed to create directory /tmp/ishinobu", err)
		return err
	}

	outputFileName := utils.GetOutputFileName(moduleName+"-downloads-"+profileUsr, params.ExportFormat, params.OutputDir)
	writer, err := utils.NewDataWriter(params.LogsDir, outputFileName, params.ExportFormat)
	if err != nil {
		return err
	}

	userProfile := strings.Split(profile, "/")[len(strings.Split(profile, "/"))-1]
	dst := "/tmp/ishinobu/" + userProfile + "_download_chrome_history"
	err = utils.CopyFile(profile, dst)
	if err != nil {
		return fmt.Errorf("error copying file: %v", err)
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
		return fmt.Errorf("error querying SQLite: %v", err)
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

	// Remove temporary folder to store collected logs
	err = os.RemoveAll(ishinobuDir)
	if err != nil {
		return fmt.Errorf("error removing directory /tmp/ishinobu: %v", err)
	}
	return nil
}

func chromeProfiles(location string, moduleName string, params mod.ModuleParams) ([]string, error) {
	userProfile := strings.Split(location, "/")[2]

	// Define the path to the Local State file
	localStatePath := filepath.Join(location, "Local State")

	// Read the contents of the Local State file
	data, err := os.ReadFile(localStatePath)
	if err != nil {
		params.Logger.Debug("Failed to read Local State file: %v", err)
		return nil, err
	}

	// Unmarshal the JSON data
	var localState map[string]interface{}
	if err := json.Unmarshal(data, &localState); err != nil {
		params.Logger.Debug("Failed to parse JSON: %v", err)
		return nil, err
	}

	// Navigate to the "profile" -> "info_cache" section
	profileSection, ok := localState["profile"].(map[string]interface{})
	if !ok {
		params.Logger.Debug("No profile data found")
		return nil, err
	}

	infoCache, ok := profileSection["info_cache"].(map[string]interface{})
	if !ok {
		params.Logger.Debug("No info_cache data found")
		return nil, err
	}

	outputFileName := utils.GetOutputFileName(moduleName+"profiles", params.ExportFormat, params.OutputDir)
	writer, err := utils.NewDataWriter(params.LogsDir, outputFileName, params.ExportFormat)
	if err != nil {
		return nil, err
	}

	// Stores the list of profilesDir
	profilesDir := make([]string, 0)

	// Collect and display profile information
	for profileDir, info := range infoCache {
		profilesDir = append(profilesDir, profileDir)
		profileInfo, ok := info.(map[string]interface{})
		if !ok {
			continue
		}

		recordData := make(map[string]interface{})
		recordData["os_user_name"] = userProfile
		recordData["profile_directory"] = profileDir
		if name, ok := profileInfo["name"].(string); ok {
			recordData["name"] = name
		}
		if userName, ok := profileInfo["user_name"].(string); ok {
			recordData["user_name"] = userName
		}
		if gaiaName, ok := profileInfo["gaia_name"].(string); ok {
			recordData["gaia_name"] = gaiaName
		}
		if gaiaGivenName, ok := profileInfo["gaia_given_name"].(string); ok {
			recordData["gaia_given_name"] = gaiaGivenName
		}
		if gaiaID, ok := profileInfo["gaia_id"].(string); ok {
			recordData["gaia_id"] = gaiaID
		}
		if isConsentedPrimaryAccount, ok := profileInfo["is_consented_primary_account"].(bool); ok {
			recordData["is_consented_primary_account"] = fmt.Sprintf("%t", isConsentedPrimaryAccount)
		}
		if isEphemeral, ok := profileInfo["is_ephemeral"].(bool); ok {
			recordData["is_ephemeral"] = fmt.Sprintf("%t", isEphemeral)
		}
		if isUsingDefaultName, ok := profileInfo["is_using_default_name"].(bool); ok {
			recordData["is_using_default_name"] = fmt.Sprintf("%t", isUsingDefaultName)
		}
		if avatarIcon, ok := profileInfo["avatar_icon"].(string); ok {
			recordData["avatar_icon"] = avatarIcon
		}
		if backgroundAppsEnabled, ok := profileInfo["background_apps"].(bool); ok {
			recordData["background_apps_enabled"] = fmt.Sprintf("%t", backgroundAppsEnabled)
		}
		if gaiaPictureFileName, ok := profileInfo["gaia_picture_file_name"].(string); ok {
			recordData["gaia_picture_file_name"] = gaiaPictureFileName
		}
		if metricsBucketIndex, ok := profileInfo["metrics_bucket_index"].(float64); ok {
			recordData["metrics_bucket_index"] = fmt.Sprintf("%v", metricsBucketIndex)
		}

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

	return profilesDir, nil
}

func getChromeExtensions(location string, profileUsr string, moduleName string, params mod.ModuleParams) error {
	extensions, err := os.ReadDir(filepath.Join(location, profileUsr, "Extensions/"))
	if err != nil {
		return fmt.Errorf("failed to read directory: %v", err)
	}

	outputFileName := utils.GetOutputFileName(moduleName+"-extensions-"+profileUsr, params.ExportFormat, params.OutputDir)
	writer, err := utils.NewDataWriter(params.LogsDir, outputFileName, params.ExportFormat)
	if err != nil {
		return fmt.Errorf("failed to create data writer: %v", err)
	}

	for _, extension := range extensions {
		manifestFiles, err := utils.ListFiles(filepath.Join(location, profileUsr, "Extensions", extension.Name(), "*", "manifest.json"))
		if err != nil {
			params.Logger.Debug("Failed to list manifest files: %v", err)
		}

		if len(manifestFiles) == 0 {
			continue
		}

		data, err := os.ReadFile(manifestFiles[0])
		if err != nil {
			params.Logger.Debug("Failed to read manifest file: %v", err)
		}

		var manifest map[string]interface{}
		if err := json.Unmarshal(data, &manifest); err != nil {
			params.Logger.Debug("Failed to parse JSON: %v", err)
		}

		recordData := make(map[string]interface{})
		if name, ok := manifest["name"].(string); ok {
			recordData["name"] = name
		}
		if version, ok := manifest["version"].(string); ok {
			recordData["version"] = version
		}
		if author, ok := manifest["author"].(string); ok {
			recordData["author"] = author
		}
		if description, ok := manifest["description"].(string); ok {
			recordData["description"] = description
		}
		if permissions, ok := manifest["permissions"].([]interface{}); ok {
			recordData["permissions"] = permissions
		}
		if scripts, ok := manifest["scripts"].([]interface{}); ok {
			recordData["scripts"] = scripts
		}
		if persistent, ok := manifest["persistent"].(bool); ok {
			recordData["persistent"] = persistent
		}
		if scopes, ok := manifest["scopes"].([]interface{}); ok {
			recordData["scopes"] = scopes
		}
		if updateURL, ok := manifest["update_url"].(string); ok {
			recordData["update_url"] = updateURL
		}
		if defaultLocale, ok := manifest["default_locale"].(string); ok {
			recordData["default_locale"] = defaultLocale
		}

		record := utils.Record{
			CollectionTimestamp: params.CollectionTimestamp,
			EventTimestamp:      params.CollectionTimestamp,
			Data:                recordData,
			SourceFile:          manifestFiles[0],
		}

		err = writer.WriteRecord(record)
		if err != nil {
			params.Logger.Debug("Failed to write record: %v", err)
		}
	}

	return nil
}

func getPopupChromeSettings(location string, profileUsr string, moduleName string, params mod.ModuleParams) error {
	// read the preferences file
	preferencesFile := filepath.Join(location, profileUsr, "Preferences")
	data, err := os.ReadFile(preferencesFile)
	if err != nil {
		return fmt.Errorf("failed to read preferences file: %v", err)
	}

	// unmarshal the JSON data
	var preferences map[string]interface{}
	if err := json.Unmarshal(data, &preferences); err != nil {
		return fmt.Errorf("failed to parse JSON: %v", err)
	}

	outputFileName := utils.GetOutputFileName(moduleName+"-settings-popup-"+profileUsr, params.ExportFormat, params.OutputDir)
	writer, err := utils.NewDataWriter(params.LogsDir, outputFileName, params.ExportFormat)
	if err != nil {
		return fmt.Errorf("failed to create data writer: %v", err)
	}

	// collect and display popup settings
	recordData := make(map[string]interface{})
	for key, value := range preferences["profile"].(map[string]interface{})["content_settings"].(map[string]interface{})["exceptions"].(map[string]interface{})["popups"].(map[string]interface{}) {
		recordData["profile"] = profileUsr
		recordData["url"] = key
		recordData["setting"] = value.(map[string]interface{})["setting"]
		if value.(map[string]interface{})["last_modified"] != nil {
			recordData["last_modified"] = utils.ParseChromeTimestamp(value.(map[string]interface{})["last_modified"].(string))
		} else {
			recordData["last_modified"] = params.CollectionTimestamp
		}

		if recordData["setting"] == "1" {
			recordData["setting"] = "Allowed"
		} else {
			recordData["setting"] = "Blocked"
		}

		record := utils.Record{
			CollectionTimestamp: params.CollectionTimestamp,
			EventTimestamp:      recordData["last_modified"].(string),
			Data:                recordData,
			SourceFile:          preferencesFile,
		}

		err = writer.WriteRecord(record)
		if err != nil {
			params.Logger.Debug("Failed to write record: %v", err)
		}
	}
	return nil
}

func getExtensionDomains(location string, profileUsr string, moduleName string, params mod.ModuleParams) error {
	params.Logger.Info("Running getExtensionDomains for profile: %s", profileUsr)

	// Define possible paths to the Network State file - Chrome may store it in different locations
	possiblePaths := []string{
		filepath.Join(location, profileUsr, "Network", "Network Persistent State"),
		filepath.Join(location, profileUsr, "Network Persistent State"),
		filepath.Join(location, profileUsr, "Network", "Network State"),
	}

	// Check which path exists
	var networkStatePath string
	var exists bool

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			networkStatePath = path
			exists = true
			params.Logger.Info("Found Network State at: %s", path)
			break
		}
	}

	// If none of the paths exist, log a warning and return nil (not an error)
	if !exists {
		params.Logger.Info("Network State file not found for profile %s - this is normal for some Chrome versions", profileUsr)
		return nil
	}

	// Setup the output file
	outputFileName := utils.GetOutputFileName(moduleName+"-extension-domains-"+profileUsr, params.ExportFormat, params.OutputDir)
	writer, err := utils.NewDataWriter(params.LogsDir, outputFileName, params.ExportFormat)
	if err != nil {
		params.Logger.Error("Failed to create data writer: %v", err)
		return err
	}

	// Read the Network State file as JSON
	data, err := os.ReadFile(networkStatePath)
	if err != nil {
		params.Logger.Error("Failed to read Network State file: %v", err)
		return nil
	}

	// Parse JSON data
	var networkState map[string]interface{}
	if err := json.Unmarshal(data, &networkState); err != nil {
		params.Logger.Error("Failed to parse Network State JSON: %v", err)
		return nil
	}

	connections := []map[string]interface{}{}

	// Process the data in "net" -> "http_server_properties" section
	if netData, ok := networkState["net"].(map[string]interface{}); ok {
		if httpProps, ok := netData["http_server_properties"].(map[string]interface{}); ok {
			// Process active connections
			if servers, ok := httpProps["servers"].([]interface{}); ok {
				for _, serverInterface := range servers {
					server, ok := serverInterface.(map[string]interface{})
					if !ok {
						continue
					}

					// Process anonymization data (extension IDs)
					anonymizationArray, ok := server["anonymization"].([]interface{})
					if !ok || len(anonymizationArray) == 0 {
						continue
					}

					// Decode the anonymization data to get extension ID
					extID := decodeAnonymization(fmt.Sprintf("%v", anonymizationArray[0]))
					if extID == "" {
						continue
					}

					// Extract domain from server field
					serverStr, ok := server["server"].(string)
					if !ok {
						continue
					}

					domain := cleanServerURL(serverStr)

					// Get timestamp using the standard utils.ParseChromeTimestamp function
					var timestamp string
					if spdy, ok := server["supports_spdy"].(float64); ok {
						// Convert to string first as utils.ParseChromeTimestamp expects a string
						spdyStr := fmt.Sprintf("%d", int64(spdy))
						timestamp = utils.ParseChromeTimestamp(spdyStr)
						if timestamp == "" {
							timestamp = params.CollectionTimestamp
						}
					} else {
						timestamp = params.CollectionTimestamp
					}

					// Get extension name
					extensionName, err := getExtensionName(location, profileUsr, extID)
					if err != nil {
						extensionName = extID
					}

					// Create record
					recordData := map[string]interface{}{
						"profile":              profileUsr,
						"extension_id":         extID,
						"extension_name":       extensionName,
						"domain":               domain,
						"connection_type":      "Active",
						"last_connection_time": timestamp,
					}

					connections = append(connections, recordData)
				}
			}

			// Process broken connections
			if broken, ok := httpProps["broken_alternative_services"].([]interface{}); ok {
				for _, brokenInterface := range broken {
					brokenConn, ok := brokenInterface.(map[string]interface{})
					if !ok {
						continue
					}

					// Process anonymization data (extension IDs)
					anonymizationArray, ok := brokenConn["anonymization"].([]interface{})
					if !ok || len(anonymizationArray) == 0 {
						continue
					}

					// Decode the anonymization data to get extension ID
					extID := decodeAnonymization(fmt.Sprintf("%v", anonymizationArray[0]))
					if extID == "" {
						continue
					}

					// Extract domain from host field
					host, ok := brokenConn["host"].(string)
					if !ok {
						continue
					}

					// Get timestamp using the standard utils.ParseChromeTimestamp function
					var timestamp string
					if brokenUntil, ok := brokenConn["broken_until"].(float64); ok {
						// Convert to string first as utils.ParseChromeTimestamp expects a string
						params.Logger.Info("brokenUntil: %v", brokenUntil)
						brokenUntilStr := fmt.Sprintf("%d", int64(brokenUntil))
						timestamp = utils.ParseChromeTimestamp(brokenUntilStr)
						if timestamp == "" {
							timestamp = params.CollectionTimestamp
						}
					} else {
						timestamp = params.CollectionTimestamp
					}

					// Get extension name
					extensionName, err := getExtensionName(location, profileUsr, extID)
					if err != nil {
						extensionName = extID
					}

					// Create record
					recordData := map[string]interface{}{
						"profile":              profileUsr,
						"extension_id":         extID,
						"extension_name":       extensionName,
						"domain":               host,
						"connection_type":      "Broken",
						"last_connection_time": timestamp,
					}

					connections = append(connections, recordData)
				}
			}
		}
	}

	// Write all connections to the output file
	for _, connData := range connections {
		record := utils.Record{
			CollectionTimestamp: params.CollectionTimestamp,
			EventTimestamp:      connData["last_connection_time"].(string),
			Data:                connData,
			SourceFile:          networkStatePath,
		}

		err = writer.WriteRecord(record)
		if err != nil {
			params.Logger.Error("Failed to write record: %v", err)
		}
	}

	params.Logger.Info("getExtensionDomains completed for profile %s, found %d connection records",
		profileUsr, len(connections))

	return nil
}

// Helper function to decode the anonymization string to extract extension ID
func decodeAnonymization(anonStr string) string {
	decoded, err := base64.StdEncoding.DecodeString(anonStr)
	if err != nil {
		return ""
	}

	decodedStr := string(decoded)
	if strings.Contains(decodedStr, "chrome-extension://") {
		parts := strings.Split(decodedStr, "chrome-extension://")
		if len(parts) > 1 {
			return parts[1]
		}
	}

	return ""
}

// Helper function to clean server URL and extract domain
func cleanServerURL(serverURL string) string {
	domain := strings.Replace(serverURL, "https://", "", 1)
	domain = strings.Replace(domain, "http://", "", 1)

	// Split on colon to remove port number if present
	parts := strings.Split(domain, ":")
	if len(parts) > 0 {
		return parts[0]
	}

	return domain
}

// Helper function to get the extension name from its ID
func getExtensionName(location string, profileUsr string, extensionID string) (string, error) {
	// Try to find the manifest file for this extension
	manifestPath := filepath.Join(location, profileUsr, "Extensions", extensionID, "*", "manifest.json")
	manifestFiles, err := utils.ListFiles(manifestPath)
	if err != nil || len(manifestFiles) == 0 {
		return extensionID, fmt.Errorf("manifest not found for extension: %s", extensionID)
	}

	// Read manifest file
	data, err := os.ReadFile(manifestFiles[0])
	if err != nil {
		return extensionID, err
	}

	// Parse JSON
	var manifest map[string]interface{}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return extensionID, err
	}

	// Get name
	if name, ok := manifest["name"].(string); ok {
		return name, nil
	}

	return extensionID, nil
}
