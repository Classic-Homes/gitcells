package models

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/Classic-Homes/gitcells/internal/config"
	"github.com/Classic-Homes/gitcells/internal/converter"
	"github.com/Classic-Homes/gitcells/internal/tui/messages"
	"github.com/Classic-Homes/gitcells/internal/tui/styles"
	"github.com/Classic-Homes/gitcells/internal/watcher"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sirupsen/logrus"
)

type WatcherModel struct {
	width  int
	height int
	config *config.Config
	logger *logrus.Logger

	// Watcher state
	watcher   *watcher.FileWatcher
	isRunning bool
	startTime time.Time
	events    []WatcherEvent
	maxEvents int

	// UI state
	showHelp      bool
	scrollOffset  int
	selectedDir   int
	showDirPicker bool

	// Current status
	watchedDirs []string
	totalFiles  int
	lastEvent   time.Time

	// Background context
	ctx    context.Context
	cancel context.CancelFunc
}

type WatcherEvent struct {
	Path      string
	Type      string
	Timestamp time.Time
	Status    string
	Error     error
}

type watcherEventMsg struct {
	event WatcherEvent
}

type watcherStatusMsg struct {
	isRunning   bool
	watchedDirs []string
	totalFiles  int
}

func NewWatcherModel() tea.Model {
	ctx, cancel := context.WithCancel(context.Background())

	m := &WatcherModel{
		maxEvents: 50,
		events:    []WatcherEvent{},
		ctx:       ctx,
		cancel:    cancel,
	}

	// Try to load configuration
	if cfg, err := config.Load("."); err == nil {
		m.config = cfg
		m.watchedDirs = cfg.Watcher.Directories
	}

	// Initialize logger
	m.logger = logrus.New()
	m.logger.SetLevel(logrus.InfoLevel)

	return m
}

func (m WatcherModel) Init() tea.Cmd {
	return tea.Batch(
		m.loadWatcherStatus(),
		watcherTick(),
	)
}

func (m WatcherModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case watcherTickMsg:
		if m.isRunning {
			cmds = append(cmds, m.refreshStatus())
		}
		cmds = append(cmds, watcherTick())

	case watcherEventMsg:
		m.addEvent(msg.event)

	case watcherStatusMsg:
		m.isRunning = msg.isRunning
		m.watchedDirs = msg.watchedDirs
		m.totalFiles = msg.totalFiles

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.watcher != nil {
				_ = m.watcher.Stop()
			}
			m.cancel()
			return m, tea.Quit
		case "esc":
			if m.showDirPicker {
				m.showDirPicker = false
			} else {
				return m, messages.RequestMainMenu()
			}
		case "?":
			m.showHelp = !m.showHelp
		case "s":
			if m.isRunning {
				cmds = append(cmds, m.stopWatcher())
			} else {
				cmds = append(cmds, m.startWatcher())
			}
		case "r":
			cmds = append(cmds, m.refreshStatus())
		case "c":
			m.clearEvents()
		case "d":
			m.showDirPicker = !m.showDirPicker
		case "up", "k":
			if m.showDirPicker && m.selectedDir > 0 {
				m.selectedDir--
			} else if m.scrollOffset > 0 {
				m.scrollOffset--
			}
		case "down", "j":
			if m.showDirPicker && m.selectedDir < len(m.watchedDirs)-1 {
				m.selectedDir++
			} else {
				m.scrollOffset++
			}
		case "enter":
			if m.showDirPicker {
				m.showDirPicker = false
			}
		}
	}

	return m, tea.Batch(cmds...)
}

func (m WatcherModel) View() string {
	if m.showHelp {
		return m.renderHelp()
	}

	if m.showDirPicker {
		return m.renderDirectoryPicker()
	}

	// Main container
	containerStyle := lipgloss.NewStyle().
		Padding(1, 2).
		Width(m.width).
		Height(m.height)

	// Header
	header := m.renderHeader()

	// Status section
	status := m.renderStatus()

	// Events section
	events := m.renderEvents()

	// Footer
	footer := m.renderFooter()

	// Combine all sections
	fullView := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		status,
		events,
		footer,
	)

	return containerStyle.Render(fullView)
}

