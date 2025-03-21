// This module collects information about network configurations from plist files.
// It collects the following information:
// - Airport preferences
// - Network interfaces
package netcfgplists

import (
	"os"

	"github.com/gnzdotmx/ishinobu/pkg/mod"
	"github.com/gnzdotmx/ishinobu/pkg/utils"
	"howett.net/plist"
)

type NetworkConfigPlistsModule struct {
	Name        string
	Description string
}

func init() {
	module := &NetworkConfigPlistsModule{
		Name:        "netcfgplists",
		Description: "Collects information about network configurations from plist files"}
	mod.RegisterModule(module)
}

func (m *NetworkConfigPlistsModule) GetName() string {
	return m.Name
}

func (m *NetworkConfigPlistsModule) GetDescription() string {
	return m.Description
}

func (m *NetworkConfigPlistsModule) Run(params mod.ModuleParams) error {
	err := parseAirportPreferences(m.GetName(), params)
	if err != nil {
		params.Logger.Debug("Error parsing airport preferences: %v", err)
	}

	err = parseNetworkInterfaces(m.GetName(), params)
	if err != nil {
		params.Logger.Debug("Error parsing network interfaces: %v", err)
	}

	return nil
}

func parseAirportPreferences(moduleName string, params mod.ModuleParams) error {
	file := "/Library/Preferences/SystemConfiguration/com.apple.airport.preferences.plist"

	outputFileName := utils.GetOutputFileName(moduleName+"-airport", params.ExportFormat, params.OutputDir)
	writer, err := utils.NewDataWriter(params.LogsDir, outputFileName, params.ExportFormat)
	if err != nil {
		return err
	}

	data, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	var plistData map[string]interface{}
	_, err = plist.Unmarshal(data, &plistData)
	if err != nil {
		return err
	}

	if knownNetworks, ok := plistData["KnownNetworks"].(map[string]interface{}); ok {
		for _, network := range knownNetworks {
			if networkMap, ok := network.(map[string]interface{}); ok {
				recordData := make(map[string]interface{})
				recordData["src_name"] = "airport"
				recordData["type"] = "Airport"

				if ssid, ok := networkMap["SSIDString"].(string); ok {
					recordData["name"] = ssid
				}
				if lastConnected, ok := networkMap["LastConnected"].(string); ok {
					recordData["last_connected"] = lastConnected
				}
				if security, ok := networkMap["SecurityType"].(string); ok {
					recordData["security"] = security
				}
				if hotspot, ok := networkMap["PersonalHotspot"].(bool); ok {
					recordData["hotspot"] = hotspot
				}

				if addedAt, ok := networkMap["AddedAt"].(string); ok {
					recordData["added_at"] = addedAt
				}
				if roaming, ok := networkMap["RoamingProfileType"].(string); ok {
					recordData["roaming_profile"] = roaming
				}
				if autoLogin, ok := networkMap["AutoLogin"].(bool); ok {
					recordData["auto_login"] = autoLogin
				}
				if captive, ok := networkMap["CaptiveBypass"].(bool); ok {
					recordData["captive_bypass"] = captive
				}
				if disabled, ok := networkMap["Disabled"].(bool); ok {
					recordData["disabled"] = disabled
				}

				record := utils.Record{
					CollectionTimestamp: params.CollectionTimestamp,
					EventTimestamp:      params.CollectionTimestamp,
					Data:                recordData,
					SourceFile:          file,
				}

				err = writer.WriteRecord(record)
				if err != nil {
					params.Logger.Debug("Failed to write record: %v", err)
				}
			}
		}
	}

	return nil
}

func parseNetworkInterfaces(moduleName string, params mod.ModuleParams) error {
	file := "/Library/Preferences/SystemConfiguration/NetworkInterfaces.plist"

	outputFileName := utils.GetOutputFileName(moduleName+"-interfaces", params.ExportFormat, params.OutputDir)
	writer, err := utils.NewDataWriter(params.LogsDir, outputFileName, params.ExportFormat)
	if err != nil {
		return err
	}

	data, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	var plistData map[string]interface{}
	_, err = plist.Unmarshal(data, &plistData)
	if err != nil {
		return err
	}

	if interfaces, ok := plistData["Interfaces"].([]interface{}); ok {
		for _, iface := range interfaces {
			if ifaceMap, ok := iface.(map[string]interface{}); ok {
				recordData := make(map[string]interface{})
				recordData["src_name"] = "network_interfaces"

				if bsdName, ok := ifaceMap["BSD Name"].(string); ok {
					recordData["type"] = bsdName
				}

				if info, ok := ifaceMap["SCNetworkInterfaceInfo"].(map[string]interface{}); ok {
					if name, ok := info["UserDefinedName"].(string); ok {
						recordData["name"] = name
					}
				}

				if active, ok := ifaceMap["Active"].(bool); ok {
					recordData["active"] = active
				}
				if builtIn, ok := ifaceMap["Built-In"].(bool); ok {
					recordData["built_in"] = builtIn
				}
				if macAddress, ok := ifaceMap["MAC Address"].(string); ok {
					recordData["mac_address"] = macAddress
				}
				if product, ok := ifaceMap["Product"].(string); ok {
					recordData["product"] = product
				}
				if vendor, ok := ifaceMap["Vendor"].(string); ok {
					recordData["vendor"] = vendor
				}
				if model, ok := ifaceMap["Model"].(string); ok {
					recordData["model"] = model
				}
				if interfaceType, ok := ifaceMap["Type"].(string); ok {
					recordData["interface_type"] = interfaceType
				}

				record := utils.Record{
					CollectionTimestamp: params.CollectionTimestamp,
					EventTimestamp:      params.CollectionTimestamp,
					Data:                recordData,
					SourceFile:          file,
				}

				err = writer.WriteRecord(record)
				if err != nil {
					params.Logger.Debug("Failed to write record: %v", err)
				}
			}
		}
	}

	return nil
}
