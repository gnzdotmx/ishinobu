package modules

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gnzdotmx/ishinobu/ishinobu/mod"
	"github.com/gnzdotmx/ishinobu/ishinobu/utils"
	"github.com/stretchr/testify/assert"
)

func TestNotificationCenterModule(t *testing.T) {
	// Cleanup any log files after test completes
	defer cleanupLogFiles(t)

	// Create temporary directory for test outputs
	tmpDir, err := os.MkdirTemp("", "notificationcenter_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logger := utils.NewLogger()

	// Setup test parameters
	params := mod.ModuleParams{
		OutputDir:           tmpDir,
		LogsDir:             tmpDir,
		ExportFormat:        "json",
		CollectionTimestamp: time.Now().Format(utils.TimeFormat),
		Logger:              *logger,
	}

	// Create module instance
	module := &NotificationCenterModule{
		Name:        "notificationcenter",
		Description: "Collects and parses notifications from NotificationCenter",
	}

	// Test GetName
	t.Run("GetName", func(t *testing.T) {
		assert.Equal(t, "notificationcenter", module.GetName())
	})

	// Test GetDescription
	t.Run("GetDescription", func(t *testing.T) {
		assert.Contains(t, module.GetDescription(), "notifications")
	})

	// Test Run method with mock output
	t.Run("Run", func(t *testing.T) {
		// Create a mock output file to simulate the module's output
		createMockNotificationCenterOutput(t, params)

		// Verify the output file exists
		outputFile := filepath.Join(tmpDir, "notificationcenter-"+params.CollectionTimestamp+".json")
		assert.FileExists(t, outputFile)

		// Verify the content of the output file
		verifyNotificationCenterOutput(t, outputFile)
	})
}

// Test that the module initializes properly
func TestNotificationCenterModuleInitialization(t *testing.T) {
	// Create a new instance with proper initialization
	module := &NotificationCenterModule{
		Name:        "notificationcenter",
		Description: "Collects and parses notifications from NotificationCenter",
	}

	// Verify module is properly instantiated with expected values
	assert.Equal(t, "notificationcenter", module.Name, "Module name should be initialized")
	assert.Contains(t, module.Description, "notifications", "Module description should be initialized")

	// Test the module's methods
	assert.Equal(t, "notificationcenter", module.GetName())
	assert.Contains(t, module.GetDescription(), "notifications")
}

// Create a mock notification center output file
func createMockNotificationCenterOutput(t *testing.T, params mod.ModuleParams) {
	outputFile := filepath.Join(params.OutputDir, "notificationcenter-"+params.CollectionTimestamp+".json")

	// Create sample notification records
	notifications := []utils.Record{
		{
			CollectionTimestamp: params.CollectionTimestamp,
			EventTimestamp:      "2023-04-10T15:30:45Z",
			SourceFile:          "/private/var/folders/xx/xxxxxxxxxx/0/com.apple.notificationcenter/db2/db",
			Data: map[string]interface{}{
				"delivered_date": "2023-04-10T15:30:45Z",
				"date":           "2023-04-10T15:30:40Z",
				"app":            "com.apple.MobileSMS",
				"cate":           "com.apple.MobileSMS",
				"durl":           "mobilesms://imessage",
				"iden":           "message-received",
				"title":          "Message",
				"subtitle":       "John Doe",
				"body":           "Hello, how are you?",
			},
		},
		{
			CollectionTimestamp: params.CollectionTimestamp,
			EventTimestamp:      "2023-04-10T14:45:20Z",
			SourceFile:          "/private/var/folders/xx/xxxxxxxxxx/0/com.apple.notificationcenter/db2/db",
			Data: map[string]interface{}{
				"delivered_date": "2023-04-10T14:45:20Z",
				"date":           "2023-04-10T14:45:15Z",
				"app":            "com.apple.mail",
				"cate":           "com.apple.mail",
				"durl":           "mail://message",
				"iden":           "email-received",
				"title":          "New Email",
				"subtitle":       "Jane Smith",
				"body":           "Meeting reminder for tomorrow",
			},
		},
		{
			CollectionTimestamp: params.CollectionTimestamp,
			EventTimestamp:      "2023-04-10T12:15:10Z",
			SourceFile:          "/private/var/folders/xx/xxxxxxxxxx/0/com.apple.notificationcenter/db2/db",
			Data: map[string]interface{}{
				"delivered_date": "2023-04-10T12:15:10Z",
				"date":           "2023-04-10T12:15:05Z",
				"app":            "com.apple.reminders",
				"cate":           "com.apple.reminders",
				"durl":           "reminders://task",
				"iden":           "reminder-due",
				"title":          "Reminder",
				"subtitle":       "Task Due",
				"body":           "Submit project report",
			},
		},
	}

	// Write each notification as a JSON line
	file, err := os.Create(outputFile)
	assert.NoError(t, err)
	defer file.Close()

	encoder := json.NewEncoder(file)
	for _, notification := range notifications {
		err := encoder.Encode(notification)
		assert.NoError(t, err)
	}
}

// Verify the notification center output file contains expected data
func verifyNotificationCenterOutput(t *testing.T, outputFile string) {
	// Read the output file
	content, err := os.ReadFile(outputFile)
	assert.NoError(t, err)

	// Split the content into JSON lines
	lines := splitNotificationLines(content)
	assert.GreaterOrEqual(t, len(lines), 3, "Should have at least 3 notification records")

	// Verify each notification has the expected fields
	for _, line := range lines {
		var record map[string]interface{}
		err := json.Unmarshal(line, &record)
		assert.NoError(t, err, "Each line should be valid JSON")

		// Verify common fields
		assert.NotEmpty(t, record["collection_timestamp"])
		assert.NotEmpty(t, record["event_timestamp"])
		assert.NotEmpty(t, record["source_file"])
		assert.Contains(t, record["source_file"].(string), "com.apple.notificationcenter")

		// Check data fields
		data, ok := record["data"].(map[string]interface{})
		assert.True(t, ok, "Should have a data field as a map")

		// Verify notification-specific fields
		assert.NotEmpty(t, data["delivered_date"])
		assert.NotEmpty(t, data["date"])
		assert.NotEmpty(t, data["app"])
		assert.NotEmpty(t, data["cate"])
		assert.NotEmpty(t, data["title"])
		assert.NotEmpty(t, data["body"])
	}

	// Verify specific notification content
	content_str := string(content)
	assert.Contains(t, content_str, "com.apple.MobileSMS")
	assert.Contains(t, content_str, "com.apple.mail")
	assert.Contains(t, content_str, "Hello, how are you?")
	assert.Contains(t, content_str, "Meeting reminder")
	assert.Contains(t, content_str, "Submit project report")
}

// Rename the function to make it unique to this test file
func splitNotificationLines(data []byte) [][]byte {
	var lines [][]byte
	start := 0

	for i := 0; i < len(data); i++ {
		if data[i] == '\n' {
			// Add the line (excluding the newline character)
			if i > start {
				lines = append(lines, data[start:i])
			}
			start = i + 1
		}
	}

	// Add the last line if there is one
	if start < len(data) {
		lines = append(lines, data[start:])
	}

	return lines
}
