package autostart

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"howett.net/plist"

	"github.com/gnzdotmx/ishinobu/pkg/mod"
	"github.com/gnzdotmx/ishinobu/pkg/testutils"
	"github.com/gnzdotmx/ishinobu/pkg/utils"
)

func TestParseLaunchItems(t *testing.T) {
	// Create temporary directory for test plist files
	tmpDir, err := os.MkdirTemp("", "launch_items_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create a test plist file
	testPlistPath := filepath.Join(tmpDir, "test.plist")
	testPlist := map[string]interface{}{
		"Label":            "com.test.launch",
		"RunAtLoad":        true,
		"KeepAlive":        true,
		"ProgramArguments": []interface{}{"/usr/bin/test", "-arg1", "-arg2"},
	}

	plistData, err := plist.Marshal(testPlist, plist.XMLFormat)
	assert.NoError(t, err)
	err = os.WriteFile(testPlistPath, plistData, 0644)
	assert.NoError(t, err)

	// Set up test parameters
	logger := testutils.NewTestLogger()
	params := mod.ModuleParams{
		LogsDir:             tmpDir,
		OutputDir:           tmpDir,
		ExportFormat:        "json",
		CollectionTimestamp: "2024-01-01T00:00:00Z",
		Logger:              *logger,
	}

	// Create test writer
	testWriter := &testutils.TestDataWriter{Records: []utils.Record{}}

	// Run parse function with test file
	err = parseLaunchItems([]string{testPlistPath}, testWriter, params)
	assert.NoError(t, err)

	// Verify output
	assert.Equal(t, 1, len(testWriter.Records))
	record := testWriter.Records[0]

	data, ok := record.Data.(map[string]interface{})
	assert.True(t, ok)
	fmt.Println(data)
	assert.Equal(t, testPlistPath, data["src_file"])
	assert.Equal(t, "launch_items", data["src_name"])
	assert.Equal(t, "com.test.launch", data["prog_name"])
	assert.Equal(t, "/usr/bin/test", data["program"])
	assert.Equal(t, "-arg1 -arg2", data["args"])
}

func TestParseLoginItems(t *testing.T) {
	// Create temporary directory for test plist files
	tmpDir, err := os.MkdirTemp("", "login_items_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create a test login items plist file
	testLoginPlistPath := filepath.Join(tmpDir, "com.apple.loginitems.plist")

	// Create a login items plist structure
	testLoginItems := []map[string]interface{}{
		{
			"SessionItems": map[string]interface{}{
				"CustomListItems": []map[string]interface{}{
					{
						"Name":    "TestApp",
						"Program": "/Applications/TestApp.app",
					},
				},
			},
		},
	}

	plistData, err := plist.Marshal(testLoginItems, plist.XMLFormat)
	assert.NoError(t, err)
	err = os.WriteFile(testLoginPlistPath, plistData, 0644)
	assert.NoError(t, err)

	// Set up test parameters
	logger := testutils.NewTestLogger()
	params := mod.ModuleParams{
		LogsDir:             tmpDir,
		OutputDir:           tmpDir,
		ExportFormat:        "json",
		CollectionTimestamp: "2024-01-01T00:00:00Z",
		Logger:              *logger,
	}

	// Create test writer
	testWriter := &testutils.TestDataWriter{Records: []utils.Record{}}

	// Run parse function with test file
	err = parseLoginItems(testLoginPlistPath, testWriter, params)
	assert.NoError(t, err)

	// Verify output
	assert.Equal(t, 1, len(testWriter.Records))
	record := testWriter.Records[0]

	data, ok := record.Data.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, testLoginPlistPath, data["src_file"])
	assert.Equal(t, "login_items", data["src_name"])
	assert.Equal(t, "TestApp", data["prog_name"])
	assert.Equal(t, "/Applications/TestApp.app", data["program"])
}

func TestParseStartupItems(t *testing.T) {
	// Create temporary directory for test startup items
	tmpDir, err := os.MkdirTemp("", "startup_items_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create test startup item file
	startupDir := filepath.Join(tmpDir, "StartupItems", "TestStartup")
	err = os.MkdirAll(startupDir, 0755)
	assert.NoError(t, err)

	startupFile := filepath.Join(startupDir, "TestStartup")
	err = os.WriteFile(startupFile, []byte("#!/bin/sh\n# Test startup item"), 0755)
	assert.NoError(t, err)

	// Set up test parameters
	logger := testutils.NewTestLogger()
	params := mod.ModuleParams{
		LogsDir:             tmpDir,
		OutputDir:           tmpDir,
		ExportFormat:        "json",
		CollectionTimestamp: "2024-01-01T00:00:00Z",
		Logger:              *logger,
	}

	// Create test writer
	testWriter := &testutils.TestDataWriter{Records: []utils.Record{}}

	// Run parse function with test file
	pattern := filepath.Join(tmpDir, "StartupItems", "*", "*")
	err = parseStartupItems([]string{pattern}, testWriter, params)
	assert.NoError(t, err)

	// Verify output
	assert.Equal(t, 1, len(testWriter.Records))
	record := testWriter.Records[0]

	data, ok := record.Data.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, startupFile, data["src_file"])
	assert.Equal(t, "startup_items", data["src_name"])
}

func TestParseCronJobs(t *testing.T) {
	// Create temporary directory for test cron jobs
	tmpDir, err := os.MkdirTemp("", "cron_jobs_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create test cron file
	cronFile := filepath.Join(tmpDir, "testuser")
	cronContent := `# Test cron file
0 * * * * /usr/bin/test -arg
# Comment line
15 2 * * * /bin/echo "test"`

	err = os.WriteFile(cronFile, []byte(cronContent), 0644)
	assert.NoError(t, err)

	// Set up test parameters
	logger := testutils.NewTestLogger()
	params := mod.ModuleParams{
		LogsDir:             tmpDir,
		OutputDir:           tmpDir,
		ExportFormat:        "json",
		CollectionTimestamp: "2024-01-01T00:00:00Z",
		Logger:              *logger,
	}

	// Create test writer
	testWriter := &testutils.TestDataWriter{Records: []utils.Record{}}

	// Run parse function with test file
	pattern := filepath.Join(tmpDir, "*")
	err = parseCronJobs(pattern, testWriter, params)
	assert.NoError(t, err)

	// Verify output
	assert.Equal(t, 2, len(testWriter.Records))

	// Check first non-comment line
	record1 := testWriter.Records[0]
	data1, ok := record1.Data.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, cronFile, data1["src_file"])
	assert.Equal(t, "cron", data1["src_name"])
	assert.Equal(t, "0 * * * * /usr/bin/test -arg", data1["program"])

	// Check second non-comment line
	record2 := testWriter.Records[1]
	data2, ok := record2.Data.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, cronFile, data2["src_file"])
	assert.Equal(t, "cron", data2["src_name"])
	assert.Equal(t, `15 2 * * * /bin/echo "test"`, data2["program"])
}

