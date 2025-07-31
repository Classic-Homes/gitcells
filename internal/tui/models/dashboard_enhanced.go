package models

import (
	"fmt"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/Classic-Homes/gitcells/internal/config"
	"github.com/Classic-Homes/gitcells/internal/tui/adapter"
	"github.com/Classic-Homes/gitcells/internal/tui/components"
	"github.com/Classic-Homes/gitcells/internal/tui/styles"
)

type DashboardEnhancedModel struct {
	width      int
	height     int
	config     *config.Config
	gitAdapter *adapter.GitAdapter
	convAdapter *adapter.ConverterAdapter
	
	// Dashboard data
	watching      []string
	totalFiles    int
	syncStatus    SyncStatus
	operations    []FileOperation
	recentCommits []CommitInfo
	
	// UI state
	selectedTab   int
	scrollOffset  int
	showHelp      bool
	
	// Progress tracking
	progressBars  components.MultiProgress
	lastUpdate    time.Time
}

type SyncStatus struct {
	Branch      string
	IsSynced    bool
	LastCommit  time.Time
	HasChanges  bool
	RemoteAhead int
	RemoteBehind int
}

type FileOperation struct {
	ID        string
	Type      OperationType
	FileName  string
	Status    OperationStatus
	Progress  int
	StartTime time.Time
	Error     error
}

type OperationType int
const (
	OpConvert OperationType = iota
	OpSync
	OpWatch
)

type OperationStatus int
const (
	StatusPending OperationStatus = iota
	StatusInProgress
	StatusCompleted
	StatusFailed
)

type CommitInfo struct {
	Hash    string
	Message string
	Time    time.Time
	Files   int
}

func NewDashboardEnhancedModel() tea.Model {
	m := &DashboardEnhancedModel{
		operations:   []FileOperation{},
		progressBars: components.NewMultiProgress(),
		lastUpdate:   time.Now(),
	}
	
	// Try to load configuration
	if cfg, err := config.Load("."); err == nil {
		m.config = cfg
		m.watching = cfg.Watcher.Directories
	}
	
	// Initialize adapters
	if gitAdapter, err := adapter.NewGitAdapter("."); err == nil {
		m.gitAdapter = gitAdapter
	}
	m.convAdapter = adapter.NewConverterAdapter()
	
	return m
}

func (m DashboardEnhancedModel) Init() tea.Cmd {
	return tea.Batch(
		m.loadInitialData(),
		dashboardTick(),
	)
}

func (m DashboardEnhancedModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case dashboardTickMsg:
		// Update progress bars
		for i := range m.operations {
			if m.operations[i].Status == StatusInProgress {
				m.operations[i].Progress += 10
				if m.operations[i].Progress >= 100 {
					m.operations[i].Progress = 100
					m.operations[i].Status = StatusCompleted
				}
				m.progressBars.UpdateBar(m.operations[i].ID, m.operations[i].Progress)
			}
		}
		
		// Refresh data every 5 seconds
		if time.Since(m.lastUpdate) > 5*time.Second {
			cmds = append(cmds, m.refreshData())
			m.lastUpdate = time.Now()
		}
		
		cmds = append(cmds, dashboardTick())

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "tab":
			m.selectedTab = (m.selectedTab + 1) % 3
		case "?":
			m.showHelp = !m.showHelp
		case "r":
			cmds = append(cmds, m.refreshData())
		case "c":
			cmds = append(cmds, m.startConversion())
		case "s":
			cmds = append(cmds, m.syncRepository())
		case "up", "k":
			if m.scrollOffset > 0 {
				m.scrollOffset--
			}
		case "down", "j":
			m.scrollOffset++
		}
		
	case dataLoadedMsg:
		m.applyLoadedData(msg)
		
	case operationUpdateMsg:
		m.updateOperation(msg)
	}

	return m, tea.Batch(cmds...)
}