func (m WatcherModel) renderHeader() string {
	titleStyle := styles.TitleStyle.MarginBottom(1)

	statusIcon := "⏸️"
	statusText := "Stopped"
	statusColor := styles.Muted

	if m.isRunning {
		statusIcon = "▶️"
		statusText = "Running"
		statusColor = styles.Success
	}

	statusStyle := lipgloss.NewStyle().Foreground(statusColor)

	title := titleStyle.Render("File Watcher")
	status := statusStyle.Render(fmt.Sprintf("%s %s", statusIcon, statusText))

	if m.isRunning && !m.startTime.IsZero() {
		uptime := time.Since(m.startTime)
		status += styles.MutedStyle.Render(fmt.Sprintf(" • Uptime: %s", formatDuration(uptime)))
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		title,
		lipgloss.NewStyle().Width(m.width-40).Render(" "),
		status,
	)
}

func (m WatcherModel) renderStatus() string {
	// Status boxes
	boxStyle := styles.BoxStyle.
		Width(25).
		Height(6).
		MarginRight(2).
		MarginBottom(1)

	// Directories box
	dirCount := len(m.watchedDirs)
	dirStatus := "Not configured"
	if dirCount > 0 {
		dirStatus = fmt.Sprintf("%d directories", dirCount)
	}

	dirsBox := boxStyle.Render(
		styles.SubtitleStyle.Render("Watching") + "\n" +
			dirStatus + "\n" +
			fmt.Sprintf("%d Excel files", m.totalFiles),
	)

	// Configuration box
	debounce := "2s"
	extensions := ".xlsx, .xls"

	if m.config != nil {
		if m.config.Watcher.DebounceDelay > 0 {
			debounce = m.config.Watcher.DebounceDelay.String()
		}
		if len(m.config.Watcher.FileExtensions) > 0 {
			extensions = fmt.Sprintf("%v", m.config.Watcher.FileExtensions)
		}
	}

	configBox := boxStyle.Render(
		styles.SubtitleStyle.Render("Configuration") + "\n" +
			fmt.Sprintf("Debounce: %s", debounce) + "\n" +
			fmt.Sprintf("Types: %s", extensions),
	)

	// Activity box
	eventCount := len(m.events)
	lastActivity := "None"
	if !m.lastEvent.IsZero() {
		lastActivity = formatDuration(time.Since(m.lastEvent)) + " ago"
	}

	activityBox := boxStyle.Render(
		styles.SubtitleStyle.Render("Activity") + "\n" +
			fmt.Sprintf("Events: %d", eventCount) + "\n" +
			fmt.Sprintf("Last: %s", lastActivity),
	)

	return lipgloss.JoinHorizontal(lipgloss.Top, dirsBox, configBox, activityBox)
}

func (m WatcherModel) renderEvents() string {
	eventStyle := styles.BoxStyle.
		Width(m.width - 8).
		Height(m.height - 15)

	content := styles.SubtitleStyle.Render("Recent Events") + "\n\n"

	if len(m.events) == 0 {
		content += styles.MutedStyle.Render("No events yet...")
		if !m.isRunning {
			content += "\n\n" + styles.MutedStyle.Render("Press 's' to start watching")
		}
	} else {
		// Show recent events (most recent first)
		start := 0
		if len(m.events) > 20 {
			start = len(m.events) - 20
		}

		for i := len(m.events) - 1; i >= start; i-- {
			event := m.events[i]

			// Event icon
			icon := "📄"
			switch event.Type {
			case "create":
				icon = "➕"
			case "modify":
				icon = "✏️"
			case "delete":
				icon = "❌"
			}

			// Status color
			statusStyle := lipgloss.NewStyle()
			switch {
			case event.Error != nil:
				statusStyle = statusStyle.Foreground(styles.Error)
			case event.Status == "completed":
				statusStyle = statusStyle.Foreground(styles.Success)
			default:
				statusStyle = statusStyle.Foreground(styles.Warning)
			}

			timeStr := event.Timestamp.Format("15:04:05")
			fileName := filepath.Base(event.Path)
			status := event.Status
			if event.Error != nil {
				status = fmt.Sprintf("error: %s", event.Error.Error())
			}

			content += fmt.Sprintf("%s %s %s %s - %s\n",
				styles.MutedStyle.Render(timeStr),
				icon,
				styles.SubtitleStyle.Render(event.Type),
				fileName,
				statusStyle.Render(status),
			)
		}
	}

	return eventStyle.Render(content)
}

