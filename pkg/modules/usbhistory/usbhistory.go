// This module collects USB device history from various macOS sources.
// IOREG collects:
// - Current USB devices
// - USB device metadata (serial numbers, vendor IDs, etc.)
// - Connection/disconnection timestamps
// - Device mount points
// - Device power states
// log show collects:
// - USB device connection history
// - USB device disconnection history
// - USB device metadata (serial numbers, vendor IDs, etc.)
// - Connection/disconnection timestamps
// - Device mount points
// - Device power states
// SYSTEM_PROFILER collects:
// - USB device metadata (serial numbers, vendor IDs, etc.)
// - Device hierarchy: parent, child, sibling, etc.
package usbhistory

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gnzdotmx/ishinobu/pkg/mod"
	"github.com/gnzdotmx/ishinobu/pkg/utils"
)

const (
	// LogCollectionHours defines how many hours of logs to collect
	LogCollectionHours = 24
)

var (
	// Static errors
	errNoUSBDataType  = errors.New("SPUSBDataType key not found in system_profiler output")
	errInvalidUSBData = errors.New("SPUSBDataType is not an array as expected")
	errNoUSBDevices   = errors.New("no USB devices found in ioreg output")
)

type USBHistoryModule struct {
	Name        string
	Description string
}

func init() {
	module := &USBHistoryModule{
		Name:        "usbhistory",
		Description: "Collects USB device connection history and metadata",
	}
	mod.RegisterModule(module)
}

func (m *USBHistoryModule) GetName() string {
	return m.Name
}

func (m *USBHistoryModule) GetDescription() string {
	return m.Description
}

func (m *USBHistoryModule) Run(params mod.ModuleParams) error {
	params.Logger.Debug("Starting USB history collection")

	// Collect all data first
	params.Logger.Debug("Collecting all USB data")
	systemProfilerData, err := collectSystemProfilerData(params)
	if err != nil {
		params.Logger.Error("Error collecting system profiler data: %v", err)
	}

	logData, err := collectLogData(params)
	if err != nil {
		params.Logger.Error("Error collecting log data: %v", err)
	}

	ioregData, err := collectIORegData(params)
	if err != nil {
		params.Logger.Error("Error collecting ioreg data: %v", err)
	}

	// Write individual reports
	if err := writeUSBList(params, systemProfilerData); err != nil {
		params.Logger.Error("Error writing USB list: %v", err)
	}

	if err := writeUSBHistory(params, logData); err != nil {
		params.Logger.Error("Error writing USB history: %v", err)
	}

	if err := writeUSBRegistry(params, ioregData); err != nil {
		params.Logger.Error("Error writing USB registry: %v", err)
	}

	// Generate and write summary using the collected data
	if err := writeUSBSummary(params, systemProfilerData, logData, ioregData); err != nil {
		params.Logger.Error("Error writing USB summary: %v", err)
	}

	params.Logger.Debug("USB history collection completed")
	return nil
}

func collectSystemProfilerData(params mod.ModuleParams) ([]interface{}, error) {
	params.Logger.Debug("Running system_profiler SPUSBDataType")
	output, err := utils.ExecCommand("system_profiler", "SPUSBDataType", "-json")
	if err != nil {
		return nil, fmt.Errorf("error running system_profiler: %w", err)
	}

	params.Logger.Debug("Parsing system_profiler output")
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(output), &data); err != nil {
		return nil, fmt.Errorf("error parsing system_profiler output: %w", err)
	}

	usbDataRaw, ok := data["SPUSBDataType"]
	if !ok {
		return nil, errNoUSBDataType
	}

	usbList, ok := usbDataRaw.([]interface{})
	if !ok {
		return nil, errInvalidUSBData
	}

	params.Logger.Debug("Found %d USB hubs/controllers", len(usbList))
	return usbList, nil
}

