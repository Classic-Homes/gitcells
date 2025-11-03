package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/Classic-Homes/gitcells/internal/tui/components"
	"github.com/Classic-Homes/gitcells/internal/tui/messages"
	"github.com/Classic-Homes/gitcells/internal/tui/styles"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ErrorLogModelV2 struct {
	width        int
	height       int
	logs         []ErrLogEntry
	filteredLogs []ErrLogEntry
	cursor       int
	scrollOffset int
	filter       string
	searchMode   bool
	searchInput  string
	showDetails  bool
}

func NewErrorLogModel() *ErrorLogModelV2 {
	return &ErrorLogModelV2{
		filter: "all",
		width:  80,
		height: 24,
	}
}

func (m *ErrorLogModelV2) Init() tea.Cmd {
	return m.loadLogs()
}

func (m *ErrorLogModelV2) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case logsLoadedMsg:
		m.logs = msg.logs
		m.applyFilter()
		return m, nil

	case tea.KeyMsg:
		if m.searchMode {
			return m.handleSearchMode(msg)
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "esc":
			if m.showDetails {
				m.showDetails = false
				return m, nil
			}
			return m, messages.RequestMainMenu()

		// Navigation
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				m.adjustScroll()
			}

		case "down", "j":
			if m.cursor < len(m.filteredLogs)-1 {
				m.cursor++
				m.adjustScroll()
			}

		case "pgup":
			m.cursor -= 10
			if m.cursor < 0 {
				m.cursor = 0
			}
			m.adjustScroll()

		case "pgdown":
			m.cursor += 10
			if m.cursor >= len(m.filteredLogs) {
				m.cursor = len(m.filteredLogs) - 1
			}
			m.adjustScroll()

		case "home", "g":
			m.cursor = 0
			m.scrollOffset = 0

		case "end", "G":
			m.cursor = len(m.filteredLogs) - 1
			m.adjustScroll()

		// Actions
		case "enter", " ":
			m.showDetails = !m.showDetails

		case "/":
			m.searchMode = true
			m.searchInput = ""

		case "f":
			// Cycle through filters
			filters := []string{"all", "error", "warn", "info", "debug"}
			for i, f := range filters {
				if f == m.filter {
					m.filter = filters[(i+1)%len(filters)]
					break
				}
			}
			m.applyFilter()

		case "c":
			// Clear logs
			m.logs = nil
			m.filteredLogs = nil
			m.cursor = 0
			m.scrollOffset = 0
			// Would actually clear log file here

		case "r":
			// Refresh logs
			return m, m.loadLogs()
		}
	}

	return m, nil
}

func (m *ErrorLogModelV2) View() string {
	// Header with breadcrumb
	breadcrumb := []string{"Dashboard", "Error Logs"}
	header := components.RenderHeader("Error Logs", breadcrumb, m.width)

	// Filter status
	filterStatus := fmt.Sprintf("Filter: %s | %d entries", m.filter, len(m.filteredLogs))
	if m.searchInput != "" {
		filterStatus += fmt.Sprintf(" | Search: %s", m.searchInput)
	}
	filterBar := styles.InfoStyle.Render(filterStatus)

	// Content
	var content string
	switch {
	case len(m.filteredLogs) == 0:
		content = styles.MutedStyle.Render("No logs to display")
	case m.showDetails && m.cursor < len(m.filteredLogs):
		content = m.renderDetailView()
	default:
		content = m.renderListView()
	}

	// Action bar
	actions := []components.ActionBarItem{
		{Key: "/", Label: "Search", Active: true},
		{Key: "f", Label: "Filter", Active: true},
		{Key: "r", Label: "Refresh", Active: true},
		{Key: "c", Label: "Clear", Active: len(m.logs) > 0},
		{Key: "esc", Label: "Back", Active: true},
	}

	helpText := "↑↓: Navigate • Enter: Details • g/G: Top/Bottom"
	actionBar := components.RenderActionBar(m.width, actions, helpText)

	// Calculate heights
	headerHeight := lipgloss.Height(header)
	filterBarHeight := lipgloss.Height(filterBar)
	actionBarHeight := lipgloss.Height(actionBar)
	contentHeight := m.height - headerHeight - filterBarHeight - actionBarHeight - 2

	// Style content with fixed height
	contentStyle := lipgloss.NewStyle().
		Width(m.width).
		Height(contentHeight).
		Padding(0, 1)

	styledContent := contentStyle.Render(content)

	// Combine all parts
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		filterBar,
		styledContent,
		actionBar,
	)
}

func (m *ErrorLogModelV2) renderListView() string {
	var lines []string

	// Calculate visible range
	visibleHeight := m.height - 10
	start := m.scrollOffset
	end := start + visibleHeight
	if end > len(m.filteredLogs) {
		end = len(m.filteredLogs)
	}

	for i := start; i < end; i++ {
		log := m.filteredLogs[i]

		cursor := "  "
		if i == m.cursor {
			cursor = styles.HighlightStyle.Render("▶ ")
		}

		// Format log entry
		timestamp := log.Timestamp.Format("15:04:05")
		level := m.formatLevel(log.Level)
		message := truncateStringEL(log.Message, 60)

		line := fmt.Sprintf("%s%s %s %s",
			cursor,
			styles.LogTimestampStyle.Render(timestamp),
			level,
			message,
		)

		lines = append(lines, line)
	}

	// Scroll indicators
	if m.scrollOffset > 0 {
		lines = append([]string{styles.MutedStyle.Render("  ↑ More above")}, lines...)
	}
	if end < len(m.filteredLogs) {
		lines = append(lines, styles.MutedStyle.Render("  ↓ More below"))
	}

	return strings.Join(lines, "\n")
}

