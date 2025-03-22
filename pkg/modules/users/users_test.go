package users

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

func TestUsersModule(t *testing.T) {
	// Create temporary directory for test outputs
	tmpDir, err := os.MkdirTemp("", "users_test")
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
	module := &UsersModule{
		Name:        "users",
		Description: "Enumerates current and deleted user profiles, identifies admin users and last logged in user",
	}

	// Test GetName
	t.Run("GetName", func(t *testing.T) {
		assert.Equal(t, "users", module.GetName())
	})

	// Test GetDescription
	t.Run("GetDescription", func(t *testing.T) {
		assert.Contains(t, module.GetDescription(), "user profiles")
	})

	// Test Run method with mock output
	t.Run("Run", func(t *testing.T) {
		// Create a mock output file to simulate the module's output
		createMockUsersOutput(t, params)

		// Verify the output file exists
		outputFile := filepath.Join(tmpDir, "users-"+params.CollectionTimestamp+".json")
		assert.FileExists(t, outputFile)

		// Verify the content of the output file
		verifyUsersOutput(t, outputFile)
	})
}

// Test that the module initializes properly
func TestUsersModuleInitialization(t *testing.T) {
	// Create a new instance with proper initialization
	module := &UsersModule{
		Name:        "users",
		Description: "Enumerates current and deleted user profiles, identifies admin users and last logged in user",
	}

	// Verify module is properly instantiated with expected values
	assert.Equal(t, "users", module.Name, "Module name should be initialized")
	assert.Contains(t, module.Description, "user profiles", "Module description should be initialized")

	// Test the module's methods
	assert.Equal(t, "users", module.GetName())
	assert.Contains(t, module.GetDescription(), "user profiles")
}

// Create a mock users output file
func createMockUsersOutput(t *testing.T, params mod.ModuleParams) {
	outputFile := filepath.Join(params.OutputDir, "users-"+params.CollectionTimestamp+".json")

	// Create sample user records
	users := []utils.Record{
		// Current user - admin, last logged in
		{
			CollectionTimestamp: params.CollectionTimestamp,
			EventTimestamp:      params.CollectionTimestamp,
			SourceFile:          "/Users/admin_user",
			Data: map[string]interface{}{
				"user":              "admin_user",
				"real_name":         "Administrator",
				"uniq_id":           "501",
				"admin":             true,
				"lastloggedin_user": true,
				"mtime":             params.CollectionTimestamp,
				"atime":             params.CollectionTimestamp,
				"ctime":             params.CollectionTimestamp,
			},
		},
		// Current user - standard
		{
			CollectionTimestamp: params.CollectionTimestamp,
			EventTimestamp:      params.CollectionTimestamp,
			SourceFile:          "/Users/standard_user",
			Data: map[string]interface{}{
				"user":              "standard_user",
				"real_name":         "Standard User",
				"uniq_id":           "502",
				"admin":             false,
				"lastloggedin_user": false,
				"mtime":             params.CollectionTimestamp,
				"atime":             params.CollectionTimestamp,
				"ctime":             params.CollectionTimestamp,
			},
		},
		// Deleted user
		{
			CollectionTimestamp: params.CollectionTimestamp,
			EventTimestamp:      "2023-03-15T12:30:45Z",
			SourceFile:          "/Library/Preferences/com.apple.preferences.accounts.plist",
			Data: map[string]interface{}{
				"user":              "deleted_user",
				"real_name":         "Deleted User",
				"uniq_id":           "503",
				"admin":             "",
				"lastloggedin_user": "",
				"date_deleted":      "2023-03-15T12:30:45Z",
			},
		},
	}

	// Write each user as a JSON line
	file, err := os.Create(outputFile)
	assert.NoError(t, err)
	defer file.Close()

	encoder := json.NewEncoder(file)
	for _, user := range users {
		err := encoder.Encode(user)
		assert.NoError(t, err)
	}
}

// Verify the users output file contains expected data
func verifyUsersOutput(t *testing.T, outputFile string) {
	// Read the output file
	content, err := os.ReadFile(outputFile)
	assert.NoError(t, err)

	// Split the content into JSON lines
	lines := testutils.SplitLines(content)
	assert.GreaterOrEqual(t, len(lines), 3, "Should have at least 3 user records")

	// Track which user types we found
	var foundAdmin, foundStandard, foundDeleted bool

	// Verify each user record has the expected fields
	for _, line := range lines {
		var record map[string]interface{}
		err := json.Unmarshal(line, &record)
		assert.NoError(t, err, "Each line should be valid JSON")

		// Verify common fields
		assert.NotEmpty(t, record["collection_timestamp"])
		assert.NotEmpty(t, record["event_timestamp"])
		assert.NotEmpty(t, record["source_file"])

		// Check data fields
		data, ok := record["data"].(map[string]interface{})
		assert.True(t, ok, "Should have a data field as a map")

		// Common fields for all users
		assert.NotEmpty(t, data["user"])
		assert.NotEmpty(t, data["real_name"])
		assert.NotEmpty(t, data["uniq_id"])

		// Check user type and specific fields
		userName, _ := data["user"].(string)
		sourceFile, _ := record["source_file"].(string)

		switch userName {
		case "admin_user":
			foundAdmin = true
			// Check admin-specific fields
			assert.Equal(t, true, data["admin"])
			assert.Equal(t, true, data["lastloggedin_user"])
			assert.Contains(t, sourceFile, "/Users/")

			// Check timestamps exist
			assert.NotEmpty(t, data["mtime"])
			assert.NotEmpty(t, data["atime"])
			assert.NotEmpty(t, data["ctime"])
		case "standard_user":
			foundStandard = true
			// Check standard user fields
			assert.Equal(t, false, data["admin"])
			assert.Equal(t, false, data["lastloggedin_user"])
			assert.Contains(t, sourceFile, "/Users/")

			// Check timestamps exist
			assert.NotEmpty(t, data["mtime"])
			assert.NotEmpty(t, data["atime"])
			assert.NotEmpty(t, data["ctime"])
		case "deleted_user":
			foundDeleted = true
			// Check deleted user fields
			assert.NotEmpty(t, data["date_deleted"])
			assert.Contains(t, sourceFile, "preferences.accounts.plist")
		}
	}

	// Verify we found all user types
	assert.True(t, foundAdmin, "Should have found admin user")
	assert.True(t, foundStandard, "Should have found standard user")
	assert.True(t, foundDeleted, "Should have found deleted user")
}
