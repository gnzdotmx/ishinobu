// This module enumerates autostart locations for plist configuration files,
// parses them, and checks code signatures on programs that run on login/startup.
// It checks:
// - LaunchAgents and LaunchDaemons
// - Login items
// - Startup items
// - Scripting additions
// - Periodic tasks
// - Cron jobs
// - Sandboxed login items
package autostart

import (
	"os"
	"path/filepath"
	"strings"

	"howett.net/plist"

	"github.com/gnzdotmx/ishinobu/pkg/mod"
	"github.com/gnzdotmx/ishinobu/pkg/utils"
)

type AutostartModule struct {
	Name        string
	Description string
}

func init() {
	module := &AutostartModule{
		Name:        "autostart",
		Description: "Collects information about programs configured to run at startup"}
	mod.RegisterModule(module)
}

func (m *AutostartModule) GetName() string {
	return m.Name
}

func (m *AutostartModule) GetDescription() string {
	return m.Description
}

func (m *AutostartModule) Run(params mod.ModuleParams) error {
	// Parse each type of autostart location
	// Find all LaunchAgents and LaunchDaemons plists
	patternsLaunchItems := []string{
		"/System/Library/LaunchAgents/*.plist",
		"/Library/LaunchAgents/*.plist",
		"/Users/*/Library/LaunchAgents/*.plist",
		"/System/Library/LaunchDaemons/*.plist",
		"/Library/LaunchDaemons/*.plist",
	}
	outputFileNameLaunchItems := utils.GetOutputFileName(m.GetName()+"-launch-items", params.ExportFormat, params.OutputDir)
	writerLaunchItems, err := utils.NewDataWriter(params.LogsDir, outputFileNameLaunchItems, params.ExportFormat)
	if err != nil {
		return err
	}
	err = parseLaunchItems(patternsLaunchItems, writerLaunchItems, params)
	if err != nil {
		params.Logger.Debug("Error parsing launch items: %v", err)
	}

	// Find all login items plists
	patternLoginItems := "/Users/*/Library/Preferences/com.apple.loginitems.plist"
	outputFileNameLoginItems := utils.GetOutputFileName(m.GetName()+"-login-items", params.ExportFormat, params.OutputDir)
	writerLoginItems, err := utils.NewDataWriter(params.LogsDir, outputFileNameLoginItems, params.ExportFormat)
	if err != nil {
		return err
	}
	err = parseLoginItems(patternLoginItems, writerLoginItems, params)
	if err != nil {
		params.Logger.Debug("Error parsing login items: %v", err)
	}

	// Find all startup items plists
	patternsStartupItems := []string{
		"/System/Library/StartupItems/*/*",
		"/Library/StartupItems/*/*",
	}
	outputFileNameStartupItems := utils.GetOutputFileName(m.GetName()+"-startup-items", params.ExportFormat, params.OutputDir)
	writerStartupItems, err := utils.NewDataWriter(params.LogsDir, outputFileNameStartupItems, params.ExportFormat)
	if err != nil {
		return err
	}
	err = parseStartupItems(patternsStartupItems, writerStartupItems, params)
	if err != nil {
		params.Logger.Debug("Error parsing startup items: %v", err)
	}

	// Find all scripting additions
	patternsScriptingAdditions := []string{
		"/System/Library/ScriptingAdditions/*.osax",
		"/Library/ScriptingAdditions/*.osax",
	}
	outputFileNameScriptingAdditions := utils.GetOutputFileName(m.GetName()+"-scripting-additions", params.ExportFormat, params.OutputDir)
	writerScriptingAdditions, err := utils.NewDataWriter(params.LogsDir, outputFileNameScriptingAdditions, params.ExportFormat)
	if err != nil {
		return err
	}
	err = parseScriptingAdditions(patternsScriptingAdditions, writerScriptingAdditions, params)
	if err != nil {
		params.Logger.Debug("Error parsing scripting additions: %v", err)
	}

	// Find all periodic tasks
	patternsPeriodicTasks := []string{
		"/etc/periodic.conf",
		"/etc/periodic/*/*",
		"/etc/*.local",
		"/etc/rc.common",
		"/etc/emond.d/*",
		"/etc/emond.d/*/*",
	}
	outputFileNamePeriodicTasks := utils.GetOutputFileName(m.GetName()+"-periodic-tasks", params.ExportFormat, params.OutputDir)
	writerPeriodicTasks, err := utils.NewDataWriter(params.LogsDir, outputFileNamePeriodicTasks, params.ExportFormat)
	if err != nil {
		return err
	}
	err = parsePeriodicTasks(patternsPeriodicTasks, writerPeriodicTasks, params)
	if err != nil {
		params.Logger.Debug("Error parsing periodic tasks: %v", err)
	}

	// Find all cron jobs
	patternCronJobs := "/var/at/tabs/*"
	outputFileNameCronJobs := utils.GetOutputFileName(m.GetName()+"-cron-jobs", params.ExportFormat, params.OutputDir)
	writerCronJobs, err := utils.NewDataWriter(params.LogsDir, outputFileNameCronJobs, params.ExportFormat)
	if err != nil {
		return err
	}
	err = parseCronJobs(patternCronJobs, writerCronJobs, params)
	if err != nil {
		params.Logger.Debug("Error parsing cron jobs: %v", err)
	}

	// Find all sandboxed login items
	patternSandboxedLoginItems := "/var/db/com.apple.xpc.launchd/disabled.*.plist"
	outputFileNameSandboxedLoginItems := utils.GetOutputFileName(m.GetName()+"-sandboxed-login-items", params.ExportFormat, params.OutputDir)
	writerSandboxedLoginItems, err := utils.NewDataWriter(params.LogsDir, outputFileNameSandboxedLoginItems, params.ExportFormat)
	if err != nil {
		return err
	}
	err = parseSandboxedLoginItems(patternSandboxedLoginItems, writerSandboxedLoginItems, params)
	if err != nil {
		params.Logger.Debug("Error parsing sandboxed login items: %v", err)
	}

	return nil
}

