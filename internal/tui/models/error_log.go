package models

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/Classic-Homes/gitcells/internal/tui/messages"
	"github.com/Classic-Homes/gitcells/internal/tui/styles"
	"github.com/Classic-Homes/gitcells/internal/utils"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
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

type SearchMode int

const (
	SearchModeOff SearchMode = iota
	SearchModeActive
)

type ErrorLogKeyMap struct {
	Up       key.Binding
	Down     key.Binding
	PageUp   key.Binding
	PageDown key.Binding
	Home     key.Binding
	End      key.Binding
	Filter   key.Binding
	Clear    key.Binding
	Refresh  key.Binding
	Export   key.Binding
	Search   key.Binding
	Details  key.Binding
	Quit     key.Binding
	Back     key.Binding
}

// Message types
type RefreshLogsMsg struct {
	Time time.Time
}
type LogsLoadedMsg struct {
	Entries []LogEntry
}
type LogsClearedMsg struct{}

type exportCompleteMsg struct {
	err error
}

type ErrorLogModel struct {
	viewport        viewport.Model
	entries         []LogEntry
	filteredEntries []LogEntry
	selectedIndex   int
	showDetails     bool
	filterLevel     string
	searchMode      SearchMode
	searchInput     textinput.Model
	searchPattern   string
	refreshInterval time.Duration
	lastRefresh     time.Time
	width           int
	height          int
	keyMap          ErrorLogKeyMap
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
			key.WithHelp("pgdn/f", "page down"),
		),
		Home: key.NewBinding(
			key.WithKeys("home", "g"),
			key.WithHelp("home/g", "go to start"),
		),
		End: key.NewBinding(
			key.WithKeys("end", "G"),
			key.WithHelp("end/G", "go to end"),
		),
		Filter: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "cycle filter"),
		),
		Clear: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "clear logs"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
		Export: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "export logs"),
		),
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search"),
		),
		Details: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "view details"),
		),
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c", "q"),
			key.WithHelp("q", "quit"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
	}
}

