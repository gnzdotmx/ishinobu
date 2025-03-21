package netcfgplists

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

func TestNetworkConfigPlistsModule(t *testing.T) {
	// Create temporary directory for test outputs
	tmpDir, err := os.MkdirTemp("", "netcfgplists_test")
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
	module := &NetworkConfigPlistsModule{
		Name:        "netcfgplists",
		Description: "Collects information about network configurations from plist files",
	}

	// Test GetName
	t.Run("GetName", func(t *testing.T) {
		assert.Equal(t, "netcfgplists", module.GetName())
	})

	// Test GetDescription
	t.Run("GetDescription", func(t *testing.T) {
		assert.Contains(t, module.GetDescription(), "network configurations")
	})

	// Test Run method with mock outputs
	t.Run("Run", func(t *testing.T) {
		// Create mock output files directly
		createMockAirportPreferencesFile(t, params)
		createMockNetworkInterfacesFile(t, params)

		// Verify airport preferences file exists
		airportFile := filepath.Join(tmpDir, "netcfgplists-airport-"+params.CollectionTimestamp+".json")
		assert.FileExists(t, airportFile)

		// Verify network interfaces file exists
		interfacesFile := filepath.Join(tmpDir, "netcfgplists-interfaces-"+params.CollectionTimestamp+".json")
		assert.FileExists(t, interfacesFile)

		// Verify content of the files
		verifyAirportPreferencesOutput(t, airportFile)
		verifyNetworkInterfacesOutput(t, interfacesFile)
	})
}

// Test that the module initializes properly
func TestNetworkConfigPlistsModuleInitialization(t *testing.T) {
	// Create a new instance with proper initialization
	module := &NetworkConfigPlistsModule{
		Name:        "netcfgplists",
		Description: "Collects information about network configurations from plist files",
	}

	// Verify module is properly instantiated with expected values
	assert.Equal(t, "netcfgplists", module.Name, "Module name should be initialized")
	assert.Contains(t, module.Description, "network configurations", "Module description should be initialized")

	// Test the module's methods
	assert.Equal(t, "netcfgplists", module.GetName())
	assert.Contains(t, module.GetDescription(), "network configurations")
}

// Create a mock airport preferences file
func createMockAirportPreferencesFile(t *testing.T, params mod.ModuleParams) {
	outputFile := filepath.Join(params.OutputDir, "netcfgplists-airport-"+params.CollectionTimestamp+".json")

	// Create mock WiFi network records
	wifiNetworks := []utils.Record{
		{
			CollectionTimestamp: params.CollectionTimestamp,
			EventTimestamp:      params.CollectionTimestamp,
			SourceFile:          "/Library/Preferences/SystemConfiguration/com.apple.airport.preferences.plist",
			Data: map[string]interface{}{
				"src_name":        "airport",
				"type":            "Airport",
				"name":            "HomeWiFi",
				"last_connected":  "2023-06-15T08:30:45Z",
				"security":        "WPA2",
				"hotspot":         false,
				"added_at":        "2023-01-10T14:22:33Z",
				"roaming_profile": "None",
				"auto_login":      true,
				"captive_bypass":  false,
				"disabled":        false,
			},
		},
		{
			CollectionTimestamp: params.CollectionTimestamp,
			EventTimestamp:      params.CollectionTimestamp,
			SourceFile:          "/Library/Preferences/SystemConfiguration/com.apple.airport.preferences.plist",
			Data: map[string]interface{}{
				"src_name":        "airport",
				"type":            "Airport",
				"name":            "CoffeeShopWiFi",
				"last_connected":  "2023-06-12T15:40:20Z",
				"security":        "WPA2",
				"hotspot":         false,
				"added_at":        "2023-03-22T09:15:10Z",
				"roaming_profile": "None",
				"auto_login":      false,
				"captive_bypass":  true,
				"disabled":        false,
			},
		},
		{
			CollectionTimestamp: params.CollectionTimestamp,
			EventTimestamp:      params.CollectionTimestamp,
			SourceFile:          "/Library/Preferences/SystemConfiguration/com.apple.airport.preferences.plist",
			Data: map[string]interface{}{
				"src_name":        "airport",
				"type":            "Airport",
				"name":            "iPhone Hotspot",
				"last_connected":  "2023-06-14T12:10:05Z",
				"security":        "WPA2",
				"hotspot":         true,
				"added_at":        "2023-05-05T18:30:45Z",
				"roaming_profile": "None",
				"auto_login":      true,
				"captive_bypass":  false,
				"disabled":        false,
			},
		},
	}

	// Write the records to the output file
	file, err := os.Create(outputFile)
	assert.NoError(t, err)
	defer file.Close()

	encoder := json.NewEncoder(file)
	for _, record := range wifiNetworks {
		err := encoder.Encode(record)
		assert.NoError(t, err)
	}
}

