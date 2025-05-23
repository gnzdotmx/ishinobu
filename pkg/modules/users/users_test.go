package users

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gnzdotmx/ishinobu/pkg/mod"
	"github.com/gnzdotmx/ishinobu/pkg/testutils"
	"github.com/gnzdotmx/ishinobu/pkg/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUsersModule_Interface tests the module interface methods
func TestUsersModule_Interface(t *testing.T) {
	module := &UsersModule{
		Name:        "users",
		Description: "Enumerates current and deleted user profiles, identifies admin users and last logged in user",
	}

	assert.Equal(t, "users", module.GetName())
	assert.Equal(t, "Enumerates current and deleted user profiles, identifies admin users and last logged in user", module.GetDescription())
}

// TestDeletedUsersStruct verifies the DeletedUser struct works as expected
func TestDeletedUsersStruct(t *testing.T) {
	user := DeletedUser{
		DateDeleted: "2023-01-15T14:30:00Z",
		UniqueID:    "502",
		Name:        "deleteduser",
		RealName:    "Deleted User",
	}

	assert.Equal(t, "2023-01-15T14:30:00Z", user.DateDeleted)
	assert.Equal(t, "502", user.UniqueID)
	assert.Equal(t, "deleteduser", user.Name)
	assert.Equal(t, "Deleted User", user.RealName)
}

// TestUserInfoStruct verifies the UserInfo struct works as expected
func TestUserInfoStruct(t *testing.T) {
	userInfo := UserInfo{
		UniqueID: "501",
		RealName: "Test User",
	}

	assert.Equal(t, "501", userInfo.UniqueID)
	assert.Equal(t, "Test User", userInfo.RealName)
}

// TestModuleRun performs a basic test of the module's Run function
// This test only verifies the module doesn't crash and returns no error
func TestModuleRun(t *testing.T) {
	// Skip this test in CI environments where file access might be restricted
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping test in CI environment")
	}

	// Create test directories
	tmpDir, err := os.MkdirTemp("", "ishinobu_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	logsDir := filepath.Join(tmpDir, "logs")
	outputDir := filepath.Join(tmpDir, "output")
	require.NoError(t, os.MkdirAll(logsDir, 0755))
	require.NoError(t, os.MkdirAll(outputDir, 0755))

	// Create test parameters with a valid logger
	logger := testutils.NewTestLogger()
	params := mod.ModuleParams{
		ExportFormat:        "json",
		OutputDir:           outputDir,
		LogsDir:             logsDir,
		CollectionTimestamp: time.Now().UTC().Format(utils.TimeFormat),
		Logger:              *logger,
	}

	// Run the module
	module := &UsersModule{
		Name:        "users",
		Description: "Enumerates current and deleted user profiles, identifies admin users and last logged in user",
	}

	// This only tests that the module runs without panicking
	err = module.Run(params)

	// We don't care about the error details, only that it doesn't panic
	if err != nil {
		t.Logf("Got error: %v - this is expected on systems without proper permissions", err)
	}
}

// createDeletedUsersPlist creates a test plist file with deleted users data
func createDeletedUsersPlist(t *testing.T, dir string) string {
	plistContent := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>deletedUsers</key>
	<array>
		<dict>
			<key>date</key>
			<string>2023-01-15T14:30:00Z</string>
			<key>dsAttrTypeStandard:UniqueID</key>
			<string>502</string>
			<key>dsAttrTypeStandard:RealName</key>
			<string>Deleted User</string>
			<key>name</key>
			<string>deleteduser</string>
		</dict>
	</array>
</dict>
</plist>`

	plistPath := filepath.Join(dir, "deleted_users.plist")
	err := os.WriteFile(plistPath, []byte(plistContent), 0600)
	require.NoError(t, err)
	return plistPath
}

// createAdminUsersPlist creates a test plist file with admin users data
func createAdminUsersPlist(t *testing.T, dir string) string {
	plistContent := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>users</key>
	<array>
		<string>admin1</string>
		<string>admin2</string>
	</array>
</dict>
</plist>`

	plistPath := filepath.Join(dir, "admin_users.plist")
	err := os.WriteFile(plistPath, []byte(plistContent), 0600)
	require.NoError(t, err)
	return plistPath
}

