package aiconfig

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/gnzdotmx/ishinobu/pkg/mod"
	"github.com/gnzdotmx/ishinobu/pkg/utils"
)

func collectMCPConfigs(writer utils.DataWriter, params mod.ModuleParams) error {
	paths := globPaths(mcpGlobPatterns)

	for _, path := range paths {
		tool := mcpToolFromPath(path)
		if err := parseMCPFile(writer, params, path, tool); err != nil {
			params.Logger.Debug("Skipping MCP config %s: %v", path, err)
		}
	}

	if err := collectMCPAuthEntries(writer, params); err != nil {
		params.Logger.Debug("Error collecting MCP auth entries: %v", err)
	}

	return nil
}

func mcpToolFromPath(path string) string {
	lower := strings.ToLower(path)
	switch {
	case strings.Contains(lower, "/claude/"):
		return "claude_desktop"
	case strings.Contains(lower, "/.cursor/"):
		return "cursor"
	case strings.Contains(lower, "codeium") || strings.Contains(lower, "windsurf"):
		return "windsurf"
	case strings.Contains(lower, "saoudrizwan.claude-dev") || strings.Contains(lower, "/.cline/"):
		return "cline"
	case strings.Contains(lower, "opencode"):
		return "opencode"
	case strings.Contains(lower, "/.continue/"):
		return "continue"
	default:
		return "unknown"
	}
}

func parseMCPFile(writer utils.DataWriter, params mod.ModuleParams, path, tool string) error {
	data, err := readConfigFile(path)
	if err != nil {
		return err
	}

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".yaml", ".yml":
		return parseMCPYAML(writer, params, path, tool, data)
	default:
		return parseMCPJSON(writer, params, path, tool, data)
	}
}

func parseMCPJSON(writer utils.DataWriter, params mod.ModuleParams, path, tool string, data []byte) error {
	var root map[string]interface{}
	if err := json.Unmarshal(data, &root); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	servers := extractMCPServerMaps(root)
	if len(servers) == 0 {
		return nil
	}

	username := extractUsername(path)
	for name, serverMap := range servers {
		record := mcpRecordFromMap(tool, username, name, serverMap)
		if err := writeRecord(writer, params, path, record); err != nil {
			params.Logger.Debug("Failed to write MCP record: %v", err)
		}
	}

	return nil
}

func parseMCPYAML(writer utils.DataWriter, params mod.ModuleParams, path, tool string, data []byte) error {
	var root map[string]interface{}
	if err := yaml.Unmarshal(data, &root); err != nil {
		return fmt.Errorf("invalid YAML: %w", err)
	}

	servers := extractMCPServerMaps(root)
	if len(servers) == 0 {
		return nil
	}

	username := extractUsername(path)
	for name, serverMap := range servers {
		record := mcpRecordFromMap(tool, username, name, serverMap)
		if err := writeRecord(writer, params, path, record); err != nil {
			params.Logger.Debug("Failed to write MCP record: %v", err)
		}
	}

	return nil
}

func extractMCPServerMaps(root map[string]interface{}) map[string]map[string]interface{} {
	result := make(map[string]map[string]interface{})

	for _, key := range []string{"mcpServers", "mcp_servers", "servers"} {
		if raw, ok := root[key].(map[string]interface{}); ok {
			for name, val := range raw {
				if serverMap, ok := val.(map[string]interface{}); ok {
					result[name] = serverMap
				}
			}
		}
	}

	return result
}

func mcpRecordFromMap(tool, username, serverName string, server map[string]interface{}) map[string]interface{} {
	command := stringField(server, "command")
	args := stringSliceField(server, "args")
	url := stringField(server, "url")
	transport := stringField(server, "type")
	if transport == "" {
		transport = stringField(server, "transport")
	}
	if transport == "" {
		if url != "" {
			transport = "http"
		} else if command != "" {
			transport = "stdio"
		}
	}

	env := stringMapField(server["env"])
	record := map[string]interface{}{
		"artifact_type": "mcp",
		"tool":          tool,
		"server_name":   serverName,
		"command":       command,
		"args":          args,
		"url":           url,
		"transport":     transport,
		"env_keys":      envKeys(env),
		"username":      username,
	}

	if indicators := riskIndicators(command, args); len(indicators) > 0 {
		record["risk_indicators"] = indicators
	}

	return record
}

func collectMCPAuthEntries(writer utils.DataWriter, params mod.ModuleParams) error {
	paths := globPaths(mcpAuthPatterns)
	for _, path := range paths {
		if !isReadableFile(path) {
			continue
		}

		record := map[string]interface{}{
			"artifact_type": "mcp_auth",
			"tool":          "mcp",
			"username":      extractUsername(path),
			"file_name":     path[strings.LastIndex(path, "/")+1:],
		}

		if err := writeRecord(writer, params, path, record); err != nil {
			params.Logger.Debug("Failed to write MCP auth record: %v", err)
		}
	}

	return nil
}

func stringField(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func stringSliceField(m map[string]interface{}, key string) []string {
	raw, ok := m[key].([]interface{})
	if !ok {
		return nil
	}
	var out []string
	for _, item := range raw {
		if s, ok := item.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

func stringMapField(raw interface{}) map[string]string {
	m, ok := raw.(map[string]interface{})
	if !ok {
		return nil
	}
	out := make(map[string]string, len(m))
	for k, v := range m {
		if s, ok := v.(string); ok {
			out[k] = s
		}
	}
	return out
}
