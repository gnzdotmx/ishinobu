package cmd

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gnzdotmx/ishinobu/ishinobu/mod"
	_ "github.com/gnzdotmx/ishinobu/ishinobu/modules"
	"github.com/gnzdotmx/ishinobu/ishinobu/utils"
)

const (
	logsDir     = "./.logs"
	modInputDir = "./modules"
	outputDir   = "./"
)

func Execute() {
	// Command-line flags
	modulesFlag := flag.String("m", "all", "Modules to run (comma-separated or 'all')")
	exportFormat := flag.String("e", "json", "Export format (json or csv)")
	parallelism := flag.Int("p", 4, "Number of modules to run in parallel")
	verbosity := flag.Int("v", 1, "Verbosity level (0=Error, 1=Info, 2=Debug)")
	flag.Parse()

	// Initialize logger
	logger := utils.NewLogger()
	logger.SetVerbosity(*verbosity)
	defer logger.Close()

	// Create a temporary folder to store log files
	if err := os.MkdirAll(logsDir, os.ModePerm); err != nil {
		logger.Error("Failed to create directory %s: %v", logsDir, err)
	}

	// Get hostnames
	hostname, err := utils.GetHostname()
	if err != nil {
		logger.Error("Failed to get hostname: %v", err)
		return
	}

	// Collection timestamp
	collectionTimestamp := utils.Now()

	// Parse modules
	var selectedModules []string
	if *modulesFlag == "all" {
		logger.Info("Running all modules")
		selectedModules = mod.AllModules()
	} else {
		logger.Info("Running selected modules: %s", *modulesFlag)
		selectedModules = strings.Split(*modulesFlag, ",")
	}

	fmt.Printf("Selected modules: %v\n", selectedModules)

	// Prepare module parameters
	params := mod.ModuleParams{
		ExportFormat:        *exportFormat,
		CollectionTimestamp: collectionTimestamp,
		Logger:              *logger,
		LogsDir:             logsDir,
		InputDir:            modInputDir,
		OutputDir:           outputDir,
		Verbosity:           *verbosity,
	}

	// Run modules
	var wg sync.WaitGroup
	sem := make(chan struct{}, *parallelism)

	for _, moduleName := range selectedModules {
		wg.Add(1)
		sem <- struct{}{}

		go func(moduleName string) {
			defer wg.Done()
			logger.Info("Starting module: %s", moduleName)

			err := mod.RunModule(moduleName, params)
			if err != nil {
				logger.Error("Module %s failed: %v", moduleName, err)
			} else {
				logger.Info("Module %s completed", moduleName)
			}

			<-sem
		}(moduleName)
	}

	wg.Wait()

	// Compress output
	outputName := fmt.Sprintf("%s.%s.tar.gz", hostname, collectionTimestamp)
	outputFilename := filepath.Join(outputDir, outputName)
	err = utils.CompressOutput(logsDir, outputFilename)
	if err != nil {
		logger.Error("Failed to compress output: %v", err)
	} else {
		logger.Info("Output compressed to %s", outputName)
	}

	// Remove temporary folder to store collected logs
	err = os.RemoveAll(logsDir)
	if err != nil {
		logger.Error("Error removing %s. %v", logsDir, err)
	}

	logger.Info("Data collection completed")
}
