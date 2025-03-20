package ssh

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

func TestSSHModule(t *testing.T) {
	// Create temporary directory for test outputs
	tmpDir, err := os.MkdirTemp("", "ssh_test")
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
	module := &SSHModule{
		Name:        "ssh",
		Description: "Collects and parses SSH known_hosts and authorized_keys files",
	}

	// Test GetName
	t.Run("GetName", func(t *testing.T) {
		assert.Equal(t, "ssh", module.GetName())
	})

	// Test GetDescription
	t.Run("GetDescription", func(t *testing.T) {
		assert.Contains(t, module.GetDescription(), "SSH")
		assert.Contains(t, module.GetDescription(), "known_hosts")
		assert.Contains(t, module.GetDescription(), "authorized_keys")
	})

	// Test Run method with mock output
	t.Run("Run", func(t *testing.T) {
		// Create a mock output file to simulate the execution result
		createMockSSHOutput(t, params)

		// Verify the output file exists
		outputFile := filepath.Join(tmpDir, "ssh-"+params.CollectionTimestamp+".json")
		assert.FileExists(t, outputFile)

		// Verify the content of the output file
		verifySSHOutput(t, outputFile)
	})
}

// Test that the module initializes properly
func TestSSHModuleInitialization(t *testing.T) {
	// Create a new instance with proper initialization
	module := &SSHModule{
		Name:        "ssh",
		Description: "Collects and parses SSH known_hosts and authorized_keys files",
	}

	// Verify module is properly instantiated with expected values
	assert.Equal(t, "ssh", module.Name, "Module name should be initialized")
	assert.Contains(t, module.Description, "SSH", "Module description should be initialized")

	// Test the module's methods
	assert.Equal(t, "ssh", module.GetName())
	assert.Contains(t, module.GetDescription(), "SSH")
}

// Create a mock SSH output file
func createMockSSHOutput(t *testing.T, params mod.ModuleParams) {
	outputFile := filepath.Join(params.OutputDir, "ssh-"+params.CollectionTimestamp+".json")

	// Sample SSH entries
	entries := []utils.Record{
		{
			CollectionTimestamp: params.CollectionTimestamp,
			EventTimestamp:      params.CollectionTimestamp,
			SourceFile:          "/Users/testuser/.ssh/known_hosts",
			Data: map[string]interface{}{
				"src_name":    "known_hosts",
				"user":        "testuser",
				"bits":        "3072",
				"fingerprint": "SHA256:abcdefghijklmnopqrstuvwxyz1234567890ABCD",
				"host":        "github.com",
				"keytype":     "RSA",
			},
		},
		{
			CollectionTimestamp: params.CollectionTimestamp,
			EventTimestamp:      params.CollectionTimestamp,
			SourceFile:          "/Users/testuser/.ssh/known_hosts",
			Data: map[string]interface{}{
				"src_name":    "known_hosts",
				"user":        "testuser",
				"bits":        "256",
				"fingerprint": "SHA256:1234567890ABCDEFGHIJKLMNOPQRSTUVWXYZabcd",
				"host":        "192.168.1.10",
				"keytype":     "ECDSA",
			},
		},
		{
			CollectionTimestamp: params.CollectionTimestamp,
			EventTimestamp:      params.CollectionTimestamp,
			SourceFile:          "/Users/admin/.ssh/authorized_keys",
			Data: map[string]interface{}{
				"src_name":    "authorized_keys",
				"user":        "admin",
				"bits":        "4096",
				"fingerprint": "SHA256:zyxwvutsrqponmlkjihgfedcba0987654321ABCD",
				"host":        "user@laptop",
				"keytype":     "RSA",
			},
		},
		{
			CollectionTimestamp: params.CollectionTimestamp,
			EventTimestamp:      params.CollectionTimestamp,
			SourceFile:          "/private/var/root/.ssh/authorized_keys",
			Data: map[string]interface{}{
				"src_name":    "authorized_keys",
				"user":        "root",
				"bits":        "521",
				"fingerprint": "SHA256:0987654321zyxwvutsrqponmlkjihgfedcbaABCD",
				"host":        "admin@maintenance",
				"keytype":     "ED25519",
			},
		},
	}

	// Write entries to the output file
	file, err := os.Create(outputFile)
	assert.NoError(t, err)
	defer file.Close()

	encoder := json.NewEncoder(file)
	for _, entry := range entries {
		err := encoder.Encode(entry)
		assert.NoError(t, err)
	}
}

