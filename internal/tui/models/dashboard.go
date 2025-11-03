package models

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/Classic-Homes/gitcells/internal/config"
	"github.com/Classic-Homes/gitcells/internal/tui/adapter"
	"github.com/Classic-Homes/gitcells/internal/tui/messages"
	"github.com/Classic-Homes/gitcells/internal/tui/styles"

	// "github.com/Classic-Homes/gitcells/internal/watcher"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type DashboardTab int

const (
	TabOverview DashboardTab = iota
	TabFiles
	TabActivity
)

var tabNames = []string{"Overview", "Files", "Activity"}

type UnifiedDashboardModel struct {
	width  int
	height int

	// Adapters
	config      *config.Config
	gitAdapter  *adapter.GitAdapter
	convAdapter *adapter.ConverterAdapter
	// watcher     *watcher.Watcher

	// State
	activeTab    DashboardTab
	syncStatus   SyncStatus
	watcherState WatcherState
	files        []FileInfo
	activities   []Activity

	// UI state
	scrollOffset    int
	showQuickAction bool
	quickActionType string
}

type SyncStatus struct {
	IsSynced   bool
	Branch     string
	HasChanges bool
}

type WatcherState struct {
	IsRunning     bool
	StartTime     time.Time
	FilesWatched  int
	LastEvent     string
	LastEventTime time.Time
}

type FileInfo struct {
	Path         string
	Size         int64
	LastModified time.Time
	IsTracked    bool
	HasChanges   bool
	JSONPath     string
}

type Activity struct {
	Time    time.Time
	Type    string // "convert", "commit", "watch", "error"
	Message string
	Details string
}

func NewDashboardModel() *UnifiedDashboardModel {
	cfg, _ := config.Load("")
	gitAdapter, _ := adapter.NewGitAdapter("")
	convAdapter := adapter.NewConverterAdapter()

	return &UnifiedDashboardModel{
		config:      cfg,
		gitAdapter:  gitAdapter,
		convAdapter: convAdapter,
		activeTab:   TabOverview,
	}
}

func (m *UnifiedDashboardModel) Init() tea.Cmd {
	return tea.Batch(
		m.loadDashboardData(),
		tea.Tick(time.Second*5, func(t time.Time) tea.Msg {
			return tickMsg(t)
		}),
	)
}

type tickMsg time.Time

func (m *UnifiedDashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		// Global shortcuts
		switch msg.String() {
		case "q", "ctrl+c":
			if m.showQuickAction {
				m.showQuickAction = false
				return m, nil
			}
			return m, tea.Quit

		case "esc":
			if m.showQuickAction {
				m.showQuickAction = false
				return m, nil
			}
			return m, messages.RequestMainMenu()

		case "tab":
			m.activeTab = (m.activeTab + 1) % 3
			return m, nil

		case "shift+tab":
			m.activeTab = (m.activeTab + 2) % 3
			return m, nil

		// Quick actions
		case "w":
			if !m.showQuickAction {
				return m, m.toggleWatcher()
			}

		case "c":
			if !m.showQuickAction {
				m.showQuickAction = true
				m.quickActionType = "convert"
				return m, nil
			}

		case "d":
			if !m.showQuickAction {
				// Switch to diff viewer
				return m, messages.RequestModeChange("diff")
			}

		case "s":
			if !m.showQuickAction {
				// Switch to settings
				return m, messages.RequestModeChange("settings")
			}

		case "?", "h":
			if !m.showQuickAction {
				// Show help/logs
				return m, messages.RequestModeChange("errorlog")
			}

		// Navigation within tabs
		case "up", "k":
			if m.scrollOffset > 0 {
				m.scrollOffset--
			}

		case "down", "j":
			m.scrollOffset++

		case "enter":
			// Handle selection based on current tab
			return m, m.handleSelection()
		}

	case tickMsg:
		// Auto-refresh data
		return m, tea.Batch(
			m.loadDashboardData(),
			tea.Tick(time.Second*5, func(t time.Time) tea.Msg {
				return tickMsg(t)
			}),
		)
	}

	return m, nil
}

