package models

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/Classic-Homes/gitcells/internal/config"
	"github.com/Classic-Homes/gitcells/internal/constants"
	"github.com/Classic-Homes/gitcells/internal/tui/messages"
	"github.com/Classic-Homes/gitcells/internal/updater"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type SettingsModel struct {
	width       int
	height      int
	cursor      int
	status      string
	updating    bool
	showConfirm bool
	confirmType string
	config      *config.Config
	currentView settingsView
}

type settingsView int

const (
	viewMain settingsView = iota
	viewFeatures
	viewUpdates
)

type updateCheckMsg struct {
	release   *updater.GitHubRelease
	hasUpdate bool
	err       error
}

type updateCompleteMsg struct {
	err error
}

type uninstallCompleteMsg struct {
	err error
}

type configLoadedMsg struct {
	config *config.Config
	err    error
}

type configSavedMsg struct {
	err error
}

var mainSettingsItems = []struct {
	title string
	desc  string
	key   string
}{
	{"Feature Settings", "Configure experimental features and beta options", "features"},
	{"Update Settings", "Configure automatic updates and preferences", "updates"},
	{"Check for Updates", "Check if a newer version of GitCells is available", "update_check"},
	{"Update GitCells", "Download and install the latest version", "update"},
	{"Uninstall GitCells", "Remove GitCells from your system", "uninstall"},
	{"Version Info", "Display current version and system information", "version"},
}

var featureSettingsItems = []struct {
	title string
	desc  string
	key   string
}{
	{"Experimental Features", "Enable cutting-edge features (may be unstable)", "experimental"},
	{"Beta Updates", "Receive beta version updates", "beta_updates"},
	{"Telemetry", "Send anonymous usage data to help improve GitCells", "telemetry"},
}

var updateSettingsItems = []struct {
	title string
	desc  string
	key   string
}{
	{"Auto Check Updates", "Automatically check for updates on startup", "auto_check"},
	{"Include Prereleases", "Include pre-release versions when checking for updates", "prereleases"},
	{"Auto Download Updates", "Automatically download updates when available", "auto_download"},
	{"Notify on Update", "Show notifications when updates are available", "notify"},
}

func NewSettingsModel() SettingsModel {
	return SettingsModel{
		cursor:      0,
		status:      "Loading configuration...",
		currentView: viewMain,
	}
}

func (m SettingsModel) Init() tea.Cmd {
	return m.loadConfig()
}

func (m SettingsModel) loadConfig() tea.Cmd {
	return func() tea.Msg {
		cfg, err := config.Load("")
		return configLoadedMsg{config: cfg, err: err}
	}
}

func (m SettingsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case configLoadedMsg:
		if msg.err != nil {
			m.status = fmt.Sprintf("Error loading config: %v", msg.err)
			m.config = config.GetDefault()
		} else {
			m.config = msg.config
			m.status = "Ready"
		}
		return m, nil

	case configSavedMsg:
		if msg.err != nil {
			m.status = fmt.Sprintf("Error saving config: %v", msg.err)
		} else {
			m.status = "Settings saved successfully!"
		}
		return m, nil

	case tea.KeyMsg:
		if m.showConfirm {
			switch msg.String() {
			case "y", "Y":
				m.showConfirm = false
				if m.confirmType == "update" {
					m.updating = true
					m.status = "Updating GitCells..."
					return m, m.performUpdate()
				} else if m.confirmType == "uninstall" {
					m.status = "Uninstalling GitCells..."
					return m, m.performUninstall()
				}
			case "n", "N", "esc":
				m.showConfirm = false
				m.status = "Operation cancelled"
			}
			return m, nil
		}

		if m.updating {
			return m, nil
		}

		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc":
			if m.currentView != viewMain {
				m.currentView = viewMain
				m.cursor = 0
				m.status = "Returned to main settings"
				return m, nil
			} else {
				return m, messages.RequestMainMenu()
			}
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			maxItems := m.getMaxCursor()
			if m.cursor < maxItems-1 {
				m.cursor++
			}
		case "enter", " ":
			return m.handleSelectionAndReturn()
		}

	case updateCheckMsg:
		m.updating = false
		if msg.err != nil {
			m.status = fmt.Sprintf("Error checking for updates: %v", msg.err)
		} else if msg.hasUpdate {
			m.status = fmt.Sprintf("Update available: %s → %s", constants.Version, msg.release.TagName)
		} else {
			m.status = "GitCells is up to date"
		}

	case updateCompleteMsg:
		m.updating = false
		if msg.err != nil {
			m.status = fmt.Sprintf("Update failed: %v", msg.err)
		} else {
			m.status = "Update completed successfully! Please restart GitCells."
		}

	case uninstallCompleteMsg:
		if msg.err != nil {
			m.status = fmt.Sprintf("Uninstall failed: %v", msg.err)
		} else {
			m.status = "GitCells has been uninstalled. Goodbye!"
		}
	}

	return m, nil
}