func NewErrorLogModel() ErrorLogModel {
	vp := viewport.New(80, 20)

	searchInput := textinput.New()
	searchInput.Placeholder = "Search logs..."
	searchInput.CharLimit = 100
	searchInput.Width = 50

	return ErrorLogModel{
		viewport:        vp,
		entries:         []LogEntry{},
		filteredEntries: []LogEntry{},
		selectedIndex:   0,
		showDetails:     false,
		filterLevel:     "all",
		searchMode:      SearchModeOff,
		searchInput:     searchInput,
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

func (m ErrorLogModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width - 4
		m.viewport.Height = msg.Height - 10 // More space for header/footer
		m.searchInput.Width = min(m.width-20, 60)
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
		m.applyFilters()
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
		if m.searchMode == SearchModeActive {
			switch msg.String() {
			case "esc":
				m.searchMode = SearchModeOff
				m.searchPattern = ""
				m.searchInput.SetValue("")
				m.applyFilters()
				m.updateViewport()
				return m, nil
			case "enter":
				m.searchPattern = m.searchInput.Value()
				m.searchMode = SearchModeOff
				m.applyFilters()
				m.updateViewport()
				return m, nil
			default:
				m.searchInput, cmd = m.searchInput.Update(msg)
				cmds = append(cmds, cmd)
				// Live search
				m.searchPattern = m.searchInput.Value()
				m.applyFilters()
				m.updateViewport()
				return m, tea.Batch(cmds...)
			}
		}

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
			if len(m.filteredEntries) > 0 {
				m.showDetails = !m.showDetails
				m.updateViewport()
			}

		case key.Matches(msg, m.keyMap.Filter):
			return m, m.cycleFilter()

		case key.Matches(msg, m.keyMap.Refresh):
			return m, m.refreshLogs()

		case key.Matches(msg, m.keyMap.Clear):
			return m, m.clearLogs()

		case msg.String() == "s", msg.String() == "/":
			m.searchMode = SearchModeActive
			m.searchInput.Focus()
			cmds = append(cmds, textinput.Blink)
			return m, tea.Batch(cmds...)

		case msg.String() == "e":
			// Export logs
			return m, m.exportLogs()

		case key.Matches(msg, m.keyMap.Quit):
			return m, tea.Quit

		case msg.String() == "esc":
			return m, messages.RequestMainMenu()
		}
	}

	if m.searchMode == SearchModeActive {
		m.searchInput, cmd = m.searchInput.Update(msg)
		cmds = append(cmds, cmd)
	} else {
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m ErrorLogModel) View() string {
	if m.searchMode == SearchModeActive {
		return m.renderSearchView()
	}

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
	// Title bar
	title := styles.TitleStyle.Render("📋 Error Logs")

	// Stats
	stats := []string{}
	if m.filterLevel != "all" {
		stats = append(stats, styles.HighlightStyle.Render(fmt.Sprintf("Filter: %s", m.filterLevel)))
	}
	if m.searchPattern != "" {
		stats = append(stats, styles.InfoStyle.Render(fmt.Sprintf("Search: \"%s\"", m.searchPattern)))
	}
	stats = append(stats, styles.MutedStyle.Render(fmt.Sprintf("%d entries", len(m.filteredEntries))))

	statsText := strings.Join(stats, " • ")

	// Last update time
	lastUpdate := styles.MutedStyle.Render(fmt.Sprintf("Updated: %s", m.lastRefresh.Format("15:04:05")))

	// Calculate spacing
	titleWidth := lipgloss.Width(title)
	statsWidth := lipgloss.Width(statsText)
	updateWidth := lipgloss.Width(lastUpdate)
	spacing := max(0, m.width-titleWidth-statsWidth-updateWidth-6)

	headerLine := lipgloss.JoinHorizontal(
		lipgloss.Left,
		title,
		"  ",
		statsText,
		strings.Repeat(" ", spacing),
		lastUpdate,
	)

	// Separator
	separator := styles.LogSeparatorStyle.Render(strings.Repeat("─", m.width))

	return headerLine + "\n" + separator + "\n"
}

func (m ErrorLogModel) renderFooter() string {
	separator := styles.LogSeparatorStyle.Render(strings.Repeat("─", m.width))

	var help []string
	if m.showDetails {
		help = []string{
			"↑/↓: navigate",
			"enter: hide details",
			"f: filter",
			"s: search",
			"e: export",
			"r: refresh",
			"c: clear",
			"esc: back",
		}
	} else {
		help = []string{
			"↑/↓: navigate",
			"enter: show details",
			"f: filter (" + m.filterLevel + ")",
			"s: search",
			"e: export",
			"r: refresh",
			"esc: back",
		}
	}

	helpText := styles.MutedStyle.Render(strings.Join(help, " • "))

	// Position indicator
	position := ""
	if len(m.filteredEntries) > 0 {
		position = styles.MutedStyle.Render(fmt.Sprintf("%d/%d", m.selectedIndex+1, len(m.filteredEntries)))
	}

	posWidth := lipgloss.Width(position)
	helpWidth := lipgloss.Width(helpText)
	spacing := max(0, m.width-helpWidth-posWidth-2)

	footer := lipgloss.JoinHorizontal(
		lipgloss.Left,
		helpText,
		strings.Repeat(" ", spacing),
		position,
	)

	return "\n" + separator + "\n" + footer
}

func (m ErrorLogModel) renderSearchView() string {
	title := styles.TitleStyle.Render("🔍 Search Logs")
	prompt := styles.SubtitleStyle.Render("Enter search term (ESC to cancel, Enter to search):")

	searchBox := styles.FocusedBoxStyle.
		Width(m.width - 4).
		Render(m.searchInput.View())

	help := styles.MutedStyle.Render("Search supports regex patterns • Case insensitive by default")

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		prompt,
		searchBox,
		"",
		help,
	)

	return styles.Center(m.width, m.height, content)
}

