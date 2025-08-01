package main

import (
	"fmt"

	"github.com/Classic-Homes/gitcells/internal/updater"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func newUpdateCommand(logger *logrus.Logger) *cobra.Command {
	var checkOnly bool
	var force bool
	var prerelease bool

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update GitCells to the latest version",
		Long: `Check for and install the latest version of GitCells from GitHub releases.
		
This command will:
- Check GitHub for the latest release
- Download and verify the update
- Replace the current binary with the new version

Use --prerelease to include pre-release versions as update targets.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			u := updater.NewWithPrerelease(version, prerelease)

			logger.Info("Checking for updates...")
			release, hasUpdate, err := u.CheckForUpdate()
			if err != nil {
				return fmt.Errorf("failed to check for updates: %w", err)
			}

			if !hasUpdate {
				if prerelease {
					fmt.Printf("GitCells is already up to date (version %s, including pre-releases)\n", version)
				} else {
					fmt.Printf("GitCells is already up to date (version %s)\n", version)
				}
				return nil
			}

			fmt.Printf("New version available: %s -> %s\n", version, release.TagName)
			if release.Prerelease {
				fmt.Printf("⚠️  Pre-release: %s\n", release.Name)
			} else {
				fmt.Printf("Release: %s\n", release.Name)
			}

			if release.Body != "" {
				fmt.Printf("Release Notes:\n%s\n\n", release.Body)
			}

			if checkOnly {
				fmt.Println("Use 'gitcells update' without --check to install the update")
				return nil
			}

			if !force {
				fmt.Print("Do you want to update? (y/N): ")
				var response string
				if _, err := fmt.Scanln(&response); err != nil {
					fmt.Println("Update cancelled")
					return nil
				}
				if response != "y" && response != "Y" && response != "yes" && response != "Yes" {
					fmt.Println("Update cancelled")
					return nil
				}
			}

			logger.Info("Downloading and installing update...")
			fmt.Printf("Updating GitCells from %s to %s...\n", version, release.TagName)

			if err := u.Update(release); err != nil {
				return fmt.Errorf("failed to update: %w", err)
			}

			fmt.Printf("✅ Successfully updated to version %s\n", release.TagName)
			fmt.Println("Please restart GitCells to use the new version")

			return nil
		},
	}

	cmd.Flags().BoolVar(&checkOnly, "check", false, "Only check for updates, don't install")
	cmd.Flags().BoolVar(&force, "force", false, "Skip confirmation prompt and update automatically")
	cmd.Flags().BoolVar(&prerelease, "prerelease", false, "Include pre-release versions as update targets")

	return cmd
}

func newVersionCommand(logger *logrus.Logger) *cobra.Command {
	var checkUpdate bool
	var prerelease bool

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Long:  `Display the current version of GitCells and optionally check for updates`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("GitCells version %s (built %s)\n", version, buildTime)

			if checkUpdate {
				u := updater.NewWithPrerelease(version, prerelease)
				logger.Debug("Checking for updates...")

				release, hasUpdate, err := u.CheckForUpdate()
				if err != nil {
					logger.Warnf("Failed to check for updates: %v", err)
					return nil
				}

				if hasUpdate {
					if release.Prerelease {
						fmt.Printf("📦 New pre-release available: %s\n", release.TagName)
						fmt.Println("Run 'gitcells update --prerelease' to install the latest pre-release")
					} else {
						fmt.Printf("📦 New version available: %s\n", release.TagName)
						fmt.Println("Run 'gitcells update' to install the latest version")
					}
				} else {
					if prerelease {
						fmt.Println("✅ You are running the latest version (including pre-releases)")
					} else {
						fmt.Println("✅ You are running the latest version")
					}
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&checkUpdate, "check-update", false, "Check for available updates")
	cmd.Flags().BoolVar(&prerelease, "prerelease", false, "Include pre-release versions when checking for updates")

	return cmd
}