func (m DashboardEnhancedModel) View() string {
	if m.showHelp {
		return m.renderHelp()
	}

	// Main container
	containerStyle := lipgloss.NewStyle().
		Padding(1, 2).
		Width(m.width).
		Height(m.height)

	// Header
	header := m.renderHeader()
	
	// Tab bar
	tabs := m.renderTabs()
	
	// Content based on selected tab
	var content string
	switch m.selectedTab {
	case 0:
		content = m.renderOverview()
	case 1:
		content = m.renderOperations()
	case 2:
		content = m.renderCommits()
	}
	
	// Footer
	footer := m.renderFooter()
	
	// Combine all sections
	fullView := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		tabs,
		content,
		footer,
	)
	
	return containerStyle.Render(fullView)
}

func (m DashboardEnhancedModel) renderHeader() string {
	titleStyle := styles.TitleStyle.Copy().
		MarginBottom(1)
		
	statusIcon := "🔴"
	statusText := "Not Connected"
	statusColor := styles.Error
	
	if m.gitAdapter != nil {
		if m.syncStatus.IsSynced {
			statusIcon = "🟢"
			statusText = "Synced"
			statusColor = styles.Success
		} else if m.syncStatus.HasChanges {
			statusIcon = "🟡"
			statusText = "Changes Pending"
			statusColor = styles.Warning
		}
	}
	
	statusStyle := lipgloss.NewStyle().
		Foreground(statusColor)
		
	title := titleStyle.Render("GitCells Dashboard")
	status := statusStyle.Render(fmt.Sprintf("%s %s | Branch: %s", statusIcon, statusText, m.syncStatus.Branch))
	
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		title,
		lipgloss.NewStyle().Width(m.width-40).Render(" "),
		status,
	)
}

