package systeminfo

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"howett.net/plist"

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
		OutputDir:           ".",
		LogsDir:             tmpDir,
		ExportFormat:        "json",
		CollectionTimestamp: time.Now().Format(utils.TimeFormat),
		Logger:              *logger,
	}

	// Create module instance
	module := &SystemInfoModule{
		Name:        "systeminfo",
		Description: "Collects system information",
	}

	// Test GetName
	t.Run("GetName", func(t *testing.T) {
		assert.Equal(t, "systeminfo", module.GetName())
	})

	// Test GetDescription
	t.Run("GetDescription", func(t *testing.T) {
		assert.Contains(t, module.GetDescription(), "system information")
	})

	// Test Run method
	t.Run("Run", func(t *testing.T) {
		// Create mock system files
		globalPrefsPath := filepath.Join(tmpDir, "Library", "Preferences", ".GlobalPreferences.plist")
		systemConfigPath := filepath.Join(tmpDir, "Library", "Preferences", "SystemConfiguration", "preferences.plist")
		systemVersionPath := filepath.Join(tmpDir, "System", "Library", "CoreServices", "SystemVersion.plist")

		// Create directories
		err := os.MkdirAll(filepath.Dir(globalPrefsPath), 0755)
		assert.NoError(t, err)
		err = os.MkdirAll(filepath.Dir(systemConfigPath), 0755)
		assert.NoError(t, err)
		err = os.MkdirAll(filepath.Dir(systemVersionPath), 0755)
		assert.NoError(t, err)

		// Create mock plist files
		globalPrefs := map[string]interface{}{
			"AppleLanguages": []string{"en-US"},
			"AppleLocale":    "en_US",
		}
		systemConfig := map[string]interface{}{
			"System": map[string]interface{}{
				"ComputerName": "TestComputer",
				"HostName":     "testcomputer.local",
			},
		}
		systemVersion := map[string]interface{}{
			"ProductBuildVersion": "22E261",
			"ProductName":         "macOS",
			"ProductVersion":      "13.3.1",
		}

		// Write mock files
		err = writePlistFile(globalPrefsPath, globalPrefs)
		assert.NoError(t, err)
		err = writePlistFile(systemConfigPath, systemConfig)
		assert.NoError(t, err)
		err = writePlistFile(systemVersionPath, systemVersion)
		assert.NoError(t, err)

		// Set mock paths
		module.GlobalPrefsPath = globalPrefsPath
		module.SystemConfigPath = systemConfigPath
		module.SystemVersionPath = systemVersionPath

		// Run the module
		err = module.Run(params)
		assert.NoError(t, err)

		// Verify output file exists
		outputFile := filepath.Join(tmpDir, "systeminfo.json")
		assert.FileExists(t, outputFile)

		// Read and verify output content
		data, err := os.ReadFile(outputFile)
		assert.NoError(t, err)

		var output map[string]interface{}
		err = json.Unmarshal(data, &output)
		assert.NoError(t, err)

		// Verify expected fields
		assert.Equal(t, "en-US", output["system_language"])
		assert.Equal(t, "en_US", output["system_locale"])
		assert.Contains(t, output, "product_build_version", "product_build_version should be present")
		assert.Contains(t, output, "product_version", "product_version should be present")
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

// TestRunWithMockFiles tests the Run method with mock system files
func TestRunWithMockFiles(t *testing.T) {
	// Create temporary directory for test outputs
	tmpDir, err := os.MkdirTemp("", "systeminfo_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logger := testutils.NewTestLogger()

	// Setup test parameters
	params := mod.ModuleParams{
		OutputDir:           ".",
		LogsDir:             tmpDir,
		ExportFormat:        "json",
		CollectionTimestamp: time.Now().Format(utils.TimeFormat),
		Logger:              *logger,
	}

	// Create mock system files
	globalPrefsPath := filepath.Join(tmpDir, "Library", "Preferences", ".GlobalPreferences.plist")
	systemConfigPath := filepath.Join(tmpDir, "Library", "Preferences", "SystemConfiguration", "preferences.plist")
	systemVersionPath := filepath.Join(tmpDir, "System", "Library", "CoreServices", "SystemVersion.plist")

	// Create directories
	err = os.MkdirAll(filepath.Dir(globalPrefsPath), 0755)
	assert.NoError(t, err)
	err = os.MkdirAll(filepath.Dir(systemConfigPath), 0755)
	assert.NoError(t, err)
	err = os.MkdirAll(filepath.Dir(systemVersionPath), 0755)
	assert.NoError(t, err)

	// Create mock plist files
	globalPrefs := map[string]interface{}{
		"AppleLanguages":      []string{"en-US"},
		"AppleLocale":         "en_US",
		"AppleInterfaceStyle": "Dark",
	}
	systemConfig := map[string]interface{}{
		"System": map[string]interface{}{
			"ComputerName": "TestComputer",
			"HostName":     "testcomputer.local",
		},
	}
	systemVersion := map[string]interface{}{
		"ProductBuildVersion": "22E261",
		"ProductName":         "macOS",
		"ProductVersion":      "13.3.1",
	}

	// Write mock files
	err = writePlistFile(globalPrefsPath, globalPrefs)
	assert.NoError(t, err)
	err = writePlistFile(systemConfigPath, systemConfig)
	assert.NoError(t, err)
	err = writePlistFile(systemVersionPath, systemVersion)
	assert.NoError(t, err)

	// Create module instance
	module := &SystemInfoModule{
		Name:              "systeminfo",
		Description:       "Collects system information",
		GlobalPrefsPath:   globalPrefsPath,
		SystemConfigPath:  systemConfigPath,
		SystemVersionPath: systemVersionPath,
	}

	// Run the module
	err = module.Run(params)
	assert.NoError(t, err)

	// Verify output file exists
	outputFile := filepath.Join(tmpDir, "systeminfo.json")
	assert.FileExists(t, outputFile)

	// Read and verify output content
	data, err := os.ReadFile(outputFile)
	assert.NoError(t, err)

	var output map[string]interface{}
	err = json.Unmarshal(data, &output)
	assert.NoError(t, err)

	// Verify expected fields
	assert.Equal(t, "en-US", output["system_language"])
	assert.Equal(t, "en_US", output["system_locale"])
	assert.Equal(t, "Dark", output["system_appearance"])
	assert.Contains(t, output, "product_build_version", "product_build_version should be present")
	assert.Contains(t, output, "product_version", "product_version should be present")
}

// TestProcessGlobalPrefs tests the processGlobalPrefs function
func TestProcessGlobalPrefs(t *testing.T) {
	tests := []struct {
		name        string
		globalPrefs map[string]interface{}
		want        map[string]interface{}
	}{
		{
			name: "Dark mode with all preferences",
			globalPrefs: map[string]interface{}{
				"AppleInterfaceStyle": "Dark",
				"AppleLanguages":      []interface{}{"en-US", "ja-JP"},
				"AppleLocale":         "en_US",
				"AppleKeyboardLayout": "com.apple.keylayout.US",
			},
			want: map[string]interface{}{
				"system_appearance": "Dark",
				"system_language":   "en-US",
				"system_locale":     "en_US",
				"keyboard_layout":   "com.apple.keylayout.US",
			},
		},
		{
			name: "Light mode with missing preferences",
			globalPrefs: map[string]interface{}{
				"AppleLanguages": []interface{}{"fr-FR"},
			},
			want: map[string]interface{}{
				"system_appearance": "Light",
				"system_language":   "fr-FR",
			},
		},
		{
			name:        "Empty preferences",
			globalPrefs: map[string]interface{}{},
			want: map[string]interface{}{
				"system_appearance": "Light",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recordData := make(map[string]interface{})
			processGlobalPrefs(tt.globalPrefs, recordData)
			assert.Equal(t, tt.want, recordData)
		})
	}
}

// Helper function to write plist files
func writePlistFile(path string, data map[string]interface{}) error {
	plistData, err := plist.Marshal(data, plist.BinaryFormat)
	if err != nil {
		return err
	}
	return os.WriteFile(path, plistData, 0600)
}
