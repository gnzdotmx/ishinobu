// This module is useful to investigate open files, network connections, and processes that opened them.
// Command: lsof -n -P
// It collects the following information:
// - PID
// - User
// - Command
// - File
// - File descriptor
// - File type
package modules

import (
	"fmt"

	"github.com/gnzdotmx/ishinobu/ishinobu/mod"
)

type LsofModule struct {
	Name        string
	Description string
}

func init() {
	module := &LsofModule{
		Name:        "lsof",
		Description: "Collects information about open files and their processes",
	}
	mod.RegisterModule(module)
}

func (m *LsofModule) GetName() string {
	return m.Name
}

func (m *LsofModule) GetDescription() string {
	return m.Description
}

func (m *LsofModule) Run(params mod.ModuleParams) error {
	cmdMod := &mod.CommandModule{
		ModuleName:  m.GetName(),
		Description: m.GetDescription(),
		Command:     "lsof",
		Args:        []string{"-n", "-P"}, // -n: no DNS lookups, -P: no port name resolution
	}

	err := cmdMod.Run(params)
	if err != nil {
		return fmt.Errorf("error running command: %v", err)
	}

	return nil
}
