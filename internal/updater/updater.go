package updater

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/inconshreveable/go-update"
	"github.com/mattn/go-isatty"
)

// isInteractive reports whether stdin is connected to a terminal. Used to
// avoid the updater hanging on `fmt.Scanln` when running under CI, piped
// input, the TUI, or any other non-TTY context.
func isInteractive() bool {
	return isatty.IsTerminal(os.Stdin.Fd()) || isatty.IsCygwinTerminal(os.Stdin.Fd())
}

// promptYesNo returns true if stdin is a TTY and the user typed y/Y.
// Returns false (without consuming stdin) on non-TTY contexts.
func promptYesNo(question string) bool {
	if !isInteractive() {
		return false
	}
	fmt.Print(question)
	var response string
	if _, err := fmt.Scanln(&response); err != nil {
		return false
	}
	return response == "y" || response == "Y"
}

// isSafeArchiveBinaryName returns true if name is a permitted gitcells
// binary entry inside a release archive: a basename only (no directories,
// no path traversal), exactly "gitcells" or "gitcells.exe".
func isSafeArchiveBinaryName(name string) bool {
	clean := filepath.Clean(name)
	if clean != filepath.Base(clean) {
		return false
	}
	if strings.ContainsAny(clean, "/\\") || strings.Contains(clean, "..") {
		return false
	}
	return clean == "gitcells" || clean == "gitcells.exe"
}

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

	// Accept only canonical checksum filenames. A loose match would allow a
	// release with multiple "checksums*"-named assets to shadow the real
	// one with attacker-controlled hashes.
	allowedChecksumNames := map[string]bool{
		"checksums.txt":    true,
		"checksums.sha256": true,
		"SHA256SUMS":       true,
	}

	for _, asset := range release.Assets {
		if asset.Name == assetName+".tar.gz" || asset.Name == assetName+".zip" {
			downloadURL = asset.BrowserDownloadURL
			expectedSize = asset.Size
		}
		if allowedChecksumNames[asset.Name] {
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
	var binaryData []byte

	// Determine archive type by the asset name in the URL
	if strings.HasSuffix(downloadURL, ".zip") {
		reader, err := u.extractBinaryFromZip(data, assetName)
		if err != nil {
			return fmt.Errorf("failed to extract binary: %w", err)
		}
		binaryData, err = io.ReadAll(reader)
		if err != nil {
			return fmt.Errorf("failed to read binary data: %w", err)
		}
	} else {
		reader, err := u.extractBinaryFromTarGz(strings.NewReader(string(data)), assetName)
		if err != nil {
			return fmt.Errorf("failed to extract binary: %w", err)
		}
		binaryData, err = io.ReadAll(reader)
		if err != nil {
			return fmt.Errorf("failed to read binary data: %w", err)
		}
	}

	// Apply the update with backup
	err = u.applyUpdateWithPermissions(bytes.NewReader(binaryData))
	if err != nil {
		return err
	}

	return nil
}

func (u *Updater) getAssetName() string {
	os := runtime.GOOS
	arch := runtime.GOARCH

	// Map to the actual asset naming convention used in releases
	switch os {
	case "darwin":
		if arch == "arm64" {
			return "gitcells-macos-apple-silicon"
		} else {
			return "gitcells-macos-intel"
		}
	case "linux":
		if arch == "arm64" {
			return "gitcells-linux-arm64"
		} else {
			return "gitcells-linux"
		}
	case "windows":
		return "gitcells-windows"
	default:
		// Fallback to old format for compatibility
		switch arch {
		case "amd64":
			arch = "x86_64"
		case "386":
			arch = "i386"
		}
		return fmt.Sprintf("%s_%s", os, arch)
	}
}

func (u *Updater) extractBinaryFromTarGz(reader io.Reader, assetName string) (io.Reader, error) {
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

		if !isSafeArchiveBinaryName(header.Name) {
			continue
		}
		// Limit file size to prevent decompression bombs (100MB max)
		const maxSize = 100 * 1024 * 1024
		// LimitReader takes maxSize+1 so we can distinguish "exactly at
		// limit" from "truncated at limit".
		limitReader := io.LimitReader(tr, maxSize+1)

		var buf bytes.Buffer
		n, err := io.Copy(&buf, limitReader)
		if err != nil {
			return nil, fmt.Errorf("failed to read binary from archive: %w", err)
		}
		if n > maxSize {
			return nil, fmt.Errorf("binary file too large (exceeds %d bytes)", maxSize)
		}
		return bytes.NewReader(buf.Bytes()), nil
	}

	return nil, fmt.Errorf("gitcells binary not found in archive")
}

