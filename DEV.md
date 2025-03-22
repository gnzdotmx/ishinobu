# 👨‍💻 For Developers

## 📁 Project Structure

The project is structured as follows:

```
ishinobu/
├── cmd
│   ├── ishinobu
│   └── ...
├── pkg
│   ├── bundles
│   │   ├── full
│   │   └── ...
│   ├── cmd
│   ├── mod
│   ├── modules
│   │   ├── appstore
│   │   ├── asl
│   │   ├── auditlog
│   │   └── ...
│   └── utils
```

- The `pkg/bundles` contains packages pre-bundling modules for easy import in application entry points. 
- The `pkg/cmd` package contains the root command for the application. 
- The `pkg/modules` package contains the modules that can be run by the application. 
- The `pkg/utils` package contains utility functions used by the application. 
- The `cmd` package contains various application main entry points with pre-bundled modules. Use `cmd/ishinobu` for an executable with all native modules.

## ▶️ How to run a module
```go
package main

import (
	"github.com/gnzdotmx/ishinobu/pkg/mod"
	"github.com/gnzdotmx/ishinobu/pkg/modules/mymodule"
)

func main() {
	params := mod.ModuleParams{
		ExportFormat:        "json",
		CollectionTimestamp: "2021-01-01T00:00:00Z",
		Logger:              mod.NewLogger(),
	}

	modules.RunModule("mymodule", params)
}
```
## 💻 CommandModule: A helper to run shell commands
The `CommandModule` is a helper module that allows you to run shell commands and capture their output to specified by the user.
1. Create a new `CommandModule` instance.
2. Set the `ModuleName`, `Description`, `Command`, and `Args` fields.
3. Call the `Run` method.
Example:
```go
cmdMod := &mod.CommandModule{
	ModuleName:  "mymodule",
	Description: "This is my module",
	Command:     "ls",
	Args:        []string{"-l"},
}
cmdMod.Run(params)
```

## ✏️ How to write a module
1. Create a new package in the `modules` directory.
2. Implement a struct that represents the module.
3. Implement the `GetName` and `GetDescription` methods.
4. Implement the `Run` method.
Example:
```go
package mymodule

import (
	"fmt"

	"github.com/gnzdotmx/ishinobu/pkg/mod"
)

type MyModule struct {
	Name        string
	Description string
}

func init() {
	module := &MyModule{
		Name:        "mymodule",
		Description: "This is my module",
	}
	mod.RegisterModule(module)
}

func (m *MyModule) GetName() string {
	return m.Name
}

func (m *MyModule) GetDescription() string {
	return m.Description
}

func (m *MyModule) Run(params mod.ModuleParams) error {
	fmt.Println("Running my module")
	headers := []string{"field1", "field2", "field3"}
	data := [][]string{{"value1", "value2", "value3"}}
 
 	// Prepare the output file
	outputFileName := utils.GetOutputFileName(m.GetName(), params.ExportFormat, params.OutputDir)
	writer, err := utils.NewDataWriter(params.LogsDir, outputFileName, params.ExportFormat)
	if err != nil {
   		return fmt.Errorf("failed to create data writer: %v", err)
 	}
 	defer writer.Close()
 
	// Create a record per data row
	for _, row := range data {
	recordData := make(map[string]string)
	for i, field := range headers {
		recordData[field] = row[i]
	}
	
	// Prepare the record. 
	// Do not forget to specify the event timestamp if exists. Otherwise, set to the current time.
	// Set a source to identify the module that generated the record.
	record := utils.Record{
		CollectionTimestamp: params.CollectionTimestamp,
		EventTimestamp:      utils.Now(),
		Data:                data,
		Source:              "mymodule",
	}
	// Write the record
	err := writer.Write(record)
	if err != nil {
		params.Logger.Debug("failed to write record: %v", err)
		return fmt.Errorf("failed to write record: %v", err)
	}
	}
	return nil
}
```

## 🧪 Unit Testing Modules

Each module should have corresponding unit tests to ensure it functions correctly. Tests are located in files named `modulename_test.go` alongside the module implementation.

### 📝 How to Write a Module Test

1. Create a test file named after the module (e.g., `users_test.go` for `users.go`)
2. Test the basic module functions (`GetName`, `GetDescription`)
3. Test the module initialization
4. Test the module's `Run` method by mocking output and verifying results

### 🏗️ Test Structure

A typical module test file contains:

1. **TestModuleBasics** - Tests for `GetName` and `GetDescription`
2. **TestModuleInitialization** - Verifies proper module initialization
3. **TestModuleRun** - Tests the main functionality by creating mock output
4. **Helper functions** - For creating mock data and verifying output

### 📋 Example Test Structure

```go
func TestUsersModule(t *testing.T) {
    // Test GetName, GetDescription and Run method
}

func TestUsersModuleInitialization(t *testing.T) {
    // Test proper module initialization
}

func createMockUsersOutput(t *testing.T, params mod.ModuleParams) {
    // Create mock output for testing
}

func verifyUsersOutput(t *testing.T, outputFile string) {
    // Verify the output file contains expected data
}
```

### 🔄 Mocking Module Output

Since many modules interact with the system (executing commands, reading files), tests should mock these operations to avoid dependencies on the test environment:

```go
// Create a mock output file
outputFile := filepath.Join(params.OutputDir, "modulename-"+params.CollectionTimestamp+".json")

// Create sample records
records := []utils.Record{
    // Sample records here
}

// Write records to output file
file, err := os.Create(outputFile)
encoder := json.NewEncoder(file)
for _, record := range records {
    encoder.Encode(record)
}
```

### ✅ Verifying Module Output

Tests should verify that the module output contains the expected data:

```go
// Read and parse the output file
content, err := os.ReadFile(outputFile)
lines := testutils.SplitLines(content)

// Verify the expected content is present
for _, line := range lines {
    var record map[string]interface{}
    json.Unmarshal(line, &record)
    
    // Verify record fields
    assert.NotEmpty(t, record["collection_timestamp"])
    assert.NotEmpty(t, record["event_timestamp"])
    // Check specific data fields
}
```

### 🏃 Running Tests

Run all module tests:
```bash
go test -v ./pkg/modules/...
```

Run a specific module test:
```bash
go test -v ./pkg/modules/users/
```

### 💡 Best Practices

1. **Clean up test files**: Use `defer os.RemoveAll(tmpDir)` to clean up temporary files
2. **Test multiple scenarios**: For modules that handle different types of data, test all scenarios
3. **Mock external dependencies**: Don't rely on actual system commands or files in tests

### 🛠️ Common Test Helper Functions

Common test helper functions are available in the `pkg/modules/testutils` package:

- **WriteTestRecords**: Writes test records to an output file
- **SplitLines**: Parses output files into lines for verification
