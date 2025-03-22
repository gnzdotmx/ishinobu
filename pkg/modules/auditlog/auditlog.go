// This module collects and parses audit logs using praudit command over the files in /private/var/audit directory.
// It collects the following information:
// - Event
// - Modifier
// - Time
// - Msec
// - Arg-num
// - Value
package auditlog

import (
	"bufio"
	"encoding/xml"
	"os/exec"
	"path/filepath"

	"github.com/gnzdotmx/ishinobu/pkg/mod"
	"github.com/gnzdotmx/ishinobu/pkg/utils"
)

type AuditLogModule struct {
	Name        string
	Description string
}

func init() {
	module := &AuditLogModule{
		Name:        "auditlog",
		Description: "Collects and parses audit logs using praudit",
	}
	mod.RegisterModule(module)
}

func (m *AuditLogModule) GetName() string {
	return m.Name
}

func (m *AuditLogModule) GetDescription() string {
	return m.Description
}

func (m *AuditLogModule) Run(params mod.ModuleParams) error {
	files, err := filepath.Glob("/private/var/audit/*")
	if err != nil {
		return err
	}
	outputFileName := utils.GetOutputFileName(m.GetName(), params.ExportFormat, params.OutputDir)
	writer, err := utils.NewDataWriter(params.LogsDir, outputFileName, params.ExportFormat)
	if err != nil {
		return err
	}
	defer writer.Close()

	for _, file := range files {
		cmd := exec.Command("praudit", "-x", "-l", file)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return err
		}

		if err := cmd.Start(); err != nil {
			return err
		}

		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			record, err := parsePrauditXMLLine(line)
			if err != nil {
				params.Logger.Debug("Failed to parse line: %v", err)
				continue
			}

			record.CollectionTimestamp = params.CollectionTimestamp
			record.SourceFile = file

			err = writer.WriteRecord(record)
			if err != nil {
				params.Logger.Debug("Failed to write record: %v", err)
			}
		}
	}

	return nil
}

func parsePrauditXMLLine(line string) (utils.Record, error) {
	header := []string{"event", "modifier", "time", "msec", "arg-num", "value", "desc", "audit-uid", "uid", "gid", "ruid", "rgid", "pid", "sid", "tid", "errval", "retval"}

	type Argument struct {
		ArgNum string `xml:"arg-num,attr"`
		Value  string `xml:"value,attr"`
		Desc   string `xml:"desc,attr"`
	}
	type Subject struct {
		AuditUID string `xml:"audit-uid,attr"`
		UID      string `xml:"uid,attr"`
		GID      string `xml:"gid,attr"`
		RUID     string `xml:"ruid,attr"`
		RGID     string `xml:"rgid,attr"`
		PID      string `xml:"pid,attr"`
		SID      string `xml:"sid,attr"`
		TID      string `xml:"tid,attr"`
	}
	type Return struct {
		Errval string `xml:"errval,attr"`
		Retval string `xml:"retval,attr"`
	}
	type AuditRecord struct {
		XMLName   xml.Name   `xml:"record"`
		Version   string     `xml:"version,attr"`
		Event     string     `xml:"event,attr"`
		Modifier  string     `xml:"modifier,attr"`
		Time      string     `xml:"time,attr"`
		Msec      string     `xml:"msec,attr"`
		Arguments []Argument `xml:"argument"`
		Subject   Subject    `xml:"subject"`
		Return    Return     `xml:"return"`
	}

	parsedXML := AuditRecord{}
	err := xml.Unmarshal([]byte(line), &parsedXML)
	if err != nil {
		return utils.Record{}, err
	}

	recordData := make(map[string]interface{})

	for _, field := range header {
		switch field {
		case "event":
			recordData[field] = parsedXML.Event
		case "modifier":
			recordData[field] = parsedXML.Modifier
		case "time":
			recordData[field] = parsedXML.Time
		case "msec":
			recordData[field] = parsedXML.Msec
		case "arg-num":
			for _, arg := range parsedXML.Arguments {
				recordData[arg.ArgNum] = arg.Value
			}
		case "audit-uid":
			recordData[field] = parsedXML.Subject.AuditUID
		case "uid":
			recordData[field] = parsedXML.Subject.UID
		case "gid":
			recordData[field] = parsedXML.Subject.GID
		case "ruid":
			recordData[field] = parsedXML.Subject.RUID
		case "rgid":
			recordData[field] = parsedXML.Subject.RGID
		case "pid":
			recordData[field] = parsedXML.Subject.PID
		case "sid":
			recordData[field] = parsedXML.Subject.SID
		case "tid":
			recordData[field] = parsedXML.Subject.TID
		case "errval":
			recordData[field] = parsedXML.Return.Errval
		case "retval":
			recordData[field] = parsedXML.Return.Retval
		}
	}

	record := utils.Record{
		EventTimestamp: utils.Now(),
		Data:           recordData,
	}
	return record, nil
}
