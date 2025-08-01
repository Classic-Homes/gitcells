package models

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/Classic-Homes/gitcells/internal/tui/messages"
	"github.com/Classic-Homes/gitcells/internal/tui/styles"
	"github.com/Classic-Homes/gitcells/internal/utils"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type LogEntry struct {
	Timestamp time.Time
	Level     string
	Message   string
	File      string
	Line      int
	Error     error
	Fields    map[string]interface{}
}

type ErrorLogModel struct {
	viewport        viewport.Model
	entries         []LogEntry
	filteredEntries []LogEntry
	selectedIndex   int
	showDetails     bool
	filterLevel     string
	refreshInterval time.Duration
	lastRefresh     time.Time
	width           int
	height          int
	keyMap          ErrorLogKeyMap
}

type ErrorLogKeyMap struct {
	Up       key.Binding
	Down     key.Binding
	PageUp   key.Binding
	PageDown key.Binding
	Home     key.Binding
	End      key.Binding
	Details  key.Binding
	Filter   key.Binding
	Refresh  key.Binding
	Clear    key.Binding
	Back     key.Binding
	Quit     key.Binding
}

func DefaultErrorLogKeyMap() ErrorLogKeyMap {
	return ErrorLogKeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup", "b"),
			key.WithHelp("pgup/b", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("pgdown", "f"),
			key.WithHelp("pgdown/f", "page down"),
		),
		Home: key.NewBinding(
			key.WithKeys("home", "g"),
			key.WithHelp("home/g", "go to start"),
		),
		End: key.NewBinding(
			key.WithKeys("end", "G"),
			key.WithHelp("end/G", "go to end"),
		),
		Details: key.NewBinding(
			key.WithKeys("enter", " "),
			key.WithHelp("enter", "toggle details"),
		),
		Filter: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "filter level"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
		Clear: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "clear logs"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
	}
}

func NewErrorLogModel() ErrorLogModel {
	vp := viewport.New(80, 20)
	vp.Style = styles.ViewportStyle

	return ErrorLogModel{
		viewport:        vp,
		entries:         []LogEntry{},
		filteredEntries: []LogEntry{},
		selectedIndex:   0,
		showDetails:     false,
		filterLevel:     "all",
		refreshInterval: 5 * time.Second,
		keyMap:          DefaultErrorLogKeyMap(),
	}
}

func (m ErrorLogModel) Init() tea.Cmd {
	return tea.Batch(
		m.refreshLogs(),
		tea.Tick(m.refreshInterval, func(t time.Time) tea.Msg {
			return RefreshLogsMsg{Time: t}
		}),
	)
}

type RefreshLogsMsg struct {
	Time time.Time
}

type LogsLoadedMsg struct {
	Entries []LogEntry
}

type LogsClearedMsg struct{}

func (m ErrorLogModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width - 4
		m.viewport.Height = msg.Height - 8
		m.updateViewport()

	case RefreshLogsMsg:
		if time.Since(m.lastRefresh) >= m.refreshInterval {
			return m, m.refreshLogs()
		}
		return m, tea.Tick(m.refreshInterval, func(t time.Time) tea.Msg {
			return RefreshLogsMsg{Time: t}
		})

	case LogsLoadedMsg:
		m.entries = msg.Entries
		m.lastRefresh = time.Now()
		m.applyFilter()
		m.updateViewport()
		return m, tea.Tick(m.refreshInterval, func(t time.Time) tea.Msg {
			return RefreshLogsMsg{Time: t}
		})

	case LogsClearedMsg:
		m.entries = []LogEntry{}
		m.filteredEntries = []LogEntry{}
		m.selectedIndex = 0
		m.updateViewport()

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.Up):
			if m.selectedIndex > 0 {
				m.selectedIndex--
				m.updateViewport()
			}

		case key.Matches(msg, m.keyMap.Down):
			if m.selectedIndex < len(m.filteredEntries)-1 {
				m.selectedIndex++
				m.updateViewport()
			}

		case key.Matches(msg, m.keyMap.PageUp):
			m.selectedIndex = max(0, m.selectedIndex-10)
			m.updateViewport()

		case key.Matches(msg, m.keyMap.PageDown):
			m.selectedIndex = min(len(m.filteredEntries)-1, m.selectedIndex+10)
			m.updateViewport()

		case key.Matches(msg, m.keyMap.Home):
			m.selectedIndex = 0
			m.updateViewport()

		case key.Matches(msg, m.keyMap.End):
			if len(m.filteredEntries) > 0 {
				m.selectedIndex = len(m.filteredEntries) - 1
			}
			m.updateViewport()

		case key.Matches(msg, m.keyMap.Details):
			m.showDetails = !m.showDetails
			m.updateViewport()

		case key.Matches(msg, m.keyMap.Filter):
			return m, m.cycleFilter()

		case key.Matches(msg, m.keyMap.Refresh):
			return m, m.refreshLogs()

		case key.Matches(msg, m.keyMap.Clear):
			return m, m.clearLogs()

		case key.Matches(msg, m.keyMap.Quit):
			return m, tea.Quit

		case msg.String() == "esc":
			return m, messages.RequestMainMenu()
		}
	}

	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m ErrorLogModel) View() string {
	if len(m.filteredEntries) == 0 {
		return m.renderEmptyState()
	}

	header := m.renderHeader()
	content := m.viewport.View()
	footer := m.renderFooter()

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		content,
		footer,
	)
}

