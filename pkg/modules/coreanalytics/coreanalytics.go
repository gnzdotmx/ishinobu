// This module collects and parses CoreAnalytics artifacts.
// It collects the following information:
// - Analytics files
// - Aggregate files
package coreanalytics

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gnzdotmx/ishinobu/pkg/mod"
	"github.com/gnzdotmx/ishinobu/pkg/utils"
)

type CoreAnalyticsModule struct {
	Name        string
	Description string
}

type CoreAnalyticsMessage struct {
	Message map[string]interface{} `json:"message"`
	Name    string                 `json:"name"`
	UUID    string                 `json:"uuid"`
}

type CoreAnalyticsMarker struct {
	Marker         string `json:"_marker"`
	StartTimestamp string `json:"startTimestamp"`
}

type CoreAnalyticsTimestamp struct {
	Timestamp string `json:"timestamp"`
}

func init() {
	module := &CoreAnalyticsModule{
		Name:        "coreanalytics",
		Description: "Collects and parses CoreAnalytics artifacts"}
	mod.RegisterModule(module)
}

func (m *CoreAnalyticsModule) GetName() string {
	return m.Name
}

func (m *CoreAnalyticsModule) GetDescription() string {
	return m.Description
}

func getMacOSVersion() (string, error) {
	out, err := utils.ExecuteCommand("sw_vers", "-productVersion")
	if err != nil {
		return "", fmt.Errorf("failed to get macOS version: %v", err)
	}
	return strings.TrimSpace(out), nil
}

func (m *CoreAnalyticsModule) Run(params mod.ModuleParams) error {
	// Get macOS version
	osVersion, err := getMacOSVersion()
	if err != nil {
		params.Logger.Debug("Failed to get macOS version: %v", err)
		return err
	}

	// Check for macOS version compatibility
	version := strings.Split(osVersion, ".")
	if len(version) > 1 {
		majorVer := version[0]
		minorVer := version[1]
		if majorVer == "10" && minorVer < "13" {
			params.Logger.Debug("Artifact is not present below OS version 10.13")
			return nil
		}
		if majorVer == "10" && minorVer >= "15" {
			params.Logger.Debug("Artifact contents have changed for macOS 10.15+")
			return nil
		}
	}

	// Parse Analytics files
	err = parseAnalyticsFiles(m.GetName(), params)
	if err != nil {
		return err
	}

	// Parse Aggregate files
	err = parseAggregateFiles(m.GetName(), params)
	if err != nil {
		return err
	}

	return nil
}

