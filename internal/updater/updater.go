// Package updater provides read-only update checking against GitHub
// releases. It does NOT self-replace the running binary: that path is
// unreliable on Windows (locked exe) and rarely worth the complexity
// elsewhere. Users install updates via package managers or by downloading
// the asset directly from the release page.
package updater

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
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
	HTMLURL    string `json:"html_url"`
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
	return NewWithPrerelease(currentVersion, false)
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

// CheckForUpdate queries GitHub for the latest applicable release and
// reports whether it's newer than the running version. Returns the
// release (always non-nil on success), a hasUpdate flag, and any error.
func (u *Updater) CheckForUpdate() (*GitHubRelease, bool, error) {
	var release GitHubRelease

	if u.AllowPrerelease {
		url := fmt.Sprintf("%s/repos/%s/releases", GitHubAPI, u.Repository)
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

		var latest *GitHubRelease
		for i := range releases {
			if releases[i].Draft {
				continue
			}
			if latest == nil || releases[i].PublishedAt.After(latest.PublishedAt) {
				latest = &releases[i]
			}
		}
		if latest == nil {
			return nil, false, fmt.Errorf("no releases found")
		}
		release = *latest
	} else {
		url := fmt.Sprintf("%s/repos/%s/releases/latest", GitHubAPI, u.Repository)
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
		if release.Draft || release.Prerelease {
			return &release, false, nil
		}
	}

	return &release, u.isNewerVersion(release.TagName, u.CurrentVersion), nil
}

// ReleasePageURL returns the human-facing release page on GitHub.
func ReleasePageURL() string {
	return fmt.Sprintf("https://github.com/%s/releases", Repository)
}

func (u *Updater) isNewerVersion(latest, current string) bool {
	// Dev builds and short hex commit hashes always look "older".
	if current == "dev" || current == "unknown" {
		return true
	}
	if len(current) >= 6 && len(current) <= 10 && isHexString(current) {
		return true
	}

	latest = strings.TrimPrefix(latest, "v")
	current = strings.TrimPrefix(current, "v")

	latestVer, err := semver.NewVersion(latest)
	if err != nil {
		return latest != current && latest > current
	}
	currentVer, err := semver.NewVersion(current)
	if err != nil {
		return true
	}
	return latestVer.GreaterThan(currentVer)
}

func isHexString(s string) bool {
	for _, r := range s {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')) {
			return false
		}
	}
	return true
}