func (m SettingsModel) ResetToMainView() SettingsModel {
	m.currentView = viewMain
	m.cursor = 0
	m.status = "Ready"
	m.showConfirm = false
	m.updating = false
	return m
}

func (m SettingsModel) getMaxCursor() int {
	switch m.currentView {
	case viewMain:
		return len(mainSettingsItems)
	case viewFeatures:
		return len(featureSettingsItems)
	case viewUpdates:
		return len(updateSettingsItems)
	default:
		return len(mainSettingsItems)
	}
}

func (m SettingsModel) View() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("99")).
		MarginBottom(1)

	subtitleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginBottom(2)

	menuStyle := lipgloss.NewStyle().
		Padding(2, 4)

	cursorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("99"))

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("82")).
		MarginTop(1).
		MarginBottom(1)

	confirmStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("214")).
		Background(lipgloss.Color("236")).
		Padding(1, 2).
		MarginTop(1).
		MarginBottom(1)

	toggleOnStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("82"))

	toggleOffStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("203"))

	// Build the title and breadcrumb based on current view
	var title, subtitle, breadcrumb string
	switch m.currentView {
	case viewMain:
		title = "GitCells Settings"
		subtitle = "System Management and Configuration"
		breadcrumb = "Main Menu"
	case viewFeatures:
		title = "Feature Settings"
		subtitle = "Experimental Features and Beta Options"
		breadcrumb = "Main Menu > Feature Settings"
	case viewUpdates:
		title = "Update Settings"
		subtitle = "Automatic Updates and Preferences"
		breadcrumb = "Main Menu > Update Settings"
	}

	breadcrumbStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("243")).
		Faint(true)

	s := titleStyle.Render(title) + "\n"
	s += subtitleStyle.Render(subtitle) + "\n"
	s += breadcrumbStyle.Render("Navigation: "+breadcrumb) + "\n\n"

	if m.showConfirm {
		var confirmMsg string
		if m.confirmType == "update" {
			confirmMsg = "Are you sure you want to update GitCells? This will download and replace the current binary."
		} else if m.confirmType == "uninstall" {
			confirmMsg = "Are you sure you want to uninstall GitCells? This will remove the binary and configuration files."
		}
		s += confirmStyle.Render(confirmMsg+"\n\n[Y]es / [N]o") + "\n\n"
	} else {
		// Render items based on current view
		var items []struct {
			title string
			desc  string
			key   string
		}

		switch m.currentView {
		case viewMain:
			items = mainSettingsItems
		case viewFeatures:
			items = featureSettingsItems
		case viewUpdates:
			items = updateSettingsItems
		}

		for i, item := range items {
			cursor := "  "
			if i == m.cursor {
				cursor = cursorStyle.Render("▶ ")
			}

			// Add toggle status for boolean settings
			itemTitle := item.title
			if m.config != nil && (m.currentView == viewFeatures || m.currentView == viewUpdates) {
				toggleValue := m.getToggleValue(item.key)
				if toggleValue {
					itemTitle += " " + toggleOnStyle.Render("[ON]")
				} else {
					itemTitle += " " + toggleOffStyle.Render("[OFF]")
				}
			}

			s += fmt.Sprintf("%s%s\n", cursor, itemTitle)
			s += fmt.Sprintf("    %s\n\n", descStyle.Render(item.desc))
		}
	}

	s += statusStyle.Render("Status: "+m.status) + "\n"

	if m.updating {
		s += descStyle.Render("Please wait...")
	} else if !m.showConfirm {
		if m.currentView == viewMain {
			s += descStyle.Render("Use ↑/↓ or j/k to navigate, Enter to select, q to quit")
		} else {
			s += descStyle.Render("Use ↑/↓ or j/k to navigate, Enter/Space to toggle, Esc to go back")
		}
	}

	return menuStyle.Render(s)
}

