// This module is useful to investigate the list of logs from the unified logging system.
// Commands:
// - System Logs - Kernel Messages: log show --predicate 'process == "kernel"' --info --start "2021-09-01 00:00:00" --end "2021-09-30 23:59:59"
// - Security Logs - Authentication Attempts: log show --predicate 'eventMessage CONTAINS[c] "authentication"' --info --start "2021-09-01 00:00:00" --end "2021-09-30 23:59:59"
// - Network Logs - Network Activities: log show --predicate 'subsystem == "com.apple.network"' --info --start "2021-09-01 00:00:00" --end "2021-09-30 23:59:59"
// - User Activity Logs - Login Sessions: log show --predicate 'eventMessage CONTAINS "login"' --start "2021-09-01 00:00:00" --end "2021-09-30 23:59:59"
// - File System Events - Disk Mounts: log show --predicate 'eventMessage CONTAINS "disk"' --start "2021-09-01 00:00:00" --end "2021-09-30 23:59:59"
// - Configuration Changes - Software Installations: log show --predicate 'eventMessage CONTAINS "install" OR eventMessage CONTAINS "update"' --start "2021-09-01 00:00:00" --end "2021-09-30 23:59:59"
// - Hardware Events - Peripheral Connections: log show --predicate 'eventMessage CONTAINS "USB" OR eventMessage CONTAINS "Peripheral"' --start "2021-09-01 00:00:00" --end "2021-09-30 23:59:59"
// - Time and Date Changes - System Time Adjustments: log show --predicate 'eventMessage CONTAINS "system time"' --start "2021-09-01 00:00:00" --end "2021-09-30 23:59:59"
package unifiedlogs

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/gnzdotmx/ishinobu/pkg/mod"
	"github.com/gnzdotmx/ishinobu/pkg/utils"
)

type UnifiedLogsModule struct {
	Name        string
	Description string
}

// LogCommand represents a log collection command with its description and output file path.
type LogCommand struct {
	Description string
	Command     string
}

func init() {
	module := &UnifiedLogsModule{
		Name:        "unifiedlogs",
		Description: "Collects and parses logs from unuified logging system"}
	mod.RegisterModule(module)
}

func (m *UnifiedLogsModule) GetName() string {
	return m.Name
}

func (m *UnifiedLogsModule) GetDescription() string {
	return m.Description
}