func collectLogData(params mod.ModuleParams) ([]map[string]interface{}, error) {
	params.Logger.Debug("Running log show command for USB events")
	output, err := utils.ExecCommand("log", "show",
		"--predicate", `(process == "usbd" OR process == "USBAgent" OR subsystem == "com.apple.iokit.IOUSBHostFamily" OR subsystem == "com.apple.usb") AND (eventMessage CONTAINS[c] "attach" OR eventMessage CONTAINS[c] "detach" OR eventMessage CONTAINS[c] "connect" OR eventMessage CONTAINS[c] "disconnect" OR eventMessage CONTAINS[c] "enumerate" OR eventMessage CONTAINS[c] "vendor" OR eventMessage CONTAINS[c] "product")`,
		"--style", "json",
		"--info",
		"--debug",
		"--last", fmt.Sprintf("%dh", LogCollectionHours))
	if err != nil {
		return nil, fmt.Errorf("error running log show: %w", err)
	}

	params.Logger.Debug("Parsing log show output")
	var logs []map[string]interface{}
	if err := json.Unmarshal([]byte(output), &logs); err != nil {
		params.Logger.Debug("Error parsing log output: %v", err)
		return nil, fmt.Errorf("error parsing log output: %w", err)
	}

	params.Logger.Debug("Found %d USB event logs", len(logs))
	return logs, nil
}

func collectIORegData(params mod.ModuleParams) ([]map[string]interface{}, error) {
	params.Logger.Debug("Running ioreg command")
	output, err := utils.ExecCommand("ioreg", "-p", "IOUSB", "-l", "-w", "0")
	if err != nil {
		return nil, fmt.Errorf("error running ioreg: %w", err)
	}

	params.Logger.Debug("Parsing ioreg output")
	devices, err := parseIORegOutput(string(output))
	if err != nil {
		params.Logger.Debug("Error parsing ioreg output: %v", err)
		return nil, fmt.Errorf("error parsing ioreg output: %w", err)
	}

	params.Logger.Debug("Found %d USB devices in registry", len(devices))
	return devices, nil
}

func writeUSBList(params mod.ModuleParams, systemProfilerData []interface{}) error {
	outputFileNameUsbList := utils.GetOutputFileName("usbhistoryUsbList", params.ExportFormat, params.OutputDir)
	params.Logger.Debug("Creating USB list writer: %s", outputFileNameUsbList)
	writerUsbList, err := utils.NewDataWriter(params.LogsDir, outputFileNameUsbList, params.ExportFormat)
	if err != nil {
		return fmt.Errorf("failed to create data writer: %w", err)
	}
	defer writerUsbList.Close()

	for _, usb := range systemProfilerData {
		usbMap, ok := usb.(map[string]interface{})
		if !ok {
			continue
		}

		record := utils.Record{
			CollectionTimestamp: params.CollectionTimestamp,
			EventTimestamp:      params.CollectionTimestamp,
			SourceFile:          "system_profiler",
			Data:                usbMap,
		}

		if err := writerUsbList.WriteRecord(record); err != nil {
			params.Logger.Debug("Error writing USB device record: %v", err)
		}

		if items, ok := usbMap["_items"].([]interface{}); ok {
			params.Logger.Debug("Found %d USB devices in hub", len(items))
			for _, item := range items {
				itemMap, ok := item.(map[string]interface{})
				if !ok {
					continue
				}

				itemRecord := utils.Record{
					CollectionTimestamp: params.CollectionTimestamp,
					EventTimestamp:      params.CollectionTimestamp,
					SourceFile:          "system_profiler",
					Data:                itemMap,
				}

				if err := writerUsbList.WriteRecord(itemRecord); err != nil {
					params.Logger.Debug("Error writing USB device item record: %v", err)
				}
			}
		}
	}

	return nil
}

func writeUSBHistory(params mod.ModuleParams, logData []map[string]interface{}) error {
	outputFileNameUsbHistory := utils.GetOutputFileName("usbhistoryUsbHistory", params.ExportFormat, params.OutputDir)
	params.Logger.Debug("Creating USB history writer: %s", outputFileNameUsbHistory)
	writerUsbHistory, err := utils.NewDataWriter(params.LogsDir, outputFileNameUsbHistory, params.ExportFormat)
	if err != nil {
		return fmt.Errorf("failed to create data writer: %w", err)
	}
	defer writerUsbHistory.Close()

	for _, log := range logData {
		msg, ok := log["eventMessage"].(string)
		if !ok {
			continue
		}

		// Only process messages containing relevant USB events
		if !containsAnyI(msg, []string{
			"attach", "detach", "connect", "disconnect", "enumerate",
			"vendor", "product", "serial",
		}) {
			continue
		}

		timestamp, ok := log["timestamp"].(string)
		if !ok {
			continue
		}

		eventTimeStr, err := utils.ParseTimestampWithFormats(timestamp)
		if err != nil {
			continue
		}

		eventTime, err := time.Parse(utils.TimeFormat, eventTimeStr)
		if err != nil {
			continue
		}

		params.Logger.Debug("Processing USB event: %s at %s", msg, eventTime)
		record := utils.Record{
			CollectionTimestamp: params.CollectionTimestamp,
			EventTimestamp:      eventTime.Format(time.RFC3339),
			SourceFile:          "system_log",
			Data:                log,
		}

		if err := writerUsbHistory.WriteRecord(record); err != nil {
			params.Logger.Debug("Error writing USB history record: %v", err)
		}
	}

	return nil
}

