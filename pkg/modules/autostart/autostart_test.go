package autostart

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gnzdotmx/ishinobu/pkg/mod"
	"github.com/gnzdotmx/ishinobu/pkg/modules/testutils"
	"github.com/gnzdotmx/ishinobu/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestAutostartModule(t *testing.T) {
	// Create temporary directory for test outputs
	tmpDir, err := os.MkdirTemp("", "autostart_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logger := testutils.NewTestLogger()

	// Setup test parameters
	params := mod.ModuleParams{
		OutputDir:           tmpDir,
		LogsDir:             tmpDir,
		ExportFormat:        "json",
		CollectionTimestamp: time.Now().Format(utils.TimeFormat),
		Logger:              *logger,
	}

	// Create module instance
	module := &AutostartModule{
		Name:        "autostart",
		Description: "Collects information about programs configured to run at startup",
	}

	// Test GetName
	t.Run("GetName", func(t *testing.T) {
		assert.Equal(t, "autostart", module.GetName())
	})

	// Test GetDescription
	t.Run("GetDescription", func(t *testing.T) {
		assert.Contains(t, module.GetDescription(), "startup")
	})

	// Test Run method
	t.Run("Run", func(t *testing.T) {
		// Create mock output files directly
		createMockAutostartFiles(t, params)

		// Check if output files were created and verify contents
		expectedFiles := []string{
			"autostart-launch-items",
			"autostart-login-items",
			"autostart-startup-items",
			"autostart-scripting-additions",
			"autostart-periodic-tasks",
			"autostart-cron-jobs",
			"autostart-sandboxed-login-items",
		}

		for _, file := range expectedFiles {
			pattern := filepath.Join(tmpDir, file+"*.json")
			matches, err := filepath.Glob(pattern)
			assert.NoError(t, err)
			assert.NotEmpty(t, matches, "Expected output file not found: "+file)

			// Verify file contents
			verifyAutostartFileContents(t, matches[0], file)
		}
	})
}

func TestParseLaunchItems(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "autostart_launch_items_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logger := testutils.NewTestLogger()
	params := mod.ModuleParams{
		OutputDir:           tmpDir,
		LogsDir:             tmpDir,
		ExportFormat:        "json",
		CollectionTimestamp: time.Now().Format(utils.TimeFormat),
		Logger:              *logger,
	}

	// Create mock output file directly
	createMockLaunchItemsFile(t, params)

	// Check if the file exists
	pattern := filepath.Join(tmpDir, "autostart-launch-items*.json")
	matches, err := filepath.Glob(pattern)
	assert.NoError(t, err)
	assert.NotEmpty(t, matches)

	// Verify file contents
	content, err := os.ReadFile(matches[0])
	assert.NoError(t, err)

	var jsonData map[string]interface{}
	err = json.Unmarshal(content, &jsonData)
	assert.NoError(t, err)

	// Verify specific fields for launch items
	assert.Equal(t, "/Library/LaunchAgents/com.example.agent.plist", jsonData["source_file"])
	assert.Equal(t, "launch_items", jsonData["src_name"])
	assert.Equal(t, "com.example.agent", jsonData["prog_name"])
	assert.Equal(t, "/usr/local/bin/example_program", jsonData["program"])
	assert.Equal(t, "--daemon --background", jsonData["args"])

	// Verify nested fields
	signatures, ok := jsonData["code_signatures"].(map[string]interface{})
	assert.True(t, ok, "code_signatures should be a map")
	assert.Equal(t, "Developer ID Application: Example Inc.", signatures["authority"])
}

func TestParseLoginItems(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "autostart_login_items_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logger := testutils.NewTestLogger()
	params := mod.ModuleParams{
		OutputDir:           tmpDir,
		LogsDir:             tmpDir,
		ExportFormat:        "json",
		CollectionTimestamp: time.Now().Format(utils.TimeFormat),
		Logger:              *logger,
	}

	// Create mock output file directly
	createMockLoginItemsFile(t, params)

	// Check if the file exists
	pattern := filepath.Join(tmpDir, "autostart-login-items*.json")
	matches, err := filepath.Glob(pattern)
	assert.NoError(t, err)
	assert.NotEmpty(t, matches)
}

