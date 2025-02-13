package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gnzdotmx/ishinobu/ishinobu/mod"
	_ "github.com/gnzdotmx/ishinobu/ishinobu/modules"
	"github.com/gnzdotmx/ishinobu/ishinobu/utils"
	"github.com/spf13/cobra"
)

const (
	logsDir     = "./.logs"
	modInputDir = "./modules"
	outputDir   = "./"
)

var (
	// Command line flags
	modulesFlag  string
	exportFormat string
	parallelism  int
	verbosity    int

	// Root command
	rootCmd = &cobra.Command{
		Use:   "ishinobu",
		Short: "Ishinobu is a data collection tool",
		Long:  `A tool for collecting and exporting system data in various formats`,
		RunE:  run,
	}
)

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Define flags
	rootCmd.Flags().StringVarP(&modulesFlag, "modules", "m", "all", "Modules to run (comma-separated or 'all')")
	rootCmd.Flags().StringVarP(&exportFormat, "export", "e", "json", "Export format (json or csv)")
	rootCmd.Flags().IntVarP(&parallelism, "parallel", "p", 4, "Number of modules to run in parallel")
	rootCmd.Flags().IntVarP(&verbosity, "verbosity", "v", 1, "Verbosity level (0=Error, 1=Info, 2=Debug)")
}

func run(cmd *cobra.Command, args []string) error {
	// Initialize logger
	logger := utils.NewLogger()
	logger.SetVerbosity(verbosity)
	defer logger.Close()

	// Create a temporary folder to store log files
	if err := os.MkdirAll(logsDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directory %s: %v", logsDir, err)
	}

	// Get hostnames
	hostname, err := utils.GetHostname()
	if err != nil {
		return fmt.Errorf("failed to get hostname: %v", err)
	}

	// Collection timestamp
	collectionTimestamp := utils.Now()

	// Parse modules
	var selectedModules []string
	if modulesFlag == "all" {
		logger.Info("Running all modules")
		selectedModules = mod.AllModules()
	} else {
		logger.Info("Running selected modules: %s", modulesFlag)
		selectedModules = strings.Split(modulesFlag, ",")
	}

	fmt.Printf("Selected modules: %v\n", selectedModules)

	// Prepare module parameters
	params := mod.ModuleParams{
		ExportFormat:        exportFormat,
		CollectionTimestamp: collectionTimestamp,
		Logger:              *logger,
		LogsDir:             logsDir,
		InputDir:            modInputDir,
		OutputDir:           outputDir,
		Verbosity:           verbosity,
	}

	// Run modules
	var wg sync.WaitGroup
	sem := make(chan struct{}, parallelism)

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
	return nil
}