func (m WatcherModel) renderDirectoryPicker() string {
	pickerStyle := styles.BoxStyle.
		Width(60).
		Height(20)

	content := styles.TitleStyle.Render("Watched Directories") + "\n\n"

	if len(m.watchedDirs) == 0 {
		content += styles.MutedStyle.Render("No directories configured")
	} else {
		for i, dir := range m.watchedDirs {
			cursor := "  "
			if i == m.selectedDir {
				cursor = lipgloss.NewStyle().Foreground(styles.Primary).Render("▶ ")
			}
			content += fmt.Sprintf("%s%s\n", cursor, dir)
		}
	}

	content += "\n" + styles.HelpStyle.Render("Use ↑/↓ to navigate, Enter to close")

	return styles.Center(m.width, m.height, pickerStyle.Render(content))
}

func (m WatcherModel) renderFooter() string {
	var commands []string

	if m.isRunning {
		commands = append(commands, "[s] Stop")
	} else {
		commands = append(commands, "[s] Start")
	}

	commands = append(commands,
		"[r] Refresh",
		"[c] Clear Events",
		"[d] Directories",
		"[?] Help",
		"[q] Quit",
	)

	helpText := lipgloss.JoinHorizontal(lipgloss.Top, commands...)
	return styles.HelpStyle.MarginTop(1).Render(helpText)
}

func (m WatcherModel) renderHelp() string {
	helpBox := styles.BoxStyle.
		Width(70).
		Render(
			styles.TitleStyle.Render("File Watcher Help") + "\n\n" +
				"Controls:\n" +
				"  s         - Start/Stop watcher\n" +
				"  r         - Refresh status\n" +
				"  c         - Clear event log\n" +
				"  d         - Show watched directories\n" +
				"  ↑/↓ or j/k - Scroll events\n" +
				"  esc       - Return to main menu\n" +
				"  q         - Quit application\n\n" +
				"About the Watcher:\n" +
				"The file watcher monitors Excel files (.xlsx, .xls) in\n" +
				"configured directories and automatically converts them\n" +
				"to JSON format when changes are detected. Events are\n" +
				"debounced to avoid excessive processing during rapid\n" +
				"file changes.\n\n" +
				"Configuration is loaded from .gitcells/config.yaml\n" +
				"in your project directory.\n\n" +
				"Press ? to close this help",
		)

	return styles.Center(m.width, m.height, helpBox)
}

// Command methods
func (m *WatcherModel) loadWatcherStatus() tea.Cmd {
	return func() tea.Msg {
		// Count Excel files in watched directories
		totalFiles := 0
		if m.config != nil {
			for _, dir := range m.config.Watcher.Directories {
				for _, ext := range m.config.Watcher.FileExtensions {
					pattern := filepath.Join(dir, "*"+ext)
					if files, err := filepath.Glob(pattern); err == nil {
						totalFiles += len(files)
					}
				}
			}
		}

		return watcherStatusMsg{
			isRunning:   m.isRunning,
			watchedDirs: m.watchedDirs,
			totalFiles:  totalFiles,
		}
	}
}

func (m *WatcherModel) refreshStatus() tea.Cmd {
	return m.loadWatcherStatus()
}