func writeUSBRegistry(params mod.ModuleParams, ioregData []map[string]interface{}) error {
	outputFileNameUsbRegistry := utils.GetOutputFileName("usbhistoryUsbRegistry", params.ExportFormat, params.OutputDir)
	params.Logger.Debug("Creating USB registry writer: %s", outputFileNameUsbRegistry)
	writerUsbRegistry, err := utils.NewDataWriter(params.LogsDir, outputFileNameUsbRegistry, params.ExportFormat)
	if err != nil {
		return fmt.Errorf("failed to create data writer: %w", err)
	}
	defer writerUsbRegistry.Close()

	for _, device := range ioregData {
		record := utils.Record{
			CollectionTimestamp: params.CollectionTimestamp,
			EventTimestamp:      params.CollectionTimestamp,
			SourceFile:          "ioreg",
			Data:                device,
		}

		if err := writerUsbRegistry.WriteRecord(record); err != nil {
			params.Logger.Debug("Error writing USB registry record: %v", err)
		}
	}

	return nil
}

func writeUSBSummary(params mod.ModuleParams, systemProfilerData []interface{}, logData []map[string]interface{}, ioregData []map[string]interface{}) error {
	outputFileNameUsbSummary := utils.GetOutputFileName("usbhistoryUsbSummary", params.ExportFormat, params.OutputDir)
	params.Logger.Debug("Creating USB summary writer: %s", outputFileNameUsbSummary)
	writerUsbSummary, err := utils.NewDataWriter(params.LogsDir, outputFileNameUsbSummary, params.ExportFormat)
	if err != nil {
		return fmt.Errorf("failed to create data writer for summary: %w", err)
	}
	defer writerUsbSummary.Close()

	deviceMap := make(map[string]*USBDeviceSummary)

	// Process system profiler data
	processSystemProfilerData(systemProfilerData, deviceMap, params)

	// Process log data
	processLogData(logData, deviceMap)

	// Process ioreg data
	processIORegData(ioregData, deviceMap)

	// Write summaries
	for key, summary := range deviceMap {
		params.Logger.Debug("Adding device to summary: %s", key)

		// Format timestamps for events
		formattedEvents := make([]map[string]interface{}, len(summary.Events))
		for i, event := range summary.Events {
			formattedEvents[i] = map[string]interface{}{
				"timestamp":  event.Timestamp.Format(time.RFC3339),
				"event_type": event.EventType,
				"speed":      event.Speed,
			}
		}

		// Convert the summary to a map for writing
		summaryMap := map[string]interface{}{
			"name":                   summary.Name,
			"vendor_id":              summary.VendorID,
			"product_id":             summary.ProductID,
			"serial_number":          summary.SerialNumber,
			"last_connected":         summary.LastConnected.Format(time.RFC3339),
			"last_disconnected":      summary.LastDisconnected.Format(time.RFC3339),
			"is_currently_connected": summary.IsCurrentlyConnected,
			"connection_speed":       summary.ConnectionSpeed,
			"mount_point":            summary.MountPoint,
			"manufacturer":           summary.Manufacturer,
			"product":                summary.Product,
			"events":                 formattedEvents,
			"total_bytes_in":         summary.TotalBytesIn,
			"total_bytes_out":        summary.TotalBytesOut,
			"last_interface":         summary.LastInterface,
			"speed":                  summary.Speed,
		}

		record := utils.Record{
			CollectionTimestamp: params.CollectionTimestamp,
			EventTimestamp:      params.CollectionTimestamp,
			SourceFile:          "summary",
			Data:                summaryMap,
		}

		if err := writerUsbSummary.WriteRecord(record); err != nil {
			params.Logger.Debug("Error writing USB summary record: %v", err)
		}
	}

	return nil
}

