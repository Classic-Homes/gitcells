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
	"github.com/Classic-Homes/gitcells/internal/tui/components"
	"github.com/Classic-Homes/gitcells/internal/tui/messages"
	"github.com/Classic-Homes/gitcells/internal/tui/styles"
	"github.com/Classic-Homes/gitcells/internal/updater"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type SettingsTab int

const (
	TabGeneral SettingsTab = iota
	TabGit
	TabWatcher
	TabAdvanced
)

var settingsTabNames = []string{"General", "Git", "Watcher", "Advanced"}

type SettingsModelV2 struct {
	width        int
	height       int
	activeTab    SettingsTab
	cursor       int
	config       *config.Config
	status       string
	updating     bool
	showConfirm  bool
	confirmType  string
	editMode     bool
	editValue    string
	editKey      string
	scrollOffset int
}

type generalSetting struct {
	title  string
	key    string
	action string // "edit", "toggle", "action"
	value  string
}

// Message types for settings model
type configLoadedMsg struct {
	config *config.Config
	err    error
}
type configSavedMsg struct {
	err error
}
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

func NewSettingsModel() *SettingsModelV2 {
	cfg, _ := config.Load("")
	return &SettingsModelV2{
		config:    cfg,
		activeTab: TabGeneral,
		width:     80,
		height:    24,
	}
}

func (m *SettingsModelV2) Init() tea.Cmd {
	return m.loadConfig()
}

func (m *SettingsModelV2) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		if m.editMode {
			return m.handleEditMode(msg)
		}
		if m.showConfirm {
			return m.handleConfirmMode(msg)
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "esc":
			return m, messages.RequestMainMenu()

		case "tab":
			m.activeTab = (m.activeTab + 1) % 4
			m.cursor = 0
			m.scrollOffset = 0
			return m, nil

		case "shift+tab":
			m.activeTab = (m.activeTab + 3) % 4
			m.cursor = 0
			m.scrollOffset = 0
			return m, nil

		// Quick actions
		case "u":
			// Check for updates
			return m, m.checkForUpdates()

		case "s":
			// Save config
			return m, m.saveConfig()

		case "r":
			// Reload config
			return m, m.loadConfig()

		// Navigation
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				// Adjust scroll if needed
				if m.cursor < m.scrollOffset {
					m.scrollOffset = m.cursor
				}
			}

		case "down", "j":
			items := m.getCurrentTabItems()
			if m.cursor < len(items)-1 {
				m.cursor++
				// Adjust scroll if needed
				visibleHeight := m.height - 12 // Account for header/footer
				if m.cursor >= m.scrollOffset+visibleHeight {
					m.scrollOffset = m.cursor - visibleHeight + 1
				}
			}

		case "enter":
			return m.handleSelection()
		}

	case configLoadedMsg:
		if msg.err != nil {
			m.status = fmt.Sprintf("Error loading config: %v", msg.err)
		} else {
			m.config = msg.config
			m.status = "Configuration loaded"
		}
		return m, nil

	case configSavedMsg:
		if msg.err != nil {
			m.status = fmt.Sprintf("Error saving config: %v", msg.err)
		} else {
			m.status = "Configuration saved successfully"
		}
		return m, nil

	case updateCheckMsg:
		m.updating = false
		switch {
		case msg.err != nil:
			m.status = fmt.Sprintf("Error checking for updates: %v", msg.err)
		case msg.hasUpdate:
			m.status = fmt.Sprintf("Update available: %s → %s", constants.Version, msg.release.TagName)
		default:
			m.status = "You're running the latest version"
		}
		return m, nil
	}

	return m, nil
}