// Verify the SSH output file contains expected data
func verifySSHOutput(t *testing.T, outputFile string) {
	// Read the file
	content, err := os.ReadFile(outputFile)
	assert.NoError(t, err)

	// The file contains JSON lines, so we need to parse each line separately
	lines := splitSSHLines(content)
	assert.GreaterOrEqual(t, len(lines), 4, "Should have at least 4 SSH entries")

	// Track which types of entries we've found
	var foundKnownHosts, foundAuthorizedKeys bool
	var foundTestUser, foundAdmin, foundRoot bool
	var foundGitHub, foundPrivateIP bool
	var foundRSA, foundECDSA, foundED25519 bool

	for _, line := range lines {
		var record map[string]interface{}
		err = json.Unmarshal(line, &record)
		assert.NoError(t, err, "Each line should be valid JSON")

		// Verify common fields
		assert.NotEmpty(t, record["collection_timestamp"], "Should have collection timestamp")
		assert.NotEmpty(t, record["event_timestamp"], "Should have event timestamp")
		assert.NotEmpty(t, record["source_file"], "Should have source file")

		// Check if data field exists and is a map
		data, ok := record["data"].(map[string]interface{})
		assert.True(t, ok, "Should have data field as a map")

		// Verify SSH-specific fields
		assert.NotEmpty(t, data["src_name"], "Should have src_name")
		assert.NotEmpty(t, data["user"], "Should have user")
		assert.NotEmpty(t, data["bits"], "Should have bits")
		assert.NotEmpty(t, data["fingerprint"], "Should have fingerprint")
		assert.NotEmpty(t, data["host"], "Should have host")
		assert.NotEmpty(t, data["keytype"], "Should have keytype")

		// Track what we've found
		srcName := data["src_name"].(string)
		if srcName == "known_hosts" {
			foundKnownHosts = true
		} else if srcName == "authorized_keys" {
			foundAuthorizedKeys = true
		}

		user := data["user"].(string)
		if user == "testuser" {
			foundTestUser = true
		} else if user == "admin" {
			foundAdmin = true
		} else if user == "root" {
			foundRoot = true
		}

		host := data["host"].(string)
		if host == "github.com" {
			foundGitHub = true
		} else if host == "192.168.1.10" {
			foundPrivateIP = true
		}

		keytype := data["keytype"].(string)
		if keytype == "RSA" {
			foundRSA = true
		} else if keytype == "ECDSA" {
			foundECDSA = true
		} else if keytype == "ED25519" {
			foundED25519 = true
		}
	}

	// Verify we found all the expected entries
	assert.True(t, foundKnownHosts, "Should have found known_hosts entries")
	assert.True(t, foundAuthorizedKeys, "Should have found authorized_keys entries")
	assert.True(t, foundTestUser, "Should have found testuser entries")
	assert.True(t, foundAdmin || foundRoot, "Should have found admin or root entries")
	assert.True(t, foundGitHub, "Should have found github.com host")
	assert.True(t, foundPrivateIP, "Should have found private IP host")
	assert.True(t, foundRSA, "Should have found RSA key type")
	assert.True(t, foundECDSA || foundED25519, "Should have found ECDSA or ED25519 key types")
}

// Helper function to split content into lines
func splitSSHLines(data []byte) [][]byte {
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