// containsAnyI checks if str contains any of the substrings case-insensitively
func containsAnyI(str string, substrings []string) bool {
	strLower := strings.ToLower(str)
	for _, sub := range substrings {
		if strings.Contains(strLower, strings.ToLower(sub)) {
			return true
		}
	}
	return false
}

func processSystemProfilerData(data []interface{}, deviceMap map[string]*USBDeviceSummary, params mod.ModuleParams) {
	for _, hub := range data {
		if hubMap, ok := hub.(map[string]interface{}); ok {
			// Process the hub itself if it's a device
			if name, ok := hubMap["_name"].(string); ok {
				processUSBDevice(hubMap, deviceMap)
				params.Logger.Debug("Processed USB hub: %s", name)
			}

			// Process items in the hub
			if items, ok := hubMap["_items"].([]interface{}); ok {
				for _, item := range items {
					if device, ok := item.(map[string]interface{}); ok {
						if name, ok := device["_name"].(string); ok {
							params.Logger.Debug("Processing USB device: %s", name)
						}
						processUSBDevice(device, deviceMap)
					}
				}
			}
		}
	}
}

func processUSBDevice(device map[string]interface{}, deviceMap map[string]*USBDeviceSummary) {
	name, _ := device["_name"].(string)
	vendorID, _ := device["vendor_id"].(string)
	productID, _ := device["product_id"].(string)
	serialNum, _ := device["serial_num"].(string)
	manufacturer, _ := device["manufacturer"].(string)

	// Generate a unique key - if no serial number, use vendor:product:name
	key := serialNum
	if key == "" {
		key = fmt.Sprintf("%s:%s:%s", vendorID, productID, name)
	}
	if key == "::" {
		if name != "" {
			key = name // Fallback to just the name if no identifiers available
		} else {
			return // Skip if we can't identify the device at all
		}
	}

	if _, exists := deviceMap[key]; !exists {
		deviceMap[key] = &USBDeviceSummary{
			Name:                 name,
			VendorID:             vendorID,
			ProductID:            productID,
			SerialNumber:         serialNum,
			Manufacturer:         manufacturer,
			IsCurrentlyConnected: true,
			Events:               make([]USBDeviceEvent, 0),
		}
	}

	// Update additional fields if available
	if speed, ok := device["speed"].(string); ok {
		deviceMap[key].ConnectionSpeed = speed
	}
	if mount, ok := device["mount_point"].(string); ok {
		deviceMap[key].MountPoint = mount
	}
	if product, ok := device["product"].(string); ok {
		deviceMap[key].Product = product
	}
}

// processLogMessage processes a single log message and updates the device map
func processLogMessage(msg string, eventTime time.Time, deviceMap map[string]*USBDeviceSummary) {
	switch {
	case strings.Contains(msg, "enumerated"):
		processEnumerationMessage(msg, eventTime, deviceMap)
	case strings.Contains(msg, "disconnected") || strings.Contains(msg, "destroying") || strings.Contains(msg, "cable disconnected"):
		processDisconnectionMessage(msg, eventTime, deviceMap)
	case strings.Contains(msg, "cable connected"):
		processConnectionMessage(msg, eventTime, deviceMap)
	}
}