func (m *WatcherModel) startWatcher() tea.Cmd {
	return func() tea.Msg {
		if m.config == nil {
			return watcherEventMsg{
				event: WatcherEvent{
					Path:      "config",
					Type:      "error",
					Timestamp: time.Now(),
					Status:    "failed",
					Error:     fmt.Errorf("no configuration found"),
				},
			}
		}

		if len(m.config.Watcher.Directories) == 0 {
			return watcherEventMsg{
				event: WatcherEvent{
					Path:      "config",
					Type:      "error",
					Timestamp: time.Now(),
					Status:    "failed",
					Error:     fmt.Errorf("no directories configured"),
				},
			}
		}

		// Create event handler
		handler := func(event watcher.FileEvent) error {
			// Convert Excel to JSON
			conv := converter.NewConverter(m.logger)
			convertOptions := converter.ConvertOptions{
				PreserveFormulas: m.config.Converter.PreserveFormulas,
				PreserveStyles:   m.config.Converter.PreserveStyles,
				PreserveComments: m.config.Converter.PreserveComments,
				CompactJSON:      m.config.Converter.CompactJSON,
				IgnoreEmptyCells: m.config.Converter.IgnoreEmptyCells,
				MaxCellsPerSheet: m.config.Converter.MaxCellsPerSheet,
			}

			_, err := conv.ExcelToJSON(event.Path, convertOptions)

			// Log the processing result
			if err != nil {
				m.logger.Errorf("Failed to process %s: %v", event.Path, err)
			} else {
				m.logger.Infof("Successfully processed %s: %s", event.Type, event.Path)
			}

			return err
		}

		// Setup watcher
		watcherConfig := &watcher.Config{
			IgnorePatterns: m.config.Watcher.IgnorePatterns,
			DebounceDelay:  m.config.Watcher.DebounceDelay,
			FileExtensions: m.config.Watcher.FileExtensions,
		}

		fw, err := watcher.NewFileWatcher(watcherConfig, handler, m.logger)
		if err != nil {
			return watcherEventMsg{
				event: WatcherEvent{
					Path:      "watcher",
					Type:      "error",
					Timestamp: time.Now(),
					Status:    "failed",
					Error:     err,
				},
			}
		}

		// Add directories to watch
		for _, dir := range m.config.Watcher.Directories {
			if err := fw.AddDirectory(dir); err != nil {
				m.logger.Warnf("Failed to add directory %s: %v", dir, err)
			}
		}

		// Start watching
		if err := fw.Start(); err != nil {
			return watcherEventMsg{
				event: WatcherEvent{
					Path:      "watcher",
					Type:      "error",
					Timestamp: time.Now(),
					Status:    "failed",
					Error:     err,
				},
			}
		}

		m.watcher = fw
		m.startTime = time.Now()

		return watcherEventMsg{
			event: WatcherEvent{
				Path:      "watcher",
				Type:      "start",
				Timestamp: time.Now(),
				Status:    "started",
			},
		}
	}
}

func (m *WatcherModel) stopWatcher() tea.Cmd {
	return func() tea.Msg {
		if m.watcher != nil {
			err := m.watcher.Stop()
			m.watcher = nil

			if err != nil {
				return watcherEventMsg{
					event: WatcherEvent{
						Path:      "watcher",
						Type:      "error",
						Timestamp: time.Now(),
						Status:    "failed",
						Error:     err,
					},
				}
			}
		}

		return watcherEventMsg{
			event: WatcherEvent{
				Path:      "watcher",
				Type:      "stop",
				Timestamp: time.Now(),
				Status:    "stopped",
			},
		}
	}
}

func (m *WatcherModel) addEvent(event WatcherEvent) {
	m.events = append(m.events, event)
	m.lastEvent = event.Timestamp

	// Keep only the most recent events
	if len(m.events) > m.maxEvents {
		m.events = m.events[len(m.events)-m.maxEvents:]
	}

	// Update running status based on event type
	if event.Type == "start" && event.Error == nil {
		m.isRunning = true
	} else if event.Type == "stop" {
		m.isRunning = false
		m.startTime = time.Time{}
	}
}

func (m *WatcherModel) clearEvents() {
	m.events = []WatcherEvent{}
	m.lastEvent = time.Time{}
}

// Message types
type watcherTickMsg time.Time

func watcherTick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return watcherTickMsg(t)
	})
}