func (m *SettingsModelV2) View() string {
	// Header with breadcrumb
	breadcrumb := []string{"Dashboard", "Settings", settingsTabNames[m.activeTab]}
	header := components.RenderHeader("Settings", breadcrumb, m.width)

	// Tabs
	tabs := make([]string, 0, len(settingsTabNames))
	for i, name := range settingsTabNames {
		style := styles.TabStyle
		if SettingsTab(i) == m.activeTab {
			style = styles.ActiveTabStyle
		}
		tabs = append(tabs, style.Render(name))
	}
	tabBar := lipgloss.JoinHorizontal(lipgloss.Left, tabs...)

	// Content
	content := m.renderCurrentTab()

	// Status bar
	statusBar := ""
	if m.status != "" {
		statusBar = styles.StatusStyle.Render(m.status)
	}

	// Action bar
	actions := []components.ActionBarItem{
		{Key: "u", Label: "Check Updates", Active: true},
		{Key: "s", Label: "Save", Active: m.config != nil},
		{Key: "r", Label: "Reload", Active: true},
		{Key: "esc", Label: "Back", Active: true},
	}

	helpText := "Tab: Switch tabs • ↑↓: Navigate • Enter: Select"
	actionBar := components.RenderActionBar(m.width, actions, helpText)

	// Calculate heights
	headerHeight := lipgloss.Height(header)
	tabBarHeight := lipgloss.Height(tabBar)
	statusBarHeight := lipgloss.Height(statusBar)
	actionBarHeight := lipgloss.Height(actionBar)
	contentHeight := m.height - headerHeight - tabBarHeight - statusBarHeight - actionBarHeight - 3

	// Style content with fixed height
	contentStyle := lipgloss.NewStyle().
		Width(m.width).
		Height(contentHeight).
		Padding(1, 2)

	styledContent := contentStyle.Render(content)

	// Combine all parts
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		tabBar,
		styledContent,
		statusBar,
		actionBar,
	)
}

func (m *SettingsModelV2) renderCurrentTab() string {
	switch m.activeTab {
	case TabGeneral:
		return m.renderGeneralTab()
	case TabGit:
		return m.renderGitTab()
	case TabWatcher:
		return m.renderWatcherTab()
	case TabAdvanced:
		return m.renderAdvancedTab()
	}
	return ""
}

func (m *SettingsModelV2) renderGeneralTab() string {
	settings := []generalSetting{
		{title: "Check for Updates", key: "update_check", action: "action"},
		{title: "Update GitCells", key: "update", action: "action"},
		{title: "Auto-check Updates", key: "auto_update", action: "toggle", value: m.getConfigBool("updates.auto_check")},
		{title: "Version Info", key: "version", action: "action", value: constants.Version},
		{title: "Uninstall GitCells", key: "uninstall", action: "action"},
	}

	return m.renderSettingsList(settings)
}

func (m *SettingsModelV2) renderGitTab() string {
	settings := []generalSetting{
		{title: "Auto-commit", key: "git.auto_commit", action: "toggle", value: m.getConfigBool("git.auto_commit")},
		{title: "Commit Template", key: "git.commit_template", action: "edit", value: m.getConfigString("git.commit_template")},
		{title: "Default Branch", key: "git.branch", action: "edit", value: m.getConfigString("git.branch")},
		{title: "Push After Commit", key: "git.push_after_commit", action: "toggle", value: m.getConfigBool("git.push_after_commit")},
	}

	return m.renderSettingsList(settings)
}

func (m *SettingsModelV2) renderWatcherTab() string {
	settings := []generalSetting{
		{title: "Watch on Startup", key: "watcher.auto_start", action: "toggle", value: m.getConfigBool("watcher.auto_start")},
		{title: "Debounce Delay", key: "watcher.debounce", action: "edit", value: m.getConfigString("watcher.debounce_seconds")},
		{title: "Include Patterns", key: "watcher.patterns", action: "edit", value: strings.Join(m.getConfigStringSlice("watcher.patterns"), ", ")},
		{title: "Watched Directories", key: "watched_dirs", action: "action", value: fmt.Sprintf("%d directories", len(m.getConfigStringSlice("watched_dirs")))},
	}

	return m.renderSettingsList(settings)
}

