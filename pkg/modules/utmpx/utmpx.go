// This module collects and parses utmpx login records.
// It collects the following information:
// - User
// - ID
// - Terminal type
// - PID
// - Logon type
package utmpx

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"time"

	"github.com/gnzdotmx/ishinobu/pkg/mod"
	"github.com/gnzdotmx/ishinobu/pkg/utils"
)

type UtmpxModule struct {
	Name        string
	Description string
}

type UtmpxRecord struct {
	User         [256]byte // user login name
	ID           [4]byte   // identifier
	TerminalType [32]byte  // terminal type
	Pid          int32     // process id creating the entry
	LogonType    int16     // type of login
	Padding      [2]byte   // padding
	Timestamp    int32     // seconds
	Microseconds int32     // microseconds
	Hostname     [256]byte // host name
	Padding2     [64]byte  // padding
}

func init() {
	module := &UtmpxModule{
		Name:        "utmpx",
		Description: "Collects and parses utmpx login records"}
	mod.RegisterModule(module)
}

func (m *UtmpxModule) GetName() string {
	return m.Name
}

func (m *UtmpxModule) GetDescription() string {
	return m.Description
}

func decodeLogonType(code int16) string {
	switch code {
	case 2:
		return "BOOT_TIME"
	case 7:
		return "USER_PROCESS"
	default:
		return "UNKNOWN"
	}
}

func (m *UtmpxModule) Run(params mod.ModuleParams) error {
	utmpxPath := "/private/var/run/utmpx"

	outputFileName := utils.GetOutputFileName(m.GetName(), params.ExportFormat, params.OutputDir)
	writer, err := utils.NewDataWriter(params.LogsDir, outputFileName, params.ExportFormat)
	if err != nil {
		return err
	}

	err = parseUtmpx(params, utmpxPath, writer)
	if err != nil {
		params.Logger.Debug("Error parsing utmpx: %v", err)
		return fmt.Errorf("error parsing utmpx: %v", err)
	}

	return nil
}

func parseUtmpx(params mod.ModuleParams, utmpxPath string, writer utils.DataWriter) error {
	// Read the utmpx file
	data, err := os.ReadFile(utmpxPath)
	if err != nil {
		return fmt.Errorf("error reading utmpx file: %v", err)
	}

	// Get record size
	recordSize := binary.Size(UtmpxRecord{})

	// Check if file is too small to even contain a header
	if len(data) < recordSize {
		params.Logger.Debug("File too small to contain valid utmpx records: %s", utmpxPath)
		return nil
	}

	// Skip the header (first record)
	data = data[recordSize:]

	// Process each record
	for len(data) >= recordSize {
		var record UtmpxRecord
		reader := bytes.NewReader(data[:recordSize])
		if err := binary.Read(reader, binary.LittleEndian, &record); err != nil {
			params.Logger.Debug("Error reading record: %v", err)
			break
		}

		// Create timestamp
		timestamp := time.Unix(int64(record.Timestamp), int64(record.Microseconds*1000)).UTC()

		// Process hostname
		hostname := string(bytes.TrimRight(record.Hostname[:], "\x00"))
		if hostname == "" {
			hostname = "localhost"
		}

		eventTimestamp := timestamp.Format(utils.TimeFormat)

		// Create record data
		recordData := map[string]interface{}{
			"user":          string(bytes.TrimRight(record.User[:], "\x00")),
			"id":            string(bytes.TrimRight(record.ID[:], "\x00")),
			"terminal_type": string(bytes.TrimRight(record.TerminalType[:], "\x00")),
			"pid":           record.Pid,
			"logon_type":    decodeLogonType(record.LogonType),
			"timestamp":     eventTimestamp,
			"hostname":      hostname,
		}

		// Write the record
		outputRecord := utils.Record{
			CollectionTimestamp: params.CollectionTimestamp,
			EventTimestamp:      eventTimestamp,
			Data:                recordData,
			SourceFile:          utmpxPath,
		}

		if err := writer.WriteRecord(outputRecord); err != nil {
			params.Logger.Debug("Failed to write record: %v", err)
		}

		// Move to next record
		data = data[recordSize:]
	}

	return nil
}
