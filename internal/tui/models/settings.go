package models

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

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
	editMode    bool
	editValue   string
	editKey     string
}

type settingsView int

const (
	viewMain settingsView = iota
	viewFeatures
	viewUpdates
	viewGit
	viewWatcher
	viewConverter
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
	{"Git Settings", "Configure Git integration and repository settings", "git"},
	{"Watcher Settings", "Configure file watching and monitoring", "watcher"},
	{"Converter Settings", "Configure Excel to JSON conversion options", "converter"},
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

var gitSettingsItems = []struct {
	title string
	desc  string
	key   string
}{
	{"Branch", "Default Git branch for commits", "branch"},
	{"Auto Push", "Automatically push changes to remote", "auto_push"},
	{"Auto Pull", "Automatically pull changes from remote", "auto_pull"},
	{"User Name", "Git commit author name", "user_name"},
	{"User Email", "Git commit author email", "user_email"},
	{"Commit Template", "Template for commit messages", "commit_template"},
}

var watcherSettingsItems = []struct {
	title string
	desc  string
	key   string
}{
	{"Debounce Delay", "Time to wait before processing file changes", "debounce_delay"},
	{"File Extensions", "Excel file extensions to watch", "file_extensions"},
	{"Ignore Patterns", "File patterns to ignore during watching", "ignore_patterns"},
}

var converterSettingsItems = []struct {
	title string
	desc  string
	key   string
}{
	{"Preserve Formulas", "Keep Excel formulas in JSON output", "preserve_formulas"},
	{"Preserve Styles", "Keep Excel cell styling in JSON output", "preserve_styles"},
	{"Preserve Comments", "Keep Excel cell comments in JSON output", "preserve_comments"},
	{"Compact JSON", "Generate compact JSON without extra whitespace", "compact_json"},
	{"Ignore Empty Cells", "Skip empty cells in JSON output", "ignore_empty_cells"},
	{"Max Cells Per Sheet", "Maximum number of cells to process per sheet", "max_cells_per_sheet"},
	{"Chunking Strategy", "Strategy for handling large Excel files", "chunking_strategy"},
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
				switch m.confirmType {
				case "update":
					m.updating = true
					m.status = "Updating GitCells..."
					return m, m.performUpdate()
				case "uninstall":
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

		if m.editMode {
			switch msg.String() {
			case "enter":
				return m.saveEdit()
			case "esc":
				m.editMode = false
				m.editValue = ""
				m.editKey = ""
				m.status = "Edit cancelled"
				return m, nil
			case "backspace":
				if len(m.editValue) > 0 {
					m.editValue = m.editValue[:len(m.editValue)-1]
				}
			default:
				// Add typed characters to edit value
				if len(msg.String()) == 1 {
					m.editValue += msg.String()
				}
			}
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
		switch {
		case msg.err != nil:
			m.status = fmt.Sprintf("Error checking for updates: %v", msg.err)
		case msg.hasUpdate:
			m.status = fmt.Sprintf("Update available: %s → %s", constants.Version, msg.release.TagName)
		default:
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
	case viewGit:
		return len(gitSettingsItems)
	case viewWatcher:
		return len(watcherSettingsItems)
	case viewConverter:
		return len(converterSettingsItems)
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
	case viewGit:
		title = "Git Settings"
		subtitle = "Git Integration and Repository Configuration"
		breadcrumb = "Main Menu > Git Settings"
	case viewWatcher:
		title = "Watcher Settings"
		subtitle = "File Monitoring and Watch Configuration"
		breadcrumb = "Main Menu > Watcher Settings"
	case viewConverter:
		title = "Converter Settings"
		subtitle = "Excel to JSON Conversion Options"
		breadcrumb = "Main Menu > Converter Settings"
	}

	breadcrumbStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("243")).
		Faint(true)

	s := titleStyle.Render(title) + "\n"
	s += subtitleStyle.Render(subtitle) + "\n"
	s += breadcrumbStyle.Render("Navigation: "+breadcrumb) + "\n\n"

	if m.showConfirm {
		var confirmMsg string
		switch m.confirmType {
		case "update":
			confirmMsg = "Are you sure you want to update GitCells? This will download and replace the current binary."
		case "uninstall":
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
		case viewGit:
			items = gitSettingsItems
		case viewWatcher:
			items = watcherSettingsItems
		case viewConverter:
			items = converterSettingsItems
		}

		for i, item := range items {
			cursor := "  "
			if i == m.cursor {
				cursor = cursorStyle.Render("▶ ")
			}

			// Add value display for settings
			itemTitle := item.title
			if m.config != nil {
				currentValue := m.getCurrentValue(item.key)
				if m.editMode && m.editKey == item.key && i == m.cursor {
					// Show edit input
					itemTitle += " " + cursorStyle.Render(fmt.Sprintf("[%s]", m.editValue))
				} else {
					// Show current value
					itemTitle += " " + descStyle.Render(fmt.Sprintf("(%s)", currentValue))
				}
			}

			s += fmt.Sprintf("%s%s\n", cursor, itemTitle)
			s += fmt.Sprintf("    %s\n\n", descStyle.Render(item.desc))
		}
	}

	s += statusStyle.Render("Status: "+m.status) + "\n"

	switch {
	case m.updating:
		s += descStyle.Render("Please wait...")
	case !m.showConfirm && m.editMode:
		s += descStyle.Render("Type to edit, Enter to save, Esc to cancel")
	case !m.showConfirm && m.currentView == viewMain:
		s += descStyle.Render("Use ↑/↓ or j/k to navigate, Enter to select, q to quit")
	case !m.showConfirm:
		s += descStyle.Render("Use ↑/↓ or j/k to navigate, Enter to edit/toggle, Esc to go back")
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
	case viewGit:
		selectedItem = gitSettingsItems[m.cursor]
	case viewWatcher:
		selectedItem = watcherSettingsItems[m.cursor]
	case viewConverter:
		selectedItem = converterSettingsItems[m.cursor]
	}

	switch selectedItem.key {
	// Main menu navigation
	case "git":
		m.currentView = viewGit
		m.cursor = 0
		m.status = "Entered Git settings"
		return m, nil
	case "watcher":
		m.currentView = viewWatcher
		m.cursor = 0
		m.status = "Entered Watcher settings"
		return m, nil
	case "converter":
		m.currentView = viewConverter
		m.cursor = 0
		m.status = "Entered Converter settings"
		return m, nil
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

	// Handle boolean toggles
	case "experimental", "beta_updates", "telemetry", "auto_check", "prereleases",
		"auto_download", "notify", "auto_push", "auto_pull", "preserve_formulas",
		"preserve_styles", "preserve_comments", "compact_json", "ignore_empty_cells":
		return m.toggleBooleanSetting(selectedItem.key)

	// Handle text/number edits for all other fields
	default:
		if m.config != nil && m.isEditableField(selectedItem.key) {
			m.editMode = true
			m.editKey = selectedItem.key
			m.editValue = m.getCurrentValue(selectedItem.key)
			m.status = fmt.Sprintf("Editing %s (Enter to save, Esc to cancel)", selectedItem.title)
			return m, nil
		}
	}

	return m, nil
}

func (m SettingsModel) getCurrentValue(key string) string {
	if m.config == nil {
		return "N/A"
	}

	switch key {
	// Git settings
	case "branch":
		return m.config.Git.Branch
	case "auto_push":
		return fmt.Sprintf("%t", m.config.Git.AutoPush)
	case "auto_pull":
		return fmt.Sprintf("%t", m.config.Git.AutoPull)
	case "user_name":
		return m.config.Git.UserName
	case "user_email":
		return m.config.Git.UserEmail
	case "commit_template":
		return m.config.Git.CommitTemplate

	// Watcher settings
	case "debounce_delay":
		return m.config.Watcher.DebounceDelay.String()
	case "file_extensions":
		return strings.Join(m.config.Watcher.FileExtensions, ", ")
	case "ignore_patterns":
		return strings.Join(m.config.Watcher.IgnorePatterns, ", ")

	// Converter settings
	case "preserve_formulas":
		return fmt.Sprintf("%t", m.config.Converter.PreserveFormulas)
	case "preserve_styles":
		return fmt.Sprintf("%t", m.config.Converter.PreserveStyles)
	case "preserve_comments":
		return fmt.Sprintf("%t", m.config.Converter.PreserveComments)
	case "compact_json":
		return fmt.Sprintf("%t", m.config.Converter.CompactJSON)
	case "ignore_empty_cells":
		return fmt.Sprintf("%t", m.config.Converter.IgnoreEmptyCells)
	case "max_cells_per_sheet":
		return fmt.Sprintf("%d", m.config.Converter.MaxCellsPerSheet)
	case "chunking_strategy":
		return m.config.Converter.ChunkingStrategy

	// Feature settings
	case "experimental":
		return fmt.Sprintf("%t", m.config.Features.EnableExperimentalFeatures)
	case "beta_updates":
		return fmt.Sprintf("%t", m.config.Features.EnableBetaUpdates)
	case "telemetry":
		return fmt.Sprintf("%t", m.config.Features.EnableTelemetry)

	// Update settings
	case "auto_check":
		return fmt.Sprintf("%t", m.config.Updates.AutoCheckUpdates)
	case "prereleases":
		return fmt.Sprintf("%t", m.config.Updates.IncludePrereleases)
	case "auto_download":
		return fmt.Sprintf("%t", m.config.Updates.AutoDownloadUpdates)
	case "notify":
		return fmt.Sprintf("%t", m.config.Updates.NotifyOnUpdate)

	default:
		return "Unknown"
	}
}

func (m SettingsModel) toggleBooleanSetting(key string) (tea.Model, tea.Cmd) {
	if m.config == nil {
		return m, nil
	}

	switch key {
	// Git settings
	case "auto_push":
		m.config.Git.AutoPush = !m.config.Git.AutoPush
	case "auto_pull":
		m.config.Git.AutoPull = !m.config.Git.AutoPull

	// Converter settings
	case "preserve_formulas":
		m.config.Converter.PreserveFormulas = !m.config.Converter.PreserveFormulas
	case "preserve_styles":
		m.config.Converter.PreserveStyles = !m.config.Converter.PreserveStyles
	case "preserve_comments":
		m.config.Converter.PreserveComments = !m.config.Converter.PreserveComments
	case "compact_json":
		m.config.Converter.CompactJSON = !m.config.Converter.CompactJSON
	case "ignore_empty_cells":
		m.config.Converter.IgnoreEmptyCells = !m.config.Converter.IgnoreEmptyCells

	// Feature settings
	case "experimental":
		m.config.Features.EnableExperimentalFeatures = !m.config.Features.EnableExperimentalFeatures
	case "beta_updates":
		m.config.Features.EnableBetaUpdates = !m.config.Features.EnableBetaUpdates
	case "telemetry":
		m.config.Features.EnableTelemetry = !m.config.Features.EnableTelemetry

	// Update settings
	case "auto_check":
		m.config.Updates.AutoCheckUpdates = !m.config.Updates.AutoCheckUpdates
	case "prereleases":
		m.config.Updates.IncludePrereleases = !m.config.Updates.IncludePrereleases
	case "auto_download":
		m.config.Updates.AutoDownloadUpdates = !m.config.Updates.AutoDownloadUpdates
	case "notify":
		m.config.Updates.NotifyOnUpdate = !m.config.Updates.NotifyOnUpdate
	}

	return m, m.saveConfig()
}

func (m SettingsModel) isEditableField(key string) bool {
	editableFields := []string{
		// Git text fields
		"branch", "user_name", "user_email", "commit_template",
		// Watcher fields
		"debounce_delay", "file_extensions", "ignore_patterns",
		// Converter numeric/text fields
		"max_cells_per_sheet", "chunking_strategy",
	}

	for _, field := range editableFields {
		if field == key {
			return true
		}
	}
	return false
}

func (m SettingsModel) saveEdit() (tea.Model, tea.Cmd) {
	if m.config == nil || m.editKey == "" {
		m.editMode = false
		m.status = "Edit failed - no configuration loaded"
		return m, nil
	}

	// Parse and validate the input based on field type
	switch m.editKey {
	// Git string fields
	case "branch", "user_name", "user_email", "commit_template":
		if err := m.setStringValue(m.editKey, m.editValue); err != nil {
			m.status = fmt.Sprintf("Invalid value: %v", err)
			return m, nil
		}

	// Watcher fields
	case "debounce_delay":
		if err := m.setDurationValue(m.editKey, m.editValue); err != nil {
			m.status = fmt.Sprintf("Invalid duration: %v", err)
			return m, nil
		}
	case "file_extensions", "ignore_patterns":
		if err := m.setStringSliceValue(m.editKey, m.editValue); err != nil {
			m.status = fmt.Sprintf("Invalid list: %v", err)
			return m, nil
		}

	// Converter fields
	case "max_cells_per_sheet":
		if err := m.setIntValue(m.editKey, m.editValue); err != nil {
			m.status = fmt.Sprintf("Invalid number: %v", err)
			return m, nil
		}
	case "chunking_strategy":
		if err := m.setStringValue(m.editKey, m.editValue); err != nil {
			m.status = fmt.Sprintf("Invalid value: %v", err)
			return m, nil
		}
	}

	// Reset edit mode
	m.editMode = false
	m.editKey = ""
	m.editValue = ""
	m.status = "Setting updated"

	return m, m.saveConfig()
}

func (m SettingsModel) setStringValue(key, value string) error {
	switch key {
	case "branch":
		m.config.Git.Branch = value
	case "user_name":
		m.config.Git.UserName = value
	case "user_email":
		m.config.Git.UserEmail = value
	case "commit_template":
		m.config.Git.CommitTemplate = value
	case "chunking_strategy":
		// Validate chunking strategy
		validStrategies := []string{"sheet-based", "size-based", "row-based"}
		valid := false
		for _, strategy := range validStrategies {
			if value == strategy {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid chunking strategy. Valid options: %s", strings.Join(validStrategies, ", "))
		}
		m.config.Converter.ChunkingStrategy = value
	default:
		return fmt.Errorf("unknown string field: %s", key)
	}
	return nil
}

func (m SettingsModel) setDurationValue(key, value string) error {
	duration, err := time.ParseDuration(value)
	if err != nil {
		return err
	}

	switch key {
	case "debounce_delay":
		m.config.Watcher.DebounceDelay = duration
	default:
		return fmt.Errorf("unknown duration field: %s", key)
	}
	return nil
}

func (m SettingsModel) setStringSliceValue(key, value string) error {
	// Split comma-separated values and trim whitespace
	var slice []string
	if value != "" {
		parts := strings.Split(value, ",")
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				slice = append(slice, trimmed)
			}
		}
	}

	switch key {
	case "file_extensions":
		m.config.Watcher.FileExtensions = slice
	case "ignore_patterns":
		m.config.Watcher.IgnorePatterns = slice
	default:
		return fmt.Errorf("unknown string slice field: %s", key)
	}
	return nil
}

func (m SettingsModel) setIntValue(key, value string) error {
	intVal, err := strconv.Atoi(value)
	if err != nil {
		return err
	}

	switch key {
	case "max_cells_per_sheet":
		if intVal <= 0 {
			return fmt.Errorf("value must be positive")
		}
		m.config.Converter.MaxCellsPerSheet = intVal
	default:
		return fmt.Errorf("unknown integer field: %s", key)
	}
	return nil
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