func (m *SettingsModelV2) renderAdvancedTab() string {
	settings := []generalSetting{
		{title: "Experimental Features", key: "features.experimental", action: "toggle", value: m.getConfigBool("features.experimental")},
		{title: "Beta Updates", key: "updates.beta", action: "toggle", value: m.getConfigBool("updates.beta")},
		{title: "Telemetry", key: "telemetry.enabled", action: "toggle", value: m.getConfigBool("telemetry.enabled")},
		{title: "Chunk Size", key: "converter.chunk_size", action: "edit", value: m.getConfigString("converter.chunk_size")},
		{title: "Debug Mode", key: "debug", action: "toggle", value: m.getConfigBool("debug")},
	}

	return m.renderSettingsList(settings)
}

func (m *SettingsModelV2) renderSettingsList(settings []generalSetting) string {
	var lines []string

	// Apply scrolling
	start := m.scrollOffset
	visibleHeight := m.height - 12
	end := start + visibleHeight
	if end > len(settings) {
		end = len(settings)
	}

	for i := start; i < end; i++ {
		setting := settings[i]
		actualIndex := i

		cursor := "  "
		if actualIndex == m.cursor {
			cursor = styles.HighlightStyle.Render("▶ ")
		}

		// Format the setting line
		var line string
		switch setting.action {
		case "toggle":
			checkbox := "☐"
			if setting.value == "true" {
				checkbox = "☑"
			}
			line = fmt.Sprintf("%s%s %s", cursor, checkbox, setting.title)
		case "edit":
			valueStr := setting.value
			if valueStr == "" {
				valueStr = styles.MutedStyle.Render("(not set)")
			} else if len(valueStr) > 30 {
				valueStr = valueStr[:27] + "..."
			}
			line = fmt.Sprintf("%s%s: %s", cursor, setting.title, styles.InfoStyle.Render(valueStr))
		case "action":
			line = fmt.Sprintf("%s%s", cursor, setting.title)
			if setting.value != "" {
				line += " " + styles.MutedStyle.Render(fmt.Sprintf("(%s)", setting.value))
			}
		}

		lines = append(lines, line)
	}

	// Show scroll indicators
	if m.scrollOffset > 0 {
		lines = append([]string{styles.MutedStyle.Render("  ↑ More items above")}, lines...)
	}
	if end < len(settings) {
		lines = append(lines, styles.MutedStyle.Render("  ↓ More items below"))
	}

	return strings.Join(lines, "\n")
}

// Helper methods
func (m *SettingsModelV2) getCurrentTabItems() []generalSetting {
	switch m.activeTab {
	case TabGeneral:
		return []generalSetting{
			{title: "Check for Updates", key: "update_check", action: "action"},
			{title: "Update GitCells", key: "update", action: "action"},
			{title: "Auto-check Updates", key: "auto_update", action: "toggle"},
			{title: "Version Info", key: "version", action: "action"},
			{title: "Uninstall GitCells", key: "uninstall", action: "action"},
		}
	case TabGit:
		return []generalSetting{
			{title: "Auto Push", key: "git.auto_push", action: "toggle"},
			{title: "Auto Pull", key: "git.auto_pull", action: "toggle"},
			{title: "Commit Template", key: "git.commit_template", action: "edit"},
			{title: "Default Branch", key: "git.branch", action: "edit"},
			{title: "User Name", key: "git.user_name", action: "edit"},
			{title: "User Email", key: "git.user_email", action: "edit"},
		}
	case TabWatcher:
		return []generalSetting{
			{title: "Debounce Delay", key: "watcher.debounce_delay", action: "edit"},
			{title: "File Extensions", key: "watcher.file_extensions", action: "edit"},
			{title: "Ignore Patterns", key: "watcher.ignore_patterns", action: "edit"},
			{title: "Watched Directories", key: "watcher.directories", action: "action"},
		}
	case TabAdvanced:
		return []generalSetting{
			{title: "Experimental Features", key: "features.enable_experimental_features", action: "toggle"},
			{title: "Beta Updates", key: "features.enable_beta_updates", action: "toggle"},
			{title: "Telemetry", key: "features.enable_telemetry", action: "toggle"},
			{title: "Max Cells per Sheet", key: "converter.max_cells_per_sheet", action: "edit"},
			{title: "Preserve Formulas", key: "converter.preserve_formulas", action: "toggle"},
			{title: "Preserve Styles", key: "converter.preserve_styles", action: "toggle"},
		}
	}
	return nil
}

