package autostart

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"howett.net/plist"

	"github.com/gnzdotmx/ishinobu/pkg/mod"
	"github.com/gnzdotmx/ishinobu/pkg/modules/testutils"
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
