package claude

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnzdotmx/ishinobu/pkg/testutils"
	"github.com/gnzdotmx/ishinobu/pkg/utils"
)

func TestClaudeModuleBasics(t *testing.T) {
	module := &ClaudeModule{
		Name:        "claude",
		Description: "Collects information about installed Claude MCP servers",
	}

	assert.Equal(t, "claude", module.GetName())
	assert.Equal(t, "Collects information about installed Claude MCP servers", module.GetDescription())
}

func TestCollectMCPServers(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir, err := os.MkdirTemp("", "claude_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create mock config files for multiple users
	users := []string{"user1", "user2"}
	mockConfigs := make(map[string]ClaudeConfig)

	for _, user := range users {
		// Create user directory structure
		userDir := filepath.Join(tmpDir, "Users", user)
		configDir := filepath.Join(userDir, "Library/Application Support/Claude")
		err = os.MkdirAll(configDir, 0755)
		require.NoError(t, err)

		// Create mock config for this user
		mockConfig := ClaudeConfig{
			MCPServers: map[string]MCPServer{
				"filesystem": {
					Command: "npx",
					Args: []string{
						"-y",
						"@modelcontextprotocol/server-filesystem",
						"/Users/" + user + "/Desktop",
						"/Users/" + user + "/Downloads",
					},
					Env: map[string]string{
						"APPDATA": "C:\\Users\\" + user + "\\AppData\\Roaming\\",
					},
				},
				"brave-search": {
					Command: "npx",
					Args: []string{
						"-y",
						"@modelcontextprotocol/server-brave-search",
					},
					Env: map[string]string{
						"BRAVE_API_KEY": "test-key-" + user,
					},
				},
			},
		}
		mockConfigs[user] = mockConfig

		// Write config file
		configPath := filepath.Join(configDir, "claude_desktop_config.json")
		configData, err := json.MarshalIndent(mockConfig, "", "  ")
		require.NoError(t, err)
		err = os.WriteFile(configPath, configData, 0600)
		require.NoError(t, err)
	}

	// Create test writer
	testWriter := &testutils.TestDataWriter{Records: []utils.Record{}}

	// Test collecting from each user's config
	for _, user := range users {
		configPath := filepath.Join(tmpDir, "Users", user, "Library/Application Support/Claude/claude_desktop_config.json")
		collectionTimestamp := time.Now().Format(time.RFC3339)
		err = collectMCPServers(testWriter, configPath, collectionTimestamp)
		require.NoError(t, err)
	}

	// Verify records
	expectedRecords := len(users) * len(mockConfigs[users[0]].MCPServers)
	assert.Equal(t, expectedRecords, len(testWriter.Records), "Number of records should match total servers across all users")

	// Verify each record
	for _, record := range testWriter.Records {
		data := record.Data.(map[string]interface{})
		username := data["username"].(string)
		serverName := data["server_name"].(string)
		mockServer := mockConfigs[username].MCPServers[serverName]

		assert.Equal(t, mockServer.Command, data["command"])
		assert.Equal(t, mockServer.Args, data["args"])
		assert.Equal(t, mockServer.Env, data["env"])
		assert.Contains(t, record.SourceFile, username)
	}
}

func TestCollectMCPServersErrors(t *testing.T) {
	testWriter := &testutils.TestDataWriter{Records: []utils.Record{}}

	// Test with non-existent file
	err := collectMCPServers(testWriter, "/nonexistent/file.json", time.Now().Format(time.RFC3339))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error reading config file")

	// Test with invalid JSON
	tmpFile := filepath.Join(t.TempDir(), "invalid.json")
	err = os.WriteFile(tmpFile, []byte("invalid json"), 0600)
	require.NoError(t, err)

	err = collectMCPServers(testWriter, tmpFile, time.Now().Format(time.RFC3339))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error parsing config file")
}
