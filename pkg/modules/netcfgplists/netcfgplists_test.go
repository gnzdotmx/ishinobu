package netcfgplists

import (
	"bufio"
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gnzdotmx/ishinobu/pkg/mod"
	"github.com/gnzdotmx/ishinobu/pkg/modules/testutils"
	"github.com/gnzdotmx/ishinobu/pkg/utils"
	"github.com/stretchr/testify/require"
	"howett.net/plist"
)

func TestParseAirportPreferences(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "netcfgplists_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create test plist data
	airportPlist := map[string]interface{}{
		"KnownNetworks": map[string]interface{}{
			"NetworkKey1": map[string]interface{}{
				"SSIDString":         "TestWiFi",
				"LastConnected":      "2023-01-01 12:00:00",
				"SecurityType":       "WPA2",
				"PersonalHotspot":    true,
				"AddedAt":            "2022-12-01 10:00:00",
				"RoamingProfileType": "None",
				"AutoLogin":          true,
				"CaptiveBypass":      false,
				"Disabled":           false,
			},
			"NetworkKey2": map[string]interface{}{
				"SSIDString":         "HomeWiFi",
				"LastConnected":      "2023-02-01 09:00:00",
				"SecurityType":       "WPA3",
				"PersonalHotspot":    false,
				"AddedAt":            "2022-11-15 15:30:00",
				"RoamingProfileType": "Automatic",
				"AutoLogin":          false,
				"CaptiveBypass":      true,
				"Disabled":           true,
			},
		},
	}

	// Create temporary plist file
	airportPath := filepath.Join(tmpDir, "airport.plist")
	plistData, err := plist.Marshal(airportPlist, plist.XMLFormat)
	require.NoError(t, err)
	err = os.WriteFile(airportPath, plistData, 0600)
	require.NoError(t, err)

	// Create output files for the records
	outputDir := filepath.Join(tmpDir, "output")
	err = os.MkdirAll(outputDir, 0755)
	require.NoError(t, err)

	// Get test records
	collectionTime := "2023-03-01T10:00:00Z"

	// Create the expected records
	record1 := utils.Record{
		CollectionTimestamp: collectionTime,
		EventTimestamp:      collectionTime,
		SourceFile:          airportPath,
		Data: map[string]interface{}{
			"src_name":        "airport",
			"type":            "Airport",
			"name":            "TestWiFi",
			"last_connected":  "2023-01-01 12:00:00",
			"security":        "WPA2",
			"hotspot":         true,
			"added_at":        "2022-12-01 10:00:00",
			"roaming_profile": "None",
			"auto_login":      true,
			"captive_bypass":  false,
			"disabled":        false,
		},
	}

	record2 := utils.Record{
		CollectionTimestamp: collectionTime,
		EventTimestamp:      collectionTime,
		SourceFile:          airportPath,
		Data: map[string]interface{}{
			"src_name":        "airport",
			"type":            "Airport",
			"name":            "HomeWiFi",
			"last_connected":  "2023-02-01 09:00:00",
			"security":        "WPA3",
			"hotspot":         false,
			"added_at":        "2022-11-15 15:30:00",
			"roaming_profile": "Automatic",
			"auto_login":      false,
			"captive_bypass":  true,
			"disabled":        true,
		},
	}

	// Create JSON files for these records
	record1File := filepath.Join(outputDir, "record1.json")
	record2File := filepath.Join(outputDir, "record2.json")

	testutils.WriteTestRecord(t, record1File, record1)
	testutils.WriteTestRecord(t, record2File, record2)

	// Create test module params and run the function directly
	logger := testutils.NewTestLogger()
	params := mod.ModuleParams{
		Logger:              *logger,
		OutputDir:           "./",
		LogsDir:             tmpDir,
		CollectionTimestamp: collectionTime,
		ExportFormat:        "json",
	}

	// This test isn't meant to fully execute the function since we're not mocking DataWriter
	// Just verify it doesn't panic
	err = parseAirportPreferences(airportPath, "netcfgplists", params)
	require.NoError(t, err)

	// Read the output file netcfgplists-airport.json
	outputFile := filepath.Join(tmpDir, "netcfgplists-airport.json")
	require.FileExists(t, outputFile)

	// Read line by line since the file contains separate JSON objects
	fileContent, err := os.ReadFile(outputFile)
	require.NoError(t, err)

	var records []map[string]interface{}
	scanner := bufio.NewScanner(bytes.NewReader(fileContent))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		var record map[string]interface{}
		err = json.Unmarshal([]byte(line), &record)
		require.NoError(t, err)
		records = append(records, record)
	}
	require.NoError(t, scanner.Err())
	require.Equal(t, 2, len(records))

	// Check each record
	for _, record := range records {
		// Check airport records properly
		switch record["name"] {
		case "TestWiFi":
			require.Equal(t, "Airport", record["type"])
			require.Equal(t, "WPA2", record["security"])
			require.Equal(t, true, record["hotspot"])
			require.Equal(t, "None", record["roaming_profile"])
		case "HomeWiFi":
			require.Equal(t, "Airport", record["type"])
			require.Equal(t, "WPA3", record["security"])
			require.Equal(t, false, record["hotspot"])
			require.Equal(t, "Automatic", record["roaming_profile"])
		default:
			t.Fatalf("Unexpected network name: %v", record["name"])
		}
	}
}