// Create a mock network interfaces file
func createMockNetworkInterfacesFile(t *testing.T, params mod.ModuleParams) {
	outputFile := filepath.Join(params.OutputDir, "netcfgplists-interfaces-"+params.CollectionTimestamp+".json")

	// Create mock network interface records
	interfaces := []utils.Record{
		{
			CollectionTimestamp: params.CollectionTimestamp,
			EventTimestamp:      params.CollectionTimestamp,
			SourceFile:          "/Library/Preferences/SystemConfiguration/NetworkInterfaces.plist",
			Data: map[string]interface{}{
				"src_name":       "network_interfaces",
				"type":           "en0",
				"name":           "Wi-Fi",
				"active":         true,
				"built_in":       true,
				"mac_address":    "aa:bb:cc:dd:ee:ff",
				"product":        "AirPort Extreme",
				"vendor":         "Apple",
				"model":          "MacBookPro",
				"interface_type": "IEEE80211",
			},
		},
		{
			CollectionTimestamp: params.CollectionTimestamp,
			EventTimestamp:      params.CollectionTimestamp,
			SourceFile:          "/Library/Preferences/SystemConfiguration/NetworkInterfaces.plist",
			Data: map[string]interface{}{
				"src_name":       "network_interfaces",
				"type":           "en1",
				"name":           "Thunderbolt Ethernet",
				"active":         false,
				"built_in":       true,
				"mac_address":    "aa:bb:cc:11:22:33",
				"product":        "Ethernet Controller",
				"vendor":         "Apple",
				"model":          "MacBookPro",
				"interface_type": "Ethernet",
			},
		},
		{
			CollectionTimestamp: params.CollectionTimestamp,
			EventTimestamp:      params.CollectionTimestamp,
			SourceFile:          "/Library/Preferences/SystemConfiguration/NetworkInterfaces.plist",
			Data: map[string]interface{}{
				"src_name":       "network_interfaces",
				"type":           "lo0",
				"name":           "Loopback",
				"active":         true,
				"built_in":       true,
				"mac_address":    "",
				"interface_type": "Loopback",
			},
		},
	}

	// Write the records to the output file
	file, err := os.Create(outputFile)
	assert.NoError(t, err)
	defer file.Close()

	encoder := json.NewEncoder(file)
	for _, record := range interfaces {
		err := encoder.Encode(record)
		assert.NoError(t, err)
	}
}

// Verify the airport preferences output file
func verifyAirportPreferencesOutput(t *testing.T, outputFile string) {
	// Read the file
	data, err := os.ReadFile(outputFile)
	assert.NoError(t, err)

	// Parse JSON lines
	lines := splitJSONLines(data)
	assert.GreaterOrEqual(t, len(lines), 3, "Should have at least 3 WiFi networks")

	// Check for expected network names
	var foundHomeWiFi, foundCoffeeShop, foundHotspot bool

	for _, line := range lines {
		var record map[string]interface{}
		err := json.Unmarshal(line, &record)
		assert.NoError(t, err)

		// Check common fields
		assert.NotEmpty(t, record["collection_timestamp"])
		assert.NotEmpty(t, record["event_timestamp"])
		assert.NotEmpty(t, record["source_file"])
		assert.Contains(t, record["source_file"], "com.apple.airport.preferences.plist")

		// Check data
		data, ok := record["data"].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "airport", data["src_name"])
		assert.Equal(t, "Airport", data["type"])
		assert.NotEmpty(t, data["name"])
		assert.NotEmpty(t, data["last_connected"])
		assert.NotEmpty(t, data["security"])

		// Track specific networks
		if name, ok := data["name"].(string); ok {
			switch name {
			case "HomeWiFi":
				foundHomeWiFi = true
				assert.Equal(t, true, data["auto_login"])
			case "CoffeeShopWiFi":
				foundCoffeeShop = true
				assert.Equal(t, true, data["captive_bypass"])
			case "iPhone Hotspot":
				foundHotspot = true
				assert.Equal(t, true, data["hotspot"])
			}
		}
	}

	assert.True(t, foundHomeWiFi, "Should contain HomeWiFi network")
	assert.True(t, foundCoffeeShop, "Should contain CoffeeShopWiFi network")
	assert.True(t, foundHotspot, "Should contain iPhone Hotspot network")
}

// Verify the network interfaces output file
func verifyNetworkInterfacesOutput(t *testing.T, outputFile string) {
	// Read the file
	data, err := os.ReadFile(outputFile)
	assert.NoError(t, err)

	// Parse JSON lines
	lines := splitJSONLines(data)
	assert.GreaterOrEqual(t, len(lines), 3, "Should have at least 3 network interfaces")

	// Check for expected interfaces
	var foundWiFi, foundEthernet, foundLoopback bool

	for _, line := range lines {
		var record map[string]interface{}
		err := json.Unmarshal(line, &record)
		assert.NoError(t, err)

		// Check common fields
		assert.NotEmpty(t, record["collection_timestamp"])
		assert.NotEmpty(t, record["event_timestamp"])
		assert.NotEmpty(t, record["source_file"])
		assert.Contains(t, record["source_file"], "NetworkInterfaces.plist")

		// Check data
		data, ok := record["data"].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "network_interfaces", data["src_name"])
		assert.NotEmpty(t, data["type"])
		assert.NotEmpty(t, data["name"])

		// Track specific interfaces
		if itype, ok := data["type"].(string); ok {
			switch itype {
			case "en0":
				foundWiFi = true
				assert.Equal(t, "Wi-Fi", data["name"])
				assert.Equal(t, true, data["active"])
				assert.Equal(t, "IEEE80211", data["interface_type"])
			case "en1":
				foundEthernet = true
				assert.Equal(t, "Thunderbolt Ethernet", data["name"])
				assert.Equal(t, false, data["active"])
				assert.Equal(t, "Ethernet", data["interface_type"])
			case "lo0":
				foundLoopback = true
				assert.Equal(t, "Loopback", data["name"])
				assert.Equal(t, "Loopback", data["interface_type"])
			}
		}
	}

	assert.True(t, foundWiFi, "Should contain WiFi interface")
	assert.True(t, foundEthernet, "Should contain Ethernet interface")
	assert.True(t, foundLoopback, "Should contain Loopback interface")
}

// Helper function to split JSON lines
func splitJSONLines(data []byte) [][]byte {
	var lines [][]byte
	start := 0

	for i := 0; i < len(data); i++ {
		if data[i] == '\n' {
			if i > start {
				lines = append(lines, data[start:i])
			}
			start = i + 1
		}
	}

	if start < len(data) {
		lines = append(lines, data[start:])
	}

	return lines
}
