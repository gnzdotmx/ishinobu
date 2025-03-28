// This module collects App Store installation history and receipt information.
// It collects:
// - Installed applications and their metadata
// - Installation dates and updates
// - App Store receipts
// - Purchase history metadata
// - App bundle IDs and versions
// - Download sizes and locations
// - App Store account information (without sensitive data)
package appstore

import (
	"os"
	"path/filepath"

	"github.com/gnzdotmx/ishinobu/pkg/mod"
	"github.com/gnzdotmx/ishinobu/pkg/utils"
)

type AppStoreModule struct {
	Name        string
	Description string
}

func init() {
	module := &AppStoreModule{
		Name:        "appstore",
		Description: "Collects App Store installation history and receipt information"}
	mod.RegisterModule(module)
}

func (m *AppStoreModule) GetName() string {
	return m.Name
}

func (m *AppStoreModule) GetDescription() string {
	return m.Description
}

func (m *AppStoreModule) Run(params mod.ModuleParams) error {
	err := collectAppStoreHistory(m.GetName(), params)
	if err != nil {
		params.Logger.Debug("Error collecting App Store history: %v", err)
	}

	err = collectAppReceipts(m.GetName(), params)
	if err != nil {
		params.Logger.Debug("Error collecting app receipts: %v", err)
	}

	err = collectStoreConfiguration(m.GetName(), params)
	if err != nil {
		params.Logger.Debug("Error collecting store configuration: %v", err)
	}

	return nil
}

func collectAppStoreHistory(moduleName string, params mod.ModuleParams) error {
	outputFileName := utils.GetOutputFileName(moduleName+"-history", params.ExportFormat, params.OutputDir)
	writer, err := utils.NewDataWriter(params.LogsDir, outputFileName, params.ExportFormat)
	if err != nil {
		return err
	}

	// Paths to check for App Store history
	historyPaths := []string{
		"/Library/Application Support/App Store/",
		"/Users/*/Library/Application Support/App Store/",
	}

	for _, basePath := range historyPaths {
		paths, err := filepath.Glob(basePath)
		if err != nil {
			continue
		}

		for _, path := range paths {
			// Get username from path
			username := utils.GetUsernameFromPath(path)

			// Check history database
			historyDB := filepath.Join(path, "storeagent.db")
			if _, err := os.Stat(historyDB); err == nil {
				rows, err := utils.QuerySQLite(historyDB, `
					SELECT 
						item_id,
						bundle_id,
						title,
						version,
						download_size,
						purchase_date,
						download_date,
						first_launch_date,
						last_launch_date
					FROM history
					ORDER BY purchase_date DESC
				`)
				if err != nil {
					params.Logger.Debug("Error querying history database: %v", err)
					continue
				}

				for rows.Next() {
					var itemID, bundleID, title, version string
					var downloadSize int64
					var purchaseDate, downloadDate, firstLaunchDate, lastLaunchDate string

					err := rows.Scan(
						&itemID,
						&bundleID,
						&title,
						&version,
						&downloadSize,
						&purchaseDate,
						&downloadDate,
						&firstLaunchDate,
						&lastLaunchDate,
					)
					if err != nil {
						params.Logger.Debug("Error scanning row: %v", err)
						continue
					}

					recordData := make(map[string]interface{})
					recordData["username"] = username
					recordData["item_id"] = itemID
					recordData["bundle_id"] = bundleID
					recordData["title"] = title
					recordData["version"] = version
					recordData["download_size"] = downloadSize
					recordData["purchase_date"] = purchaseDate
					recordData["download_date"] = downloadDate
					recordData["first_launch_date"] = firstLaunchDate
					recordData["last_launch_date"] = lastLaunchDate

					record := utils.Record{
						CollectionTimestamp: params.CollectionTimestamp,
						EventTimestamp:      purchaseDate,
						Data:                recordData,
						SourceFile:          historyDB,
					}

					err = writer.WriteRecord(record)
					if err != nil {
						params.Logger.Debug("Failed to write record: %v", err)
					}
				}
				rows.Close()
			}
		}
	}

	return nil
}

