package aiconfig

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/gnzdotmx/ishinobu/pkg/mod"
	"github.com/gnzdotmx/ishinobu/pkg/utils"
)

func collectHooks(writer utils.DataWriter, params mod.ModuleParams) error {
	paths := globPaths(hooksGlobPatterns)

	for _, path := range paths {
		tool := hooksToolFromPath(path)
		if strings.HasSuffix(path, "tasks.json") {
			if err := parseVSCodeTasks(writer, params, path, tool); err != nil {
				params.Logger.Debug("Skipping tasks config %s: %v", path, err)
			}
			continue
		}

		if err := parseHooksFile(writer, params, path, tool); err != nil {
			params.Logger.Debug("Skipping hooks config %s: %v", path, err)
		}
	}

	return nil
}

func hooksToolFromPath(path string) string {
	lower := strings.ToLower(path)
	switch {
	case strings.Contains(lower, "/.claude/"):
		return "claude_code"
	case strings.Contains(lower, "/.gemini/"):
		return "gemini_cli"
	case strings.Contains(lower, "/.codex/"):
		return "codex_cli"
	case strings.Contains(lower, "/.cline/") || strings.Contains(lower, "saoudrizwan.claude-dev"):
		return "cline"
	case strings.Contains(lower, "/.vscode/"):
		return "vscode"
	default:
		return "unknown"
	}
}

func parseHooksFile(writer utils.DataWriter, params mod.ModuleParams, path, tool string) error {
	data, err := readConfigFile(path)
	if err != nil {
		return err
	}

	var root map[string]interface{}
	if err := json.Unmarshal(data, &root); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	hooksRoot, ok := root["hooks"].(map[string]interface{})
	if !ok {
		return nil
	}

	username := extractUsername(path)
	projectPath := extractProjectPath(path, string(os.PathSeparator)+".claude"+string(os.PathSeparator))
	if projectPath == "" {
		projectPath = extractProjectPath(path, string(os.PathSeparator)+".gemini"+string(os.PathSeparator))
	}

	for eventName, eventHooks := range hooksRoot {
		hookGroups, ok := eventHooks.([]interface{})
		if !ok {
			continue
		}

		for _, group := range hookGroups {
			groupMap, ok := group.(map[string]interface{})
			if !ok {
				continue
			}
			matcher := stringField(groupMap, "matcher")
			innerHooks, ok := groupMap["hooks"].([]interface{})
			if !ok {
				continue
			}

			for _, hook := range innerHooks {
				hookMap, ok := hook.(map[string]interface{})
				if !ok {
					continue
				}
				command := stringField(hookMap, "command")
				if command == "" {
					continue
				}

				record := map[string]interface{}{
					"artifact_type": "hook",
					"tool":          tool,
					"hook_event":    eventName,
					"hook_name":     stringField(hookMap, "name"),
					"hook_type":     stringField(hookMap, "type"),
					"matcher":       matcher,
					"command":       command,
					"username":      username,
					"project_path":  projectPath,
				}

				if desc := stringField(hookMap, "description"); desc != "" {
					record["description"] = truncateString(desc, 500)
				}

				if indicators := riskIndicators(command, nil); len(indicators) > 0 {
					record["risk_indicators"] = indicators
				}

				if err := writeRecord(writer, params, path, record); err != nil {
					params.Logger.Debug("Failed to write hook record: %v", err)
				}
			}
		}
	}

	return nil
}

func parseVSCodeTasks(writer utils.DataWriter, params mod.ModuleParams, path, tool string) error {
	data, err := readConfigFile(path)
	if err != nil {
		return err
	}

	var root map[string]interface{}
	if err := json.Unmarshal(data, &root); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	tasks, ok := root["tasks"].([]interface{})
	if !ok {
		return nil
	}

	username := extractUsername(path)
	projectPath := extractProjectPath(path, string(os.PathSeparator)+".vscode"+string(os.PathSeparator))

	for _, task := range tasks {
		taskMap, ok := task.(map[string]interface{})
		if !ok {
			continue
		}

		command := stringField(taskMap, "command")
		if command == "" {
			continue
		}

		args := stringSliceField(taskMap, "args")
		record := map[string]interface{}{
			"artifact_type": "hook",
			"tool":          tool,
			"hook_event":    "task",
			"hook_name":     stringField(taskMap, "label"),
			"hook_type":     stringField(taskMap, "type"),
			"command":       command,
			"args":          args,
			"username":      username,
			"project_path":  projectPath,
		}

		if indicators := riskIndicators(command, args); len(indicators) > 0 {
			record["risk_indicators"] = indicators
		}

		if err := writeRecord(writer, params, path, record); err != nil {
			params.Logger.Debug("Failed to write VS Code task record: %v", err)
		}
	}

	return nil
}