func TestParseNetworkInterfaces(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "netcfgplists_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create test plist data
	interfacesPlist := map[string]interface{}{
		"Interfaces": []interface{}{
			map[string]interface{}{
				"BSD Name": "en0",
				"SCNetworkInterfaceInfo": map[string]interface{}{
					"UserDefinedName": "Wi-Fi",
				},
				"Active":      true,
				"Built-In":    true,
				"MAC Address": "00:11:22:33:44:55",
				"Product":     "AirPort Extreme",
				"Vendor":      "Apple Inc.",
				"Model":       "AirPort",
				"Type":        "Ethernet",
			},
			map[string]interface{}{
				"BSD Name": "en1",
				"SCNetworkInterfaceInfo": map[string]interface{}{
					"UserDefinedName": "Ethernet Adapter",
				},
				"Active":      false,
				"Built-In":    false,
				"MAC Address": "aa:bb:cc:dd:ee:ff",
				"Product":     "USB Ethernet",
				"Vendor":      "Third Party",
				"Model":       "USB-C Ethernet",
				"Type":        "Ethernet",
			},
		},
	}

	// Create temporary plist file
	interfacesPath := filepath.Join(tmpDir, "interfaces.plist")
	plistData, err := plist.Marshal(interfacesPlist, plist.XMLFormat)
	require.NoError(t, err)
	err = os.WriteFile(interfacesPath, plistData, 0600)
	require.NoError(t, err)

	// Create output files for the records
	outputDir := filepath.Join(tmpDir, "output")
	err = os.MkdirAll(outputDir, 0755)
	require.NoError(t, err)

	// Get test records
	collectionTime := "2023-03-01T10:00:00Z"

	// Create the expected records
	record1 := utils.Record{
		CollectionTimestamp: collectionTime,
		EventTimestamp:      collectionTime,
		SourceFile:          interfacesPath,
		Data: map[string]interface{}{
			"src_name":       "network_interfaces",
			"type":           "en0",
			"name":           "Wi-Fi",
			"active":         true,
			"built_in":       true,
			"mac_address":    "00:11:22:33:44:55",
			"product":        "AirPort Extreme",
			"vendor":         "Apple Inc.",
			"model":          "AirPort",
			"interface_type": "Ethernet",
		},
	}

	record2 := utils.Record{
		CollectionTimestamp: collectionTime,
		EventTimestamp:      collectionTime,
		SourceFile:          interfacesPath,
		Data: map[string]interface{}{
			"src_name":       "network_interfaces",
			"type":           "en1",
			"name":           "Ethernet Adapter",
			"active":         false,
			"built_in":       false,
			"mac_address":    "aa:bb:cc:dd:ee:ff",
			"product":        "USB Ethernet",
			"vendor":         "Third Party",
			"model":          "USB-C Ethernet",
			"interface_type": "Ethernet",
		},
	}

	// Create JSON files for these records
	record1File := filepath.Join(outputDir, "record1.json")
	record2File := filepath.Join(outputDir, "record2.json")

	testutils.WriteTestRecord(t, record1File, record1)
	testutils.WriteTestRecord(t, record2File, record2)

	// Create test module params and run the function directly
	logger := testutils.NewTestLogger()
	params := mod.ModuleParams{
		Logger:              *logger,
		OutputDir:           "./",
		LogsDir:             tmpDir,
		CollectionTimestamp: collectionTime,
		ExportFormat:        "json",
	}

	// This test isn't meant to fully execute the function since we're not mocking DataWriter
	// Just verify it doesn't panic
	err = parseNetworkInterfaces(interfacesPath, "netcfgplists", params)
	require.NoError(t, err)

	// Read the output file netcfgplists-interfaces.json
	outputFile := filepath.Join(tmpDir, "netcfgplists-interfaces.json")
	require.FileExists(t, outputFile)

	// Read line by line since the file contains separate JSON objects
	fileContent, err := os.ReadFile(outputFile)
	require.NoError(t, err)

	var records []map[string]interface{}
	scanner := bufio.NewScanner(bytes.NewReader(fileContent))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		var record map[string]interface{}
		err = json.Unmarshal([]byte(line), &record)
		require.NoError(t, err)
		records = append(records, record)
	}
	require.NoError(t, scanner.Err())
	require.Equal(t, 2, len(records))

	// Check each record
	for _, record := range records {
		// Check if this is the first or second interface based on name
		if record["name"] == "Wi-Fi" {
			require.Equal(t, "en0", record["type"])
			require.Equal(t, true, record["active"])
			require.Equal(t, "AirPort Extreme", record["product"])
		} else {
			require.Equal(t, "en1", record["type"])
			require.Equal(t, "Ethernet Adapter", record["name"])
			require.Equal(t, false, record["active"])
			require.Equal(t, "USB Ethernet", record["product"])
		}
	}
}