func TestParseSandboxedLoginItems(t *testing.T) {
	// Create temporary directory for test plist files
	tmpDir, err := os.MkdirTemp("", "sandboxed_login_items_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create a test sandboxed login items plist file
	testPlistPath := filepath.Join(tmpDir, "disabled.1.plist")

	// Create a sandboxed login items plist structure
	testSandboxedItems := map[string]interface{}{
		"com.test.app":    false,
		"com.example.app": false,
		"com.enabled.app": true, // This should be ignored since it's true
	}

	plistData, err := plist.Marshal(testSandboxedItems, plist.XMLFormat)
	assert.NoError(t, err)
	err = os.WriteFile(testPlistPath, plistData, 0644)
	assert.NoError(t, err)

	// Set up test parameters
	logger := testutils.NewTestLogger()
	params := mod.ModuleParams{
		LogsDir:             tmpDir,
		OutputDir:           tmpDir,
		ExportFormat:        "json",
		CollectionTimestamp: "2024-01-01T00:00:00Z",
		Logger:              *logger,
	}

	// Create test writer
	testWriter := &testutils.TestDataWriter{Records: []utils.Record{}}

	// Run parse function with test file
	pattern := filepath.Join(tmpDir, "disabled.*.plist")
	err = parseSandboxedLoginItems(pattern, testWriter, params)
	assert.NoError(t, err)

	// Verify output - should only have 2 records (the false ones)
	assert.Equal(t, 2, len(testWriter.Records))

	// Check both apps were properly captured
	appNames := []string{}
	for _, record := range testWriter.Records {
		data, ok := record.Data.(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, testPlistPath, data["src_file"])
		assert.Equal(t, "sandboxed_loginitems", data["src_name"])
		appNames = append(appNames, data["prog_name"].(string))
	}

	assert.Contains(t, appNames, "com.test.app")
	assert.Contains(t, appNames, "com.example.app")
	assert.NotContains(t, appNames, "com.enabled.app")
}

func TestParseScriptingAdditions(t *testing.T) {
	// Create temporary directory for test scripting additions
	tmpDir, err := os.MkdirTemp("", "scripting_additions_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create a test scripting addition file
	scriptingDir := filepath.Join(tmpDir, "ScriptingAdditions")
	err = os.MkdirAll(scriptingDir, 0755)
	assert.NoError(t, err)

	scriptingFile := filepath.Join(scriptingDir, "TestAddition.osax")
	err = os.WriteFile(scriptingFile, []byte("test scripting addition file content"), 0644)
	assert.NoError(t, err)

	// Set up test parameters
	logger := testutils.NewTestLogger()
	params := mod.ModuleParams{
		LogsDir:             tmpDir,
		OutputDir:           tmpDir,
		ExportFormat:        "json",
		CollectionTimestamp: "2024-01-01T00:00:00Z",
		Logger:              *logger,
	}

	// Create test writer
	testWriter := &testutils.TestDataWriter{Records: []utils.Record{}}

	// Run parse function with test file
	pattern := filepath.Join(scriptingDir, "*.osax")
	err = parseScriptingAdditions([]string{pattern}, testWriter, params)
	assert.NoError(t, err)

	// Verify output
	assert.Equal(t, 1, len(testWriter.Records))
	record := testWriter.Records[0]

	data, ok := record.Data.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, scriptingFile, data["src_file"])
	assert.Equal(t, "scripting_additions", data["src_name"])
}

func TestParsePeriodicTasks(t *testing.T) {
	// Create temporary directory for test periodic tasks
	tmpDir, err := os.MkdirTemp("", "periodic_tasks_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create test periodic task files
	periodicDir := filepath.Join(tmpDir, "periodic")
	dailyDir := filepath.Join(periodicDir, "daily")
	err = os.MkdirAll(dailyDir, 0755)
	assert.NoError(t, err)

	// Create a test periodic task file
	periodicFile := filepath.Join(dailyDir, "clean.sh")
	err = os.WriteFile(periodicFile, []byte("#!/bin/sh\n# Test periodic task"), 0755)
	assert.NoError(t, err)

	// Create a test periodic.conf file
	periodicConfFile := filepath.Join(tmpDir, "periodic.conf")
	err = os.WriteFile(periodicConfFile, []byte("# Test periodic conf\ndaily_clean_enable=\"YES\"\n"), 0644)
	assert.NoError(t, err)

	// Set up test parameters
	logger := testutils.NewTestLogger()
	params := mod.ModuleParams{
		LogsDir:             tmpDir,
		OutputDir:           tmpDir,
		ExportFormat:        "json",
		CollectionTimestamp: "2024-01-01T00:00:00Z",
		Logger:              *logger,
	}

	// Create test writer
	testWriter := &testutils.TestDataWriter{Records: []utils.Record{}}

	// Run parse function with test files
	patterns := []string{
		filepath.Join(tmpDir, "periodic.conf"),
		filepath.Join(periodicDir, "*", "*"),
	}
	err = parsePeriodicTasks(patterns, testWriter, params)
	assert.NoError(t, err)

	// Verify output
	assert.Equal(t, 2, len(testWriter.Records))

	// Check both files were properly captured
	fileNames := []string{periodicFile, periodicConfFile}
	for _, record := range testWriter.Records {
		data, ok := record.Data.(map[string]interface{})
		assert.True(t, ok)
		assert.Contains(t, fileNames, data["src_file"])
		assert.Equal(t, "periodic_rules_items", data["src_name"])
	}
}

func TestAutostartModuleRun(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "autostart_module_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create necessary directory structure for each type of autostart item

	// 1. Launch Agents
	launchAgentsDir := filepath.Join(tmpDir, "Library", "LaunchAgents")
	err = os.MkdirAll(launchAgentsDir, 0755)
	assert.NoError(t, err)

	// Create test plist file
	launchAgentPlist := map[string]interface{}{
		"Label":            "com.test.launch",
		"RunAtLoad":        true,
		"KeepAlive":        true,
		"ProgramArguments": []interface{}{"/usr/bin/test", "-arg1", "-arg2"},
	}
	launchAgentPlistData, err := plist.Marshal(launchAgentPlist, plist.XMLFormat)
	assert.NoError(t, err)
	launchAgentPath := filepath.Join(launchAgentsDir, "com.test.launch.plist")
	err = os.WriteFile(launchAgentPath, launchAgentPlistData, 0644)
	assert.NoError(t, err)

	// 2. Login Items
	userLibraryDir := filepath.Join(tmpDir, "Users", "testuser", "Library")
	prefsDir := filepath.Join(userLibraryDir, "Preferences")
	err = os.MkdirAll(prefsDir, 0755)
	assert.NoError(t, err)

	// Create login items plist
	loginItemsList := []map[string]interface{}{
		{
			"SessionItems": map[string]interface{}{
				"CustomListItems": []map[string]interface{}{
					{
						"Name":    "TestApp",
						"Program": "/Applications/TestApp.app",
					},
				},
			},
		},
	}
	loginItemsData, err := plist.Marshal(loginItemsList, plist.XMLFormat)
	assert.NoError(t, err)
	loginItemsPath := filepath.Join(prefsDir, "com.apple.loginitems.plist")
	err = os.WriteFile(loginItemsPath, loginItemsData, 0644)
	assert.NoError(t, err)

	// Set up test parameters
	logger := testutils.NewTestLogger()
	params := mod.ModuleParams{
		LogsDir:             tmpDir,
		OutputDir:           tmpDir,
		ExportFormat:        "json",
		CollectionTimestamp: "2024-01-01T00:00:00Z",
		Logger:              *logger,
	}

	// Create the module
	module := &AutostartModule{
		Name:        "autostart",
		Description: "Test autostart module",
	}

	// Test GetName and GetDescription
	assert.Equal(t, "autostart", module.GetName())
	assert.Equal(t, "Test autostart module", module.GetDescription())

	// Instead of trying to mock filepath.Glob, we'll test the individual parse functions
	// which have been tested separately in other test functions

	// Test parseLaunchItems directly
	launchItemsWriter, err := utils.NewDataWriter(tmpDir, "test-launch-items.json", "json")
	assert.NoError(t, err)
	err = parseLaunchItems([]string{launchAgentPath}, launchItemsWriter, params)
	assert.NoError(t, err)

	// Test parseLoginItems directly
	loginItemsWriter, err := utils.NewDataWriter(tmpDir, "test-login-items.json", "json")
	assert.NoError(t, err)
	err = parseLoginItems(loginItemsPath, loginItemsWriter, params)
	assert.NoError(t, err)

	// Verify output files were created
	outputPatterns := []string{
		filepath.Join(tmpDir, "test-launch-items.json"),
		filepath.Join(tmpDir, "test-login-items.json"),
	}

	for _, pattern := range outputPatterns {
		_, err := os.Stat(pattern)
		assert.NoError(t, err, "Output file %s should exist", pattern)
	}
}

func TestParseErrorHandling(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "error_handling_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Set up test parameters
	logger := testutils.NewTestLogger()
	params := mod.ModuleParams{
		LogsDir:             tmpDir,
		OutputDir:           tmpDir,
		ExportFormat:        "json",
		CollectionTimestamp: "2024-01-01T00:00:00Z",
		Logger:              *logger,
	}

	// 1. Test error handling in parseLaunchItems - Invalid plist
	invalidPlistDir := filepath.Join(tmpDir, "invalid_plists")
	err = os.MkdirAll(invalidPlistDir, 0755)
	assert.NoError(t, err)
	invalidPlistPath := filepath.Join(invalidPlistDir, "invalid.plist")
	err = os.WriteFile(invalidPlistPath, []byte("this is not a valid plist file"), 0644)
	assert.NoError(t, err)

	testWriter := &testutils.TestDataWriter{Records: []utils.Record{}}
	err = parseLaunchItems([]string{invalidPlistPath}, testWriter, params)
	assert.NoError(t, err, "parseLaunchItems should handle invalid plists without returning an error")
	assert.Empty(t, testWriter.Records, "No records should be written for invalid plists")

	// 2. Test error handling in parseLoginItems - Invalid path
	err = parseLoginItems("/non/existent/path.plist", testWriter, params)
	assert.NoError(t, err, "parseLoginItems should handle missing files without returning an error")

	// 3. Test error handling in parseStartupItems - Empty pattern
	err = parseStartupItems([]string{}, testWriter, params)
	assert.NoError(t, err, "parseStartupItems should handle empty patterns without returning an error")

	// 4. Test error handling in parseScriptingAdditions - Invalid pattern
	err = parseScriptingAdditions([]string{"[invalid-glob-pattern"}, testWriter, params)
	assert.NoError(t, err, "parseScriptingAdditions should handle invalid patterns without returning an error")

	// 5. Test error handling in parsePeriodicTasks - Invalid glob pattern
	err = parsePeriodicTasks([]string{"[invalid-glob-pattern"}, testWriter, params)
	assert.NoError(t, err, "parsePeriodicTasks should handle invalid patterns without returning an error")

	// 6. Test error handling in parseCronJobs - Invalid path
	err = parseCronJobs("/non/existent/path", testWriter, params)
	assert.NoError(t, err, "parseCronJobs should handle invalid paths without returning an error")

	// 7. Test error handling in parseSandboxedLoginItems - Invalid pattern
	err = parseSandboxedLoginItems("[invalid-glob-pattern", testWriter, params)
	assert.Error(t, err, "parseSandboxedLoginItems should return an error with an invalid pattern")

	// 8. Test error handling with an invalid output path - should return error
	badWriter, err := utils.NewDataWriter("/invalid/path", "test.json", "json")
	assert.Error(t, err, "NewDataWriter should return an error with an invalid path")
	assert.Nil(t, badWriter, "Writer should be nil when error occurs")
}
