// This module collects information about installed Claude MCP servers.
// It reads the configuration file at ~/Library/Application Support/Claude/claude_desktop_config.json
// and extracts information about configured MCP servers.
// It collects:
// - Server name
// - Command
// - Arguments
// - Environment variables
package claude

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gnzdotmx/ishinobu/pkg/mod"
	"github.com/gnzdotmx/ishinobu/pkg/utils"
)

type ClaudeModule struct {
	Name        string
	Description string
}

// MCPServer represents a single MCP server configuration
type MCPServer struct {
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env,omitempty"`
}

// ClaudeConfig represents the structure of the Claude desktop config file
type ClaudeConfig struct {
	MCPServers map[string]MCPServer `json:"mcpServers"`
}

// Error variables
var (
	errNoMCPServers = errors.New("no MCP servers found in config file")
)

func init() {
	module := &ClaudeModule{
		Name:        "claude",
		Description: "Collects information about installed Claude MCP servers",
	}
	mod.RegisterModule(module)
}

func (m *ClaudeModule) GetName() string {
	return m.Name
}

func (m *ClaudeModule) GetDescription() string {
	return m.Description
}

func (m *ClaudeModule) Run(params mod.ModuleParams) error {
	// Get all Claude config locations
	locations, err := filepath.Glob("/Users/*/Library/Application Support/Claude")
	if err != nil {
		params.Logger.Error("Error listing Claude locations: %v", err)
		return fmt.Errorf("error finding Claude locations: %w", err)
	}

	params.Logger.Debug("Found %d Claude locations", len(locations))

	// Process each location's config
	for _, location := range locations {
		username := filepath.Base(filepath.Dir(filepath.Dir(filepath.Dir(location))))
		params.Logger.Info("Processing Claude configuration for user: %s", username)

		// Create data writer for this user
		outputFileName := utils.GetOutputFileName(m.GetName()+"-"+username, params.ExportFormat, params.OutputDir)
		writer, err := utils.NewDataWriter(params.OutputDir, outputFileName, params.ExportFormat)
		if err != nil {
			params.Logger.Error("Error creating data writer for user %s: %v", username, err)
			continue
		}
		defer writer.Close()

		// Construct path to Claude config file
		configPath := filepath.Join(location, "claude_desktop_config.json")
		params.Logger.Debug("Looking for config file at: %s", configPath)

		// Check if config file exists and is accessible
		if info, err := os.Stat(configPath); err != nil {
			switch {
			case os.IsNotExist(err):
				params.Logger.Debug("Claude config file not found at %s", configPath)
			case os.IsPermission(err):
				params.Logger.Debug("Permission denied accessing Claude config at %s", configPath)
			default:
				params.Logger.Debug("Error accessing Claude config at %s: %v", configPath, err)
			}
			continue
		} else {
			params.Logger.Debug("Found config file: %s (size: %d bytes)", configPath, info.Size())
		}

		// Collect MCP servers for this user
		if err := collectMCPServers(writer, configPath, params.CollectionTimestamp); err != nil {
			params.Logger.Error("Error collecting MCP servers for %s: %v", username, err)
			continue
		} else {
			params.Logger.Debug("Successfully collected MCP servers for user %s", username)
		}
	}

	return nil
}

// collectMCPServers reads and parses the Claude desktop config file
func collectMCPServers(writer utils.DataWriter, configPath string, collectionTimestamp string) error {
	// Read the config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("error reading config file: %w", err)
	}

	// Parse the config file
	var config ClaudeConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("error parsing config file: %w", err)
	}

	// Check if there are any servers to process
	if len(config.MCPServers) == 0 {
		return errNoMCPServers
	}

	// Process each server
	for name, server := range config.MCPServers {
		// Create a record for each server
		record := utils.Record{
			CollectionTimestamp: collectionTimestamp,
			EventTimestamp:      utils.Now(),
			SourceFile:          configPath,
			Data: map[string]interface{}{
				"server_name": name,
				"command":     server.Command,
				"args":        server.Args,
				"env":         server.Env,
				"username":    filepath.Base(filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(configPath))))),
			},
		}

		if err := writer.WriteRecord(record); err != nil {
			return fmt.Errorf("failed to write record: %w", err)
		}
	}

	return nil
}
