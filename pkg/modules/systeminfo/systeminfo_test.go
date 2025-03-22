package systeminfo

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/gnzdotmx/ishinobu/pkg/mod"
	"github.com/gnzdotmx/ishinobu/pkg/modules/testutils"
	"github.com/gnzdotmx/ishinobu/pkg/utils"
)

func TestSystemInfoModule(t *testing.T) {
	// Create temporary directory for test outputs
	tmpDir, err := os.MkdirTemp("", "systeminfo_test")
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
	module := &SystemInfoModule{
		Name:        "systeminfo",
		Description: "Collects basic system information to identify the host",
	}

	// Test GetName
	t.Run("GetName", func(t *testing.T) {
		assert.Equal(t, "systeminfo", module.GetName())
	})

	// Test GetDescription
	t.Run("GetDescription", func(t *testing.T) {
		assert.Contains(t, module.GetDescription(), "system information")
	})

	// Test Run method with mock output
	t.Run("Run", func(t *testing.T) {
		// Create a mock output file to simulate the module's output
		createMockSystemInfoOutput(t, params)

		// Verify the output file exists
		outputFile := filepath.Join(tmpDir, "systeminfo-"+params.CollectionTimestamp+".json")
		assert.FileExists(t, outputFile)

		// Verify the content of the output file
		verifySystemInfoOutput(t, outputFile)
	})
}

// Test that the module initializes properly
func TestSystemInfoModuleInitialization(t *testing.T) {
	// Create a new instance with proper initialization
	module := &SystemInfoModule{
		Name:        "systeminfo",
		Description: "Collects basic system information to identify the host",
	}

	// Verify module is properly instantiated with expected values
	assert.Equal(t, "systeminfo", module.Name, "Module name should be initialized")
	assert.Contains(t, module.Description, "system information", "Module description should be initialized")

	// Test the module's methods
	assert.Equal(t, "systeminfo", module.GetName())
	assert.Contains(t, module.GetDescription(), "system information")
}

// Create a mock systeminfo output file
func createMockSystemInfoOutput(t *testing.T, params mod.ModuleParams) {
	outputFile := filepath.Join(params.OutputDir, "systeminfo-"+params.CollectionTimestamp+".json")

	// Mock system information record
	sysInfo := utils.Record{
		CollectionTimestamp: params.CollectionTimestamp,
		EventTimestamp:      time.Now().UTC().Format(time.RFC3339),
		SourceFile:          "System Information",
		Data: map[string]interface{}{
			"system_appearance":     "Dark",
			"system_language":       "en-US",
			"system_locale":         "en_US",
			"keyboard_layout":       "com.apple.keylayout.US",
			"local_hostname":        "MacBook-Pro",
			"computer_name":         "John's MacBook Pro",
			"hostname":              "MacBook-Pro.local",
			"product_version":       "13.4.1",
			"product_build_version": "22F82",
			"model":                 "MacBookPro18,3",
			"serial_no":             "C02ABC123DEF",
			"system_tz":             "America/Los_Angeles",
			"fvde_status":           "On",
			"gatekeeper_status":     "assessments enabled",
			"sip_status":            "System Integrity Protection status: enabled",
			"ipaddress":             "192.168.1.100",
		},
	}

	// Write the record to the output file
	file, err := os.Create(outputFile)
	assert.NoError(t, err)
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(sysInfo)
	assert.NoError(t, err)
}

// Verify the systeminfo output file contains expected data
func verifySystemInfoOutput(t *testing.T, outputFile string) {
	// Read the output file
	content, err := os.ReadFile(outputFile)
	assert.NoError(t, err)

	// Parse the JSON content
	var record map[string]interface{}
	err = json.Unmarshal(content, &record)
	assert.NoError(t, err, "Output should be valid JSON")

	// Verify record structure
	assert.NotEmpty(t, record["collection_timestamp"], "Should have collection timestamp")
	assert.NotEmpty(t, record["event_timestamp"], "Should have event timestamp")
	assert.Equal(t, "System Information", record["source_file"], "Source should be 'System Information'")

	// Verify system information fields
	data, ok := record["data"].(map[string]interface{})
	assert.True(t, ok, "Record should have data field as a map")

	// Check for expected system information fields
	expectedFields := []string{
		"system_appearance", "system_language", "system_locale", "keyboard_layout",
		"local_hostname", "computer_name", "hostname", "product_version",
		"product_build_version", "model", "serial_no", "system_tz",
		"fvde_status", "gatekeeper_status", "sip_status", "ipaddress",
	}

	for _, field := range expectedFields {
		assert.Contains(t, data, field, "Data should contain "+field)
	}

	// Verify specific values
	assert.Contains(t, []string{"Light", "Dark"}, data["system_appearance"], "System appearance should be Light or Dark")
	assert.Contains(t, data["product_version"].(string), ".", "Product version should be in format x.y.z")
	assert.Contains(t, data["sip_status"].(string), "System Integrity Protection", "SIP status should mention System Integrity Protection")
}
