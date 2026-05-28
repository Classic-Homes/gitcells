package main

import (
	"fmt"
	"os"

	"github.com/Classic-Homes/gitcells/internal/constants"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	logger *logrus.Logger

	// colorsDisabled is set by PersistentPreRun based on --no-color,
	// NO_COLOR, and TTY detection. Subcommands that print their own ANSI
	// codes (currently diff) read this directly so the logger and the
	// command output stay consistent.
	colorsDisabled bool
)

func main() {
	logger = setupLogger()

	rootCmd := &cobra.Command{
		Use:   "gitcells",
		Short: "Version control for Excel files",
		Long: `GitCells converts Excel files to JSON so you can version-control them with Git.

Quickstart:
  gitcells init                  Set up GitCells in the current directory
  gitcells convert sheet.xlsx    Convert one workbook to JSON chunks
  gitcells watch ./reports       Auto-convert on every save
  gitcells status                See what's modified vs. stored
  gitcells diff sheet.xlsx       Compare live workbook with stored chunks

Run 'gitcells <command> --help' for details on any command.`,
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
	)

	// Global flags
	rootCmd.PersistentFlags().String("config", "", "config file (default: .gitcells.yaml)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "enable verbose logging")
	rootCmd.PersistentFlags().Bool("no-color", false, "disable colored output (also honors NO_COLOR env var)")

	// Apply global flags before any subcommand runs.
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if verbose, _ := cmd.Flags().GetBool("verbose"); verbose {
			logger.SetLevel(logrus.DebugLevel)
		}

		flag, _ := cmd.Flags().GetBool("no-color")
		colorsDisabled = shouldDisableColors(flag, os.Getenv("NO_COLOR"), isStderrTerminal())
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
			DisableColors: colorsDisabled,
		})
	}

	if err := rootCmd.Execute(); err != nil {
		logger.Error(err)
		os.Exit(1)
	}
}

// shouldDisableColors returns true when ANSI color output should be
// suppressed. Order of precedence:
//
//  1. --no-color flag (highest)
//  2. NO_COLOR env var, per https://no-color.org (any non-empty value)
//  3. stderr is not a TTY (piped, redirected, or running under a CI log
//     collector that strips escape codes anyway)
//
// Kept as a pure function so the truth table is unit-testable without
// touching real terminals.
func shouldDisableColors(flag bool, envValue string, isTTY bool) bool {
	if flag {
		return true
	}
	if envValue != "" {
		return true
	}
	return !isTTY
}

// isStderrTerminal reports whether stderr is connected to a terminal.
// Stderr (not stdout) is what logrus writes to by default, so basing the
// color decision on stderr matches what the user actually sees from log
// lines. Subcommands that print to stdout (diff, status) also consult
// this — keeping the two streams consistent is simpler than tracking
// them separately.
func isStderrTerminal() bool {
	return term.IsTerminal(int(os.Stderr.Fd()))
}

func setupLogger() *logrus.Logger {
	l := logrus.New()
	// Defaults are overwritten by PersistentPreRun once flags are parsed.
	// What's here is enough for early panics before Execute is called.
	l.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		DisableColors: shouldDisableColors(false, os.Getenv("NO_COLOR"), isStderrTerminal()),
	})
	return l
}