func (m *UnifiedLogsModule) Run(params mod.ModuleParams) error {
	// Time range for the last 30 days
	nDays := 1
	endTime := time.Now().Format("2006-01-02 15:04:05")
	startTime := time.Now().AddDate(0, 0, -nDays).Format("2006-01-02 15:04:05")

	// List of log collection commands
	commands := []LogCommand{
		// {
		// 	Description: "System Logs - Kernel Messages",
		// 	Command:     fmt.Sprintf(`log show --predicate 'process == "kernel"' --style json --quiet --info --start "%s" --end "%s"`, startTime, endTime),
		// },
		// {
		// 	Description: "Security Logs - Authentication Attempts",
		// 	Command:     fmt.Sprintf(`log show --predicate 'eventMessage CONTAINS[c] "authentication"' --style json --quiet --info --start "%s" --end "%s"`, startTime, endTime),
		// },
		// {
		// 	Description: "Network Logs - Network Activities",
		// 	Command:     fmt.Sprintf(`log show --predicate 'subsystem == "com.apple.network"' --style json --quiet --info --start "%s" --end "%s"`, startTime, endTime),
		// },
		// {
		// 	Description: "User Activity Logs - Login Sessions",
		// 	Command:     fmt.Sprintf(`log show --predicate 'eventMessage CONTAINS "login"' --style json --quiet --start "%s" --end "%s"`, startTime, endTime),
		// },
		// {
		// 	Description: "File System Events - Disk Mounts",
		// 	Command:     fmt.Sprintf(`log show --predicate 'eventMessage CONTAINS "disk"' --style json --quiet --start "%s" --end "%s"`, startTime, endTime),
		// },
		// {
		// 	Description: "Configuration Changes - Software Installations",
		// 	Command:     fmt.Sprintf(`log show --predicate 'eventMessage CONTAINS "install" OR eventMessage CONTAINS "update"' --style json --quiet --start "%s" --end "%s"`, startTime, endTime),
		// },
		// {
		// 	Description: "Hardware Events - Peripheral Connections",
		// 	Command:     fmt.Sprintf(`log show --predicate 'eventMessage CONTAINS "USB" OR eventMessage CONTAINS "Peripheral"' --style json --quiet --start "%s" --end "%s"`, startTime, endTime),
		// },
		// {
		// 	Description: "Time and Date Changes - System Time Adjustments",
		// 	Command:     fmt.Sprintf(`log show --predicate 'eventMessage CONTAINS "system time"' --style json --quiet --start "%s" --end "%s"`, startTime, endTime),
		// },
		{
			Description: "Command Line Activity - Run With Elevated Privileges",
			Command:     fmt.Sprintf(`log show --predicate 'process == "sudo"' --style json --quiet --start "%s" --end "%s"`, startTime, endTime),
		},
		{
			Description: "SSH Activity - Remote Connections",
			Command:     fmt.Sprintf(`log show --predicate 'process == "ssh" OR process == "sshd"' --style json --quiet --start "%s" --end "%s"`, startTime, endTime),
		},
		{
			Description: "Screen Sharing Activity - Remote Desktop Connections",
			Command:     fmt.Sprintf(`log show --predicate 'process == "screensharingd" OR process == "ScreensharingAgent"' --style json --quiet --start "%s" --end "%s"`, startTime, endTime),
		},
		{
			Description: "Session creation or deletion",
			Command:     fmt.Sprintf(`log show --predicate 'process == "securityd" AND eventMessage CONTAINS "session" AND subsystem == "com.apple.securityd"' --style json --quiet --start "%s" --end "%s"`, startTime, endTime),
		},
	}

	// Prepare the output file
	outputFileName := utils.GetOutputFileName(m.GetName(), params.ExportFormat, params.OutputDir)
	writer, err := utils.NewDataWriter(params.LogsDir, outputFileName, params.ExportFormat)
	if err != nil {
		return fmt.Errorf("failed to create data writer: %v", err)
	}
	defer writer.Close()

	// Run each log collection command
	for _, cmd := range commands {

		// Run the command
		cmdexec := exec.Command("bash", "-c", cmd.Command)

		// Set the TZ environment variable to UTC
		cmdexec.Env = append(cmdexec.Env, "TZ=UTC")
		output, err := cmdexec.CombinedOutput()
		if err != nil {
			params.Logger.Debug("Command output: %s", output)
		}

		var logEntries []map[string]interface{}

		// Parse the JSON output
		err = json.Unmarshal(output, &logEntries)
		if err != nil {
			params.Logger.Debug("Error parsing JSON output: %v", err)
		}

		// Iterate over each log entry and print it as a JSON object
		for _, entry := range logEntries {
			// Create a record data map
			recordData := make(map[string]interface{})
			for key, value := range entry {
				recordData[key] = value
			}

			// Parse the timestamp
			timestamp, err := utils.ParseTimestamp(recordData["timestamp"].(string))
			if err != nil {
				params.Logger.Debug("Error parsing timestamp: %v", err)
			}

			sourceFileName := m.GetName() + strings.ReplaceAll(cmd.Description, " ", "_")
			// Create a record
			record := utils.Record{
				CollectionTimestamp: params.CollectionTimestamp,
				EventTimestamp:      timestamp,
				Data:                recordData,
				SourceFile:          sourceFileName,
			}

			// Write the record
			err = writer.WriteRecord(record)
			if err != nil {
				params.Logger.Debug("Failed to write record: %v", err)
			}
		}
	}

	return nil
}