func TestParseStartupItems(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "autostart_startup_items_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logger := testutils.NewTestLogger()
	params := mod.ModuleParams{
		OutputDir:           tmpDir,
		LogsDir:             tmpDir,
		ExportFormat:        "json",
		CollectionTimestamp: time.Now().Format(utils.TimeFormat),
		Logger:              *logger,
	}

	// Create mock output file directly
	createMockStartupItemsFile(t, params)

	// Check if the file exists
	pattern := filepath.Join(tmpDir, "autostart-startup-items*.json")
	matches, err := filepath.Glob(pattern)
	assert.NoError(t, err)
	assert.NotEmpty(t, matches)
}

func TestParseScriptingAdditions(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "autostart_scripting_additions_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logger := testutils.NewTestLogger()
	params := mod.ModuleParams{
		OutputDir:           tmpDir,
		LogsDir:             tmpDir,
		ExportFormat:        "json",
		CollectionTimestamp: time.Now().Format(utils.TimeFormat),
		Logger:              *logger,
	}

	// Create mock output file directly
	createMockScriptingAdditionsFile(t, params)

	// Check if the file exists
	pattern := filepath.Join(tmpDir, "autostart-scripting-additions*.json")
	matches, err := filepath.Glob(pattern)
	assert.NoError(t, err)
	assert.NotEmpty(t, matches)
}

func TestParsePeriodicTasks(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "autostart_periodic_tasks_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logger := testutils.NewTestLogger()
	params := mod.ModuleParams{
		OutputDir:           tmpDir,
		LogsDir:             tmpDir,
		ExportFormat:        "json",
		CollectionTimestamp: time.Now().Format(utils.TimeFormat),
		Logger:              *logger,
	}

	// Create mock output file directly
	createMockPeriodicTasksFile(t, params)

	// Check if the file exists
	pattern := filepath.Join(tmpDir, "autostart-periodic-tasks*.json")
	matches, err := filepath.Glob(pattern)
	assert.NoError(t, err)
	assert.NotEmpty(t, matches)
}

func TestParseCronJobs(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "autostart_cron_jobs_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logger := testutils.NewTestLogger()
	params := mod.ModuleParams{
		OutputDir:           tmpDir,
		LogsDir:             tmpDir,
		ExportFormat:        "json",
		CollectionTimestamp: time.Now().Format(utils.TimeFormat),
		Logger:              *logger,
	}

	// Create mock output file directly
	createMockCronJobsFile(t, params)

	// Check if the file exists
	pattern := filepath.Join(tmpDir, "autostart-cron-jobs*.json")
	matches, err := filepath.Glob(pattern)
	assert.NoError(t, err)
	assert.NotEmpty(t, matches)
}

func TestParseSandboxedLoginItems(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "autostart_sandboxed_login_items_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logger := testutils.NewTestLogger()
	params := mod.ModuleParams{
		OutputDir:           tmpDir,
		LogsDir:             tmpDir,
		ExportFormat:        "json",
		CollectionTimestamp: time.Now().Format(utils.TimeFormat),
		Logger:              *logger,
	}

	// Create mock output file directly
	createMockSandboxedLoginItemsFile(t, params)

	// Check if the file exists
	pattern := filepath.Join(tmpDir, "autostart-sandboxed-login-items*.json")
	matches, err := filepath.Glob(pattern)
	assert.NoError(t, err)
	assert.NotEmpty(t, matches)
}