func (m ErrorLogModel) renderHeader() string {
	title := styles.TitleStyle.Render("Error Logs")

	filterInfo := fmt.Sprintf("Filter: %s", m.filterLevel)
	if m.filterLevel != "all" {
		filterInfo = styles.HighlightStyle.Render(filterInfo)
	} else {
		filterInfo = styles.MutedStyle.Render(filterInfo)
	}

	count := fmt.Sprintf("(%d entries)", len(m.filteredEntries))
	countStyle := styles.MutedStyle.Render(count)

	lastUpdate := fmt.Sprintf("Last update: %s", m.lastRefresh.Format("15:04:05"))
	updateStyle := styles.MutedStyle.Render(lastUpdate)

	headerLine := lipgloss.JoinHorizontal(
		lipgloss.Left,
		title,
		" ",
		filterInfo,
		" ",
		countStyle,
		strings.Repeat(" ", max(0, m.width-lipgloss.Width(title)-lipgloss.Width(filterInfo)-lipgloss.Width(countStyle)-lipgloss.Width(updateStyle)-3)),
		updateStyle,
	)

	return headerLine + "\n" + strings.Repeat("─", m.width)
}

func (m ErrorLogModel) renderFooter() string {
	help := []string{
		"↑/↓: navigate",
		"enter: details",
		"f: filter",
		"r: refresh",
		"c: clear",
		"esc: back",
	}

	helpText := strings.Join(help, " • ")
	return styles.MutedStyle.Render(helpText)
}

func (m ErrorLogModel) renderEmptyState() string {
	message := "No log entries found"
	if m.filterLevel != "all" {
		message = fmt.Sprintf("No %s level entries found", m.filterLevel)
	}

	emptyStyle := lipgloss.NewStyle().
		Foreground(styles.MutedColor).
		Align(lipgloss.Center).
		Width(m.width).
		Height(m.height / 2)

	return emptyStyle.Render(message)
}

func (m *ErrorLogModel) updateViewport() {
	if len(m.filteredEntries) == 0 {
		m.viewport.SetContent("")
		return
	}

	var content strings.Builder

	for i, entry := range m.filteredEntries {
		isSelected := i == m.selectedIndex
		line := m.renderLogEntry(entry, isSelected, i)

		if i > 0 {
			content.WriteString("\n")
		}
		content.WriteString(line)

		if isSelected && m.showDetails {
			details := m.renderLogDetails(entry)
			content.WriteString("\n")
			content.WriteString(details)
		}
	}

	m.viewport.SetContent(content.String())

	if m.selectedIndex < len(m.filteredEntries) {
		lineHeight := 1
		if m.showDetails {
			lineHeight = 3
		}
		m.viewport.GotoTop()
		for i := 0; i < m.selectedIndex*lineHeight; i++ {
			m.viewport.ScrollDown(1)
		}
	}
}

func (m ErrorLogModel) renderLogEntry(entry LogEntry, isSelected bool, index int) string {
	timestamp := entry.Timestamp.Format("15:04:05")
	level := m.renderLogLevel(entry.Level)
	message := entry.Message

	if len(message) > 80 {
		message = message[:77] + "..."
	}

	line := fmt.Sprintf("%s [%s] %s", timestamp, level, message)

	if isSelected {
		return styles.SelectedStyle.Render(line)
	}

	return line
}