// processEnumerationMessage handles USB device enumeration messages
func processEnumerationMessage(msg string, eventTime time.Time, deviceMap map[string]*USBDeviceSummary) {
	// Example: "enumerated 0x05ac/12a8/1601 (iPhone / 1) at 480 Mbps"
	parts := strings.Split(msg, " ")
	for i, part := range parts {
		if !strings.Contains(part, "/") {
			continue
		}

		ids := strings.Split(strings.Trim(part, "0x"), "/")
		if len(ids) < 2 {
			continue
		}

		vendorID := ids[0]
		productID := ids[1]
		serialNum := ""
		if len(ids) > 2 {
			serialNum = ids[2]
		}

		// Try to extract device name
		name := ""
		if i+2 < len(parts) && strings.HasPrefix(parts[i+1], "(") {
			name = strings.Trim(parts[i+1], "()")
		}

		key := serialNum
		if key == "" {
			key = fmt.Sprintf("%s:%s:%s", vendorID, productID, name)
		}

		device, exists := deviceMap[key]
		if !exists {
			device = &USBDeviceSummary{
				Name:         name,
				VendorID:     vendorID,
				ProductID:    productID,
				SerialNumber: serialNum,
				Events:       make([]USBDeviceEvent, 0),
			}
			deviceMap[key] = device
		}

		// Extract speed if available
		for j := i; j < len(parts); j++ {
			if strings.Contains(parts[j], "Mbps") {
				device.Speed = parts[j-1] + " " + parts[j]
				break
			}
		}

		event := USBDeviceEvent{
			Timestamp: eventTime,
			EventType: "connect",
			Speed:     device.Speed,
		}
		device.Events = append(device.Events, event)
		if device.LastConnected.IsZero() || eventTime.After(device.LastConnected) {
			device.LastConnected = eventTime
		}
		device.IsCurrentlyConnected = true
		break
	}
}

// processDisconnectionMessage handles USB device disconnection messages
func processDisconnectionMessage(msg string, eventTime time.Time, deviceMap map[string]*USBDeviceSummary) {
	for _, device := range deviceMap {
		if (device.VendorID != "" && strings.Contains(msg, device.VendorID)) ||
			(device.ProductID != "" && strings.Contains(msg, device.ProductID)) ||
			(device.SerialNumber != "" && strings.Contains(msg, device.SerialNumber)) {
			event := USBDeviceEvent{
				Timestamp: eventTime,
				EventType: "disconnect",
			}
			device.Events = append(device.Events, event)
			if device.LastDisconnected.IsZero() || eventTime.After(device.LastDisconnected) {
				device.LastDisconnected = eventTime
			}
			device.IsCurrentlyConnected = false
		}
	}
}

// processConnectionMessage handles USB device connection messages
func processConnectionMessage(msg string, eventTime time.Time, deviceMap map[string]*USBDeviceSummary) {
	for _, device := range deviceMap {
		if strings.Contains(msg, device.Name) {
			event := USBDeviceEvent{
				Timestamp: eventTime,
				EventType: "connect",
			}
			device.Events = append(device.Events, event)
			if device.LastConnected.IsZero() || eventTime.After(device.LastConnected) {
				device.LastConnected = eventTime
			}
			device.IsCurrentlyConnected = true
		}
	}
}

func processLogData(logs []map[string]interface{}, deviceMap map[string]*USBDeviceSummary) {
	for _, log := range logs {
		msg, ok := log["eventMessage"].(string)
		if !ok {
			continue
		}

		// Parse timestamp first
		timestamp, ok := log["timestamp"].(string)
		if !ok {
			continue
		}

		eventTimeStr, err := utils.ParseTimestampWithFormats(timestamp)
		if err != nil {
			continue
		}

		eventTime, err := time.Parse(utils.TimeFormat, eventTimeStr)
		if err != nil {
			continue
		}

		processLogMessage(msg, eventTime, deviceMap)
	}

	// Update final connection state based on event order
	for _, device := range deviceMap {
		if len(device.Events) > 0 {
			lastEvent := device.Events[len(device.Events)-1]
			device.IsCurrentlyConnected = lastEvent.EventType == "connect"
		}
	}
}

func processIORegData(devices []map[string]interface{}, deviceMap map[string]*USBDeviceSummary) {
	for _, device := range devices {
		if props, ok := device["properties"].(map[string]interface{}); ok {
			vendorID, _ := props["idVendor"].(string)
			productID, _ := props["idProduct"].(string)
			serialNum, _ := props["USB Serial Number"].(string)
			name := device["clean_name"].(string)

			// Generate a unique key - if no serial number, use vendor:product:name
			key := serialNum
			if key == "" {
				key = fmt.Sprintf("%s:%s:%s", vendorID, productID, name)
			}
			if key == "::" {
				key = name // Fallback to just the name if no identifiers available
			}

			if key == "" {
				continue // Skip if we can't identify the device at all
			}

			if _, exists := deviceMap[key]; !exists {
				deviceMap[key] = &USBDeviceSummary{
					Name:                 name,
					VendorID:             vendorID,
					ProductID:            productID,
					SerialNumber:         serialNum,
					Events:               make([]USBDeviceEvent, 0),
					IsCurrentlyConnected: true,
				}
			}

			// Update additional information
			if deviceInfo, ok := device["device_info"].(string); ok {
				deviceMap[key].Product = deviceInfo
			}

			// Add any additional properties that might be useful
			if speed, ok := props["Speed"].(string); ok {
				deviceMap[key].Speed = speed
			}
			if manufacturer, ok := props["USB Vendor Name"].(string); ok {
				deviceMap[key].Manufacturer = manufacturer
			}
		}
	}
}