func (u *Updater) extractBinaryFromZip(data []byte, assetName string) (io.Reader, error) {
	// For .zip files
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("failed to create zip reader: %w", err)
	}

	for _, file := range reader.File {
		if !isSafeArchiveBinaryName(file.Name) {
			continue
		}
		const maxSize = 100 * 1024 * 1024
		if file.UncompressedSize64 > maxSize {
			return nil, fmt.Errorf("binary file too large (exceeds %d bytes)", maxSize)
		}

		rc, err := file.Open()
		if err != nil {
			return nil, fmt.Errorf("failed to open file in zip: %w", err)
		}
		defer rc.Close()

		var buf bytes.Buffer
		limitReader := io.LimitReader(rc, maxSize+1)
		n, err := io.Copy(&buf, limitReader)
		if err != nil {
			return nil, fmt.Errorf("failed to read binary from zip: %w", err)
		}
		if n > maxSize {
			return nil, fmt.Errorf("binary file too large (exceeds %d bytes)", maxSize)
		}
		return bytes.NewReader(buf.Bytes()), nil
	}

	return nil, fmt.Errorf("gitcells binary not found in zip archive")
}

func (u *Updater) isNewerVersion(latest, current string) bool {
	// Handle dev version or commit hash
	if current == "dev" || current == "unknown" {
		return true
	}

	// If current looks like a commit hash (short hex string), assume update is available
	if len(current) >= 6 && len(current) <= 10 && isHexString(current) {
		return true
	}

	// Clean version strings
	latest = strings.TrimPrefix(latest, "v")
	current = strings.TrimPrefix(current, "v")

	latestVersion, err := semver.NewVersion(latest)
	if err != nil {
		// Fallback to string comparison if semver parsing fails
		return latest != current && latest > current
	}

	currentVersion, err := semver.NewVersion(current)
	if err != nil {
		// If current version is not valid semver, assume update is available
		return true
	}

	return latestVersion.GreaterThan(currentVersion)
}

// isHexString checks if a string contains only hexadecimal characters
func isHexString(s string) bool {
	for _, r := range s {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')) {
			return false
		}
	}
	return true
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

// applyUpdateWithPermissions attempts to apply the update with proper permission handling
func (u *Updater) applyUpdateWithPermissions(binary io.Reader) error {
	// Read the binary data once so we can retry if needed
	binaryData, err := io.ReadAll(binary)
	if err != nil {
		return fmt.Errorf("failed to read binary data: %w", err)
	}
	// First try to apply normally
	err = update.Apply(bytes.NewReader(binaryData), update.Options{
		TargetMode: 0755,
	})

	// If permission denied, handle based on platform
	if err != nil && strings.Contains(err.Error(), "permission denied") {
		execPath, pathErr := u.GetCurrentExecutablePath()
		if pathErr != nil {
			return fmt.Errorf("failed to get executable path: %w", pathErr)
		}

		switch runtime.GOOS {
		case "darwin", "linux":
			return u.applyUpdateWithSudo(bytes.NewReader(binaryData), execPath)
		case "windows":
			return u.applyUpdateWithElevation(bytes.NewReader(binaryData), execPath)
		default:
			return fmt.Errorf("update failed due to permissions: %w. Please run with elevated privileges", err)
		}
	}

	return err
}