func (m SettingsModel) handleSelectionAndReturn() (tea.Model, tea.Cmd) {
	var selectedItem struct {
		title string
		desc  string
		key   string
	}

	switch m.currentView {
	case viewMain:
		selectedItem = mainSettingsItems[m.cursor]
	case viewFeatures:
		selectedItem = featureSettingsItems[m.cursor]
	case viewUpdates:
		selectedItem = updateSettingsItems[m.cursor]
	}

	switch selectedItem.key {
	// Main menu navigation
	case "features":
		m.currentView = viewFeatures
		m.cursor = 0
		m.status = "Entered feature settings"
		return m, nil
	case "updates":
		m.currentView = viewUpdates
		m.cursor = 0
		m.status = "Entered update settings"
		return m, nil

	// System actions
	case "update_check":
		m.updating = true
		m.status = "Checking for updates..."
		return m, m.checkForUpdates()
	case "update":
		m.showConfirm = true
		m.confirmType = "update"
		return m, nil
	case "uninstall":
		m.showConfirm = true
		m.confirmType = "uninstall"
		return m, nil
	case "version":
		m.status = fmt.Sprintf("GitCells v%s (Go %s)", constants.Version, constants.GoVersion)
		return m, nil

	// Feature toggles
	case "experimental":
		if m.config != nil {
			m.config.Features.EnableExperimentalFeatures = !m.config.Features.EnableExperimentalFeatures
			return m, m.saveConfig()
		}
		return m, nil
	case "beta_updates":
		if m.config != nil {
			m.config.Features.EnableBetaUpdates = !m.config.Features.EnableBetaUpdates
			return m, m.saveConfig()
		}
		return m, nil
	case "telemetry":
		if m.config != nil {
			m.config.Features.EnableTelemetry = !m.config.Features.EnableTelemetry
			return m, m.saveConfig()
		}
		return m, nil

	// Update toggles
	case "auto_check":
		if m.config != nil {
			m.config.Updates.AutoCheckUpdates = !m.config.Updates.AutoCheckUpdates
			return m, m.saveConfig()
		}
		return m, nil
	case "prereleases":
		if m.config != nil {
			m.config.Updates.IncludePrereleases = !m.config.Updates.IncludePrereleases
			return m, m.saveConfig()
		}
		return m, nil
	case "auto_download":
		if m.config != nil {
			m.config.Updates.AutoDownloadUpdates = !m.config.Updates.AutoDownloadUpdates
			return m, m.saveConfig()
		}
		return m, nil
	case "notify":
		if m.config != nil {
			m.config.Updates.NotifyOnUpdate = !m.config.Updates.NotifyOnUpdate
			return m, m.saveConfig()
		}
		return m, nil
	}

	return m, nil
}

func (m SettingsModel) getToggleValue(key string) bool {
	if m.config == nil {
		return false
	}

	switch key {
	// Feature settings
	case "experimental":
		return m.config.Features.EnableExperimentalFeatures
	case "beta_updates":
		return m.config.Features.EnableBetaUpdates
	case "telemetry":
		return m.config.Features.EnableTelemetry
	// Update settings
	case "auto_check":
		return m.config.Updates.AutoCheckUpdates
	case "prereleases":
		return m.config.Updates.IncludePrereleases
	case "auto_download":
		return m.config.Updates.AutoDownloadUpdates
	case "notify":
		return m.config.Updates.NotifyOnUpdate
	default:
		return false
	}
}

func (m SettingsModel) saveConfig() tea.Cmd {
	return func() tea.Msg {
		err := m.config.Save("")
		return configSavedMsg{err: err}
	}
}

func (m SettingsModel) checkForUpdates() tea.Cmd {
	return func() tea.Msg {
		u := updater.New(constants.Version)
		release, hasUpdate, err := u.CheckForUpdate()
		return updateCheckMsg{
			release:   release,
			hasUpdate: hasUpdate,
			err:       err,
		}
	}
}

func (m SettingsModel) performUpdate() tea.Cmd {
	return func() tea.Msg {
		u := updater.New(constants.Version)
		release, hasUpdate, err := u.CheckForUpdate()
		if err != nil {
			return updateCompleteMsg{err: err}
		}
		if !hasUpdate {
			return updateCompleteMsg{err: fmt.Errorf("no update available")}
		}

		err = u.Update(release)
		return updateCompleteMsg{err: err}
	}
}

func (m SettingsModel) performUninstall() tea.Cmd {
	return func() tea.Msg {
		// Look for uninstall script in common locations
		scriptPaths := []string{
			"../scripts/uninstall.sh",
			"./scripts/uninstall.sh",
			"/usr/local/share/gitcells/uninstall.sh",
		}

		var uninstallScript string
		for _, path := range scriptPaths {
			if _, err := os.Stat(path); err == nil {
				uninstallScript = path
				break
			}
		}

		if uninstallScript == "" {
			// Fallback: create a simple uninstall command
			return uninstallCompleteMsg{err: m.performSimpleUninstall()}
		}

		// Run the uninstall script
		cmd := exec.Command("bash", uninstallScript)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()

		return uninstallCompleteMsg{err: err}
	}
}

func (m SettingsModel) performSimpleUninstall() error {
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Try to remove the binary
	err = os.Remove(executable)
	if err != nil && !strings.Contains(err.Error(), "permission denied") {
		return fmt.Errorf("failed to remove binary: %w", err)
	}

	// Try to remove common config directories
	configDirs := []string{
		os.ExpandEnv("$HOME/.config/gitcells"),
		os.ExpandEnv("$HOME/.gitcells"),
	}

	for _, dir := range configDirs {
		if _, err := os.Stat(dir); err == nil {
			os.RemoveAll(dir) // Ignore errors for config removal
		}
	}

	return nil
}
