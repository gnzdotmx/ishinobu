package mod

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/gnzdotmx/ishinobu/pkg/utils"
)

type CommandModule struct {
	ModuleName  string
	Description string
	Command     string
	Args        []string
	ParseLine   func(string) (utils.Record, error)
}

func (c *CommandModule) Name() string {
	return c.ModuleName
}

func (c *CommandModule) GetDescription() string {
	return c.Description
}

// Execute the command, parse the output and write the records to the output file
// Input: ModuleParams (collection timestamp, export format, output directory, logs directory)
// Output: error
func (c *CommandModule) Run(params ModuleParams) error {
	start := time.Now()

	// Run the command
	cmd := exec.Command(c.Command, c.Args...) // #nosec G204
	// Set the TZ environment variable to UTC
	cmd.Env = append(cmd.Env, "TZ=UTC")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("error running command: %v", err)
	}

	// Split the output into lines
	lines := strings.Split(string(output), "\n")
	if len(lines) < 1 {
		return nil // No output
	}

	// First line is the header
	header := lines[0]
	fields := strings.Fields(header)
	// Index map for fields
	fieldIndices := make(map[string]int)
	for i, field := range fields {
		fieldIndices[field] = i
	}

	// Prepare the output file
	outputFileName := utils.GetOutputFileName(c.ModuleName, params.ExportFormat, params.OutputDir)
	writer, err := utils.NewDataWriter(params.LogsDir, outputFileName, params.ExportFormat)
	if err != nil {
		return err
	}
	defer writer.Close()

	// Process each line
	for _, line := range lines[1:] {
		if strings.TrimSpace(line) == "" {
			continue
		}

		cols := strings.Fields(line)
		if len(cols) < len(fields) {
			continue
		}

		// Create a record data map
		recordData := make(map[string]interface{})
		for field, idx := range fieldIndices {
			if idx < len(cols) {
				recordData[field] = cols[idx]
			} else {
				recordData[field] = ""
			}
		}

		record := utils.Record{
			CollectionTimestamp: params.CollectionTimestamp,
			EventTimestamp:      utils.Now(),
			Data:                recordData,
			SourceFile:          c.ModuleName,
		}

		err := writer.WriteRecord(record)
		if err != nil {
			return fmt.Errorf("error writing record: %v", err)
		}
	}

	elapsed := time.Since(start)

	params.Logger.Info("✓ Module %s completed in %s", c.ModuleName, elapsed)
	return nil
}