func collectAppReceipts(moduleName string, params mod.ModuleParams) error {
	outputFileName := utils.GetOutputFileName(moduleName+"-receipts", params.ExportFormat, params.OutputDir)
	writer, err := utils.NewDataWriter(params.LogsDir, outputFileName, params.ExportFormat)
	if err != nil {
		return err
	}

	// Collect receipts from installed applications
	apps, err := filepath.Glob("/Applications/*.app")
	if err != nil {
		return err
	}

	for _, appPath := range apps {
		receiptPath := filepath.Join(appPath, "Contents", "_MASReceipt", "receipt")
		if _, err := os.Stat(receiptPath); err == nil {
			// Get app bundle identifier
			infoPlistPath := filepath.Join(appPath, "Contents", "Info.plist")
			infoPlistData, err := os.ReadFile(infoPlistPath)
			if err != nil {
				continue
			}

			plistData, err := utils.ParseBiPList(string(infoPlistData))
			if err != nil {
				continue
			}

			var bundleID, version string
			if val, ok := plistData["CFBundleIdentifier"].(string); ok {
				bundleID = val
			}
			if val, ok := plistData["CFBundleShortVersionString"].(string); ok {
				version = val
			}

			// Get receipt metadata (not the actual receipt content)
			receiptInfo, err := os.Stat(receiptPath)
			if err != nil {
				continue
			}

			recordData := make(map[string]interface{})
			recordData["app_path"] = appPath
			recordData["bundle_id"] = bundleID
			recordData["version"] = version
			recordData["receipt_path"] = receiptPath
			recordData["receipt_size"] = receiptInfo.Size()
			recordData["receipt_modified"] = receiptInfo.ModTime().Format(utils.TimeFormat)

			record := utils.Record{
				CollectionTimestamp: params.CollectionTimestamp,
				EventTimestamp:      receiptInfo.ModTime().Format(utils.TimeFormat),
				Data:                recordData,
				SourceFile:          receiptPath,
			}

			err = writer.WriteRecord(record)
			if err != nil {
				params.Logger.Debug("Failed to write record: %v", err)
			}
		}
	}

	return nil
}

func collectStoreConfiguration(moduleName string, params mod.ModuleParams) error {
	outputFileName := utils.GetOutputFileName(moduleName+"-config", params.ExportFormat, params.OutputDir)
	writer, err := utils.NewDataWriter(params.LogsDir, outputFileName, params.ExportFormat)
	if err != nil {
		return err
	}

	// Paths to check for App Store configuration
	configPaths := []string{
		"/Library/Preferences/com.apple.appstore.plist",
		"/Users/*/Library/Preferences/com.apple.appstore.plist",
	}

	for _, pattern := range configPaths {
		paths, err := filepath.Glob(pattern)
		if err != nil {
			continue
		}

		for _, path := range paths {
			username := utils.GetUsernameFromPath(path)

			data, err := os.ReadFile(path)
			if err != nil {
				continue
			}

			plistData, err := utils.ParseBiPList(string(data))
			if err != nil {
				continue
			}

			recordData := make(map[string]interface{})
			recordData["username"] = username
			if automaticDownloads, ok := plistData["AutomaticDownloadEnabled"].(bool); ok {
				recordData["automatic_downloads"] = automaticDownloads
			}
			if automaticUpdates, ok := plistData["AutomaticUpdateEnabled"].(bool); ok {
				recordData["automatic_updates"] = automaticUpdates
			}
			if freeDownloadsRequirePassword, ok := plistData["FreeDownloadsRequirePassword"].(bool); ok {
				recordData["free_downloads_require_password"] = freeDownloadsRequirePassword
			}
			if lastUpdateCheck, ok := plistData["LastUpdateCheck"].(string); ok {
				recordData["last_update_check"] = lastUpdateCheck
			}
			if passwordSettings, ok := plistData["PasswordSetting"].(string); ok {
				recordData["password_settings"] = passwordSettings
			}

			// Get file modification time for timestamp
			fileInfo, err := os.Stat(path)
			lastModified := ""
			if err == nil {
				lastModified = fileInfo.ModTime().Format(utils.TimeFormat)
				recordData["last_modified"] = lastModified
			}

			record := utils.Record{
				CollectionTimestamp: params.CollectionTimestamp,
				EventTimestamp:      lastModified,
				Data:                recordData,
				SourceFile:          path,
			}

			err = writer.WriteRecord(record)
			if err != nil {
				params.Logger.Debug("Failed to write record: %v", err)
			}
		}
	}

	return nil
}