func (m DashboardEnhancedModel) renderTabs() string {
	tabs := []string{"Overview", "Operations", "Commits"}
	
	var renderedTabs []string
	for i, tab := range tabs {
		style := lipgloss.NewStyle().
			Padding(0, 2).
			Foreground(styles.Muted)
			
		if i == m.selectedTab {
			style = style.
				Foreground(styles.Primary).
				Bold(true).
				BorderBottom(true).
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(styles.Primary)
		}
		
		renderedTabs = append(renderedTabs, style.Render(tab))
	}
	
	tabBar := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
	
	return lipgloss.NewStyle().
		MarginBottom(1).
		BorderBottom(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(styles.Muted).
		Width(m.width-4).
		Render(tabBar)
}

func (m DashboardEnhancedModel) renderOverview() string {
	// Stats boxes
	statsStyle := styles.BoxStyle.Copy().
		Width(30).
		Height(5).
		MarginRight(2)
		
	watchingBox := statsStyle.Render(
		styles.SubtitleStyle.Render("Watching") + "\n" +
		fmt.Sprintf("%d directories\n", len(m.watching)) +
		fmt.Sprintf("%d Excel files", m.totalFiles),
	)
	
	syncBox := statsStyle.Render(
		styles.SubtitleStyle.Render("Sync Status") + "\n" +
		fmt.Sprintf("Last commit: %s ago\n", formatDuration(time.Since(m.syncStatus.LastCommit))) +
		fmt.Sprintf("Remote: ↑%d ↓%d", m.syncStatus.RemoteAhead, m.syncStatus.RemoteBehind),
	)
	
	conversionStats, _ := m.convAdapter.GetConversionStats(".")
	convBox := statsStyle.Render(
		styles.SubtitleStyle.Render("Conversions") + "\n" +
		fmt.Sprintf("Converted: %d/%d\n", conversionStats.ConvertedFiles, conversionStats.TotalExcelFiles) +
		fmt.Sprintf("Pending: %d", conversionStats.PendingConversions),
	)
	
	statsRow := lipgloss.JoinHorizontal(lipgloss.Top, watchingBox, syncBox, convBox)
	
	// Recent activity
	activityStyle := styles.BoxStyle.Copy().
		Width(m.width - 8).
		MarginTop(2)
		
	activityContent := styles.SubtitleStyle.Render("Recent Activity") + "\n\n"
	
	for i, op := range m.operations {
		if i >= 5 {
			break
		}
		
		icon := m.getOperationIcon(op)
		status := m.getOperationStatus(op)
		duration := formatDuration(time.Since(op.StartTime))
		
		activityContent += fmt.Sprintf("%s %s - %s (%s)\n", icon, op.FileName, status, duration)
	}
	
	activityBox := activityStyle.Render(activityContent)
	
	return lipgloss.JoinVertical(lipgloss.Left, statsRow, activityBox)
}

func (m DashboardEnhancedModel) renderOperations() string {
	if len(m.operations) == 0 {
		return styles.MutedStyle.Render("No operations in progress")
	}
	
	// Show progress bars for active operations
	content := ""
	for _, op := range m.operations {
		if op.Status == StatusInProgress {
			content += m.renderOperation(op) + "\n\n"
		}
	}
	
	// Show completed operations
	content += styles.SubtitleStyle.Render("Completed Operations") + "\n\n"
	for _, op := range m.operations {
		if op.Status == StatusCompleted || op.Status == StatusFailed {
			content += m.renderOperation(op) + "\n"
		}
	}
	
	return content
}

func (m DashboardEnhancedModel) renderOperation(op FileOperation) string {
	icon := m.getOperationIcon(op)
	
	if op.Status == StatusInProgress {
		progressBar := components.NewProgressBar(100)
		progressBar.SetLabel(fmt.Sprintf("%s %s", icon, op.FileName))
		progressBar.SetProgress(op.Progress)
		progressBar.SetWidth(60)
		return progressBar.View()
	}
	
	status := m.getOperationStatus(op)
	duration := formatDuration(time.Since(op.StartTime))
	
	style := lipgloss.NewStyle()
	if op.Status == StatusCompleted {
		style = style.Foreground(styles.Success)
	} else if op.Status == StatusFailed {
		style = style.Foreground(styles.Error)
	}
	
	return style.Render(fmt.Sprintf("%s %s - %s (%s)", icon, op.FileName, status, duration))
}

func (m DashboardEnhancedModel) renderCommits() string {
	if len(m.recentCommits) == 0 {
		return styles.MutedStyle.Render("No recent commits")
	}
	
	content := styles.SubtitleStyle.Render("Recent Commits") + "\n\n"
	
	for _, commit := range m.recentCommits {
		timeAgo := formatDuration(time.Since(commit.Time))
		content += fmt.Sprintf("• %s\n", styles.MutedStyle.Render(commit.Hash[:7]))
		content += fmt.Sprintf("  %s\n", commit.Message)
		content += fmt.Sprintf("  %s (%d files)\n\n", 
			styles.MutedStyle.Render(timeAgo+" ago"),
			commit.Files,
		)
	}
	
	return content
}

func (m DashboardEnhancedModel) renderFooter() string {
	helpText := "[Tab] Switch tabs • [r] Refresh • [c] Convert • [s] Sync • [?] Help • [q] Quit"
	return styles.HelpStyle.Copy().
		MarginTop(1).
		Render(helpText)
}

func (m DashboardEnhancedModel) renderHelp() string {
	helpBox := styles.BoxStyle.Copy().
		Width(60).
		Render(
			styles.TitleStyle.Render("GitCells Dashboard Help") + "\n\n" +
			"Navigation:\n" +
			"  Tab        - Switch between tabs\n" +
			"  ↑/↓ or j/k - Scroll content\n" +
			"  q          - Quit dashboard\n\n" +
			"Actions:\n" +
			"  r - Refresh data\n" +
			"  c - Start conversion of pending files\n" +
			"  s - Sync with remote repository\n" +
			"  w - Start/stop file watcher\n\n" +
			"Press ? to close this help",
		)
		
	return styles.Center(m.width, m.height, helpBox)
}

// Helper methods
func (m DashboardEnhancedModel) getOperationIcon(op FileOperation) string {
	switch op.Type {
	case OpConvert:
		return "🔄"
	case OpSync:
		return "📤"
	case OpWatch:
		return "👁️"
	default:
		return "•"
	}
}

func (m DashboardEnhancedModel) getOperationStatus(op FileOperation) string {
	switch op.Status {
	case StatusPending:
		return "Pending"
	case StatusInProgress:
		return fmt.Sprintf("In Progress (%d%%)", op.Progress)
	case StatusCompleted:
		return "Completed"
	case StatusFailed:
		if op.Error != nil {
			return fmt.Sprintf("Failed: %s", op.Error.Error())
		}
		return "Failed"
	default:
		return "Unknown"
	}
}

// Command methods
func (m *DashboardEnhancedModel) loadInitialData() tea.Cmd {
	return func() tea.Msg {
		// Load initial data
		data := dataLoadedMsg{}
		
		// Count Excel files
		if m.config != nil {
			for _, dir := range m.config.Watcher.Directories {
				for _, ext := range m.config.Watcher.FileExtensions {
					pattern := filepath.Join(dir, "*"+ext)
					if files, err := filepath.Glob(pattern); err == nil {
						data.totalFiles += len(files)
					}
				}
			}
		}
		
		// Get git status
		if m.gitAdapter != nil {
			if status, err := m.gitAdapter.GetStatus(); err == nil {
				data.branch = status.Branch
				data.hasChanges = !status.Clean
			}
		}
		
		return data
	}
}

func (m *DashboardEnhancedModel) refreshData() tea.Cmd {
	return m.loadInitialData()
}

func (m *DashboardEnhancedModel) startConversion() tea.Cmd {
	return func() tea.Msg {
		// Find pending conversions
		pending, _ := m.convAdapter.GetPendingConversions(".", "*.xlsx")
		
		for _, file := range pending {
			op := FileOperation{
				ID:        fmt.Sprintf("conv-%d", time.Now().UnixNano()),
				Type:      OpConvert,
				FileName:  filepath.Base(file),
				Status:    StatusInProgress,
				Progress:  0,
				StartTime: time.Now(),
			}
			
			// In a real implementation, this would start the actual conversion
			return operationUpdateMsg{operation: op}
		}
		
		return nil
	}
}

func (m *DashboardEnhancedModel) syncRepository() tea.Cmd {
	return func() tea.Msg {
		op := FileOperation{
			ID:        fmt.Sprintf("sync-%d", time.Now().UnixNano()),
			Type:      OpSync,
			FileName:  "Repository sync",
			Status:    StatusInProgress,
			Progress:  0,
			StartTime: time.Now(),
		}
		
		// In a real implementation, this would perform git operations
		return operationUpdateMsg{operation: op}
	}
}

func (m *DashboardEnhancedModel) applyLoadedData(msg dataLoadedMsg) {
	m.totalFiles = msg.totalFiles
	m.syncStatus.Branch = msg.branch
	m.syncStatus.HasChanges = msg.hasChanges
	m.syncStatus.IsSynced = !msg.hasChanges
	m.syncStatus.LastCommit = time.Now().Add(-2 * time.Minute) // Mock data
}

func (m *DashboardEnhancedModel) updateOperation(msg operationUpdateMsg) {
	// Add or update operation
	found := false
	for i, op := range m.operations {
		if op.ID == msg.operation.ID {
			m.operations[i] = msg.operation
			found = true
			break
		}
	}
	
	if !found {
		m.operations = append(m.operations, msg.operation)
		if msg.operation.Status == StatusInProgress {
			m.progressBars.AddBar(msg.operation.ID, msg.operation.FileName, 100)
		}
	}
}

// Message types
type dataLoadedMsg struct {
	totalFiles int
	branch     string
	hasChanges bool
}

type operationUpdateMsg struct {
	operation FileOperation
}

// Message types for dashboard
type dashboardTickMsg time.Time

func dashboardTick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return dashboardTickMsg(t)
	})
}