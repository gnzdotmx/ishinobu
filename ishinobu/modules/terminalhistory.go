// Description: This module collects and parses terminal histories from the following paths:
// - /Users/*/.*_history
// - /Users/*/.bash_sessions/*
// - /private/var/*/.*_history
// - /private/var/*/.bash_sessions/*
// The module parses the terminal histories and extracts the username and command executed.
package modules

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/gnzdotmx/ishinobu/ishinobu/mod"
	"github.com/gnzdotmx/ishinobu/ishinobu/utils"
)

type TerminalModule struct {
	Name        string
	Description string
}

func init() {
	module := &TerminalModule{
		Name:        "terminalhistory",
		Description: "Collects and parses terminal histories"}
	mod.RegisterModule(module)
}

func (m *TerminalModule) GetName() string {
	return m.Name
}

func (m *TerminalModule) GetDescription() string {
	return m.Description
}

func (m *TerminalModule) Run(params mod.ModuleParams) error {
	paths := []string{"/Users/*/.*_history", "/Users/*/.bash_sessions/*",
		"/private/var/*/.*_history", "/private/var/*/.bash_sessions/*"}
	var expandedPaths []string
	for _, path := range paths {
		expandedPath, err := filepath.Glob(path)
		if err != nil {
			continue
		}
		expandedPaths = append(expandedPaths, expandedPath...)
	}

	outputFileName := utils.GetOutputFileName(m.GetName(), params.ExportFormat, params.OutputDir)
	writer, err := utils.NewDataWriter(params.LogsDir, outputFileName, params.ExportFormat)
	if err != nil {
		return err
	}

	for _, path := range expandedPaths {
		username := getUsernameFromPath(path)

		file, err := os.Open(path)
		if err != nil {
			params.Logger.Debug("Error opening file: %v", err)
			continue
		}
		defer file.Close()

		r := bufio.NewReader(file)
		for {
			line, _, err := r.ReadLine()

			if len(line) > 0 {
				recordData := make(map[string]interface{})
				recordData["username"] = username
				recordData["command"] = string(line)

				record := utils.Record{
					CollectionTimestamp: params.CollectionTimestamp,
					EventTimestamp:      params.CollectionTimestamp,
					Data:                recordData,
					SourceFile:          path,
				}

				err := writer.WriteRecord(record)
				if err != nil {
					params.Logger.Debug("Error writing record: %v", err)
				}
			}

			if err != nil {
				break
			}
		}
	}
	return nil
}

func getUsernameFromPath(path string) string {
	var user string
	if strings.Contains(path, "/Users/") {
		user = strings.Split(path, "/")[2]
	} else if strings.Contains(path, "/private/var/") {
		user = strings.Split(path, "/")[3]
	}

	return user
}