func (m *UnifiedDashboardModel) View() string {
	// Set default dimensions if not set
	if m.width == 0 {
		m.width = 80
	}
	if m.height == 0 {
		m.height = 24
	}

	// Header
	header := m.renderHeader()

	// Content area
	var content string
	switch m.activeTab {
	case TabOverview:
		content = m.renderOverview()
	case TabFiles:
		content = m.renderFiles()
	case TabActivity:
		content = m.renderActivity()
	}

	// Quick action bar
	actionBar := m.renderActionBar()

	// Calculate heights
	headerHeight := lipgloss.Height(header)
	actionBarHeight := lipgloss.Height(actionBar)
	contentHeight := m.height - headerHeight - actionBarHeight - 2

	// Style content with fixed height
	contentStyle := lipgloss.NewStyle().
		Width(m.width).
		Height(contentHeight).
		Padding(0, 2)

	styledContent := contentStyle.Render(content)

	// Combine all parts
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		styledContent,
		actionBar,
	)
}

func (m *UnifiedDashboardModel) renderHeader() string {
	// Title
	title := styles.TitleStyle.Render("GitCells Dashboard")

	// Tabs
	tabs := make([]string, 0, len(tabNames))
	for i, name := range tabNames {
		style := styles.TabStyle
		if DashboardTab(i) == m.activeTab {
			style = styles.ActiveTabStyle
		}
		tabs = append(tabs, style.Render(name))
	}
	tabBar := lipgloss.JoinHorizontal(lipgloss.Left, tabs...)

	// Status indicators
	var status []string

	// Git sync status
	syncIcon := "⚡"
	if !m.syncStatus.IsSynced {
		syncIcon = "⚠️"
	}
	status = append(status, fmt.Sprintf("%s %s", syncIcon, m.syncStatus.Branch))

	// Watcher status
	watchIcon := "👁"
	if !m.watcherState.IsRunning {
		watchIcon = "💤"
	}
	status = append(status, fmt.Sprintf("%s Watch", watchIcon))

	statusBar := styles.StatusStyle.Render(strings.Join(status, " • "))

	// Combine header elements
	headerTop := lipgloss.JoinHorizontal(
		lipgloss.Left,
		title,
		lipgloss.NewStyle().Width(m.width-lipgloss.Width(title)-lipgloss.Width(statusBar)).Render(" "),
		statusBar,
	)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		headerTop,
		tabBar,
		styles.BorderStyle.Render(strings.Repeat("─", m.width)),
	)
}

func (m *UnifiedDashboardModel) renderOverview() string {
	content := make([]string, 0, 15) // Pre-allocate for typical overview

	// Quick stats
	content = append(content, styles.SubtitleStyle.Render("📊 Quick Stats"))

	stats := []string{
		fmt.Sprintf("Excel Files: %d tracked", len(m.files)),
		fmt.Sprintf("Git Status: %s", m.syncStatus.getStatusText()),
		fmt.Sprintf("Watcher: %s", m.watcherState.getStatusText()),
		fmt.Sprintf("Last Activity: %s", m.getLastActivityTime()),
	}

	for _, stat := range stats {
		content = append(content, "  "+stat)
	}

	content = append(content, "")

	// Recent changes
	if m.syncStatus.HasChanges {
		content = append(content, styles.SubtitleStyle.Render("🔄 Pending Changes"))
		for i, file := range m.files {
			if file.HasChanges && i < 5 {
				content = append(content, fmt.Sprintf("  • %s", filepath.Base(file.Path)))
			}
		}
		content = append(content, "")
	}

	// Quick actions hint
	content = append(content, styles.SubtitleStyle.Render("⚡ Quick Actions"))
	content = append(content, "  Press 'w' to toggle watcher")
	content = append(content, "  Press 'c' to convert a file")
	content = append(content, "  Press 'd' for diff viewer")

	return strings.Join(content, "\n")
}