func (m ErrorLogModel) renderEmptyState() string {
	icon := "📭"
	message := "No log entries found"

	if m.filterLevel != "all" {
		message = fmt.Sprintf("No %s entries found", m.filterLevel)
	}
	if m.searchPattern != "" {
		icon = "🔍"
		message = fmt.Sprintf("No entries matching \"%s\"", m.searchPattern)
	}

	emptyContent := lipgloss.JoinVertical(
		lipgloss.Center,
		icon,
		"",
		styles.MutedStyle.Render(message),
		"",
		styles.MutedStyle.Render("Press 'r' to refresh • 'f' to change filter • 'esc' to go back"),
	)

	return styles.Center(m.width, m.height, emptyContent)
}

func (m *ErrorLogModel) updateViewport() {
	if len(m.filteredEntries) == 0 {
		m.viewport.SetContent("")
		return
	}

	var content strings.Builder

	for i, entry := range m.filteredEntries {
		isSelected := i == m.selectedIndex

		if i > 0 {
			content.WriteString("\n")
		}

		// Render the log entry
		entryContent := m.renderLogEntry(entry, isSelected)
		content.WriteString(entryContent)

		// Show details if selected and details are enabled
		if isSelected && m.showDetails {
			details := m.renderLogDetails(entry)
			if details != "" {
				content.WriteString("\n")
				content.WriteString(details)
			}
		}
	}

	m.viewport.SetContent(content.String())

	// Scroll to keep selected item in view
	m.scrollToSelected()
}

func (m *ErrorLogModel) scrollToSelected() {
	if m.selectedIndex < len(m.filteredEntries) {
		// Calculate approximate line position
		linesAbove := 0
		for i := 0; i < m.selectedIndex; i++ {
			linesAbove++ // Entry line
			if m.showDetails && i == m.selectedIndex-1 {
				// Add lines for details of previous selected item
				linesAbove += 5 // Approximate
			}
		}

		// Ensure selected item is visible
		viewportTop := m.viewport.YOffset
		viewportBottom := viewportTop + m.viewport.Height

		if linesAbove < viewportTop {
			m.viewport.GotoTop()
			for i := 0; i < linesAbove; i++ {
				m.viewport.ScrollDown(1)
			}
		} else if linesAbove >= viewportBottom {
			m.viewport.GotoTop()
			for i := 0; i < linesAbove-m.viewport.Height/2; i++ {
				m.viewport.ScrollDown(1)
			}
		}
	}
}

func (m ErrorLogModel) renderLogEntry(entry LogEntry, isSelected bool) string {
	// Format timestamp
	timestamp := styles.LogTimestampStyle.Render(entry.Timestamp.Format("15:04:05"))

	// Format level
	level := m.renderLogLevel(entry.Level)

	// Format message (truncate if needed)
	message := entry.Message
	maxMsgWidth := m.width - 25 // Account for timestamp, level, and padding
	if lipgloss.Width(message) > maxMsgWidth {
		message = truncateString(message, maxMsgWidth-3) + "..."
	}

	// Build the entry line
	entryLine := fmt.Sprintf("%s %s %s", timestamp, level, message)

	// Apply selection style
	if isSelected {
		return styles.LogSelectedEntryStyle.Render(entryLine)
	}

	return styles.LogEntryStyle.Render(entryLine)
}

func (m ErrorLogModel) renderLogLevel(level string) string {
	switch strings.ToUpper(level) {
	case "ERROR":
		return styles.LogLevelErrorStyle.Render("ERROR")
	case "WARN", "WARNING":
		return styles.LogLevelWarnStyle.Render("WARN")
	case "INFO":
		return styles.LogLevelInfoStyle.Render("INFO")
	case "DEBUG":
		return styles.LogLevelDebugStyle.Render("DEBUG")
	default:
		return styles.MutedStyle.Width(5).Render(level)
	}
}