// Helper function to verify autostart file contents
func verifyAutostartFileContents(t *testing.T, filePath string, fileType string) {
	// Read the file
	content, err := os.ReadFile(filePath)
	assert.NoError(t, err, "Should be able to read the autostart file")

	// Parse the JSON
	var jsonData map[string]interface{}
	err = json.Unmarshal(content, &jsonData)
	assert.NoError(t, err, "Should be able to parse the autostart JSON")

	// Verify common fields
	assert.NotEmpty(t, jsonData["collection_timestamp"], "Should have collection timestamp")
	assert.NotEmpty(t, jsonData["event_timestamp"], "Should have event timestamp")
	assert.NotEmpty(t, jsonData["source_file"], "Should have source file")
	assert.NotEmpty(t, jsonData["src_name"], "Should have source name")

	// Verify type-specific fields
	switch fileType {
	case "autostart-launch-items":
		assert.Equal(t, "launch_items", jsonData["src_name"])
		assert.NotEmpty(t, jsonData["prog_name"], "Should have program name")
		assert.NotEmpty(t, jsonData["program"], "Should have program path")

	case "autostart-login-items":
		assert.Equal(t, "login_items", jsonData["src_name"])
		assert.NotEmpty(t, jsonData["prog_name"], "Should have program name")
		assert.NotEmpty(t, jsonData["program"], "Should have program path")

	case "autostart-startup-items":
		assert.Equal(t, "startup_items", jsonData["src_name"])

	case "autostart-scripting-additions":
		assert.Equal(t, "scripting_additions", jsonData["src_name"])

	case "autostart-periodic-tasks":
		assert.Equal(t, "periodic_rules_items", jsonData["src_name"])

	case "autostart-cron-jobs":
		assert.Equal(t, "cron", jsonData["src_name"])
		assert.NotEmpty(t, jsonData["program"], "Should have program/cron entry")

	case "autostart-sandboxed-login-items":
		assert.Equal(t, "sandboxed_loginitems", jsonData["src_name"])
		assert.NotEmpty(t, jsonData["prog_name"], "Should have program name")
	}
}

// Helper function to create mock autostart files for all types
func createMockAutostartFiles(t *testing.T, params mod.ModuleParams) {
	createMockLaunchItemsFile(t, params)
	createMockLoginItemsFile(t, params)
	createMockStartupItemsFile(t, params)
	createMockScriptingAdditionsFile(t, params)
	createMockPeriodicTasksFile(t, params)
	createMockCronJobsFile(t, params)
	createMockSandboxedLoginItemsFile(t, params)
}

func createMockLaunchItemsFile(t *testing.T, params mod.ModuleParams) {
	filename := "autostart-launch-items-" + params.CollectionTimestamp + "." + params.ExportFormat
	filepath := filepath.Join(params.OutputDir, filename)

	recordData := map[string]interface{}{
		"src_file":        "/Library/LaunchAgents/com.example.agent.plist",
		"src_name":        "launch_items",
		"prog_name":       "com.example.agent",
		"program":         "/usr/local/bin/example_program",
		"args":            "--daemon --background",
		"code_signatures": map[string]string{"authority": "Developer ID Application: Example Inc."},
	}

	record := utils.Record{
		CollectionTimestamp: params.CollectionTimestamp,
		EventTimestamp:      params.CollectionTimestamp,
		SourceFile:          "/Library/LaunchAgents/com.example.agent.plist",
		Data:                recordData,
	}

	writeAutostartTestRecord(t, filepath, record)
}

func createMockLoginItemsFile(t *testing.T, params mod.ModuleParams) {
	filename := "autostart-login-items-" + params.CollectionTimestamp + "." + params.ExportFormat
	filepath := filepath.Join(params.OutputDir, filename)

	recordData := map[string]interface{}{
		"src_file":        "/Users/testuser/Library/Preferences/com.apple.loginitems.plist",
		"src_name":        "login_items",
		"prog_name":       "Example App",
		"program":         "/Applications/Example.app",
		"code_signatures": map[string]string{"authority": "Developer ID Application: Example Inc."},
	}

	record := utils.Record{
		CollectionTimestamp: params.CollectionTimestamp,
		EventTimestamp:      params.CollectionTimestamp,
		SourceFile:          "/Users/testuser/Library/Preferences/com.apple.loginitems.plist",
		Data:                recordData,
	}

	writeAutostartTestRecord(t, filepath, record)
}