// applyUpdateWithSudo handles Unix-like systems that require sudo
func (u *Updater) applyUpdateWithSudo(binary io.Reader, targetPath string) error {
	// Create a temporary file to store the new binary
	tmpFile, err := os.CreateTemp("", "gitcells-update-*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write the binary to temp file
	_, err = io.Copy(tmpFile, binary)
	if err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write update to temp file: %w", err)
	}
	tmpFile.Close()

	// Make temp file executable
	err = os.Chmod(tmpFile.Name(), 0755)
	if err != nil {
		return fmt.Errorf("failed to set permissions on temp file: %w", err)
	}

	fmt.Println("\nPermission denied. GitCells needs elevated privileges to update.")
	fmt.Println("You have two options:")
	fmt.Println()
	fmt.Println("Option 1 - Use sudo (recommended):")
	fmt.Printf("  sudo mv %s %s\n", tmpFile.Name(), targetPath)
	fmt.Println()
	fmt.Println("Option 2 - Update manually:")
	fmt.Printf("  The new version has been downloaded to: %s\n", tmpFile.Name())
	fmt.Printf("  You can manually replace %s with this file\n", targetPath)
	fmt.Println()

	if promptYesNo("Would you like to run the sudo command now? (y/N): ") {
		// #nosec G204 - tmpFile.Name() and targetPath are from our own code, not user input
		cmd := exec.Command("sudo", "mv", tmpFile.Name(), targetPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("sudo command failed: %w", err)
		}

		// Ensure the file has correct permissions
		cmd = exec.Command("sudo", "chmod", "755", targetPath)
		if err := cmd.Run(); err != nil {
			fmt.Printf("Warning: failed to set permissions: %v\n", err)
		}

		return nil
	}

	// Non-TTY or user declined: print manual instructions and exit clean.
	return fmt.Errorf("update downloaded to %s. Please manually move it to %s", tmpFile.Name(), targetPath)
}

// applyUpdateWithElevation handles Windows systems that require elevation
func (u *Updater) applyUpdateWithElevation(binary io.Reader, targetPath string) error {
	// Create a temporary file to store the new binary
	tmpFile, err := os.CreateTemp("", "gitcells-update-*.exe.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}

	// Write the binary to temp file
	_, err = io.Copy(tmpFile, binary)
	if err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return fmt.Errorf("failed to write update to temp file: %w", err)
	}
	tmpFile.Close()

	// Quote both paths defensively. tmpFile.Name() / targetPath come from
	// os.CreateTemp and our own derivation respectively, but escaping any
	// stray `"` keeps the script robust if a user has a quote in their
	// install path.
	quote := func(s string) string { return strings.ReplaceAll(s, `"`, `""`) }
	batchScript := fmt.Sprintf(`@echo off
echo Updating GitCells...
ping 127.0.0.1 -n 2 > nul
move /Y "%s" "%s"
if %%errorlevel%% equ 0 (
    echo Update successful!
) else (
    echo Update failed. Please run as administrator.
    pause
)
`, quote(tmpFile.Name()), quote(targetPath))

	batchFile, err := os.CreateTemp("", "gitcells-update-*.bat")
	if err != nil {
		os.Remove(tmpFile.Name())
		return fmt.Errorf("failed to create batch file: %w", err)
	}

	_, err = batchFile.WriteString(batchScript)
	batchFile.Close()
	if err != nil {
		os.Remove(tmpFile.Name())
		os.Remove(batchFile.Name())
		return fmt.Errorf("failed to write batch file: %w", err)
	}

	fmt.Println("\nPermission denied. GitCells needs administrator privileges to update.")
	fmt.Println("Please run one of the following commands in an elevated command prompt:")
	fmt.Println()
	fmt.Printf("Option 1 - Run the update script:\n")
	fmt.Printf("  %s\n", batchFile.Name())
	fmt.Println()
	fmt.Printf("Option 2 - Update manually:\n")
	fmt.Printf("  move /Y \"%s\" \"%s\"\n", tmpFile.Name(), targetPath)
	fmt.Println()

	// Try to run with elevation using PowerShell
	if promptYesNo("Would you like to try running with administrator privileges now? (y/N): ") {
		// Use PowerShell to run as administrator. Escape the path in single
		// quotes by doubling embedded ' characters (PowerShell rule).
		psQuoted := strings.ReplaceAll(batchFile.Name(), "'", "''")
		// #nosec G204 - batchFile.Name() is from our own temp file creation, not user input
		cmd := exec.Command("powershell", "-Command",
			fmt.Sprintf("Start-Process cmd -ArgumentList '/c \"%s\"' -Verb RunAs", psQuoted))

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to run with elevation: %w. Please run the command manually", err)
		}

		fmt.Println("Update script launched. Please check the elevated window for results.")
		return nil
	}

	return fmt.Errorf("update downloaded to %s. Please run the update script or move manually", tmpFile.Name())
}
