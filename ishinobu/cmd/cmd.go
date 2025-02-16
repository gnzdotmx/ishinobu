package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

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
	timeout      int
)

// Root command
var rootCmd = &cobra.Command{
	Use:   "ishinobu",
	Short: "Ishinobu is a data collection tool",
	Long:  `A tool for collecting and exporting system data in various formats`,
	RunE:  run,
}

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
	rootCmd.Flags().IntVarP(&timeout, "timeout", "t", 0, "Timeout in seconds for each module (0 = no timeout)")
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

	// Track total execution time
	startTime := time.Now()

	// Create context for timeout
	ctx := context.Background()
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
		defer cancel()
	}

	// Create a map to track module status
	type moduleStatus struct {
		startTime time.Time
		status    string // "running", "completed", "failed", "timeout"
		elapsed   time.Duration
	}
	statusMap := make(map[string]*moduleStatus)
	var statusMutex sync.Mutex

	// Function to print status
	printStatus := func() {
		statusMutex.Lock()
		defer statusMutex.Unlock()

		// Move cursor to the start of the status display
		fmt.Print("\033[0;0H") // Move to top-left
		fmt.Print("\033[2J")   // Clear screen

		fmt.Println("Module Status:")
		fmt.Println("==============")

		for _, moduleName := range selectedModules {
			status := statusMap[moduleName]
			if status == nil {
				fmt.Printf("  %s: \033[33m●\033[0m Waiting\n", moduleName)
				continue
			}

			// Determine icon and color
			var icon, color string
			switch status.status {
			case "running":
				icon = "●"
				color = "\033[33m" // Yellow
			case "completed":
				icon = "✓"
				color = "\033[32m" // Green
			case "failed", "timeout":
				icon = "✗"
				color = "\033[31m" // Red
			}

			fmt.Printf("  %s: %s%s\033[0m %s (%.1f/%.1fs)\n",
				moduleName,
				color,
				icon,
				status.status,
				status.elapsed.Seconds(),
				float64(timeout),
			)
		}
	}

	// Start status printer
	stopPrinter := make(chan struct{})
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				printStatus()
			case <-stopPrinter:
				return
			}
		}
	}()

	// Run modules
	var wg sync.WaitGroup
	sem := make(chan struct{}, parallelism)

	for _, moduleName := range selectedModules {
		wg.Add(1)
		sem <- struct{}{}

		go func(moduleName string) {
			defer wg.Done()

			// Initialize status
			statusMutex.Lock()
			statusMap[moduleName] = &moduleStatus{
				startTime: time.Now(),
				status:    "running",
			}
			statusMutex.Unlock()

			// Create channel for module execution
			done := make(chan error, 1)
			go func() {
				done <- mod.RunModule(moduleName, params)
			}()

			// Wait for module to finish or timeout
			select {
			case err := <-done:
				statusMutex.Lock()
				status := statusMap[moduleName]
				status.elapsed = time.Since(status.startTime)
				if err != nil {
					status.status = "failed"
				} else {
					status.status = "completed"
				}
				statusMutex.Unlock()
			case <-ctx.Done():
				statusMutex.Lock()
				status := statusMap[moduleName]
				status.elapsed = time.Since(status.startTime)
				status.status = "timeout"
				statusMutex.Unlock()
			}

			<-sem
		}(moduleName)
	}

	wg.Wait()
	close(stopPrinter)
	printStatus() // Final status update

	totalElapsed := time.Since(startTime)
	logger.Info("🏁 All modules completed in %s", totalElapsed)

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