func (m *SettingsModelV2) handleSelection() (tea.Model, tea.Cmd) {
	items := m.getCurrentTabItems()
	if m.cursor >= len(items) {
		return m, nil
	}

	selected := items[m.cursor]
	switch selected.action {
	case "toggle":
		// Toggle boolean value
		current := m.getConfigBool(selected.key)
		newValue := current != "true"
		m.setConfigBool(selected.key, newValue)
		m.status = fmt.Sprintf("Toggled %s", selected.title)
		return m, nil

	case "edit":
		// Enter edit mode
		m.editMode = true
		m.editKey = selected.key
		m.editValue = m.getConfigString(selected.key)
		return m, nil

	case "action":
		// Handle special actions
		switch selected.key {
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
			m.status = fmt.Sprintf("GitCells %s (built %s)", constants.Version, constants.BuildTime)
			return m, nil
		case "watched_dirs":
			// Could switch to a watched dirs management view
			return m, nil
		}
	}

	return m, nil
}

func (m *SettingsModelV2) handleEditMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.editMode = false
		m.editKey = ""
		m.editValue = ""
		return m, nil
	case "enter":
		// Save the edited value
		m.setConfigString(m.editKey, m.editValue)
		m.status = fmt.Sprintf("Updated %s", m.editKey)
		m.editMode = false
		m.editKey = ""
		m.editValue = ""
		return m, nil
	case "backspace":
		if len(m.editValue) > 0 {
			m.editValue = m.editValue[:len(m.editValue)-1]
		}
	default:
		// Add character to edit value
		if len(msg.String()) == 1 {
			m.editValue += msg.String()
		}
	}
	return m, nil
}

func (m *SettingsModelV2) handleConfirmMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		m.showConfirm = false
		switch m.confirmType {
		case "update":
			m.updating = true
			m.status = "Updating GitCells..."
			return m, m.performUpdate()
		case "uninstall":
			return m, m.performUninstall()
		}
	case "n", "N", "esc":
		m.showConfirm = false
		m.confirmType = ""
		return m, nil
	}
	return m, nil
}

// Config helper methods
func (m *SettingsModelV2) getConfigBool(key string) string {
	if m.config == nil {
		return "false"
	}
	parts := strings.Split(key, ".")
	var value interface{} = m.config

	// Navigate nested config
	for _, part := range parts {
		switch v := value.(type) {
		case *config.Config:
			switch part {
			case "git":
				value = v.Git
			case "watcher":
				value = v.Watcher
			case "converter":
				value = v.Converter
			default:
				return "false"
			}
		case config.GitConfig:
			switch part {
			case "auto_push":
				return fmt.Sprintf("%t", v.AutoPush)
			case "auto_pull":
				return fmt.Sprintf("%t", v.AutoPull)
			default:
				return "false"
			}
		case config.WatcherConfig:
			// No boolean fields in WatcherConfig
			return "false"
		default:
			return "false"
		}
	}
	return "false"
}

func (m *SettingsModelV2) getConfigString(key string) string {
	if m.config == nil {
		return ""
	}

	switch key {
	case "git.commit_template":
		return m.config.Git.CommitTemplate
	case "git.branch":
		return m.config.Git.Branch
	case "git.user_name":
		return m.config.Git.UserName
	case "git.user_email":
		return m.config.Git.UserEmail
	case "watcher.debounce_delay":
		return m.config.Watcher.DebounceDelay.String()
	case "converter.max_cells_per_sheet":
		return fmt.Sprintf("%d", m.config.Converter.MaxCellsPerSheet)
	}

	return ""
}