// USBDeviceEvent represents a single USB device event
type USBDeviceEvent struct {
	Timestamp    time.Time `json:"timestamp"`
	EventType    string    `json:"event_type"` // connect, disconnect, data_transfer
	ProcessName  string    `json:"process_name,omitempty"`
	ProcessID    int       `json:"process_id,omitempty"`
	BytesIn      int64     `json:"bytes_in,omitempty"`
	BytesOut     int64     `json:"bytes_out,omitempty"`
	Interface    string    `json:"interface,omitempty"`
	Duration     float64   `json:"duration,omitempty"`
	ConnectTime  float64   `json:"connect_time,omitempty"`
	Name         string    `json:"name,omitempty"`
	SerialNumber string    `json:"serial_number,omitempty"`
	Speed        string    `json:"speed,omitempty"`
}

// USBDeviceSummary represents a summarized view of a USB device's history
type USBDeviceSummary struct {
	Name                 string           `json:"name"`
	VendorID             string           `json:"vendor_id,omitempty"`
	ProductID            string           `json:"product_id,omitempty"`
	SerialNumber         string           `json:"serial_number,omitempty"`
	LastConnected        time.Time        `json:"last_connected,omitempty"`
	LastDisconnected     time.Time        `json:"last_disconnected,omitempty"`
	IsCurrentlyConnected bool             `json:"is_currently_connected"`
	ConnectionSpeed      string           `json:"connection_speed,omitempty"`
	MountPoint           string           `json:"mount_point,omitempty"`
	Manufacturer         string           `json:"manufacturer,omitempty"`
	Product              string           `json:"product,omitempty"`
	Events               []USBDeviceEvent `json:"events,omitempty"`
	TotalBytesIn         int64            `json:"total_bytes_in"`
	TotalBytesOut        int64            `json:"total_bytes_out"`
	LastInterface        string           `json:"last_interface,omitempty"`
	Speed                string           `json:"speed,omitempty"`
}

func (m *USBHistoryModule) GenerateSummary(params mod.ModuleParams) ([]USBDeviceSummary, error) {
	var summaries []USBDeviceSummary
	deviceMap := make(map[string]*USBDeviceSummary) // key: vendorID:productID:serialNumber

	// Get current devices from system_profiler
	params.Logger.Debug("Running system_profiler for summary")
	output, err := utils.ExecCommand("system_profiler", "SPUSBDataType", "-json")
	if err == nil {
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(output), &data); err == nil {
			if usbData, ok := data["SPUSBDataType"].([]interface{}); ok {
				processSystemProfilerData(usbData, deviceMap, params)
			}
		}
	}

	// Get connection history from logs
	params.Logger.Debug("Running log show for summary")
	logOutput, err := utils.ExecCommand("log", "show",
		"--predicate", `(process == "usbd" OR process == "USBAgent" OR subsystem == "com.apple.iokit.IOUSBHostFamily" OR subsystem == "com.apple.usb") AND (eventMessage CONTAINS[c] "attach" OR eventMessage CONTAINS[c] "detach" OR eventMessage CONTAINS[c] "connect" OR eventMessage CONTAINS[c] "disconnect" OR eventMessage CONTAINS[c] "enumerate" OR eventMessage CONTAINS[c] "vendor" OR eventMessage CONTAINS[c] "product")`,
		"--style", "json",
		"--info",
		"--debug",
		"--last", fmt.Sprintf("%dh", LogCollectionHours))
	if err == nil {
		var logs []map[string]interface{}
		if err := json.Unmarshal([]byte(logOutput), &logs); err == nil {
			params.Logger.Debug("Processing %d log entries for summary", len(logs))
			processLogData(logs, deviceMap)
		} else {
			params.Logger.Debug("Error parsing log output for summary: %v", err)
		}
	} else {
		params.Logger.Debug("Error running log show for summary: %v", err)
	}

	// Get additional details from ioreg
	params.Logger.Debug("Running ioreg for summary")
	ioregOutput, err := utils.ExecCommand("ioreg", "-p", "IOUSB", "-l", "-w", "0")
	if err == nil {
		devices, err := parseIORegOutput(string(ioregOutput))
		if err == nil {
			params.Logger.Debug("Processing %d ioreg devices for summary", len(devices))
			processIORegData(devices, deviceMap)
		} else {
			params.Logger.Debug("Error parsing ioreg output for summary: %v", err)
		}
	} else {
		params.Logger.Debug("Error running ioreg for summary: %v", err)
	}

	// Convert map to slice
	for key, summary := range deviceMap {
		params.Logger.Debug("Adding device to summary: %s", key)
		summaries = append(summaries, *summary)
	}

	return summaries, nil
}