func parseAnalyticsFiles(moduleName string, params mod.ModuleParams) error {
	// Find all Analytics files
	patterns := []string{
		"/Library/Logs/DiagnosticReports/Analytics*.core_analytics",
		"/Library/Logs/DiagnosticReports/Retired/Analytics*.core_analytics",
	}

	var analyticsFiles []string
	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			params.Logger.Debug("Error finding analytics files: %v", err)
			continue
		}
		analyticsFiles = append(analyticsFiles, matches...)
	}

	if len(analyticsFiles) == 0 {
		params.Logger.Debug("No .core_analytics files found")
		return nil
	}

	outputFileName := utils.GetOutputFileName(moduleName+"-analytics", params.ExportFormat, params.OutputDir)
	writer, err := utils.NewDataWriter(params.LogsDir, outputFileName, params.ExportFormat)
	if err != nil {
		return err
	}

	for _, file := range analyticsFiles {
		data, err := os.ReadFile(file)
		if err != nil {
			params.Logger.Debug("Error reading file %s: %v", file, err)
			continue
		}

		lines := strings.Split(string(data), "\n")
		var diagStart, diagEnd string

		// Find diagnostic start time
		for _, line := range lines {
			if strings.HasPrefix(line, "{\"_marker\":") && !strings.Contains(line, "end-of-file") {
				var marker CoreAnalyticsMarker
				if err := json.Unmarshal([]byte(line), &marker); err == nil {
					diagStart = marker.StartTimestamp
					break
				}
			}
		}

		// Find diagnostic end time
		for _, line := range lines {
			if strings.HasPrefix(line, "{\"timestamp\":") {
				var timestamp CoreAnalyticsTimestamp
				if err := json.Unmarshal([]byte(line), &timestamp); err == nil {
					diagEnd = timestamp.Timestamp
					break
				}
			}
		}

		// Parse message lines
		for _, line := range lines {
			if strings.HasPrefix(line, "{\"message\":") {
				var message CoreAnalyticsMessage
				if err := json.Unmarshal([]byte(line), &message); err != nil {
					params.Logger.Debug("Error parsing message line: %v", err)
					continue
				}

				recordData := make(map[string]interface{})
				recordData["src_report"] = file
				recordData["diag_start"] = diagStart
				recordData["diag_end"] = diagEnd
				recordData["name"] = message.Name
				recordData["uuid"] = message.UUID

				// Process message fields
				for k, v := range message.Message {
					switch k {
					case "processName", "appDescription", "appName", "appVersion", "foreground":
						recordData[k] = v
					case "uptime", "powerTime", "activeTime":
						if val, ok := v.(float64); ok {
							recordData[k] = val
							recordData[k+"_parsed"] = formatDuration(val)
						}
					case "activations", "launches", "activityPeriods", "idleTimeouts":
						recordData[k] = v
					}
				}

				// Parse appDescription if present
				if appDesc, ok := recordData["appDescription"].(string); ok && appDesc != "" {
					parts := strings.Split(appDesc, " ||| ")
					if len(parts) == 2 {
						recordData["appName"] = parts[0]
						recordData["appVersion"] = parts[1]
					}
				}

				record := utils.Record{
					CollectionTimestamp: params.CollectionTimestamp,
					EventTimestamp:      diagStart,
					Data:                recordData,
					SourceFile:          file,
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

func parseAggregateFiles(moduleName string, params mod.ModuleParams) error {
	// Find aggregate files
	pattern := "/private/var/db/analyticsd/aggregates/4d7c9e4a-8c8c-4971-bce3-09d38d078849"
	aggregateFiles, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}

	if len(aggregateFiles) == 0 {
		params.Logger.Debug("No aggregate files found")
		return nil
	}

	outputFileName := utils.GetOutputFileName(moduleName+"-aggregates", params.ExportFormat, params.OutputDir)
	writer, err := utils.NewDataWriter(params.LogsDir, outputFileName, params.ExportFormat)
	if err != nil {
		return err
	}

	for _, file := range aggregateFiles {
		data, err := os.ReadFile(file)
		if err != nil {
			params.Logger.Debug("Error reading aggregate file: %v", err)
			continue
		}

		fileInfo, err := os.Stat(file)
		if err != nil {
			params.Logger.Debug("Error getting file info: %v", err)
			continue
		}

		diagStart := fileInfo.ModTime().Format(time.RFC3339)
		diagEnd := fileInfo.ModTime().Format(time.RFC3339)

		// Parse the aggregate data
		var aggregateData []interface{}
		if err := json.Unmarshal(data, &aggregateData); err != nil {
			params.Logger.Debug("Error parsing aggregate data: %v", err)
			continue
		}

		for _, entry := range aggregateData {
			recordData := make(map[string]interface{})
			recordData["src_report"] = file
			recordData["diag_start"] = diagStart
			recordData["diag_end"] = diagEnd
			recordData["uuid"] = filepath.Base(file)
			recordData["entry"] = entry // Use the entry variable

			// Parse the entry data structure
			// Note: The exact structure would need to be adapted based on the actual data format

			record := utils.Record{
				CollectionTimestamp: params.CollectionTimestamp,
				EventTimestamp:      diagStart,
				Data:                recordData,
				SourceFile:          file,
			}

			err = writer.WriteRecord(record)
			if err != nil {
				params.Logger.Debug("Failed to write record: %v", err)
			}
		}
	}

	return nil
}

func formatDuration(seconds float64) string {
	duration := time.Duration(seconds) * time.Second
	return fmt.Sprintf("%02d:%02d:%02d",
		int(duration.Hours()),
		int(duration.Minutes())%60,
		int(duration.Seconds())%60)
}
