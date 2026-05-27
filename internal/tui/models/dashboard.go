package models

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Classic-Homes/gitcells/internal/config"
	"github.com/Classic-Homes/gitcells/internal/tui/adapter"
	"github.com/Classic-Homes/gitcells/internal/tui/styles"
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

const (
	maxActivities  = 100
	tickInterval   = 2 * time.Second
	tabCount       = 3
	headerOverhead = 4
)

type DashboardModel struct {
	width  int
	height int

	cfg            *config.Config
	gitAdapter     *adapter.GitAdapter
	watcherAdapter *adapter.WatcherAdapter

	activeTab    DashboardTab
	syncStatus   SyncStatus
	watcherState WatcherState
	files        []FileInfo

	activitiesMu sync.Mutex
	activities   []Activity

	scrollOffset int
}

type SyncStatus struct {
	Branch     string
	HasChanges bool
	InRepo     bool
}

type WatcherState struct {
	IsRunning          bool
	DirectoriesWatched int
	LastEvent          string
	LastEventTime      time.Time
}

type FileInfo struct {
	Path         string
	Size         int64
	LastModified time.Time
	HasChanges   bool
}

type Activity struct {
	Time    time.Time
	Type    string
	Message string
	Details string
}

type tickMsg time.Time
type watcherEventMsg adapter.WatcherEvent

func NewDashboardModel() *DashboardModel {
	cfg, cfgErr := config.Load("")
	if cfgErr != nil {
		cfg = config.GetDefault()
	}
	gitAdapter, _ := adapter.NewGitAdapter("")

	m := &DashboardModel{
		cfg:        cfg,
		gitAdapter: gitAdapter,
		activeTab:  TabOverview,
	}

	// Watcher adapter feeds events back into the Bubble Tea loop via a
	// channel-bounded callback. We capture m so the closure can push
	// activities. Wrapping in a goroutine-safe append (activitiesMu) keeps
	// it correct under concurrent fsnotify callbacks.
	wa, err := adapter.NewWatcherAdapter(cfg, nil, func(ev adapter.WatcherEvent) {
		m.recordActivity(Activity{
			Time:    ev.Timestamp,
			Type:    ev.Type,
			Message: ev.Message,
			Details: ev.Details,
		})
	})
	if err == nil {
		m.watcherAdapter = wa
	}
	if cfgErr != nil {
		m.recordActivity(Activity{
			Time:    time.Now(),
			Type:    "error",
			Message: "Config load failed; running with defaults",
			Details: cfgErr.Error(),
		})
	}
	return m
}

func (m *DashboardModel) Init() tea.Cmd {
	return tea.Batch(m.refresh(), tickEvery())
}

func tickEvery() tea.Cmd {
	return tea.Tick(tickInterval, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func (m *DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			if m.watcherAdapter != nil && m.watcherAdapter.IsRunning() {
				_ = m.watcherAdapter.Stop()
			}
			return m, tea.Quit
		case "tab":
			m.activeTab = (m.activeTab + 1) % tabCount
		case "shift+tab":
			m.activeTab = (m.activeTab + tabCount - 1) % tabCount
		case "w":
			return m, m.toggleWatcher()
		case "r":
			return m, m.refresh()
		case "up", "k":
			if m.scrollOffset > 0 {
				m.scrollOffset--
			}
		case "down", "j":
			m.scrollOffset++
		}

	case tickMsg:
		return m, tea.Batch(m.refresh(), tickEvery())
	}

	return m, nil
}

func (m *DashboardModel) View() string {
	if m.width == 0 {
		m.width = 80
	}
	if m.height == 0 {
		m.height = 24
	}

	header := m.renderHeader()
	actionBar := m.renderActionBar()

	var content string
	switch m.activeTab {
	case TabOverview:
		content = m.renderOverview()
	case TabFiles:
		content = m.renderFiles()
	case TabActivity:
		content = m.renderActivity()
	}

	contentHeight := m.height - lipgloss.Height(header) - lipgloss.Height(actionBar) - 2
	if contentHeight < 1 {
		contentHeight = 1
	}
	styledContent := lipgloss.NewStyle().Width(m.width).Height(contentHeight).Padding(0, 2).Render(content)

	return lipgloss.JoinVertical(lipgloss.Left, header, styledContent, actionBar)
}

func (m *DashboardModel) renderHeader() string {
	title := styles.TitleStyle.Render("GitCells")

	tabs := make([]string, 0, len(tabNames))
	for i, name := range tabNames {
		style := styles.TabStyle
		if DashboardTab(i) == m.activeTab {
			style = styles.ActiveTabStyle
		}
		tabs = append(tabs, style.Render(name))
	}
	tabBar := lipgloss.JoinHorizontal(lipgloss.Left, tabs...)

	branch := m.syncStatus.Branch
	if branch == "" {
		branch = "—"
	}
	watchTag := "watch:off"
	if m.watcherState.IsRunning {
		watchTag = fmt.Sprintf("watch:on (%d dirs)", m.watcherState.DirectoriesWatched)
	}
	statusBar := styles.StatusStyle.Render(fmt.Sprintf("%s • %s", branch, watchTag))

	headerTop := lipgloss.JoinHorizontal(
		lipgloss.Left,
		title,
		lipgloss.NewStyle().Width(maxInt(0, m.width-lipgloss.Width(title)-lipgloss.Width(statusBar))).Render(" "),
		statusBar,
	)
	return lipgloss.JoinVertical(
		lipgloss.Left,
		headerTop,
		tabBar,
		styles.BorderStyle.Render(strings.Repeat("─", m.width)),
	)
}