func (m ErrorLogModel) renderLogDetails(entry LogEntry) string {
	if !m.showDetails {
		return ""
	}

	var details strings.Builder

	// File location
	if entry.File != "" {
		location := fmt.Sprintf("%s:%d", entry.File, entry.Line)
		details.WriteString(fmt.Sprintf("📍 %s\n", styles.LogFileStyle.Render(location)))
	}

	// Error details
	if entry.Error != nil {
		details.WriteString(fmt.Sprintf("❌ %s\n", styles.ErrorStyle.Render(entry.Error.Error())))
	}

	// Additional fields
	if len(entry.Fields) > 0 {
		details.WriteString("📋 Additional Info:\n")

		// Sort fields for consistent display
		keys := make([]string, 0, len(entry.Fields))
		for k := range entry.Fields {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			key := styles.LogFieldKeyStyle.Render(k + ":")
			value := styles.LogFieldValueStyle.Render(fmt.Sprintf("%v", entry.Fields[k]))
			details.WriteString(fmt.Sprintf("   %s %s\n", key, value))
		}
	}

	// Full message if it was truncated
	if len(entry.Message) > 80 {
		details.WriteString("\n💬 Full Message:\n")
		wrapped := wordWrap(entry.Message, m.width-10)
		for _, line := range strings.Split(wrapped, "\n") {
			details.WriteString("   " + line + "\n")
		}
	}

	detailsStr := strings.TrimRight(details.String(), "\n")
	return styles.LogDetailsStyle.Render(detailsStr)
}

func (m *ErrorLogModel) applyFilters() {
	m.filteredEntries = []LogEntry{}

	for _, entry := range m.entries {
		// Apply level filter
		if m.filterLevel != "all" && !strings.EqualFold(entry.Level, m.filterLevel) {
			continue
		}

		// Apply search filter
		if m.searchPattern != "" {
			matched := false
			searchLower := strings.ToLower(m.searchPattern)

			// Try regex first
			if re, err := regexp.Compile("(?i)" + m.searchPattern); err == nil {
				if re.MatchString(entry.Message) ||
					re.MatchString(entry.File) ||
					(entry.Error != nil && re.MatchString(entry.Error.Error())) {
					matched = true
				}
			} else {
				// Fall back to simple contains
				if strings.Contains(strings.ToLower(entry.Message), searchLower) ||
					strings.Contains(strings.ToLower(entry.File), searchLower) ||
					(entry.Error != nil && strings.Contains(strings.ToLower(entry.Error.Error()), searchLower)) {
					matched = true
				}
			}

			if !matched {
				continue
			}
		}

		m.filteredEntries = append(m.filteredEntries, entry)
	}

	// Reset selection if needed
	if m.selectedIndex >= len(m.filteredEntries) {
		m.selectedIndex = max(0, len(m.filteredEntries)-1)
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
		m.applyFilters()
		m.updateViewport()
		return nil
	}
}

