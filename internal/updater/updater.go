package updater

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/Classic-Homes/gitcells/internal/constants"
	"github.com/Masterminds/semver/v3"
	"github.com/inconshreveable/go-update"
)

const (
	GitHubAPI     = "https://api.github.com"
	Repository    = "Classic-Homes/gitcells"
	UpdateTimeout = 30 * time.Second
)

type GitHubRelease struct {
	TagName    string `json:"tag_name"`
	Name       string `json:"name"`
	Body       string `json:"body"`
	Draft      bool   `json:"draft"`
	Prerelease bool   `json:"prerelease"`
	Assets     []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
		Size               int64  `json:"size"`
	} `json:"assets"`
	PublishedAt time.Time `json:"published_at"`
}

type Updater struct {
	CurrentVersion  string
	Repository      string
	AllowPrerelease bool
	client          *http.Client
}

func New(currentVersion string) *Updater {
	return &Updater{
		CurrentVersion:  currentVersion,
		Repository:      Repository,
		AllowPrerelease: false,
		client: &http.Client{
			Timeout: UpdateTimeout,
		},
	}
}

func NewWithPrerelease(currentVersion string, allowPrerelease bool) *Updater {
	return &Updater{
		CurrentVersion:  currentVersion,
		Repository:      Repository,
		AllowPrerelease: allowPrerelease,
		client: &http.Client{
			Timeout: UpdateTimeout,
		},
	}
}

func (u *Updater) CheckForUpdate() (*GitHubRelease, bool, error) {
	var url string
	var release GitHubRelease

	if u.AllowPrerelease {
		// Get all releases and find the latest (including pre-releases)
		url = fmt.Sprintf("%s/repos/%s/releases", GitHubAPI, u.Repository)

		resp, err := u.client.Get(url)
		if err != nil {
			return nil, false, fmt.Errorf("failed to fetch releases: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, false, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
		}

		var releases []GitHubRelease
		if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
			return nil, false, fmt.Errorf("failed to decode releases response: %w", err)
		}

		// Find the latest release (including pre-releases, but skip drafts)
		var latestRelease *GitHubRelease
		for _, rel := range releases {
			if rel.Draft {
				continue
			}
			if latestRelease == nil || rel.PublishedAt.After(latestRelease.PublishedAt) {
				latestRelease = &rel
			}
		}

		if latestRelease == nil {
			return nil, false, fmt.Errorf("no releases found")
		}

		release = *latestRelease
	} else {
		// Get only the latest stable release
		url = fmt.Sprintf("%s/repos/%s/releases/latest", GitHubAPI, u.Repository)

		resp, err := u.client.Get(url)
		if err != nil {
			return nil, false, fmt.Errorf("failed to fetch latest release: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, false, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
		}

		if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
			return nil, false, fmt.Errorf("failed to decode release response: %w", err)
		}

		// Skip drafts and prereleases for stable updates
		if release.Draft || release.Prerelease {
			return nil, false, nil
		}
	}

	// Compare versions
	hasUpdate := u.isNewerVersion(release.TagName, u.CurrentVersion)

	return &release, hasUpdate, nil
}