func (m *SettingsModelV2) getConfigStringSlice(key string) []string {
	if m.config == nil {
		return nil
	}

	switch key {
	case "watcher.file_extensions":
		return m.config.Watcher.FileExtensions
	case "watcher.directories":
		return m.config.Watcher.Directories
	case "watcher.ignore_patterns":
		return m.config.Watcher.IgnorePatterns
	}

	return nil
}

func (m *SettingsModelV2) setConfigBool(key string, value bool) {
	if m.config == nil {
		return
	}

	switch key {
	case "git.auto_push":
		m.config.Git.AutoPush = value
	case "git.auto_pull":
		m.config.Git.AutoPull = value
	case "features.enable_experimental_features":
		m.config.Features.EnableExperimentalFeatures = value
	case "features.enable_beta_updates":
		m.config.Features.EnableBetaUpdates = value
	case "features.enable_telemetry":
		m.config.Features.EnableTelemetry = value
	case "updates.auto_check_updates":
		m.config.Updates.AutoCheckUpdates = value
	case "converter.preserve_formulas":
		m.config.Converter.PreserveFormulas = value
	case "converter.preserve_styles":
		m.config.Converter.PreserveStyles = value
	}
}

func (m *SettingsModelV2) setConfigString(key string, value string) {
	if m.config == nil {
		return
	}

	switch key {
	case "git.commit_template":
		m.config.Git.CommitTemplate = value
	case "git.branch":
		m.config.Git.Branch = value
	case "git.user_name":
		m.config.Git.UserName = value
	case "git.user_email":
		m.config.Git.UserEmail = value
	case "watcher.debounce_delay":
		if duration, err := time.ParseDuration(value); err == nil {
			m.config.Watcher.DebounceDelay = duration
		}
	case "converter.max_cells_per_sheet":
		if size, err := strconv.Atoi(value); err == nil {
			m.config.Converter.MaxCellsPerSheet = size
		}
	}
}

// Command methods
func (m *SettingsModelV2) loadConfig() tea.Cmd {
	return func() tea.Msg {
		cfg, err := config.Load("")
		return configLoadedMsg{config: cfg, err: err}
	}
}

func (m *SettingsModelV2) saveConfig() tea.Cmd {
	return func() tea.Msg {
		if m.config == nil {
			return configSavedMsg{err: fmt.Errorf("no config loaded")}
		}
		// Save config to file
		err := fmt.Errorf("save not implemented") // TODO: implement config save
		return configSavedMsg{err: err}
	}
}

func (m *SettingsModelV2) checkForUpdates() tea.Cmd {
	return func() tea.Msg {
		updater := updater.New("")
		release, hasUpdate, err := updater.CheckForUpdate()
		return updateCheckMsg{release: release, hasUpdate: hasUpdate, err: err}
	}
}

func (m *SettingsModelV2) performUpdate() tea.Cmd {
	return func() tea.Msg {
		updater := updater.New("")
		_, _, err := updater.CheckForUpdate()
		if err != nil {
			return updateCompleteMsg{err: err}
		}
		// Would perform actual update here
		return updateCompleteMsg{err: nil}
	}
}

func (m *SettingsModelV2) performUninstall() tea.Cmd {
	return func() tea.Msg {
		// Get binary path
		binaryPath, err := os.Executable()
		if err != nil {
			return uninstallCompleteMsg{err: err}
		}

		// Remove binary
		if err := os.Remove(binaryPath); err != nil {
			return uninstallCompleteMsg{err: err}
		}

		// Remove from PATH if in /usr/local/bin
		if strings.Contains(binaryPath, "/usr/local/bin") {
			_ = exec.Command("sudo", "rm", "-f", binaryPath).Run()
		}

		return uninstallCompleteMsg{err: nil}
	}
}
