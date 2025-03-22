package utils

import (
	"os"
	"strings"
	"testing"
)

func TestNewLogger(t *testing.T) {
	// Create a new logger
	logger := NewLogger()
	defer func() {
		logger.Close()
		// Clean up the log file
		os.Remove(logger.file.Name())
	}()

	// Check if logger is properly initialized
	if logger.file == nil {
		t.Error("Expected file to be initialized, got nil")
	}
	if logger.log == nil {
		t.Error("Expected log to be initialized, got nil")
	}
	if logger.verbosity != 1 {
		t.Errorf("Expected default verbosity to be 1, got %d", logger.verbosity)
	}
}

func TestSetVerbosity(t *testing.T) {
	logger := NewLogger()
	defer func() {
		logger.Close()
		os.Remove(logger.file.Name())
	}()

	logger.SetVerbosity(2)
	if logger.verbosity != 2 {
		t.Errorf("Expected verbosity to be 2, got %d", logger.verbosity)
	}
}

func TestLogLevels(t *testing.T) {
	logger := NewLogger()
	defer func() {
		logger.Close()
		os.Remove(logger.file.Name())
	}()

	tests := []struct {
		name      string
		logFunc   func(string, ...interface{})
		message   string
		verbosity int
		expected  bool // whether the message should be logged
	}{
		{"Info with verbosity 1", logger.Info, "test info", 1, true},
		{"Info with verbosity 0", logger.Info, "test info", 0, false},
		{"Debug with verbosity 2", logger.Debug, "test debug", 2, true},
		{"Debug with verbosity 1", logger.Debug, "test debug", 1, false},
		{"Error with verbosity 0", logger.Error, "test error", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset file content
			err := logger.file.Truncate(0)
			if err != nil {
				t.Fatalf("Failed to truncate file: %v", err)
			}

			_, err = logger.file.Seek(0, 0)
			if err != nil {
				t.Fatalf("Failed to seek file: %v", err)
			}

			logger.SetVerbosity(tt.verbosity)
			tt.logFunc(tt.message)

			// Read file content
			content := make([]byte, 1024)
			n, _ := logger.file.ReadAt(content, 0)
			logContent := string(content[:n])

			messageExists := strings.Contains(logContent, tt.message)
			if messageExists != tt.expected {
				t.Errorf("Expected message existence to be %v, got %v", tt.expected, messageExists)
			}
		})
	}
}

func TestClose(t *testing.T) {
	logger := NewLogger()
	filename := logger.file.Name()

	err := logger.Close()
	if err != nil {
		t.Errorf("Expected no error on close, got %v", err)
	}

	// Clean up
	os.Remove(filename)
}
