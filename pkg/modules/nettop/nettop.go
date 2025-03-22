// This module is useful to investigate the amount of data transferred by processes and network interfaces.
// Command: nettop -n -P -J interface,state,bytes_in,bytes_out,packets_in,packets_out -L 1
// It collects the following information:
// - Interface
// - State
// - Bytes in
// - Bytes out
// - Packets in
// - Packets out
package nettop

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/gnzdotmx/ishinobu/pkg/mod"
	"github.com/gnzdotmx/ishinobu/pkg/utils"
)

type NettopModule struct {
	Name        string
	Description string
}

func init() {
	module := &NettopModule{Name: "nettop", Description: "Collects information about network connections"}
	mod.RegisterModule(module)
}

func (m *NettopModule) GetName() string {
	return m.Name
}

func (m *NettopModule) GetDescription() string {
	return m.Description
}

func (m *NettopModule) Run(params mod.ModuleParams) error {
	// -L output format: interface,state,bytes_in,bytes_out,packets_in,packets_out
	cmd := exec.Command("nettop", "-n", "-P", "-J", "interface,state,bytes_in,bytes_out,packets_in,packets_out", "-L", "1")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("error running command: %w", err)
	}

	// Split the output into lines
	lines := strings.Split(string(output), "\n")
	if len(lines) < 1 {
		params.Logger.Debug("No output from nettop")
		return nil // No output
	}

	// First line is the header
	header := lines[0]
	fields := strings.Split(header, ",")
	// Index map for fields
	fieldIndices := make(map[string]int)
	for i, field := range fields {
		fieldIndices[field] = i
	}

	// Prepare the output file
	outputFileName := utils.GetOutputFileName(m.GetName(), params.ExportFormat, params.OutputDir)
	writer, err := utils.NewDataWriter(params.LogsDir, outputFileName, params.ExportFormat)
	if err != nil {
		return fmt.Errorf("failed to create data writer: %w", err)
	}
	defer writer.Close()

	// Parse the lines
	for _, line := range lines[1:] {
		if strings.TrimSpace(line) == "" {
			continue
		}

		cols := strings.Split(line, ",")
		if len(cols) != len(fields) {
			continue
		}

		recordData := make(map[string]interface{})
		for index, col := range cols {

			colName := fields[index]

			// skip empty columns
			if colName == "" && col == "" {
				continue
			}

			if colName == "" && col != "" {
				recordData["process"] = col
			} else {
				recordData[string(colName)] = col
			}
		}

		record := utils.Record{
			CollectionTimestamp: params.CollectionTimestamp,
			EventTimestamp:      utils.Now(),
			Data:                recordData,
			SourceFile:          "nettop",
		}

		err = writer.WriteRecord(record)
		if err != nil {
			params.Logger.Debug("Failed to write record: %v", err)
			return fmt.Errorf("failed to write record: %w", err)
		}

	}

	return nil
}