func parseLaunchItems(patterns []string, writer utils.DataWriter, params mod.ModuleParams) error {
	for _, pattern := range patterns {
		files, err := filepath.Glob(pattern)
		if err != nil {
			continue
		}

		for _, file := range files {
			data, err := os.ReadFile(file)
			if err != nil {
				continue
			}

			var plistData map[string]interface{}
			_, err = plist.Unmarshal(data, &plistData)
			if err != nil {
				continue
			}

			recordData := make(map[string]interface{})
			recordData["src_file"] = file
			recordData["src_name"] = "launch_items"

			// Extract Label
			if label, ok := plistData["Label"].(string); ok {
				recordData["prog_name"] = label
			}

			// Extract Program or ProgramArguments
			if program, ok := plistData["Program"].(string); ok {
				recordData["program"] = program
				// Get code signature if available
				sig, err := utils.GetCodeSignature(program)
				if err == nil {
					recordData["code_signatures"] = sig
				}
			} else if args, ok := plistData["ProgramArguments"].([]interface{}); ok && len(args) > 0 {
				recordData["program"] = args[0]
				if len(args) > 1 {
					var argsList []string
					for _, arg := range args[1:] {
						if str, ok := arg.(string); ok {
							argsList = append(argsList, str)
						}
					}
					recordData["args"] = strings.Join(argsList, " ")
				}
			}

			record := utils.Record{
				CollectionTimestamp: params.CollectionTimestamp,
				EventTimestamp:      params.CollectionTimestamp,
				Data:                recordData,
				SourceFile:          file,
			}

			err = writer.WriteRecord(record)
			if err != nil {
				params.Logger.Debug("Failed to write record: %v", err)
			}
		}
	}

	return nil
}

func parseLoginItems(pattern string, writer utils.DataWriter, params mod.ModuleParams) error {
	files, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		var plistData []interface{}
		_, err = plist.Unmarshal(data, &plistData)
		if err != nil {
			continue
		}

		if len(plistData) > 0 {
			if sessionItems, ok := plistData[0].(map[string]interface{}); ok {
				if customItems, ok := sessionItems["SessionItems"].(map[string]interface{}); ok {
					if items, ok := customItems["CustomListItems"].([]interface{}); ok {
						for _, item := range items {
							if itemMap, ok := item.(map[string]interface{}); ok {
								recordData := make(map[string]interface{})
								recordData["src_file"] = file
								recordData["src_name"] = "login_items"

								if name, ok := itemMap["Name"].(string); ok {
									recordData["prog_name"] = name
								}

								// Handle both Alias and Bookmark paths
								if program, ok := itemMap["Program"].(string); ok {
									recordData["program"] = program
									sig, err := utils.GetCodeSignature(program)
									if err == nil {
										recordData["code_signatures"] = sig
									}
								}

								record := utils.Record{
									CollectionTimestamp: params.CollectionTimestamp,
									EventTimestamp:      params.CollectionTimestamp,
									Data:                recordData,
									SourceFile:          file,
								}

								err = writer.WriteRecord(record)
								if err != nil {
									params.Logger.Debug("Failed to write record: %v", err)
								}
							}
						}
					}
				}
			}
		}
	}

	return nil
}