func (m ErrorLogModel) renderLogLevel(level string) string {
	switch strings.ToUpper(level) {
	case "ERROR":
		return styles.ErrorStyle.Render("ERROR")
	case "WARN", "WARNING":
		return styles.WarningStyle.Render("WARN ")
	case "INFO":
		return styles.InfoStyle.Render("INFO ")
	case "DEBUG":
		return styles.MutedStyle.Render("DEBUG")
	default:
		return level
	}
}

func (m ErrorLogModel) renderLogDetails(entry LogEntry) string {
	var details strings.Builder

	if entry.File != "" {
		details.WriteString(fmt.Sprintf("  File: %s:%d\n", entry.File, entry.Line))
	}

	if entry.Error != nil {
		details.WriteString(fmt.Sprintf("  Error: %s\n", entry.Error.Error()))
	}

	if len(entry.Fields) > 0 {
		details.WriteString("  Fields:\n")
		for k, v := range entry.Fields {
			details.WriteString(fmt.Sprintf("    %s: %v\n", k, v))
		}
	}

	return styles.MutedStyle.Render(details.String())
}

func (m *ErrorLogModel) applyFilter() {
	if m.filterLevel == "all" {
		m.filteredEntries = m.entries
	} else {
		m.filteredEntries = []LogEntry{}
		for _, entry := range m.entries {
			if strings.EqualFold(entry.Level, m.filterLevel) {
				m.filteredEntries = append(m.filteredEntries, entry)
			}
		}
	}

	if m.selectedIndex >= len(m.filteredEntries) {
		m.selectedIndex = max(0, len(m.filteredEntries)-1)
	}
}

func (m ErrorLogModel) refreshLogs() tea.Cmd {
	return func() tea.Msg {
		entries, err := m.loadLogsFromFile()
		if err != nil {
			utils.LogErrorDefault(err, "Failed to load logs", map[string]interface{}{
				"operation": "refresh_logs",
			})
			return LogsLoadedMsg{Entries: []LogEntry{}}
		}
		return LogsLoadedMsg{Entries: entries}
	}
}

func (m *ErrorLogModel) cycleFilter() tea.Cmd {
	return func() tea.Msg {
		filters := []string{"all", "error", "warn", "info", "debug"}
		currentIndex := 0
		for i, filter := range filters {
			if filter == m.filterLevel {
				currentIndex = i
				break
			}
		}
		nextIndex := (currentIndex + 1) % len(filters)
		m.filterLevel = filters[nextIndex]
		m.applyFilter()
		return tea.Msg(nil)
	}
}

func (m ErrorLogModel) clearLogs() tea.Cmd {
	return func() tea.Msg {
		err := m.clearLogFile()
		if err != nil {
			utils.LogErrorDefault(err, "Failed to clear logs", map[string]interface{}{
				"operation": "clear_logs",
			})
		}
		return LogsClearedMsg{}
	}
}

func (m ErrorLogModel) loadLogsFromFile() ([]LogEntry, error) {
	logFile := utils.GetLogFilePath()
	file, err := os.Open(logFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []LogEntry{}, nil
		}
		return nil, err
	}
	defer file.Close()

	var entries []LogEntry
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		entry, err := m.parseLogLine(line)
		if err != nil {
			continue
		}
		entries = append(entries, entry)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Timestamp.After(entries[j].Timestamp)
	})

	if len(entries) > 1000 {
		entries = entries[:1000]
	}

	return entries, nil
}

func (m ErrorLogModel) parseLogLine(line string) (LogEntry, error) {
	parts := strings.SplitN(line, " ", 4)
	if len(parts) < 4 {
		return LogEntry{}, fmt.Errorf("invalid log format")
	}

	timestamp, err := time.Parse("2006-01-02T15:04:05.000Z07:00", parts[0])
	if err != nil {
		timestamp = time.Now()
	}

	level := strings.Trim(parts[1], "[]")
	message := parts[3]

	return LogEntry{
		Timestamp: timestamp,
		Level:     level,
		Message:   message,
	}, nil
}

func (m ErrorLogModel) clearLogFile() error {
	logFile := utils.GetLogFilePath()
	return os.Truncate(logFile, 0)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
