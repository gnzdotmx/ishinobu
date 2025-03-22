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

type DataWriter struct {
	file   *os.File
	writer interface{}
	format string
}

func NewDataWriter(outDir, filename, format string) (*DataWriter, error) {
	file, err := os.Create(filepath.Join(outDir, filename))
	if err != nil {
		return nil, err
	}

	var writer interface{}
	if format == "csv" {
		csvWriter := csv.NewWriter(file)
		// Write CSV header
		err := csvWriter.Write([]string{"collection_timestamp", "events_timestamp", "source_file", "data"})
		if err != nil {
			return nil, fmt.Errorf("error writing CSV header: %w", err)
		}
		writer = csvWriter
	} else {
		writer = json.NewEncoder(file)
	}

	return &DataWriter{
		file:   file,
		writer: writer,
		format: format,
	}, nil
}

func (dw *DataWriter) WriteRecord(record Record) error {
	if dw.format == "csv" {
		csvWriter := dw.writer.(*csv.Writer)
		cols := []string{
			record.CollectionTimestamp,
			record.EventTimestamp,
			record.SourceFile,
		}

		for k, v := range record.Data.(map[string]interface{}) {
			k = cleanKey(k)
			// Convert value to string and clean it for CSV
			strValue := fmt.Sprintf("%v", v)
			// Replace newlines with space and clean any problematic characters
			strValue = strings.ReplaceAll(strValue, "\n", " ")
			strValue = strings.ReplaceAll(strValue, "\r", " ")
			cols = append(cols, fmt.Sprintf("%v: %v", k, strValue))
		}

		if err := csvWriter.Write(cols); err != nil {
			return err
		}
		csvWriter.Flush()
		if err := csvWriter.Error(); err != nil {
			return err
		}
	} else {
		jsonEncoder := dw.writer.(*json.Encoder)
		jsonrecord := map[string]interface{}{
			"collection_timestamp": record.CollectionTimestamp,
			"event_timestamp":      record.EventTimestamp,
			"source_file":          record.SourceFile,
		}

		for k, v := range record.Data.(map[string]interface{}) {
			k = cleanKey(k)
			jsonrecord[k] = v
		}

		return jsonEncoder.Encode(jsonrecord)
	}
	return nil
}

func (dw *DataWriter) Close() error {
	return dw.file.Close()
}

func cleanKey(k string) string {
	// regex to clean k to only allow alphanumeric characters, -, _ and lowercase
	k = strings.ToLower(k)
	k = regexp.MustCompile("[^a-z0-9_-]").ReplaceAllString(k, "")
	return k
}
