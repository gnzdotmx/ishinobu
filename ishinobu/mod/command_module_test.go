package mod

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gnzdotmx/ishinobu/ishinobu/utils"
)

func TestCommandModule_Name(t *testing.T) {
	cmd := &CommandModule{
		ModuleName:  "test_module",
		Description: "Test module description",
		Command:     "echo",
		Args:        []string{"test"},
	}

	if name := cmd.Name(); name != "test_module" {
		t.Errorf("Expected module name 'test_module', got '%s'", name)
	}
}

func TestCommandModule_GetDescription(t *testing.T) {
	cmd := &CommandModule{
		ModuleName:  "test_module",
		Description: "Test module description",
		Command:     "echo",
		Args:        []string{"test"},
	}

	if desc := cmd.GetDescription(); desc != "Test module description" {
		t.Errorf("Expected description 'Test module description', got '%s'", desc)
	}
}

func setupTestDirs(t *testing.T) (string, string, string) {
	rootDir := "/tmp/ishinobu_test/"
	modDir := "./modules"
	outputDir := rootDir
	logsDir := "/"

	// Clean up any existing test directories
	os.RemoveAll(rootDir)

	// Create the root directory
	if err := os.MkdirAll(rootDir, 0755); err != nil {
		t.Fatalf("Failed to create directory %s: %v", rootDir, err)
	}

	return modDir, outputDir, logsDir
}

func cleanupTestDirs(t *testing.T) {
	os.RemoveAll("/tmp/ishinobu_test")
	// Remove ishinobu_*.log files in current directory where the test is running
	files, err := filepath.Glob("ishinobu_*.log")
	if err != nil {
		t.Fatalf("Failed to get ishinobu_*.log files: %v", err)
	}
	for _, file := range files {
		os.RemoveAll(file)
	}
}

func TestCommandModule_Run(t *testing.T) {
	modDir, outputDir, logsDir := setupTestDirs(t)
	defer cleanupTestDirs(t)

	logger := utils.NewLogger()
	defer logger.Close()

	cmd := &CommandModule{
		ModuleName:  "test_module",
		Description: "Test module description",
		Command:     "ps",
		Args:        []string{"aux"},
	}

	params := ModuleParams{
		ExportFormat:        "json",
		CollectionTimestamp: time.Now().UTC().Format(time.RFC3339),
		Logger:              *logger,
		InputDir:            modDir,
		LogsDir:             logsDir,
		OutputDir:           outputDir,
		Verbosity:           1,
		StartTime:           time.Now().Unix(),
	}

	err := cmd.Run(params)
	if err != nil {
		t.Errorf("CommandModule.Run() failed: %v", err)
	}

	expectedFile := filepath.Join(outputDir, "test_module.json")
	if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
		t.Errorf("Expected output file %s was not created", expectedFile)
	}
}

func TestCommandModule_RunWithInvalidCommand(t *testing.T) {
	modDir, outputDir, logsDir := setupTestDirs(t)
	defer cleanupTestDirs(t)

	logger := utils.NewLogger()
	defer logger.Close()

	cmd := &CommandModule{
		ModuleName:  "test_module",
		Description: "Test module description",
		Command:     "nonexistentcommand",
		Args:        []string{"test"},
	}

	params := ModuleParams{
		ExportFormat:        "json",
		CollectionTimestamp: time.Now().UTC().Format(time.RFC3339),
		Logger:              *logger,
		InputDir:            modDir,
		OutputDir:           outputDir,
		LogsDir:             logsDir,
		Verbosity:           1,
		StartTime:           time.Now().Unix(),
	}

	err := cmd.Run(params)
	if err == nil {
		t.Error("Expected error when running invalid command, got nil")
	}
}
