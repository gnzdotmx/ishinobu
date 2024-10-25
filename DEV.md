# For Developers

## Project Structure

The project is structured as follows:

```
ishinobu/
├── cmd/
│   └── cmd.go
├── modules/
│   ├── ps.go
│   ├── unifiedlogs.go
│   └── nettop.go
├── utils/
│   └── logger.go
└── main.go
```

- The `cmd` package contains the main entry point for the application. 
- The `modules` package contains the modules that can be run by the application. 
- The `utils` package contains utility functions used by the application. 
- The `main.go` file is the main entry point for the application.






## How to run a module
```go
package main

import (
	"github.com/gnzdotmx/ishinobu/ishinobu/mod"
	"github.com/gnzdotmx/ishinobu/ishinobu/modules"
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
## CommandModule: A helper to run shell commands
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

## How to write a module
1. Create a new file in the `modules` directory.
2. Implement a struct that represents the module.
3. Implement the `GetName` and `GetDescription` methods.
4. Implement the `Run` method.
Example:
```go
package modules

import (
	"fmt"

	"github.com/gnzdotmx/ishinobu/ishinobu/mod"
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