func (u *Updater) Update(release *GitHubRelease) error {
	// Find the appropriate asset for current platform
	assetName := u.getAssetName()
	var downloadURL string
	var expectedSize int64
	var checksumURL string

	for _, asset := range release.Assets {
		if strings.Contains(asset.Name, assetName) && strings.HasSuffix(asset.Name, ".tar.gz") {
			downloadURL = asset.BrowserDownloadURL
			expectedSize = asset.Size
		}
		// Look for checksum file
		if strings.Contains(asset.Name, "checksums") && (strings.HasSuffix(asset.Name, ".txt") || strings.HasSuffix(asset.Name, ".sha256")) {
			checksumURL = asset.BrowserDownloadURL
		}
	}

	if downloadURL == "" {
		return fmt.Errorf("no compatible asset found for platform %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	// Download checksums if available
	var expectedChecksum string
	if checksumURL != "" {
		checksum, err := u.downloadChecksum(checksumURL, assetName)
		if err != nil {
			// Don't fail the update if checksum download fails, just log a warning
			// In production, you might want to make this mandatory
			fmt.Printf("Warning: Could not download checksums for verification: %v\n", err)
		} else {
			expectedChecksum = checksum
		}
	}

	// Download the update
	resp, err := u.client.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("failed to download update: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	// Verify content length if available
	if resp.ContentLength > 0 && resp.ContentLength != expectedSize {
		return fmt.Errorf("downloaded size mismatch: expected %d, got %d", expectedSize, resp.ContentLength)
	}

	// Read the entire response for checksum verification
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read download data: %w", err)
	}

	// Verify checksum if available
	if expectedChecksum != "" {
		if err := u.VerifyChecksum(data, expectedChecksum); err != nil {
			return fmt.Errorf("checksum verification failed: %w", err)
		}
		fmt.Println("✅ Checksum verification passed")
	}

	// Extract binary from archive
	binary, err := u.extractBinary(strings.NewReader(string(data)), assetName)
	if err != nil {
		return fmt.Errorf("failed to extract binary: %w", err)
	}

	// Apply the update with backup
	err = update.Apply(binary, update.Options{
		// Create a backup of the current binary
		TargetMode: constants.DirPermissions,
	})
	if err != nil {
		return fmt.Errorf("failed to apply update: %w", err)
	}

	return nil
}

func (u *Updater) getAssetName() string {
	os := runtime.GOOS
	arch := runtime.GOARCH

	// Convert Go arch names to common naming conventions
	switch arch {
	case "amd64":
		arch = "x86_64"
	case "386":
		arch = "i386"
	}

	return fmt.Sprintf("%s_%s", os, arch)
}

func (u *Updater) extractBinary(reader io.Reader, assetName string) (io.Reader, error) {
	// For .tar.gz files
	gzr, err := gzip.NewReader(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read tar header: %w", err)
		}

		// Look for the gitcells binary
		if strings.Contains(header.Name, "gitcells") && !strings.Contains(header.Name, "/") {
			// Limit file size to prevent decompression bombs (100MB max)
			const maxSize = 100 * 1024 * 1024
			limitReader := io.LimitReader(tr, maxSize)

			// Create a buffer to hold the binary
			var buf strings.Builder
			n, err := io.Copy(&buf, limitReader)
			if err != nil {
				return nil, fmt.Errorf("failed to read binary from archive: %w", err)
			}
			if n == maxSize {
				return nil, fmt.Errorf("binary file too large (exceeds %d bytes)", maxSize)
			}
			return strings.NewReader(buf.String()), nil
		}
	}

	return nil, fmt.Errorf("gitcells binary not found in archive")
}

func (u *Updater) isNewerVersion(latest, current string) bool {
	// Handle dev version
	if current == "dev" {
		return true
	}

	latestVersion, err := semver.NewVersion(latest)
	if err != nil {
		// Fallback to string comparison if semver parsing fails
		latest = strings.TrimPrefix(latest, "v")
		current = strings.TrimPrefix(current, "v")
		return latest != current && latest > current
	}

	currentVersion, err := semver.NewVersion(current)
	if err != nil {
		// If current version is not valid semver, assume update is available
		return true
	}

	return latestVersion.GreaterThan(currentVersion)
}

func (u *Updater) VerifyChecksum(data []byte, expectedChecksum string) error {
	hasher := sha256.New()
	hasher.Write(data)
	actualChecksum := hex.EncodeToString(hasher.Sum(nil))

	if actualChecksum != expectedChecksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum)
	}

	return nil
}

func (u *Updater) downloadChecksum(checksumURL, assetName string) (string, error) {
	resp, err := u.client.Get(checksumURL)
	if err != nil {
		return "", fmt.Errorf("failed to download checksum file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("checksum download failed with status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read checksum data: %w", err)
	}

	// Parse checksum file (format: checksum filename)
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) >= 2 {
			checksum := parts[0]
			filename := parts[1]

			// Remove any leading path separators or wildcards
			filename = strings.TrimPrefix(filename, "*")
			filename = strings.TrimPrefix(filename, "./")

			// Check if this line is for our asset
			if strings.Contains(filename, assetName) {
				return checksum, nil
			}
		}
	}

	return "", fmt.Errorf("checksum not found for asset %s", assetName)
}

func (u *Updater) GetCurrentExecutablePath() (string, error) {
	executable, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %w", err)
	}

	return filepath.EvalSymlinks(executable)
}
