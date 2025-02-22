package mod

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gnzdotmx/ishinobu/ishinobu/utils"
)

// Mock module for testing
type MockModule struct {
	name      string
	runCount  int
	shouldErr bool
}

func (m *MockModule) GetName() string {
	return m.name
}

func (m *MockModule) Run(params ModuleParams) error {
	m.runCount++
	if m.shouldErr {
		return errors.New("mock error")
	}
	return nil
}

func TestRegisterModule(t *testing.T) {
	// Clear the registry before testing
	moduleRegistry = make(map[string]Module)

	mock := &MockModule{name: "test_module"}
	RegisterModule(mock)

	if len(moduleRegistry) != 1 {
		t.Errorf("Expected registry length 1, got %d", len(moduleRegistry))
	}

	if _, exists := moduleRegistry["test_module"]; !exists {
		t.Error("Module was not registered correctly")
	}
}

func TestAllModules(t *testing.T) {
	// Clear the registry and add test modules
	moduleRegistry = make(map[string]Module)

	modules := []string{"module1", "module2", "module3"}
	for _, name := range modules {
		RegisterModule(&MockModule{name: name})
	}

	registered := AllModules()
	if len(registered) != len(modules) {
		t.Errorf("Expected %d modules, got %d", len(modules), len(registered))
	}

	// Check if all modules are present
	for _, name := range modules {
		found := false
		for _, regName := range registered {
			if regName == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Module %s not found in registry", name)
		}
	}
}

func TestRunModule(t *testing.T) {
	// Clear the registry before testing
	moduleRegistry = make(map[string]Module)

	// Test cases
	tests := []struct {
		name      string
		module    *MockModule
		shouldErr bool
	}{
		{
			name:      "successful_run",
			module:    &MockModule{name: "success_module", shouldErr: false},
			shouldErr: false,
		},
		{
			name:      "error_run",
			module:    &MockModule{name: "error_module", shouldErr: true},
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Register the test module
			RegisterModule(tt.module)

			// Create test parameters
			params := ModuleParams{
				ExportFormat:        "json",
				CollectionTimestamp: time.Now().UTC().Format(time.RFC3339),
				Logger:              *utils.NewLogger(),
				LogsDir:             "/tmp/logs",
				OutputDir:           "/tmp/output",
				InputDir:            "/tmp/input",
				Verbosity:           1,
				StartTime:           time.Now().Unix(),
			}

			// Run the module
			err := RunModule(tt.module.GetName(), params)

			// Check error condition
			if tt.shouldErr && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			// Check if module was actually run
			if tt.module.runCount != 1 {
				t.Errorf("Expected module to run once, but ran %d times", tt.module.runCount)
			}

			RemoveIshinobuLogs(t)
		})
	}
}

func TestRunModule_NonexistentModule(t *testing.T) {
	// Clear the registry
	moduleRegistry = make(map[string]Module)

	params := ModuleParams{
		ExportFormat:        "json",
		CollectionTimestamp: time.Now().UTC().Format(time.RFC3339),
		Logger:              *utils.NewLogger(),
		LogsDir:             "/tmp/logs",
		OutputDir:           "/tmp/output",
		Verbosity:           1,
		StartTime:           time.Now().Unix(),
	}

	err := RunModule("nonexistent_module", params)
	if err == nil {
		t.Error("Expected error when running nonexistent module, got nil")
	}

	RemoveIshinobuLogs(t)
}

func RemoveIshinobuLogs(t *testing.T) {
	// Remove ishinobu_*.log files in current directory where the test is running
	files, err := filepath.Glob("ishinobu_*.log")
	if err != nil {
		t.Fatalf("Failed to get ishinobu_*.log files: %v", err)
	}
	for _, file := range files {
		os.RemoveAll(file)
	}
}
