package usbhistory

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gnzdotmx/ishinobu/pkg/mod"
	"github.com/gnzdotmx/ishinobu/pkg/testutils"
	"github.com/gnzdotmx/ishinobu/pkg/utils"

	"github.com/stretchr/testify/assert"
)

var (
	errUnexpectedIoregArgs          = errors.New("unexpected ioreg arguments")
	errUnexpectedSystemProfilerArgs = errors.New("unexpected system_profiler arguments")
	errUnexpectedLogArgs            = errors.New("unexpected log arguments")
	errUnexpectedCommand            = errors.New("unexpected command")
)

// MockLogger implements the mod.Logger interface for testing
type MockLogger struct{}

func (l *MockLogger) Debug(format string, args ...interface{})  {}
func (l *MockLogger) Info(format string, args ...interface{})   {}
func (l *MockLogger) Warn(format string, args ...interface{})   {}
func (l *MockLogger) Error(format string, args ...interface{})  {}
func (l *MockLogger) Fatal(format string, args ...interface{})  {}
func (l *MockLogger) Panic(format string, args ...interface{})  {}
func (l *MockLogger) Printf(format string, args ...interface{}) {}

// Mock ExecCommand for testing
func mockExecCommand(name string, arg ...string) ([]byte, error) {
	switch name {
	case "ioreg":
		if len(arg) != 5 || arg[0] != "-p" || arg[1] != "IOUSB" || arg[2] != "-l" || arg[3] != "-w" || arg[4] != "0" {
			return nil, errUnexpectedIoregArgs
		}
		return []byte(`+-o Root  <class IORegistryEntry, id 0x100000100, retain 35>
| +-o AppleT8112USBXHCI@00000000  <class AppleT8112USBXHCI, id 0x100000425, registered, matched, active, busy 0 (444 ms), retain 71>`), nil

	case "system_profiler":
		if len(arg) != 2 || arg[0] != "SPUSBDataType" || arg[1] != "-json" {
			return nil, errUnexpectedSystemProfilerArgs
		}
		return []byte(`{"SPUSBDataType":[{"_name":"Root","_items":[{"_name":"USB31Bus"}]}]}`), nil

	case "log":
		if len(arg) < 2 || arg[0] != "show" {
			return nil, errUnexpectedLogArgs
		}
		return []byte(`[{"timestamp":"2025-05-30 23:14:36.492903+0900","eventMessage":"enumerated 0x05ac/12a8/1601 (iPhone / 1) at 480 Mbps"}]`), nil

	default:
		return nil, errUnexpectedCommand
	}
}

