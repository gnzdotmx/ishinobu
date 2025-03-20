package mod

import (
	"fmt"

	"github.com/gnzdotmx/ishinobu/pkg/utils"
)

var moduleRegistry = make(map[string]Module)

type Module interface {
	GetName() string // Name of the module
	Run(params ModuleParams) error
}

type ModuleParams struct {
	ExportFormat        string
	CollectionTimestamp string
	Logger              utils.Logger
	LogsDir             string
	InputDir            string
	OutputDir           string
	Verbosity           int
	StartTime           int64 // Add start time tracking
}

func RegisterModule(module Module) {
	moduleRegistry[module.GetName()] = module
}

func AllModules() []string {
	keys := make([]string, 0, len(moduleRegistry))
	for k := range moduleRegistry {
		keys = append(keys, k)
	}
	return keys
}

func RunModule(name string, params ModuleParams) error {
	module, exists := moduleRegistry[name]
	if !exists {
		return fmt.Errorf("module %s not found", name)
	}
	return module.Run(params)
}
