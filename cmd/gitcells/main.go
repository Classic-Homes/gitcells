package main

import (
	"fmt"
	"os"

	"github.com/Classic-Homes/gitcells/internal/constants"
	"github.com/Classic-Homes/gitcells/internal/tui"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	logger *logrus.Logger
)

func main() {
	logger = setupLogger()

	rootCmd := &cobra.Command{
		Use:     "gitcells",
		Short:   "Version control for Excel files",
		Long:    `GitCells converts Excel files to JSON for version control and collaboration`,
		Version: fmt.Sprintf("%s (built %s)", constants.Version, constants.BuildTime),
	}

	// Add commands
	rootCmd.AddCommand(
		newInitCommand(logger),
		newWatchCommand(logger),
		newSyncCommand(logger),
		newConvertCommand(logger),
		newStatusCommand(logger),
		newDiffCommand(logger),
		newUpdateCommand(logger),
		newVersionCommand(logger),
		newTUICommand(logger),
	)

	// Global flags
	rootCmd.PersistentFlags().String("config", "", "config file (default: .gitcells.yaml)")
	rootCmd.PersistentFlags().Bool("verbose", false, "enable verbose logging")
	rootCmd.PersistentFlags().Bool("tui", false, "launch interactive TUI mode")

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

func newTUICommand(logger *logrus.Logger) *cobra.Command {
	var useV2 bool
	cmd := &cobra.Command{
		Use:   "tui",
		Short: "Launch interactive TUI mode",
		Long:  `Launch GitCells in Terminal User Interface mode for interactive operations`,
		RunE: func(cmd *cobra.Command, args []string) error {
			logger.Info("Launching GitCells TUI...")
			if useV2 {
				logger.Info("Using condensed TUI v2...")
				return tui.RunV2()
			}
			return tui.Run()
		},
	}
	cmd.Flags().BoolVar(&useV2, "v2", true, "Use the condensed v2 TUI (default: true)")
	return cmd
}