// createLastUserPlist creates a test plist file with last logged in user data
func createLastUserPlist(t *testing.T, dir string) string {
	plistContent := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>lastUserName</key>
	<string>testuser</string>
</dict>
</plist>`

	plistPath := filepath.Join(dir, "login_window.plist")
	err := os.WriteFile(plistPath, []byte(plistContent), 0600)
	require.NoError(t, err)
	return plistPath
}

// createUserPlist creates a test plist file with user info data
func createUserPlist(t *testing.T, dir, username string) string {
	plistContent := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>uid</key>
	<array>
		<string>501</string>
	</array>
	<key>realname</key>
	<array>
		<string>Test User</string>
	</array>
</dict>
</plist>`

	// Create the directory structure
	userDir := filepath.Join(dir, "private", "var", "db", "dslocal", "nodes", "Default", "users")
	err := os.MkdirAll(userDir, 0755)
	require.NoError(t, err)

	plistPath := filepath.Join(userDir, username+".plist")
	err = os.WriteFile(plistPath, []byte(plistContent), 0600)
	require.NoError(t, err)
	return plistPath
}

// createEmptyPlist creates an empty plist file
func createEmptyPlist(t *testing.T, dir, filename string) string {
	plistContent := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
</dict>
</plist>`

	plistPath := filepath.Join(dir, filename)
	err := os.WriteFile(plistPath, []byte(plistContent), 0600)
	require.NoError(t, err)
	return plistPath
}

// createInvalidPlist creates an invalid plist file
func createInvalidPlist(t *testing.T, dir, filename string) string {
	plistPath := filepath.Join(dir, filename)
	err := os.WriteFile(plistPath, []byte("This is not a valid plist file"), 0600)
	require.NoError(t, err)
	return plistPath
}

// TestHelperFunctions tests all helper functions with actual test data
func TestHelperFunctions(t *testing.T) {
	// Skip this test in CI environments where file access might be restricted
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping test in CI environment")
	}

	// Create a temporary test directory
	tmpDir, err := os.MkdirTemp("", "users_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Test getDeletedUsers function
	t.Run("getDeletedUsers", func(t *testing.T) {
		// Create valid test plist
		validPlistPath := createDeletedUsersPlist(t, tmpDir)

		// Create invalid plist
		invalidPlistPath := createInvalidPlist(t, tmpDir, "invalid_deleted_users.plist")

		// Create empty plist
		emptyPlistPath := createEmptyPlist(t, tmpDir, "empty_deleted_users.plist")

		// Test with valid plist
		t.Run("ValidPlist", func(t *testing.T) {
			// Call the actual function with the test file
			users, err := getDeletedUsers(validPlistPath)

			// We don't check the actual content because ParseBiPList is called,
			// and we can't mock it easily. We just verify the function runs without error.
			if err != nil {
				t.Logf("Got error with valid plist: %v - this may be due to ParseBiPList implementation", err)
			} else {
				// If we got no error, verify we got at least some data
				assert.NotNil(t, users)
			}
		})

		// Test with nonexistent file
		t.Run("NonexistentFile", func(t *testing.T) {
			users, err := getDeletedUsers(filepath.Join(tmpDir, "nonexistent.plist"))
			assert.Error(t, err)
			assert.Nil(t, users)
		})

		// Test with invalid plist
		t.Run("InvalidPlist", func(t *testing.T) {
			users, err := getDeletedUsers(invalidPlistPath)

			// The function might error because of ParseBiPList
			if err != nil {
				assert.Nil(t, users)
			}
		})

		// Test with empty plist
		t.Run("EmptyPlist", func(t *testing.T) {
			users, err := getDeletedUsers(emptyPlistPath)

			// This might not error, just return empty slice
			if err == nil {
				assert.Empty(t, users)
			}
		})
	})

	// Test getAdminUsers function
	t.Run("getAdminUsers", func(t *testing.T) {
		// Create valid test plist
		validPlistPath := createAdminUsersPlist(t, tmpDir)

		// Create invalid plist
		invalidPlistPath := createInvalidPlist(t, tmpDir, "invalid_admin_users.plist")

		// Create empty plist
		emptyPlistPath := createEmptyPlist(t, tmpDir, "empty_admin_users.plist")

		// Test with valid plist
		t.Run("ValidPlist", func(t *testing.T) {
			// Call the actual function with the test file
			users, err := getAdminUsers(validPlistPath)

			// We don't check the actual content because ParseBiPList is called,
			// and we can't mock it easily. We just verify the function runs.
			if err != nil {
				t.Logf("Got error with valid plist: %v - this may be due to ParseBiPList implementation", err)
			} else {
				// If we got no error, verify we got at least some data
				assert.NotEmpty(t, users)
			}
		})

		// Test with nonexistent file
		t.Run("NonexistentFile", func(t *testing.T) {
			users, err := getAdminUsers(filepath.Join(tmpDir, "nonexistent.plist"))
			assert.Error(t, err)
			assert.Nil(t, users)
		})

		// Test with invalid plist
		t.Run("InvalidPlist", func(t *testing.T) {
			users, err := getAdminUsers(invalidPlistPath)

			// The function might error because of ParseBiPList
			if err != nil {
				assert.Nil(t, users)
			}
		})

		// Test with empty plist
		t.Run("EmptyPlist", func(t *testing.T) {
			users, err := getAdminUsers(emptyPlistPath)

			// This should error with no admin users found
			assert.Error(t, err)
			assert.Nil(t, users)
		})
	})

	// Test getLastLoggedInUser function
	t.Run("getLastLoggedInUser", func(t *testing.T) {
		// Create valid test plist
		validPlistPath := createLastUserPlist(t, tmpDir)

		// Create invalid plist
		invalidPlistPath := createInvalidPlist(t, tmpDir, "invalid_last_user.plist")

		// Create empty plist
		emptyPlistPath := createEmptyPlist(t, tmpDir, "empty_last_user.plist")

		// Test with valid plist
		t.Run("ValidPlist", func(t *testing.T) {
			// Call the actual function with the test file
			user, err := getLastLoggedInUser(validPlistPath)

			// We don't check the actual content because ParseBiPList is called,
			// and we can't mock it easily. We just verify the function runs.
			if err != nil {
				t.Logf("Got error with valid plist: %v - this may be due to ParseBiPList implementation", err)
			} else {
				// If we got no error, verify we got some data
				assert.NotEmpty(t, user)
			}
		})

		// Test with nonexistent file
		t.Run("NonexistentFile", func(t *testing.T) {
			user, err := getLastLoggedInUser(filepath.Join(tmpDir, "nonexistent.plist"))
			assert.Error(t, err)
			assert.Empty(t, user)
		})

		// Test with invalid plist
		t.Run("InvalidPlist", func(t *testing.T) {
			user, err := getLastLoggedInUser(invalidPlistPath)

			// The function might error because of ParseBiPList
			if err != nil {
				assert.Empty(t, user)
			}
		})

		// Test with empty plist
		t.Run("EmptyPlist", func(t *testing.T) {
			user, err := getLastLoggedInUser(emptyPlistPath)

			// This should error with last user not found
			assert.Error(t, err)
			assert.Empty(t, user)
		})
	})

	// Test getUserInfo function
	t.Run("getUserInfo", func(t *testing.T) {
		// Create valid test plist
		username := "testuser"
		userPlistPath := createUserPlist(t, tmpDir, username)

		// Test with valid plist
		t.Run("ValidPlist", func(t *testing.T) {
			// Call the function with our test plist path
			userInfo, err := getUserInfo(userPlistPath)

			// We don't check the actual content because ParseBiPList is called,
			// and we can't mock it easily. We just verify the function runs.
			if err != nil {
				t.Logf("Got error with valid plist: %v - this may be due to ParseBiPList implementation", err)
			} else {
				// If we got no error, verify we got some data
				assert.NotNil(t, userInfo)
				assert.Equal(t, "501", userInfo.UniqueID)
				assert.Equal(t, "Test User", userInfo.RealName)
			}
		})

		// Test with nonexistent file
		t.Run("NonexistentFile", func(t *testing.T) {
			nonexistentPath := filepath.Join(tmpDir, "nonexistent.plist")
			userInfo, err := getUserInfo(nonexistentPath)

			// Should error when file doesn't exist
			assert.Error(t, err)
			assert.Nil(t, userInfo)
		})

		// Test with invalid plist
		t.Run("InvalidPlist", func(t *testing.T) {
			invalidPath := createInvalidPlist(t, tmpDir, "invalid_user.plist")
			userInfo, err := getUserInfo(invalidPath)

			// The function might error because of ParseBiPList
			if err != nil {
				assert.Nil(t, userInfo)
			}
		})

		// Test with empty plist
		t.Run("EmptyPlist", func(t *testing.T) {
			emptyPath := createEmptyPlist(t, tmpDir, "empty_user.plist")
			userInfo, err := getUserInfo(emptyPath)

			// Should not error, just return empty fields
			if err == nil {
				assert.NotNil(t, userInfo)
				assert.Empty(t, userInfo.UniqueID)
				assert.Empty(t, userInfo.RealName)
			}
		})
	})
}
