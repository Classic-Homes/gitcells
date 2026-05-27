package main

import (
	"fmt"

	"github.com/Classic-Homes/gitcells/internal/constants"
	"github.com/Classic-Homes/gitcells/internal/updater"
	"github.com/Classic-Homes/gitcells/internal/utils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func newUpdateCommand(logger *logrus.Logger) *cobra.Command {
	var prerelease bool

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Check for a newer GitCells release",
		Long: `Check GitHub for a newer release of GitCells and print download
instructions. GitCells does NOT self-replace its binary — install
updates with your package manager or by downloading from the
release page (locked-binary issues on Windows make in-place
self-updates unreliable).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			u := updater.NewWithPrerelease(constants.Version, prerelease)

			logger.Info("Checking for updates...")
			release, hasUpdate, err := u.CheckForUpdate()
			if err != nil {
				return utils.WrapError(err, utils.ErrorTypeNetwork, "update", "failed to check for updates")
			}

			if !hasUpdate {
				fmt.Printf("GitCells is up to date (version %s)\n", constants.Version)
				return nil
			}

			fmt.Printf("New version available: %s -> %s\n", constants.Version, release.TagName)
			if release.Prerelease {
				fmt.Printf("Pre-release: %s\n", release.Name)
			}
			if release.Body != "" {
				fmt.Printf("\nRelease notes:\n%s\n", release.Body)
			}
			fmt.Printf("\nDownload: %s\n", releaseURL(release))
			fmt.Println("Or install via your package manager.")
			return nil
		},
	}

	cmd.Flags().BoolVar(&prerelease, "prerelease", false, "Include pre-release versions")
	return cmd
}

func releaseURL(r *updater.GitHubRelease) string {
	if r != nil && r.HTMLURL != "" {
		return r.HTMLURL
	}
	return updater.ReleasePageURL()
}

func newVersionCommand(logger *logrus.Logger) *cobra.Command {
	var checkUpdate bool
	var prerelease bool

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Long:  `Display the current version of GitCells and optionally check for updates.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("GitCells version %s (built %s)\n", constants.Version, constants.BuildTime)

			if checkUpdate {
				u := updater.NewWithPrerelease(constants.Version, prerelease)
				logger.Debug("Checking for updates...")

				release, hasUpdate, err := u.CheckForUpdate()
				if err != nil {
					logger.Warnf("Failed to check for updates: %v", err)
					return nil
				}

				if hasUpdate {
					tag := release.TagName
					if release.Prerelease {
						fmt.Printf("New pre-release available: %s\n", tag)
					} else {
						fmt.Printf("New version available: %s\n", tag)
					}
					fmt.Printf("Download: %s\n", releaseURL(release))
				} else {
					fmt.Println("You are running the latest version")
				}
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&checkUpdate, "check-update", false, "Check for available updates")
	cmd.Flags().BoolVar(&prerelease, "prerelease", false, "Include pre-release versions when checking")
	return cmd
}
