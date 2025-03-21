// This module collects and parses Spotlight shortcuts data.
// It collects the following information:
// - Username
// - Shortcut
// - Display name
// - Last used
// - URL
package spotlight

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gnzdotmx/ishinobu/pkg/mod"
	"github.com/gnzdotmx/ishinobu/pkg/utils"
)

type SpotlightModule struct {
	Name        string
	Description string
}

func init() {
	module := &SpotlightModule{
		Name:        "spotlight",
		Description: "Collects and parses Spotlight shortcuts data"}
	mod.RegisterModule(module)
}

func (m *SpotlightModule) GetName() string {
	return m.Name
}

func (m *SpotlightModule) GetDescription() string {
	return m.Description
}

func (m *SpotlightModule) Run(params mod.ModuleParams) error {
	// Look for Spotlight shortcuts in both user directories and system directories
	patterns := []string{
		"/Users/*/Library/Application Support/com.apple.spotlight.Shortcuts",
		"/private/var/*/Library/Application Support/com.apple.spotlight.Shortcuts",
	}

	for _, pattern := range patterns {
		files, err := filepath.Glob(pattern)
		if err != nil {
			params.Logger.Debug("Error listing Spotlight locations: %v", err)
			continue
		}

		for _, file := range files {
			err = processSpotlightFile(file, m.GetName(), params)
			if err != nil {
				params.Logger.Debug("Error processing Spotlight file %s: %v", file, err)
			}
		}
	}

	return nil
}

func processSpotlightFile(file string, moduleName string, params mod.ModuleParams) error {
	// Extract username from path
	pathParts := strings.Split(file, "/")
	var username string
	for i, part := range pathParts {
		if part == "Users" || part == "var" {
			if i+1 < len(pathParts) {
				username = pathParts[i+1]
				break
			}
		}
	}

	// Setup output writer
	outputFileName := utils.GetOutputFileName(moduleName+"-"+username, params.ExportFormat, params.OutputDir)
	writer, err := utils.NewDataWriter(params.LogsDir, outputFileName, params.ExportFormat)
	if err != nil {
		return fmt.Errorf("failed to create data writer: %v", err)
	}

	// Read and parse the plist file
	data, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}

	spotlightData, err := utils.ParseBiPList(string(data))
	if err != nil {
		return fmt.Errorf("failed to parse plist: %v", err)
	}

	// Process each shortcut entry
	for shortcut, value := range spotlightData {
		shortcutData := value.(map[string]interface{})

		recordData := make(map[string]interface{})
		recordData["username"] = username
		recordData["shortcut"] = shortcut
		recordData["display_name"] = shortcutData["DISPLAY_NAME"]

		// Convert timestamp
		if lastUsed, ok := shortcutData["LAST_USED"].(float64); ok {
			timestamp, err := utils.ConvertCFAbsoluteTimeToDate(fmt.Sprintf("%f", lastUsed))
			if err != nil {
				params.Logger.Debug("Error converting timestamp: %v", err)
				continue
			}
			recordData["last_used"] = timestamp
		}

		recordData["url"] = shortcutData["URL"]

		record := utils.Record{
			CollectionTimestamp: params.CollectionTimestamp,
			EventTimestamp:      recordData["last_used"].(string),
			Data:                recordData,
			SourceFile:          file,
		}

		err = writer.WriteRecord(record)
		if err != nil {
			params.Logger.Debug("Failed to write record: %v", err)
		}
	}

	return nil
}