// parseIORegOutput parses the output of ioreg command into structured data
func parseIORegOutput(output string) ([]map[string]interface{}, error) {
	var devices []map[string]interface{}
	var currentDevice map[string]interface{}
	var currentProperties map[string]interface{}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.Contains(line, "+-o") || strings.Contains(line, "| +-o") {
			if currentDevice != nil {
				if deviceName, ok := currentDevice["name"].(string); ok {
					deviceData := processDevice(deviceName, currentProperties)
					devices = append(devices, deviceData)
				}
			}
			currentDevice = make(map[string]interface{})
			currentProperties = make(map[string]interface{})
			currentDevice["name"] = strings.TrimPrefix(line, "+-o")
			currentDevice["name"] = strings.TrimPrefix(currentDevice["name"].(string), "| +-o")
			currentDevice["name"] = strings.TrimSpace(currentDevice["name"].(string))
		} else if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				key = strings.TrimPrefix(key, "|")
				key = strings.TrimSpace(key)
				key = strings.Trim(key, "\"")

				value := strings.TrimSpace(parts[1])
				currentProperties[key] = cleanPropertyValue(value)
			}
		}
	}

	// Add the last device if exists
	if currentDevice != nil {
		if deviceName, ok := currentDevice["name"].(string); ok {
			deviceData := processDevice(deviceName, currentProperties)
			devices = append(devices, deviceData)
		}
	}

	if len(devices) == 0 {
		return nil, errNoUSBDevices
	}

	return devices, nil
}

// processDevice creates a clean device record from the device name and properties
func processDevice(deviceName string, properties map[string]interface{}) map[string]interface{} {
	cleanName := strings.Split(deviceName, "<")[0]
	cleanName = strings.TrimSpace(cleanName)
	if cleanName == "" {
		cleanName = "unknown_device"
	}

	deviceInfo := ""
	if strings.Contains(deviceName, "<") && strings.Contains(deviceName, ">") {
		info := strings.Split(deviceName, "<")[1]
		deviceInfo = strings.TrimSuffix(info, ">")
	}

	// Create device data structure
	deviceData := map[string]interface{}{
		"clean_name":  cleanName,
		"name":        deviceName,
		"device_info": deviceInfo,
	}

	// Clean up properties
	if properties != nil {
		cleanProperties := make(map[string]interface{})
		for key, value := range properties {
			cleanProperties[key] = value
		}
		deviceData["properties"] = cleanProperties
	}

	return deviceData
}

// cleanPropertyValue cleans up property values by removing unnecessary escaping
func cleanPropertyValue(value string) interface{} {
	// Try to parse JSON-like values first
	if strings.HasPrefix(value, "{") && strings.HasSuffix(value, "}") {
		// Remove escaped quotes inside JSON
		value = strings.ReplaceAll(value, "\\\"", "\"")
		var jsonValue interface{}
		if err := json.Unmarshal([]byte(value), &jsonValue); err == nil {
			return jsonValue
		}
	}

	// Remove surrounding quotes if present
	value = strings.TrimSpace(value)
	if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
		value = value[1 : len(value)-1]
	}

	// Clean up escaped quotes and backslashes
	value = strings.ReplaceAll(value, "\\\"", "\"")
	value = strings.ReplaceAll(value, "\\\\", "\\")

	// Handle boolean values
	switch strings.ToLower(value) {
	case "yes", "true":
		return true
	case "no", "false":
		return false
	}

	return value
}
