// This module collects AI coding tool configuration artifacts relevant to
// incident response, including MCP server definitions, agent lifecycle hooks,
// and project-level instruction files.
//
// It collects:
// - MCP server configurations (command, args, transport; env var names only)
// - Agent hooks and VS Code tasks used as persistence vectors
// - AI instruction and rules files (metadata and hashes, not file contents)
package aiconfig

import (
	"fmt"

	"github.com/gnzdotmx/ishinobu/pkg/mod"
	"github.com/gnzdotmx/ishinobu/pkg/utils"
)

type AIConfigModule struct {
	Name        string
	Description string
}

func init() {
	module := &AIConfigModule{
		Name:        "aiconfig",
		Description: "Collects AI tool MCP configs, agent hooks, and instruction files (Claude Desktop/Code, Cursor, Windsurf, Cline, Gemini CLI, Codex CLI, OpenCode, Continue, VS Code, GitHub Copilot, Aider, Amazon Q, Kiro, Goose, Zed)",
	}
	mod.RegisterModule(module)
}

func (m *AIConfigModule) GetName() string {
	return m.Name
}

func (m *AIConfigModule) GetDescription() string {
	return m.Description
}

func (m *AIConfigModule) Run(params mod.ModuleParams) error {
	collectors := []struct {
		suffix string
		fn     func(utils.DataWriter, mod.ModuleParams) error
	}{
		{"mcp", collectMCPConfigs},
		{"hooks", collectHooks},
		{"rules", collectRules},
	}

	for _, collector := range collectors {
		outputFileName := utils.GetOutputFileName(m.GetName()+"-"+collector.suffix, params.ExportFormat, params.OutputDir)
		writer, err := utils.NewDataWriter(params.LogsDir, outputFileName, params.ExportFormat)
		if err != nil {
			return fmt.Errorf("failed to create data writer for %s: %w", collector.suffix, err)
		}

		if err := collector.fn(writer, params); err != nil {
			params.Logger.Debug("Error collecting %s artifacts: %v", collector.suffix, err)
		}

		if err := writer.Close(); err != nil {
			params.Logger.Debug("Error closing writer for %s: %v", collector.suffix, err)
		}
	}

	return nil
}
