// This module collects basic system information to identify the host.
// It collects the following information:
// - System appearance (Dark/Light mode)
// - System language
// - System locale
// - Keyboard layout
// - Computer name
// - Hostname
package systeminfo

import (
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/gnzdotmx/ishinobu/pkg/mod"
	"github.com/gnzdotmx/ishinobu/pkg/utils"
)

type SystemInfoModule struct {
	Name        string
	Description string
}

func init() {
	module := &SystemInfoModule{
		Name:        "systeminfo",
		Description: "Collects basic system information to identify the host"}
	mod.RegisterModule(module)
}

func (m *SystemInfoModule) GetName() string {
	return m.Name
}

func (m *SystemInfoModule) GetDescription() string {
	return m.Description
}

func (m *SystemInfoModule) Run(params mod.ModuleParams) error {
	outputFileName := utils.GetOutputFileName(m.GetName(), params.ExportFormat, params.OutputDir)
	writer, err := utils.NewDataWriter(params.LogsDir, outputFileName, params.ExportFormat)
	if err != nil {
		return err
	}

	// Read system preferences
	globalPrefsPath := "/Library/Preferences/.GlobalPreferences.plist"
	systemConfigPath := "/Library/Preferences/SystemConfiguration/preferences.plist"
	systemVersionPath := "/System/Library/CoreServices/SystemVersion.plist"

	// Collect system information
	recordData := make(map[string]interface{})

	// Read and parse the plist file
	globalPrefsFile, err := os.ReadFile(globalPrefsPath)
	if err != nil {
		params.Logger.Debug("Error reading GlobalPreferences: %v", err)
	}

	globalPrefs, err := utils.ParseBiPList(string(globalPrefsFile))
	if err != nil {
		params.Logger.Debug("Error reading GlobalPreferences: %v", err)
	}

	// Add global preferences data
	if globalPrefs != nil {
		processGlobalPrefs(globalPrefs, recordData)
	}

	systemConfigFile, err := os.ReadFile(systemConfigPath)
	if err != nil {
		params.Logger.Debug("Error reading system configuration: %v", err)
	}

	systemConfig, err := utils.ParseBiPList(string(systemConfigFile))
	if err != nil {
		params.Logger.Debug("Error reading system configuration: %v", err)
	}

	systemVersionFile, err := os.ReadFile(systemVersionPath)
	if err != nil {
		params.Logger.Debug("Error reading system version: %v", err)
	}

	systemVersion, err := utils.ParseBiPList(string(systemVersionFile))
	if err != nil {
		params.Logger.Debug("Error reading system version: %v", err)
	}

	// Collect system information
	recordData = make(map[string]interface{})

	// Get computer name and hostname
	if systemConfig != nil {
		if computerName, ok := utils.GetNestedValue(systemConfig, "System", "Network", "HostNames", "LocalHostName").(string); ok {
			recordData["local_hostname"] = computerName
		}
		if computerName, ok := utils.GetNestedValue(systemConfig, "System", "Network", "HostNames", "ComputerName").(string); ok {
			recordData["computer_name"] = computerName
		}
		if hostname, ok := utils.GetNestedValue(systemConfig, "System", "Network", "HostNames", "HostName").(string); ok {
			recordData["hostname"] = hostname
		}
	}

	// Get system version information
	if systemVersion != nil {
		recordData["product_version"] = systemVersion["ProductVersion"]
		recordData["product_build_version"] = systemVersion["ProductBuildVersion"]
	}

	// Get hardware information using system_profiler
	if hwInfo, err := exec.Command("system_profiler", "SPHardwareDataType", "-json").Output(); err == nil {
		hwData, err := utils.ParseBiPList(string(hwInfo))
		if err == nil {
			if items, ok := hwData["SPHardwareDataType"].([]interface{}); ok && len(items) > 0 {
				if item, ok := items[0].(map[string]interface{}); ok {
					recordData["model"] = item["machine_model"]
					recordData["serial_no"] = item["serial_number"]
				}
			}
		}
	}

	// Get timezone
	if tz, err := exec.Command("systemsetup", "-gettimezone").Output(); err == nil {
		recordData["system_tz"] = strings.TrimPrefix(strings.TrimSpace(string(tz)), "Time Zone: ")
	}

	// Get FileVault status
	if fdeStatus, err := exec.Command("fdesetup", "status").Output(); err == nil {
		if strings.Contains(string(fdeStatus), "On") {
			recordData["fvde_status"] = "On"
		} else {
			recordData["fvde_status"] = "Off"
		}
	}

	// Get Gatekeeper status
	if gkStatus, err := exec.Command("spctl", "--status").Output(); err == nil {
		recordData["gatekeeper_status"] = strings.TrimSpace(string(gkStatus))
	}

	// Get SIP status
	if sipStatus, err := exec.Command("csrutil", "status").Output(); err == nil {
		recordData["sip_status"] = strings.TrimSpace(string(sipStatus))
	}

	// Get IP address
	if ipAddr, err := exec.Command("ipconfig", "getifaddr", "en0").Output(); err == nil {
		recordData["ipaddress"] = strings.TrimSpace(string(ipAddr))
	}

	// Create and write the record
	record := utils.Record{
		CollectionTimestamp: params.CollectionTimestamp,
		EventTimestamp:      time.Now().UTC().Format(time.RFC3339),
		Data:                recordData,
		SourceFile:          "System Information",
	}

	return writer.WriteRecord(record)
}

func processGlobalPrefs(globalPrefs map[string]interface{}, recordData map[string]interface{}) {
	// Get system appearance (Dark/Light mode)
	if appearance, ok := globalPrefs["AppleInterfaceStyle"].(string); ok {
		recordData["system_appearance"] = appearance
	} else {
		recordData["system_appearance"] = "Light"
	}

	// Get system language
	if language, ok := globalPrefs["AppleLanguages"].([]interface{}); ok && len(language) > 0 {
		if primaryLang, ok := language[0].(string); ok {
			recordData["system_language"] = primaryLang
		}
	}

	// Get system locale
	if locale, ok := globalPrefs["AppleLocale"].(string); ok {
		recordData["system_locale"] = locale
	}

	// Get system keyboard layout
	if keyboard, ok := globalPrefs["AppleKeyboardLayout"].(string); ok {
		recordData["keyboard_layout"] = keyboard
	}
}