func (m *DashboardModel) renderOverview() string {
	lines := []string{
		styles.SubtitleStyle.Render("Status"),
		fmt.Sprintf("  Git:     %s", m.syncStatusText()),
		fmt.Sprintf("  Watcher: %s", m.watcherStatusText()),
		fmt.Sprintf("  Files:   %d tracked", len(m.files)),
		"",
		styles.SubtitleStyle.Render("Keys"),
		"  w   toggle watcher",
		"  r   refresh",
		"  tab switch tab",
		"  q   quit",
	}
	return strings.Join(lines, "\n")
}

func (m *DashboardModel) renderFiles() string {
	lines := []string{styles.SubtitleStyle.Render("Tracked Excel files"), ""}
	if len(m.files) == 0 {
		lines = append(lines, styles.MutedStyle.Render("  none"))
		return strings.Join(lines, "\n")
	}
	lines = append(lines,
		styles.MutedStyle.Render(fmt.Sprintf("  %-40s %-12s %-20s %s", "File", "Size", "Modified", "Status")),
		styles.MutedStyle.Render("  "+strings.Repeat("─", 88)),
	)
	end := m.scrollOffset + (m.height - 10)
	if end > len(m.files) {
		end = len(m.files)
	}
	for i := m.scrollOffset; i < end; i++ {
		f := m.files[i]
		status := "synced"
		if f.HasChanges {
			status = "modified"
		}
		lines = append(lines, fmt.Sprintf("  %-40s %-12s %-20s %s",
			truncate(filepath.Base(f.Path), 40),
			formatBytes(f.Size),
			f.LastModified.Format("Jan 2 15:04"),
			status,
		))
	}
	return strings.Join(lines, "\n")
}

func (m *DashboardModel) renderActivity() string {
	m.activitiesMu.Lock()
	defer m.activitiesMu.Unlock()

	lines := []string{styles.SubtitleStyle.Render("Recent activity"), ""}
	if len(m.activities) == 0 {
		lines = append(lines, styles.MutedStyle.Render("  (none yet)"))
		return strings.Join(lines, "\n")
	}
	limit := 20
	if len(m.activities) < limit {
		limit = len(m.activities)
	}
	for i := 0; i < limit; i++ {
		a := m.activities[i]
		lines = append(lines, fmt.Sprintf("  %s  %s  %s",
			styles.MutedStyle.Render(a.Time.Format("15:04:05")),
			activityIcon(a.Type),
			a.Message,
		))
		if a.Details != "" {
			lines = append(lines, styles.MutedStyle.Render("           "+a.Details))
		}
	}
	return strings.Join(lines, "\n")
}

func (m *DashboardModel) renderActionBar() string {
	watchLabel := "[w] start watch"
	if m.watcherState.IsRunning {
		watchLabel = "[w] stop watch"
	}
	actions := strings.Join([]string{watchLabel, "[r] refresh", "[tab] tab", "[q] quit"}, " • ")
	return lipgloss.NewStyle().
		BorderTop(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("241")).
		Width(m.width).
		Padding(0, 2).
		Render(styles.ActionStyle.Render(actions))
}

func (m *DashboardModel) refresh() tea.Cmd {
	return func() tea.Msg {
		if m.gitAdapter != nil {
			if branch, err := m.gitAdapter.GetCurrentBranch(); err == nil {
				m.syncStatus.Branch = branch
				m.syncStatus.InRepo = true
			}
			if clean, err := m.gitAdapter.IsClean(); err == nil {
				m.syncStatus.HasChanges = !clean
			}
		}
		if m.watcherAdapter != nil {
			s := m.watcherAdapter.GetStatus()
			m.watcherState.IsRunning = s.IsRunning
			m.watcherState.DirectoriesWatched = s.DirectoriesWatched
			m.watcherState.LastEvent = s.LastEvent
			m.watcherState.LastEventTime = s.LastEventTime
		}
		return nil
	}
}

func (m *DashboardModel) toggleWatcher() tea.Cmd {
	return func() tea.Msg {
		if m.watcherAdapter == nil {
			m.recordActivity(Activity{Time: time.Now(), Type: "error", Message: "Watcher unavailable"})
			return nil
		}
		var err error
		if m.watcherAdapter.IsRunning() {
			err = m.watcherAdapter.Stop()
		} else {
			err = m.watcherAdapter.Start()
		}
		if err != nil {
			m.recordActivity(Activity{Time: time.Now(), Type: "error", Message: err.Error()})
		}
		return nil
	}
}

func (m *DashboardModel) recordActivity(a Activity) {
	m.activitiesMu.Lock()
	defer m.activitiesMu.Unlock()
	m.activities = append([]Activity{a}, m.activities...)
	if len(m.activities) > maxActivities {
		m.activities = m.activities[:maxActivities]
	}
}

func (m *DashboardModel) syncStatusText() string {
	if !m.syncStatus.InRepo {
		return "not a git repo"
	}
	if m.syncStatus.HasChanges {
		return "uncommitted changes"
	}
	return "clean"
}

func (m *DashboardModel) watcherStatusText() string {
	if m.watcherState.IsRunning {
		return fmt.Sprintf("running (%d dirs)", m.watcherState.DirectoriesWatched)
	}
	return "stopped"
}

func activityIcon(kind string) string {
	switch kind {
	case "started", "stopped":
		return "•"
	case "file_changed":
		return "→"
	case "error":
		return "!"
	default:
		return "·"
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max < 4 {
		return s[:max]
	}
	return s[:max-3] + "..."
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