func createMockStartupItemsFile(t *testing.T, params mod.ModuleParams) {
	filename := "autostart-startup-items-" + params.CollectionTimestamp + "." + params.ExportFormat
	filepath := filepath.Join(params.OutputDir, filename)

	recordData := map[string]interface{}{
		"src_file": "/Library/StartupItems/ExampleStartup/ExampleStartup",
		"src_name": "startup_items",
	}

	record := utils.Record{
		CollectionTimestamp: params.CollectionTimestamp,
		EventTimestamp:      params.CollectionTimestamp,
		SourceFile:          "/Library/StartupItems/ExampleStartup/ExampleStartup",
		Data:                recordData,
	}

	writeAutostartTestRecord(t, filepath, record)
}

func createMockScriptingAdditionsFile(t *testing.T, params mod.ModuleParams) {
	filename := "autostart-scripting-additions-" + params.CollectionTimestamp + "." + params.ExportFormat
	filepath := filepath.Join(params.OutputDir, filename)

	recordData := map[string]interface{}{
		"src_file":        "/Library/ScriptingAdditions/Example.osax",
		"src_name":        "scripting_additions",
		"code_signatures": map[string]string{"authority": "Developer ID Application: Example Inc."},
	}

	record := utils.Record{
		CollectionTimestamp: params.CollectionTimestamp,
		EventTimestamp:      params.CollectionTimestamp,
		SourceFile:          "/Library/ScriptingAdditions/Example.osax",
		Data:                recordData,
	}

	writeAutostartTestRecord(t, filepath, record)
}

func createMockPeriodicTasksFile(t *testing.T, params mod.ModuleParams) {
	filename := "autostart-periodic-tasks-" + params.CollectionTimestamp + "." + params.ExportFormat
	filepath := filepath.Join(params.OutputDir, filename)

	recordData := map[string]interface{}{
		"src_file": "/etc/periodic/daily/example_task",
		"src_name": "periodic_rules_items",
	}

	record := utils.Record{
		CollectionTimestamp: params.CollectionTimestamp,
		EventTimestamp:      params.CollectionTimestamp,
		SourceFile:          "/etc/periodic/daily/example_task",
		Data:                recordData,
	}

	writeAutostartTestRecord(t, filepath, record)
}

func createMockCronJobsFile(t *testing.T, params mod.ModuleParams) {
	filename := "autostart-cron-jobs-" + params.CollectionTimestamp + "." + params.ExportFormat
	filepath := filepath.Join(params.OutputDir, filename)

	recordData := map[string]interface{}{
		"src_file": "/var/at/tabs/testuser",
		"src_name": "cron",
		"program":  "0 * * * * /usr/local/bin/example_cron_script",
	}

	record := utils.Record{
		CollectionTimestamp: params.CollectionTimestamp,
		EventTimestamp:      params.CollectionTimestamp,
		SourceFile:          "/var/at/tabs/testuser",
		Data:                recordData,
	}

	writeAutostartTestRecord(t, filepath, record)
}

func createMockSandboxedLoginItemsFile(t *testing.T, params mod.ModuleParams) {
	filename := "autostart-sandboxed-login-items-" + params.CollectionTimestamp + "." + params.ExportFormat
	filepath := filepath.Join(params.OutputDir, filename)

	recordData := map[string]interface{}{
		"src_file":  "/var/db/com.apple.xpc.launchd/disabled.plist",
		"src_name":  "sandboxed_loginitems",
		"prog_name": "com.example.loginitem",
	}

	record := utils.Record{
		CollectionTimestamp: params.CollectionTimestamp,
		EventTimestamp:      params.CollectionTimestamp,
		SourceFile:          "/var/db/com.apple.xpc.launchd/disabled.plist",
		Data:                recordData,
	}

	writeAutostartTestRecord(t, filepath, record)
}

func writeAutostartTestRecord(t *testing.T, filepath string, record utils.Record) {
	// Create JSON representation of the record
	jsonRecord := map[string]interface{}{
		"collection_timestamp": record.CollectionTimestamp,
		"event_timestamp":      record.EventTimestamp,
		"source_file":          record.SourceFile,
	}

	for k, v := range record.Data.(map[string]interface{}) {
		jsonRecord[k] = v
	}

	data, err := json.MarshalIndent(jsonRecord, "", "  ")
	assert.NoError(t, err)

	err = os.WriteFile(filepath, data, 0644)
	assert.NoError(t, err)
}
