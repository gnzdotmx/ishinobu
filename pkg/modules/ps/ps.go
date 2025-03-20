// The module is useful to investigate the list of running processes and their details.
// Command: ps aux
// It collects the following information:
// - PID
// - User
// - Command
// - CPU usage
// - Memory usage
// - Status
package ps

import (
	"fmt"

	"github.com/gnzdotmx/ishinobu/pkg/mod"
)

type ProcessListModule struct {
	Name        string
	Description string
}

func init() {
	module := &ProcessListModule{Name: "ps", Description: "Collects the list of running processes"}
	mod.RegisterModule(module)
}

func (m *ProcessListModule) GetName() string {
	return m.Name
}

func (m *ProcessListModule) GetDescription() string {
	return m.Description
}

func (m *ProcessListModule) Run(params mod.ModuleParams) error {
	cmdMod := &mod.CommandModule{
		ModuleName:  m.GetName(),
		Description: m.GetDescription(),
		Command:     "ps",
		Args:        []string{"aux"},
	}

	err := cmdMod.Run(params)
	if err != nil {
		params.Logger.Debug("error running command: %v", err)
		return fmt.Errorf("error running command: %v", err)
	}

	return nil
}