func (m *ErrorLogModelV2) renderDetailView() string {
	if m.cursor >= len(m.filteredLogs) {
		return ""
	}

	log := m.filteredLogs[m.cursor]

	details := make([]string, 0, 20) // Pre-allocate for typical detail view

	// Header
	titleStyle := styles.TitleStyle
	details = append(details, titleStyle.MarginBottom(0).Render("Log Entry Details"))
	details = append(details, "")

	// Basic info
	details = append(details, fmt.Sprintf("%s %s",
		styles.LogFieldKeyStyle.Render("Time:"),
		log.Timestamp.Format("2006-01-02 15:04:05"),
	))

	details = append(details, fmt.Sprintf("%s %s",
		styles.LogFieldKeyStyle.Render("Level:"),
		m.formatLevel(log.Level),
	))

	if log.File != "" {
		details = append(details, fmt.Sprintf("%s %s",
			styles.LogFieldKeyStyle.Render("File:"),
			styles.LogFileStyle.Render(log.File),
		))
	}

	details = append(details, "")
	details = append(details, styles.LogFieldKeyStyle.Render("Message:"))

	// Wrap message
	wrapped := wrapText(log.Message, m.width-4)
	for _, line := range wrapped {
		details = append(details, "  "+line)
	}

	// Fields
	if len(log.Fields) > 0 {
		details = append(details, "")
		details = append(details, styles.LogFieldKeyStyle.Render("Fields:"))
		for k, v := range log.Fields {
			details = append(details, fmt.Sprintf("  %s: %v",
				styles.LogFieldKeyStyle.Render(k),
				styles.LogFieldValueStyle.Render(fmt.Sprint(v)),
			))
		}
	}

	// Error details
	if log.Error != "" {
		details = append(details, "")
		details = append(details, styles.LogFieldKeyStyle.Render("Error:"))
		errorWrapped := wrapText(log.Error, m.width-4)
		for _, line := range errorWrapped {
			details = append(details, "  "+styles.ErrorStyle.Render(line))
		}
	}

	details = append(details, "")
	details = append(details, styles.MutedStyle.Render("Press Enter or Space to return to list"))

	return strings.Join(details, "\n")
}

func (m *ErrorLogModelV2) handleSearchMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.searchMode = false
		m.searchInput = ""
		m.applyFilter()
		return m, nil

	case "enter":
		m.searchMode = false
		m.applyFilter()
		return m, nil

	case "backspace":
		if len(m.searchInput) > 0 {
			m.searchInput = m.searchInput[:len(m.searchInput)-1]
		}

	default:
		if len(msg.String()) == 1 {
			m.searchInput += msg.String()
		}
	}

	return m, nil
}

func (m *ErrorLogModelV2) applyFilter() {
	m.filteredLogs = nil

	for _, log := range m.logs {
		// Apply level filter
		if m.filter != "all" && strings.ToLower(log.Level) != m.filter {
			continue
		}

		// Apply search filter
		if m.searchInput != "" {
			searchLower := strings.ToLower(m.searchInput)
			if !strings.Contains(strings.ToLower(log.Message), searchLower) &&
				!strings.Contains(strings.ToLower(log.Error), searchLower) {
				continue
			}
		}

		m.filteredLogs = append(m.filteredLogs, log)
	}

	// Reset cursor if needed
	if m.cursor >= len(m.filteredLogs) {
		m.cursor = len(m.filteredLogs) - 1
		if m.cursor < 0 {
			m.cursor = 0
		}
	}
	m.adjustScroll()
}

func (m *ErrorLogModelV2) adjustScroll() {
	visibleHeight := m.height - 10

	if m.cursor < m.scrollOffset {
		m.scrollOffset = m.cursor
	} else if m.cursor >= m.scrollOffset+visibleHeight {
		m.scrollOffset = m.cursor - visibleHeight + 1
	}

	if m.scrollOffset < 0 {
		m.scrollOffset = 0
	}
}

func (m *ErrorLogModelV2) formatLevel(level string) string {
	switch strings.ToLower(level) {
	case "error":
		return styles.LogLevelErrorStyle.Render("ERROR")
	case "warn", "warning":
		return styles.LogLevelWarnStyle.Render("WARN ")
	case "info":
		return styles.LogLevelInfoStyle.Render("INFO ")
	case "debug":
		return styles.LogLevelDebugStyle.Render("DEBUG")
	default:
		return styles.MutedStyle.Render(level)
	}
}

func (m *ErrorLogModelV2) loadLogs() tea.Cmd {
	return func() tea.Msg {
		// Mock some log entries for now
		logs := []ErrLogEntry{
			{Timestamp: time.Now().Add(-1 * time.Hour), Level: "info", Message: "GitCells started"},
			{Timestamp: time.Now().Add(-30 * time.Minute), Level: "warn", Message: "Large Excel file detected"},
			{Timestamp: time.Now().Add(-5 * time.Minute), Level: "error", Message: "Failed to convert file", Error: "Invalid format"},
		}
		return logsLoadedMsg{logs: logs}
	}
}

type logsLoadedMsg struct {
	logs []ErrLogEntry
}

func truncateStringEL(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

func wrapText(text string, width int) []string {
	if width <= 0 {
		return []string{text}
	}

	var lines []string
	words := strings.Fields(text)

	currentLine := ""
	for _, word := range words {
		if len(currentLine)+len(word)+1 <= width {
			if currentLine != "" {
				currentLine += " "
			}
			currentLine += word
		} else {
			if currentLine != "" {
				lines = append(lines, currentLine)
			}
			currentLine = word
		}
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return lines
}

// ErrLogEntry represents a single log entry for V2
type ErrLogEntry struct {
	Timestamp time.Time
	Level     string
	Message   string
	File      string
	Line      int
	Error     string
	Fields    map[string]interface{}
}
