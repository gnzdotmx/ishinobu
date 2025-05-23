// This module collects:
// - VSCode extensions installed for each user on disk.
// Relevant fields:
// - extension_id: The unique identifier of the extension.
// - extension_name: The name of the extension.
// - publisher: The publisher of the extension.
// - version: The version of the extension.
// - location: The location of the extension on disk.
package vscode

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gnzdotmx/ishinobu/pkg/mod"
	"github.com/gnzdotmx/ishinobu/pkg/utils"
)

var (
	errNoExtensionsDir  = errors.New("no extensions directory found")
	errReadPackageJSON  = errors.New("failed to read package.json")
	errParsePackageJSON = errors.New("failed to parse package.json")
)

type VSCodeModule struct {
	Name        string
	Description string
}

func init() {
	module := &VSCodeModule{
		Name:        "vscode",
		Description: "Collects installed VSCode extensions"}
	mod.RegisterModule(module)
}

func (m *VSCodeModule) GetName() string {
	return m.Name
}

func (m *VSCodeModule) GetDescription() string {
	return m.Description
}

func (m *VSCodeModule) Run(params mod.ModuleParams) error {
	// Check common locations for VSCode extensions
	possibleLocations := []string{
		"/Users/*/Library/Application Support/Code/User/extensions",
		"/Users/*/.vscode/extensions",
	}

	for _, locationPattern := range possibleLocations {
		locations, err := filepath.Glob(locationPattern)
		if err != nil {
			params.Logger.Error("Error listing VSCode extension locations: %v", err)
			continue
		}

		for _, location := range locations {
			err = collectVSCodeExtensions(location, m.GetName(), params)
			if err != nil {
				params.Logger.Error("Error when collecting VSCode extensions: %v", err)
			}
		}
	}
	return nil
}

func collectVSCodeExtensions(location string, moduleName string, params mod.ModuleParams) error {
	// Extract username from path
	pathParts := strings.Split(location, "/")
	var username string
	for i, part := range pathParts {
		if part == "Users" && i+1 < len(pathParts) {
			username = pathParts[i+1]
			break
		}
	}

	// Create writer for output
	outputFileName := utils.GetOutputFileName(moduleName+"-extensions-"+username, params.ExportFormat, params.OutputDir)
	writer, err := utils.NewDataWriter(params.LogsDir, outputFileName, params.ExportFormat)
	if err != nil {
		return fmt.Errorf("failed to create data writer: %w", err)
	}

	// List all extension directories
	extensions, err := os.ReadDir(location)
	if err != nil {
		return fmt.Errorf("%w: %s", errNoExtensionsDir, err.Error())
	}

	for _, extension := range extensions {
		if !extension.IsDir() {
			continue
		}

		// Look for package.json which contains extension metadata
		packageJsonPath := filepath.Join(location, extension.Name(), "package.json")
		if _, err := os.Stat(packageJsonPath); os.IsNotExist(err) {
			continue
		}

		// Read and parse package.json
		data, err := os.ReadFile(packageJsonPath)
		if err != nil {
			params.Logger.Debug("%s for extension %s: %v", errReadPackageJSON.Error(), extension.Name(), err)
			continue
		}

		var packageJson map[string]interface{}
		if err := json.Unmarshal(data, &packageJson); err != nil {
			params.Logger.Debug("%s for extension %s: %v", errParsePackageJSON.Error(), extension.Name(), err)
			continue
		}

		// Extract relevant data
		recordData := make(map[string]interface{})
		recordData["username"] = username
		recordData["extension_location"] = filepath.Join(location, extension.Name())

		// Parse the extension name from the directory (typically publisher.name format)
		nameParts := strings.Split(extension.Name(), ".")
		if len(nameParts) >= 2 {
			recordData["publisher"] = nameParts[0]
			recordData["extension_id"] = strings.Join(nameParts[1:], ".")
		} else {
			recordData["extension_id"] = extension.Name()
		}

		// Extract more details from package.json
		if name, ok := packageJson["name"].(string); ok {
			recordData["extension_name"] = name
		}
		if displayName, ok := packageJson["displayName"].(string); ok {
			recordData["display_name"] = displayName
		}
		if publisher, ok := packageJson["publisher"].(string); ok {
			recordData["publisher"] = publisher
		}
		if version, ok := packageJson["version"].(string); ok {
			recordData["version"] = version
		}
		if description, ok := packageJson["description"].(string); ok {
			recordData["description"] = description
		}
		if repository, ok := packageJson["repository"].(map[string]interface{}); ok {
			if url, ok := repository["url"].(string); ok {
				recordData["repository"] = url
			}
		} else if repository, ok := packageJson["repository"].(string); ok {
			recordData["repository"] = repository
		}

		// Write the record
		record := utils.Record{
			CollectionTimestamp: params.CollectionTimestamp,
			EventTimestamp:      params.CollectionTimestamp,
			Data:                recordData,
			SourceFile:          packageJsonPath,
		}

		err = writer.WriteRecord(record)
		if err != nil {
			params.Logger.Debug("Failed to write record: %v", err)
		}
	}

	return nil
}
