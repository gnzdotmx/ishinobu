package coreanalytics

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gnzdotmx/ishinobu/pkg/mod"
	"github.com/gnzdotmx/ishinobu/pkg/modules/testutils"
	"github.com/stretchr/testify/assert"
)

func TestGetName(t *testing.T) {
	module := &CoreAnalyticsModule{
		Name:        "coreanalytics",
		Description: "Collects and parses CoreAnalytics artifacts",
	}
	assert.Equal(t, "coreanalytics", module.GetName())
}

func TestGetDescription(t *testing.T) {
	module := &CoreAnalyticsModule{
		Name:        "coreanalytics",
		Description: "Collects and parses CoreAnalytics artifacts",
	}
	assert.Equal(t, "Collects and parses CoreAnalytics artifacts", module.GetDescription())
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		seconds  float64
		expected string
	}{
		{0, "00:00:00"},
		{60, "00:01:00"},
		{3600, "01:00:00"},
		{3661, "01:01:01"},
		{86400, "24:00:00"},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%.0f seconds", test.seconds), func(t *testing.T) {
			result := formatDuration(test.seconds)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestGetMacOSVersion(t *testing.T) {
	// This test depends on the actual OS version, so it's a simple verification
	// that the function returns something reasonable
	version, err := getMacOSVersion()
	assert.NoError(t, err)
	assert.NotEmpty(t, version)
}

func TestParseAnalyticsFiles(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "coreanalytics_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create fake analytics files
	analyticsDir := filepath.Join(tmpDir, "DiagnosticReports")
	err = os.MkdirAll(analyticsDir, 0755)
	assert.NoError(t, err)

	analyticsFile := filepath.Join(analyticsDir, "Analytics1.core_analytics")

	// Create a sample analytics file content
	analyticContent := `{"_marker":"start","startTimestamp":"2023-01-01T12:00:00Z"}
{"timestamp":"2023-01-01T12:30:00Z"}
{"message":{"processName":"TestProcess","uptime":3600,"powerTime":1800,"foreground":true,"appDescription":"TestApp ||| 1.0","activations":5},"name":"AppUsage","uuid":"test-uuid-1"}
{"message":{"processName":"TestProcess2","activeTime":1200,"foreground":false,"appName":"TestApp2","appVersion":"2.0","launches":3},"name":"AppLaunch","uuid":"test-uuid-2"}
`
	err = os.WriteFile(analyticsFile, []byte(analyticContent), 0644)
	assert.NoError(t, err)

	// Setup test params
	logger := testutils.NewTestLogger()
	params := mod.ModuleParams{
		Logger:              *logger,
		LogsDir:             tmpDir,
		OutputDir:           "./",
		ExportFormat:        "json",
		CollectionTimestamp: time.Now().Format(time.RFC3339),
	}

	// Test the function
	patterns := []string{filepath.Join(tmpDir, "DiagnosticReports", "Analytics*.core_analytics")}
	err = parseAnalyticsFiles("coreanalytics", params, patterns)
	assert.NoError(t, err)

	// Verify that output file was created
	outputFile := filepath.Join(tmpDir, "coreanalytics-analytics.json")
	_, err = os.Stat(outputFile)
	assert.NoError(t, err)

	// Read the output file to verify content
	content, err := os.ReadFile(outputFile)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "TestProcess")
	assert.Contains(t, string(content), "AppUsage")
	assert.Contains(t, string(content), "test-uuid-1")
}

func TestParseAnalyticsFilesError(t *testing.T) {
	// Test with invalid pattern
	logger := testutils.NewTestLogger()
	params := mod.ModuleParams{
		Logger:              *logger,
		LogsDir:             "/nonexistent",
		OutputDir:           "/nonexistent",
		ExportFormat:        "json",
		CollectionTimestamp: time.Now().Format(time.RFC3339),
	}

	// This should not return an error, just log and continue
	err := parseAnalyticsFiles("coreanalytics", params, []string{"/invalid/[pattern"})
	assert.NoError(t, err)
}

func TestParseAggregateFiles(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "coreanalytics_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create fake aggregate files directory structure
	aggregateDir := filepath.Join(tmpDir, "db", "analyticsd", "aggregates")
	err = os.MkdirAll(aggregateDir, 0755)
	assert.NoError(t, err)

	// Create a sample aggregate file
	aggregateFile := filepath.Join(aggregateDir, "4d7c9e4a-8c8c-4971-bce3-09d38d078849")

	// Create sample aggregate data
	aggregateData := []map[string]interface{}{
		{
			"app":   "TestApp",
			"usage": 3600,
			"count": 5,
		},
		{
			"app":   "TestApp2",
			"usage": 1800,
			"count": 3,
		},
	}

	jsonData, err := json.Marshal(aggregateData)
	assert.NoError(t, err)

	err = os.WriteFile(aggregateFile, jsonData, 0644)
	assert.NoError(t, err)

	// Setup test params
	logger := testutils.NewTestLogger()
	params := mod.ModuleParams{
		Logger:              *logger,
		LogsDir:             tmpDir,
		OutputDir:           "./",
		ExportFormat:        "json",
		CollectionTimestamp: time.Now().Format(time.RFC3339),
	}

	// Test the function
	pattern := filepath.Join(tmpDir, "db", "analyticsd", "aggregates", "4d7c9e4a-8c8c-4971-bce3-09d38d078849")
	err = parseAggregateFiles("coreanalytics", params, pattern)
	assert.NoError(t, err)

	// Verify that output file was created
	outputFile := filepath.Join(tmpDir, "coreanalytics-aggregates.json")
	_, err = os.Stat(outputFile)
	assert.NoError(t, err)

	// Read the output file to verify content
	content, err := os.ReadFile(outputFile)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "TestApp")
}

func TestParseAggregateFilesError(t *testing.T) {
	// Test with non-existent pattern
	logger := testutils.NewTestLogger()
	params := mod.ModuleParams{
		Logger:              *logger,
		LogsDir:             "/nonexistent",
		OutputDir:           "/nonexistent",
		ExportFormat:        "json",
		CollectionTimestamp: time.Now().Format(time.RFC3339),
	}

	// This should not return an error, just log and continue
	err := parseAggregateFiles("coreanalytics", params, "/nonexistent/pattern")
	assert.NoError(t, err)
}

func TestRunCoreAnalyticsModule(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "coreanalytics_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Setup test params
	logger := testutils.NewTestLogger()
	params := mod.ModuleParams{
		Logger:              *logger,
		LogsDir:             tmpDir,
		OutputDir:           "./",
		ExportFormat:        "json",
		CollectionTimestamp: time.Now().Format(time.RFC3339),
	}

	// Create a module and run it
	module := &CoreAnalyticsModule{
		Name:        "coreanalytics",
		Description: "Collects and parses CoreAnalytics artifacts",
	}

	// Run the module - the actual OS version will determine if processing continues
	err = module.Run(params)
	assert.NoError(t, err)

	// Don't assert on the specific error as it depends on the OS version
	// but we can at least verify the module runs
	assert.NotPanics(t, func() {
		_ = module.Run(params)
	})
}