func TestUSBHistoryModuleBasic(t *testing.T) {
	// Create temporary directory for test outputs
	tmpDir, err := os.MkdirTemp("", "usbhistory_basic_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Save the original ExecCommand and restore it after the test
	originalExecCommand := utils.ExecCommand
	utils.ExecCommand = mockExecCommand
	defer func() { utils.ExecCommand = originalExecCommand }()

	module := &USBHistoryModule{Name: "usbhistory", Description: "Collects USB device connection history and metadata"}

	// Test module name and description
	if module.GetName() != "usbhistory" {
		t.Errorf("Expected module name 'usbhistory', got '%s'", module.GetName())
	}

	if module.GetDescription() != "Collects USB device connection history and metadata" {
		t.Errorf("Expected module description 'Collects USB device connection history and metadata', got '%s'", module.GetDescription())
	}

	// Test module run
	params := mod.ModuleParams{
		Logger:              *testutils.NewTestLogger(),
		CollectionTimestamp: "2025-05-31T11:55:43+09:00",
		LogsDir:             tmpDir,
		OutputDir:           ".",
		ExportFormat:        "json",
	}

	err = module.Run(params)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestUSBHistoryModuleOutputFiles(t *testing.T) {
	// Create temporary directory for test outputs
	tmpDir, err := os.MkdirTemp("", "usbhistory_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Save the original ExecCommand and restore it after the test
	originalExecCommand := utils.ExecCommand
	defer func() { utils.ExecCommand = originalExecCommand }()

	// Create test parameters
	params := mod.ModuleParams{
		CollectionTimestamp: "2025-05-30T23:46:41+09:00",
		OutputDir:           ".",
		LogsDir:             tmpDir,
		ExportFormat:        "json",
		Logger:              *testutils.NewTestLogger(),
	}

	// Mock the command execution
	utils.ExecCommand = func(name string, arg ...string) ([]byte, error) {
		// Log the command being executed for debugging
		t.Logf("Mock executing command: %s %v", name, arg)

		switch name {
		case "ioreg":
			if len(arg) >= 2 && arg[0] == "-p" && arg[1] == "IOUSB" {
				return []byte(`+-o USB Device@01000000  <class IOUSBDevice, id 0x100000426>
  | {
  |   "USB Vendor Name" = "Test Manufacturer"
  |   "USB Serial Number" = "1234567890"
  | }
  |
+-o USB Hub@02000000  <class IOUSBHub, id 0x100000427>
  | {
  |   "USB Product Name" = "Test Hub"
  |   "USB Vendor Name" = "Test Hub Manufacturer"
  | }`), nil
			}
			return nil, errUnexpectedIoregArgs

		case "system_profiler":
			if len(arg) >= 2 && arg[0] == "SPUSBDataType" && arg[1] == "-json" {
				return []byte(`{
					"SPUSBDataType" : [
						{
							"_name" : "USB 3.1 Bus",
							"host_controller" : "Test USB Controller",
							"_items" : [
								{
									"_name" : "USB Device",
									"manufacturer" : "Test Manufacturer",
									"serial_num" : "1234567890",
									"vendor_id" : "0x05ac",
									"product_id" : "0x8600"
								}
							]
						}
					]
				}`), nil
			}
			return nil, errUnexpectedSystemProfilerArgs

		case "log":
			if len(arg) >= 2 && arg[0] == "show" && strings.Contains(arg[2], "com.apple.iokit.IOUSBHost") {
				return []byte(`[
					{
						"timestamp" : "2025-05-30T23:46:41+09:00",
						"subsystem" : "com.apple.iokit.IOUSBHost",
						"eventMessage" : "USB device connected",
						"messageType" : "Default",
						"processImagePath" : "/System/Library/Extensions/IOUSBHostFamily.kext/Contents/MacOS/IOUSBHostFamily"
					}
				]`), nil
			}
			return nil, errUnexpectedLogArgs

		default:
			return nil, errUnexpectedCommand
		}
	}

	// Create module instance
	module := &USBHistoryModule{Name: "usbhistory"}
	err = module.Run(params)
	assert.NoError(t, err)

	// Define expected output files
	usbListFile := filepath.Join(tmpDir, "usbhistoryUsbList.json")
	usbHistoryFile := filepath.Join(tmpDir, "usbhistoryUsbHistory.json")
	usbRegistryFile := filepath.Join(tmpDir, "usbhistoryUsbRegistry.json")

	// Debug: list files in output dir
	files, err := os.ReadDir(tmpDir)
	assert.NoError(t, err)
	t.Log("Files in output directory:")
	for _, file := range files {
		t.Logf("- %s", file.Name())
	}

	// Helper function to read and verify JSON records
	readJSONRecords := func(t *testing.T, filePath string) []map[string]interface{} {
		data, err := os.ReadFile(filePath)
		assert.NoError(t, err, "Failed to read file: %s", filePath)

		// Debug: Print the raw file content
		t.Logf("Content of %s:\n%s", filePath, string(data))

		var records []map[string]interface{}
		decoder := json.NewDecoder(strings.NewReader(string(data)))
		for decoder.More() {
			var record map[string]interface{}
			err := decoder.Decode(&record)
			assert.NoError(t, err, "Failed to decode JSON from file: %s", filePath)
			records = append(records, record)
		}

		// Debug: Print the parsed records
		t.Logf("Parsed records from %s: %+v", filePath, records)

		return records
	}

	// Verify each output file exists and contains valid data
	verifyUSBListOutput(t, usbListFile, readJSONRecords)
	verifyUSBHistoryOutput(t, usbHistoryFile, readJSONRecords)
	verifyUSBRegistryOutput(t, usbRegistryFile, readJSONRecords)
}

func verifyUSBRegistryOutput(t *testing.T, filePath string, readRecords func(*testing.T, string) []map[string]interface{}) {
	records := readRecords(t, filePath)
	assert.NotEmpty(t, records, "No records found in USB registry output")

	for _, record := range records {
		assert.Equal(t, "ioreg", record["source_file"], "Unexpected source_file value")
		assert.Contains(t, record, "clean_name", "Record should contain clean_name field")
		assert.Contains(t, record, "properties", "Record should contain properties field")

		cleanName, ok := record["clean_name"].(string)
		assert.True(t, ok, "clean_name should be a string")

		if strings.Contains(cleanName, "USB Device") {
			properties, ok := record["properties"].(map[string]interface{})
			assert.True(t, ok, "Properties should be a map[string]interface{}")
			assert.Contains(t, properties, "USB Vendor Name", "Properties should contain USB Vendor Name field")
			assert.Contains(t, properties, "USB Serial Number", "Properties should contain USB Serial Number field")
		}
	}
}

func verifyUSBListOutput(t *testing.T, filePath string, readRecords func(*testing.T, string) []map[string]interface{}) {
	records := readRecords(t, filePath)
	assert.NotEmpty(t, records, "No records found in USB list output")

	for _, record := range records {
		assert.Equal(t, "system_profiler", record["source_file"], "Unexpected source_file value")
		assert.Contains(t, record, "_name", "Record should contain _name field")

		if items, ok := record["_items"].([]interface{}); ok {
			for _, item := range items {
				if deviceData, ok := item.(map[string]interface{}); ok {
					if deviceData["_name"] == "USB Device" {
						assert.Contains(t, deviceData, "manufacturer", "Device data should contain manufacturer field")
						assert.Contains(t, deviceData, "serial_num", "Device data should contain serial_num field")
						assert.Contains(t, deviceData, "vendor_id", "Device data should contain vendor_id field")
						assert.Contains(t, deviceData, "product_id", "Device data should contain product_id field")
					}
				}
			}
		}
	}
}

func verifyUSBHistoryOutput(t *testing.T, filePath string, readRecords func(*testing.T, string) []map[string]interface{}) {
	records := readRecords(t, filePath)
	assert.NotEmpty(t, records, "No records found in USB history output")

	for _, record := range records {
		assert.Equal(t, "system_log", record["source_file"], "Unexpected source_file value")
		assert.Contains(t, record, "subsystem", "Record should contain subsystem field")
		assert.Contains(t, record, "eventmessage", "Record should contain eventMessage field")
		assert.Contains(t, record, "timestamp", "Record should contain timestamp field")
		assert.Contains(t, record, "messagetype", "Record should contain messageType field")
		assert.Contains(t, record, "processimagepath", "Record should contain processImagePath field")
	}
}

func TestParseIORegOutput(t *testing.T) {
	input := `+-o TestDevice@00000000  <class TestClass, id 0x123456>
  | {
  |   "TestProperty" = "TestValue"
  |   "NumberProperty" = 123
  |   "BoolProperty" = Yes
  | }`

	devices, err := parseIORegOutput(input)
	assert.NoError(t, err)
	assert.NotEmpty(t, devices)

	device := devices[0]
	assert.Equal(t, "TestDevice@00000000", device["clean_name"])
	assert.Equal(t, "class TestClass, id 0x123456", device["device_info"])

	properties, ok := device["properties"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "TestValue", properties["TestProperty"])
	assert.Equal(t, "123", properties["NumberProperty"])
	assert.Equal(t, true, properties["BoolProperty"])
}

func TestCleanPropertyValue(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected interface{}
	}{
		{
			name:     "string with escaped quotes",
			input:    "\"test\\\"value\"",
			expected: "test\"value",
		},
		{
			name:     "number",
			input:    "123",
			expected: "123",
		},
		{
			name:     "boolean true",
			input:    "Yes",
			expected: true,
		},
		{
			name:     "boolean false",
			input:    "No",
			expected: false,
		},
		{
			name:     "json object",
			input:    "{\"key\": \"value\"}",
			expected: map[string]interface{}{"key": "value"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanPropertyValue(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