func (m *UnifiedDashboardModel) renderFiles() string {
	var content []string

	content = append(content, styles.SubtitleStyle.Render("📁 Tracked Files"))
	content = append(content, "")

	if len(m.files) == 0 {
		content = append(content, styles.MutedStyle.Render("  No Excel files found in watched directories"))
	} else {
		// Table header
		header := fmt.Sprintf("  %-40s %-15s %-20s %s", "File", "Size", "Modified", "Status")
		content = append(content, styles.MutedStyle.Render(header))
		content = append(content, styles.MutedStyle.Render(strings.Repeat("─", 90)))

		// File list
		start := m.scrollOffset
		end := start + (m.height - 10) // Leave room for header/footer
		if end > len(m.files) {
			end = len(m.files)
		}

		for i := start; i < end; i++ {
			file := m.files[i]
			status := "✓ Synced"
			if file.HasChanges {
				status = "● Modified"
			}
			if !file.IsTracked {
				status = "○ Untracked"
			}

			line := fmt.Sprintf("  %-40s %-15s %-20s %s",
				truncateStringUD(filepath.Base(file.Path), 40),
				formatFileSize(file.Size),
				file.LastModified.Format("Jan 2 15:04"),
				status,
			)
			content = append(content, line)
		}
	}

	return strings.Join(content, "\n")
}

func (m *UnifiedDashboardModel) renderActivity() string {
	var content []string

	content = append(content, styles.SubtitleStyle.Render("📋 Recent Activity"))
	content = append(content, "")

	if len(m.activities) == 0 {
		content = append(content, styles.MutedStyle.Render("  No recent activity"))
	} else {
		for i, activity := range m.activities {
			if i >= 20 { // Show last 20 activities
				break
			}

			icon := m.getActivityIcon(activity.Type)
			timeStr := activity.Time.Format("15:04:05")

			line := fmt.Sprintf("  %s %s %s",
				styles.MutedStyle.Render(timeStr),
				icon,
				activity.Message,
			)
			content = append(content, line)

			if activity.Details != "" {
				content = append(content, styles.MutedStyle.Render("       "+activity.Details))
			}
		}
	}

	return strings.Join(content, "\n")
}

func (m *UnifiedDashboardModel) renderActionBar() string {
	// Quick actions
	actions := []string{
		"[w] Watch",
		"[c] Convert",
		"[d] Diff",
		"[s] Settings",
		"[?] Help",
	}

	if m.watcherState.IsRunning {
		actions[0] = "[w] Stop Watch"
	}

	actionStr := strings.Join(actions, " • ")

	// Navigation help
	navHelp := "Tab: Switch tabs • ↑↓: Scroll • Esc: Menu • q: Quit"

	// Combine with styling
	actionBar := lipgloss.JoinHorizontal(
		lipgloss.Left,
		styles.ActionStyle.Render(actionStr),
		lipgloss.NewStyle().Width(m.width-lipgloss.Width(actionStr)-lipgloss.Width(navHelp)-4).Render(" "),
		styles.MutedStyle.Render(navHelp),
	)

	return lipgloss.NewStyle().
		BorderTop(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("241")).
		Width(m.width).
		Padding(0, 2).
		Render(actionBar)
}

// Helper methods

func (m *UnifiedDashboardModel) loadDashboardData() tea.Cmd {
	return func() tea.Msg {
		// Load data from adapters
		// This would be implemented to actually fetch data
		return nil
	}
}

func (m *UnifiedDashboardModel) toggleWatcher() tea.Cmd {
	return func() tea.Msg {
		// Toggle watcher state
		// TODO: Implement actual watcher start/stop logic
		m.watcherState.IsRunning = !m.watcherState.IsRunning
		return nil
	}
}

func (m *UnifiedDashboardModel) handleSelection() tea.Cmd {
	// Handle enter key based on current tab and selection
	return nil
}

func (m *UnifiedDashboardModel) getActivityIcon(activityType string) string {
	switch activityType {
	case "convert":
		return "🔄"
	case "commit":
		return "📝"
	case "watch":
		return "👁"
	case "error":
		return "❌"
	default:
		return "•"
	}
}

func (m *UnifiedDashboardModel) getLastActivityTime() string {
	if len(m.activities) > 0 {
		return m.activities[0].Time.Format("15:04")
	}
	return "No activity"
}

func (s SyncStatus) getStatusText() string {
	if s.IsSynced {
		return "✓ Synced with remote"
	}
	if s.HasChanges {
		return "● Local changes pending"
	}
	return "⚠️ Out of sync"
}

func (w WatcherState) getStatusText() string {
	if w.IsRunning {
		return fmt.Sprintf("Running (%d files)", w.FilesWatched)
	}
	return "Stopped"
}

func truncateStringUD(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

func formatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
