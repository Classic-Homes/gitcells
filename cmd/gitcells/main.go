package main

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	version   = "dev"
	buildTime = "unknown"
	logger    *logrus.Logger
)

func main() {
	logger = setupLogger()

	rootCmd := &cobra.Command{
		Use:     "gitcells",
		Short:   "Version control for Excel files",
		Long:    `GitCells converts Excel files to JSON for version control and collaboration`,
		Version: fmt.Sprintf("%s (built %s)", version, buildTime),
	}

	// Add commands
	rootCmd.AddCommand(
		newInitCommand(logger),
		newWatchCommand(logger),
		newSyncCommand(logger),
		newConvertCommand(logger),
		newStatusCommand(logger),
		newDiffCommand(logger),
	)

	// Global flags
	rootCmd.PersistentFlags().String("config", "", "config file (default: .gitcells.yaml)")
	rootCmd.PersistentFlags().Bool("verbose", false, "enable verbose logging")

	// Handle verbose flag
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		verbose, _ := cmd.Flags().GetBool("verbose")
		if verbose {
			logger.SetLevel(logrus.DebugLevel)
		}
	}

	if err := rootCmd.Execute(); err != nil {
		logger.Error(err)
		os.Exit(1)
	}
}

func setupLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		DisableColors: false,
	})
	return logger
}