func parseStartupItems(patterns []string, writer utils.DataWriter, params mod.ModuleParams) error {
	for _, pattern := range patterns {
		files, err := filepath.Glob(pattern)
		if err != nil {
			continue
		}

		for _, file := range files {
			recordData := make(map[string]interface{})
			recordData["src_file"] = file
			recordData["src_name"] = "startup_items"

			record := utils.Record{
				CollectionTimestamp: params.CollectionTimestamp,
				EventTimestamp:      params.CollectionTimestamp,
				Data:                recordData,
				SourceFile:          file,
			}

			err = writer.WriteRecord(record)
			if err != nil {
				params.Logger.Debug("Failed to write record: %v", err)
			}
		}
	}

	return nil
}

func parseScriptingAdditions(patterns []string, writer utils.DataWriter, params mod.ModuleParams) error {
	for _, pattern := range patterns {
		files, err := filepath.Glob(pattern)
		if err != nil {
			continue
		}

		for _, file := range files {
			recordData := make(map[string]interface{})
			recordData["src_file"] = file
			recordData["src_name"] = "scripting_additions"

			sig, err := utils.GetCodeSignature(file)
			if err == nil {
				recordData["code_signatures"] = sig
			}

			record := utils.Record{
				CollectionTimestamp: params.CollectionTimestamp,
				EventTimestamp:      params.CollectionTimestamp,
				Data:                recordData,
				SourceFile:          file,
			}

			err = writer.WriteRecord(record)
			if err != nil {
				params.Logger.Debug("Failed to write record: %v", err)
			}
		}
	}

	return nil
}

func parsePeriodicTasks(patterns []string, writer utils.DataWriter, params mod.ModuleParams) error {
	for _, pattern := range patterns {
		files, err := filepath.Glob(pattern)
		if err != nil {
			continue
		}

		for _, file := range files {
			recordData := make(map[string]interface{})
			recordData["src_file"] = file
			recordData["src_name"] = "periodic_rules_items"

			record := utils.Record{
				CollectionTimestamp: params.CollectionTimestamp,
				EventTimestamp:      params.CollectionTimestamp,
				Data:                recordData,
				SourceFile:          file,
			}

			err = writer.WriteRecord(record)
			if err != nil {
				params.Logger.Debug("Failed to write record: %v", err)
			}
		}
	}

	return nil
}

func parseCronJobs(pattern string, writer utils.DataWriter, params mod.ModuleParams) error {
	files, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if !strings.HasPrefix(line, "# ") && len(strings.TrimSpace(line)) > 0 {
				recordData := make(map[string]interface{})
				recordData["src_file"] = file
				recordData["src_name"] = "cron"
				recordData["program"] = line

				record := utils.Record{
					CollectionTimestamp: params.CollectionTimestamp,
					EventTimestamp:      params.CollectionTimestamp,
					Data:                recordData,
					SourceFile:          file,
				}

				err = writer.WriteRecord(record)
				if err != nil {
					params.Logger.Debug("Failed to write record: %v", err)
				}
			}
		}
	}

	return nil
}

func parseSandboxedLoginItems(pattern string, writer utils.DataWriter, params mod.ModuleParams) error {
	files, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		var plistData map[string]interface{}
		_, err = plist.Unmarshal(data, &plistData)
		if err != nil {
			continue
		}

		for key, value := range plistData {
			if boolVal, ok := value.(bool); ok && !boolVal {
				recordData := make(map[string]interface{})
				recordData["src_file"] = file
				recordData["src_name"] = "sandboxed_loginitems"
				recordData["prog_name"] = key

				record := utils.Record{
					CollectionTimestamp: params.CollectionTimestamp,
					EventTimestamp:      params.CollectionTimestamp,
					Data:                recordData,
					SourceFile:          file,
				}

				err = writer.WriteRecord(record)
				if err != nil {
					params.Logger.Debug("Failed to write record: %v", err)
				}
			}
		}
	}

	return nil
}
