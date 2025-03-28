package utils

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Record struct {
	CollectionTimestamp string      `json:"collection_timestamp"`
	EventTimestamp      string      `json:"event_timestamp"`
	SourceFile          string      `json:"source_file"`
	Data                interface{} `json:"data"`
}

type DataWriter interface {
	WriteRecord(record Record) error
	Close() error
}

type CSVDataWriter struct {
	file   *os.File
	writer *csv.Writer
}

func NewDataWriter(outDir, filename, format string) (DataWriter, error) {
	file, err := os.Create(filepath.Join(outDir, filename))
	if err != nil {
		return nil, err
	}

	switch format {
	case "csv":
		csvWriter := csv.NewWriter(file)

		// Write CSV header
		err := csvWriter.Write([]string{"collection_timestamp", "events_timestamp", "source_file", "data"})
		if err != nil {
			return nil, fmt.Errorf("error writing CSV header: %w", err)
		}

		return &CSVDataWriter{
			file:   file,
			writer: csvWriter,
		}, nil
	case "json":
		return &JSONDataWriter{
			file:   file,
			writer: json.NewEncoder(file),
		}, nil
	default:
		return nil, fmt.Errorf("%w: %s", errUnsupportedDataWriterFormat, format)
	}
}

func (dw *CSVDataWriter) WriteRecord(record Record) error {
	cols := []string{
		record.CollectionTimestamp,
		record.EventTimestamp,
		record.SourceFile,
	}

	recordMap, ok := record.Data.(map[string]interface{})
	if !ok {
		return errWriterInvalidRecordData
	}

	for k, v := range recordMap {
		k = cleanKey(k)
		// Convert value to string and clean it for CSV
		strValue := fmt.Sprintf("%v", v)
		// Replace newlines with space and clean any problematic characters
		strValue = strings.ReplaceAll(strValue, "\n", " ")
		strValue = strings.ReplaceAll(strValue, "\r", " ")
		cols = append(cols, fmt.Sprintf("%v: %v", k, strValue))
	}

	if err := dw.writer.Write(cols); err != nil {
		return err
	}

	dw.writer.Flush()
	if err := dw.writer.Error(); err != nil {
		return err
	}

	return nil
}

func (dw *CSVDataWriter) Close() error {
	return dw.file.Close()
}

type JSONDataWriter struct {
	file   *os.File
	writer *json.Encoder
}

func (dw *JSONDataWriter) WriteRecord(record Record) error {
	jsonrecord := map[string]interface{}{
		"collection_timestamp": record.CollectionTimestamp,
		"event_timestamp":      record.EventTimestamp,
		"source_file":          record.SourceFile,
	}

	recordMap, ok := record.Data.(map[string]interface{})
	if !ok {
		return errWriterInvalidRecordData
	}

	for k, v := range recordMap {
		k = cleanKey(k)
		jsonrecord[k] = v
	}

	return dw.writer.Encode(jsonrecord)
}

func (dw *JSONDataWriter) Close() error {
	return dw.file.Close()
}

func cleanKey(k string) string {
	// regex to clean k to only allow alphanumeric characters, -, _ and lowercase
	k = strings.ToLower(k)
	k = regexp.MustCompile("[^a-z0-9_-]").ReplaceAllString(k, "")
	return k
}
