// This module collects and parses system.log files.
// It collects the following information:
// - System name
// - Process name
// - PID
// - Message
package syslog

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/gnzdotmx/ishinobu/pkg/mod"
	"github.com/gnzdotmx/ishinobu/pkg/utils"
)

type SyslogModule struct {
	Name        string
	Description string
}

func init() {
	module := &SyslogModule{
		Name:        "syslog",
		Description: "Collects and parses system.log files"}
	mod.RegisterModule(module)
}

func (m *SyslogModule) GetName() string {
	return m.Name
}

func (m *SyslogModule) GetDescription() string {
	return m.Description
}

func (m *SyslogModule) Run(params mod.ModuleParams) error {
	// Define the pattern for system.log files
	syslogPattern := "/private/var/log/system.log*"

	// Get list of system.log files
	syslogFiles, err := utils.ListFiles(syslogPattern)
	if err != nil {
		params.Logger.Debug("Error listing syslog files: %v", err)
		return err
	}

	if len(syslogFiles) == 0 {
		params.Logger.Debug("No system.log files found in: %s", syslogPattern)
		return nil
	}

	outputFileName := utils.GetOutputFileName(m.GetName(), params.ExportFormat, params.OutputDir)
	writer, err := utils.NewDataWriter(params.LogsDir, outputFileName, params.ExportFormat)
	if err != nil {
		return err
	}

	// Process each syslog file
	for _, logFile := range syslogFiles {
		err := parseSyslogFile(logFile, writer, params)
		if err != nil {
			params.Logger.Debug("Error parsing syslog file %s: %v", logFile, err)
			continue
		}
	}

	return nil
}

func writeRecord(writer utils.DataWriter, logFile, timestamp, message string, params mod.ModuleParams) error {
	formattedTime, err := utils.ConvertDateString(timestamp)
	if err != nil {
		return err
	}

	recordData := map[string]interface{}{
		"message": message,
	}

	record := utils.Record{
		CollectionTimestamp: params.CollectionTimestamp,
		EventTimestamp:      formattedTime,
		Data:                recordData,
		SourceFile:          logFile,
	}

	return writer.WriteRecord(record)
}

func parseSyslogFile(logFile string, writer utils.DataWriter, params mod.ModuleParams) error {
	// Regular expression for parsing syslog entries
	syslogRegex := regexp.MustCompile(`(?P<month>[A-Za-z]{3})\s+(?P<day>\d{1,2})\s+(?P<time>\d{2}:\d{2}:\d{2})\s+(?P<systemname>[\S]+)\s+(?P<processName>[\S]+)\[(?P<PID>\d+)\](?:[^:]*)?:\s*(?P<message>.*)`)
	var reader io.Reader

	file, err := os.Open(logFile)
	if err != nil {
		return fmt.Errorf("error opening file %s: %v", logFile, err)
	}
	defer file.Close()

	// Check if file is gzipped
	if strings.HasSuffix(logFile, ".gz") {
		gzReader, err := gzip.NewReader(file)
		if err != nil {
			return fmt.Errorf("error creating gzip reader for %s: %v", logFile, err)
		}
		defer gzReader.Close()
		reader = gzReader
	} else {
		reader = file
	}

	scanner := bufio.NewScanner(reader)
	var multilineBuffer strings.Builder
	var lastTimestamp string

	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines
		if len(strings.TrimSpace(line)) == 0 {
			continue
		}

		// Check if line starts with month abbreviation (new log entry)
		if matches := syslogRegex.FindStringSubmatch(line); matches != nil {
			// If we have accumulated multiline content, write it first
			if multilineBuffer.Len() > 0 {
				err := writeRecord(writer, logFile, lastTimestamp, multilineBuffer.String(), params)
				if err != nil {
					params.Logger.Debug("Error writing record: %v", err)
				}
				multilineBuffer.Reset()
			}

			// Extract timestamp components
			month := matches[1]
			day := matches[2]
			timeStr := matches[3]
			timestamp := fmt.Sprintf("%s %s %s", month, day, timeStr)

			// Parse timestamp to standard format
			formattedTime, err := utils.ConvertDateString(fmt.Sprintf("%s %s %s", month, day, timeStr))
			if err != nil {
				params.Logger.Debug("Error parsing timestamp: %v", err)
				continue
			}

			// Create record data
			recordData := map[string]interface{}{
				"systemname":  matches[4],
				"processname": matches[5],
				"pid":         matches[6],
				"message":     matches[7],
			}

			record := utils.Record{
				CollectionTimestamp: params.CollectionTimestamp,
				EventTimestamp:      formattedTime,
				Data:                recordData,
				SourceFile:          logFile,
			}

			if err := writer.WriteRecord(record); err != nil {
				params.Logger.Debug("Failed to write record: %v", err)
			}

			lastTimestamp = timestamp
		} else {
			// This is a continuation of the previous message
			if multilineBuffer.Len() > 0 {
				multilineBuffer.WriteString(" ")
			}
			multilineBuffer.WriteString(strings.TrimSpace(line))
		}
	}

	// Write any remaining multiline content
	if multilineBuffer.Len() > 0 {
		err := writeRecord(writer, logFile, lastTimestamp, multilineBuffer.String(), params)
		if err != nil {
			params.Logger.Debug("Error writing record: %v", err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file %s: %v", logFile, err)
	}

	return nil
}