func (m ErrorLogModel) exportLogs() tea.Cmd {
	return func() tea.Msg {
		// Create logs in user's home directory under .gitcells/logs
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return exportCompleteMsg{err: err}
		}

		logsDir := filepath.Join(homeDir, ".gitcells", "logs")
		if err := os.MkdirAll(logsDir, 0755); err != nil {
			return exportCompleteMsg{err: err}
		}

		filename := filepath.Join(logsDir, fmt.Sprintf("gitcells_logs_%s.txt", time.Now().Format("20060102_150405")))
		file, err := os.Create(filename)
		if err != nil {
			utils.LogErrorDefault(err, "Failed to create export file", map[string]interface{}{
				"filename": filename,
			})
			return nil
		}
		defer file.Close()

		writer := bufio.NewWriter(file)
		defer writer.Flush()

		// Write header
		fmt.Fprintf(writer, "GitCells Error Log Export\n")
		fmt.Fprintf(writer, "Generated: %s\n", time.Now().Format(time.RFC3339))
		fmt.Fprintf(writer, "Total Entries: %d\n", len(m.filteredEntries))
		fmt.Fprintf(writer, "Filter: %s\n", m.filterLevel)
		if m.searchPattern != "" {
			fmt.Fprintf(writer, "Search: %s\n", m.searchPattern)
		}
		fmt.Fprintf(writer, "%s\n\n", strings.Repeat("=", 80))

		// Write entries
		for _, entry := range m.filteredEntries {
			fmt.Fprintf(writer, "Timestamp: %s\n", entry.Timestamp.Format(time.RFC3339))
			fmt.Fprintf(writer, "Level: %s\n", entry.Level)
			fmt.Fprintf(writer, "Message: %s\n", entry.Message)

			if entry.File != "" {
				fmt.Fprintf(writer, "Location: %s:%d\n", entry.File, entry.Line)
			}

			if entry.Error != nil {
				fmt.Fprintf(writer, "Error: %s\n", entry.Error.Error())
			}

			if len(entry.Fields) > 0 {
				fmt.Fprintf(writer, "Fields:\n")
				for k, v := range entry.Fields {
					fmt.Fprintf(writer, "  %s: %v\n", k, v)
				}
			}

			fmt.Fprintf(writer, "%s\n\n", strings.Repeat("-", 40))
		}

		utils.LogInfo("Exported logs", map[string]interface{}{
			"filename": filename,
			"entries":  len(m.filteredEntries),
		})

		return nil
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

	// Sort by timestamp (newest first)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Timestamp.After(entries[j].Timestamp)
	})

	// Limit to most recent 1000 entries
	if len(entries) > 1000 {
		entries = entries[:1000]
	}

	return entries, nil
}

func (m ErrorLogModel) parseLogLine(line string) (LogEntry, error) {
	// Expected format: "2006-01-02T15:04:05.000Z07:00 [LEVEL] component message {fields}"
	parts := strings.SplitN(line, " ", 4)
	if len(parts) < 4 {
		return LogEntry{}, fmt.Errorf("invalid log format")
	}

	timestamp, err := time.Parse("2006-01-02T15:04:05.000Z07:00", parts[0])
	if err != nil {
		timestamp = time.Now()
	}

	level := strings.Trim(parts[1], "[]")

	// Parse component and message
	messageParts := strings.SplitN(parts[3], " ", 2)
	component := messageParts[0]
	message := ""
	if len(messageParts) > 1 {
		message = messageParts[1]
	}

	// Extract fields if present (JSON at end)
	fields := make(map[string]interface{})
	if idx := strings.LastIndex(message, " {"); idx > 0 {
		message = message[:idx]
		// Parse JSON fields - simplified for now
	}

	return LogEntry{
		Timestamp: timestamp,
		Level:     level,
		Message:   message,
		File:      component,
		Fields:    fields,
	}, nil
}

func (m ErrorLogModel) clearLogFile() error {
	logFile := utils.GetLogFilePath()
	return os.Truncate(logFile, 0)
}

// Helper functions
func truncateString(s string, maxWidth int) string {
	if lipgloss.Width(s) <= maxWidth {
		return s
	}

	// Binary search for the right length
	left, right := 0, len(s)
	result := s

	for left < right {
		mid := (left + right + 1) / 2
		truncated := s[:mid]
		if lipgloss.Width(truncated) <= maxWidth {
			result = truncated
			left = mid
		} else {
			right = mid - 1
		}
	}

	return result
}

func wordWrap(text string, width int) string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return ""
	}

	var lines []string
	currentLine := words[0]

	for i := 1; i < len(words); i++ {
		word := words[i]
		testLine := currentLine + " " + word
		if lipgloss.Width(testLine) <= width {
			currentLine = testLine
		} else {
			lines = append(lines, currentLine)
			currentLine = word
		}
	}
	lines = append(lines, currentLine)

	return strings.Join(lines, "\n")
}
