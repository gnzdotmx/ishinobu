// This module is useful to investigate the list of current network connections and their details.
// Command: netstat -anv
// It collects the following information:
// - Protocol
// - Local Address
// - Foreign Address
// - State
// - PID
// - Program name
package netstat

import (
	"fmt"

	"github.com/gnzdotmx/ishinobu/pkg/mod"
)

type NetstatModule struct {
	Name        string
	Description string
}

func init() {
	module := &NetstatModule{
		Name:        "netstat",
		Description: "Collects and parses netstat output"}
	mod.RegisterModule(module)
}

func (m *NetstatModule) GetName() string {
	return m.Name
}

func (m *NetstatModule) GetDescription() string {
	return m.Description
}

func (m *NetstatModule) Run(params mod.ModuleParams) error {
	cmdMod := &mod.CommandModule{
		ModuleName:  m.GetName(),
		Description: m.GetDescription(),
		Command:     "netstat",
		Args:        []string{"-anv"},
	}

	err := cmdMod.Run(params)
	if err != nil {
		return fmt.Errorf("error running command: %w", err)
	}

	return nil
}
