// This module collects and parses logs from Apple System Logs (ASL).
// The module uses the syslog command to read logs from the ASL database and
// parse them as XML. The parsed logs are then printed to the console.
// It collects the following information:
// - ASLMessageID
// - Time
// - TimeNanoSec
// - Level
// - PID
// - UID
package modules

import (
	"encoding/xml"
	"io"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gnzdotmx/ishinobu/ishinobu/mod"
	"github.com/gnzdotmx/ishinobu/ishinobu/utils"
)

type AslModule struct {
	Name        string
	Description string
}

// Plist represents the root plist element containing an array of LogEntry.
type Plist struct {
	XMLName xml.Name   `xml:"plist"`
	Entries []LogEntry `xml:"array>dict"`
}

// LogEntry represents each dictionary entry in the plist array.
type LogEntry struct {
	ASLMessageID   string
	Time           string
	TimeNanoSec    string
	Level          string
	PID            string
	UID            string
	GID            string
	ReadGID        string
	Host           string
	Sender         string
	Facility       string
	Message        string
	MsgCount       string
	ShimCount      string
	SenderMachUUID string
}

func (le *LogEntry) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var currentKey string
	for {
		tok, err := d.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "key":
				// Decode the <key> element to get the current key.
				var key string
				if err := d.DecodeElement(&key, &t); err != nil {
					return err
				}
				currentKey = key
			case "string":
				// Decode the value element corresponding to the current key.
				var value string
				if err := d.DecodeElement(&value, &t); err != nil {
					return err
				}
				// Assign the value to the appropriate struct field based on the current key.
				switch currentKey {
				case "ASLMessageID":
					le.ASLMessageID = value
				case "Time":
					le.Time = value
				case "TimeNanoSec":
					le.TimeNanoSec = value
				case "Level":
					le.Level = value
				case "PID":
					le.PID = value
				case "UID":
					le.UID = value
				case "GID":
					le.GID = value
				case "ReadGID":
					le.ReadGID = value
				case "Host":
					le.Host = value
				case "Sender":
					le.Sender = value
				case "Facility":
					le.Facility = value
				case "Message":
					le.Message = value
				case "MsgCount":
					le.MsgCount = value
				case "ShimCount":
					le.ShimCount = value
				case "SenderMachUUID":
					le.SenderMachUUID = value
				}
			}
		case xml.EndElement:
			if t.Name.Local == "dict" {
				return nil
			}
		}
	}
	return nil
}

func init() {
	module := &AslModule{
		Name:        "asl",
		Description: "Collects and parses logs from Apple System Logs (ASL)"}
	mod.RegisterModule(module)
}

func (m *AslModule) GetName() string {
	return m.Name
}

func (m *AslModule) GetDescription() string {
	return m.Description
}

func (m *AslModule) Run(params mod.ModuleParams) error {
	aslFiles, err := filepath.Glob("/private/var/log/asl/*.asl")
	if err != nil {
		return err
	}

	outputFileName := utils.GetOutputFileName(m.GetName(), params.ExportFormat, params.OutputDir)
	writer, err := utils.NewDataWriter(params.LogsDir, outputFileName, params.ExportFormat)
	if err != nil {
		return err
	}
	defer writer.Close()

	for _, file := range aslFiles {
		cmd := exec.Command("syslog", "-F", "xml", "-f", file)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			continue
		}

		if err := cmd.Start(); err != nil {
			continue
		}

		var plist Plist
		buf, err := io.ReadAll(stdout)
		if err != nil {
			continue
		}

		decoder := xml.NewDecoder(strings.NewReader(string(buf)))
		if err := decoder.Decode(&plist); err != nil {
			params.Logger.Debug("Error decoding plist XML: %v", err)
			continue
		}

		if err := cmd.Wait(); err != nil {
			continue
		}

		for _, entry := range plist.Entries {
			recordData := make(map[string]interface{})
			recordData["ASLMessageID"] = entry.ASLMessageID
			recordData["Time"] = entry.Time
			recordData["TimeNanoSec"] = entry.TimeNanoSec
			recordData["Level"] = entry.Level
			recordData["PID"] = entry.PID
			recordData["UID"] = entry.UID
			recordData["GID"] = entry.GID
			recordData["ReadGID"] = entry.ReadGID
			recordData["Host"] = entry.Host
			recordData["Sender"] = entry.Sender
			recordData["Facility"] = entry.Facility
			recordData["Message"] = entry.Message
			recordData["MsgCount"] = entry.MsgCount
			recordData["ShimCount"] = entry.ShimCount
			recordData["SenderMachUUID"] = entry.SenderMachUUID

			parsedEntryTime, err := utils.ConvertDateString(entry.Time)
			if err != nil {
				params.Logger.Debug("Failed to parse timestamp: %v", err)
			}

			record := utils.Record{
				CollectionTimestamp: utils.Now(),
				EventTimestamp:      parsedEntryTime,
				SourceFile:          file,
				Data:                recordData,
			}

			err = writer.WriteRecord(record)
			if err != nil {
				params.Logger.Debug("Failed to write record: %v", err)
			}
		}
	}

	return nil
}
