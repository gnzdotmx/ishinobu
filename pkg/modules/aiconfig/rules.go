package aiconfig

import (
	"path/filepath"
	"strings"

	"github.com/gnzdotmx/ishinobu/pkg/mod"
	"github.com/gnzdotmx/ishinobu/pkg/utils"
)

func collectRules(writer utils.DataWriter, params mod.ModuleParams) error {
	files := globRuleFiles(rulesGlobPatterns)

	for _, path := range files {
		tool, ruleType := rulesToolFromPath(path)
		projectPath := extractProjectPath(path, string(filepath.Separator)+ruleTypeMarker(ruleType)+string(filepath.Separator))
		if projectPath == "" {
			projectPath = filepath.Dir(path)
		}

		record := map[string]interface{}{
			"artifact_type": "rule",
			"tool":          tool,
			"rule_type":     ruleType,
			"file_name":     filepath.Base(path),
			"username":      extractUsername(path),
			"project_path":  projectPath,
		}

		if err := writeRecord(writer, params, path, record); err != nil {
			params.Logger.Debug("Failed to write rule record: %v", err)
		}
	}

	return nil
}

func ruleTypeMarker(ruleType string) string {
	switch ruleType {
	case "cursor_rules":
		return ".cursor"
	case "claude_rules":
		return ".claude"
	case "windsurf_rules":
		return ".windsurf"
	case "cline_rules":
		return ".clinerules"
	case "copilot_instructions":
		return ".github"
	case "amazonq_rules":
		return ".amazonq"
	case "kiro_steering":
		return ".kiro"
	default:
		return ""
	}
}

func rulesToolFromPath(path string) (tool, ruleType string) {
	lower := strings.ToLower(path)
	base := strings.ToLower(filepath.Base(path))

	switch {
	case strings.Contains(lower, "/.cursor/rules/"):
		return "cursor", "cursor_rules"
	case strings.Contains(lower, "/.claude/") && base == "claude.md":
		return "claude_code", "claude_rules"
	case strings.Contains(lower, "/.gemini/") && base == "gemini.md":
		return "gemini_cli", "gemini_rules"
	case strings.Contains(lower, "/.windsurf/rules/"):
		return "windsurf", "windsurf_rules"
	case strings.Contains(lower, "/.clinerules/"):
		return "cline", "cline_rules"
	case strings.Contains(lower, "/.github/copilot-instructions.md"):
		return "github_copilot", "copilot_instructions"
	case base == "agents.md":
		return "codex_cli", "agents_md"
	case base == "conventions.md":
		return "aider", "conventions"
	case strings.Contains(lower, "/.amazonq/rules/"):
		return "amazon_q", "amazonq_rules"
	case strings.Contains(lower, "/.kiro/steering/"):
		return "kiro", "kiro_steering"
	case base == ".goosehints":
		return "goose", "goose_hints"
	case base == ".rules":
		return "zed", "zed_rules"
	default:
		return "unknown", "unknown"
	}
}
