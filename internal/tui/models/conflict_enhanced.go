package models

import (
	"fmt"
	"strings"

	"github.com/Classic-Homes/gitcells/internal/tui/components"
	"github.com/Classic-Homes/gitcells/internal/tui/styles"
	"github.com/Classic-Homes/gitcells/pkg/models"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ConflictEnhancedModel struct {
	width  int
	height int

	// Conflict data
	conflicts  []ExcelConflict
	currentIdx int

	// UI components
	diffViewer components.DiffViewer

	// State
	mode        ConflictMode
	resolution  map[string]ConflictResolution
	showDiff    bool
}

type ConflictMode int

const (
	ConflictModeList ConflictMode = iota
	ConflictModeResolve
	ConflictModePreview
	ConflictModeDiff
)

type ExcelConflict struct {
	File         string
	Sheet        string
	Cell         string
	BaseValue    interface{}
	OurValue     interface{}
	TheirValue   interface{}
	BaseFormula  string
	OurFormula   string
	TheirFormula string
	HasFormula   bool
}

type ConflictResolution struct {
	Choice        ResolutionChoice
	CustomValue   interface{}
	CustomFormula string
}

type ResolutionChoice int

const (
	ChoiceUnresolved ResolutionChoice = iota
	ChoiceOurs
	ChoiceTheirs
	ChoiceBase
	ChoiceCustom
)

func NewConflictEnhancedModel() tea.Model {
	m := &ConflictEnhancedModel{
		resolution: make(map[string]ConflictResolution),
		mode:       ConflictModeList,
		width:      80, // Default width
		height:     24, // Default height
	}

	// Load mock conflicts for demo
	m.loadMockConflicts()

	return m
}

func (m *ConflictEnhancedModel) loadMockConflicts() {
	m.conflicts = []ExcelConflict{
		{
			File:         "Budget2024.xlsx",
			Sheet:        "Summary",
			Cell:         "B15",
			BaseFormula:  "=SUM(B10:B12)",
			OurFormula:   "=SUM(B10:B13)",
			TheirFormula: "=SUM(B10:B14)",
			HasFormula:   true,
		},
		{
			File:       "Budget2024.xlsx",
			Sheet:      "Q1",
			Cell:       "D20",
			BaseValue:  1000000,
			OurValue:   1250000,
			TheirValue: 1275000,
			HasFormula: false,
		},
		{
			File:         "Reports.xlsx",
			Sheet:        "Annual",
			Cell:         "F5",
			BaseFormula:  "=AVERAGE(A1:A12)",
			OurValue:     "Manual Override: 95.5",
			TheirFormula: "=AVERAGE(A1:A12)*1.05",
			HasFormula:   true,
		},
	}
}

func (m ConflictEnhancedModel) Init() tea.Cmd {
	return nil
}

func (m ConflictEnhancedModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.diffViewer.View() != "" {
			m.diffViewer.SetDimensions(m.width-4, m.height-10)
		}

	case tea.KeyMsg:
		switch m.mode {
		case ConflictModeList:
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "up", "k":
				if m.currentIdx > 0 {
					m.currentIdx--
				}
			case "down", "j":
				if m.currentIdx < len(m.conflicts)-1 {
					m.currentIdx++
				}
			case "enter", "r":
				m.mode = ConflictModeResolve
			case "d":
				m.showDiff = !m.showDiff
				if m.showDiff {
					m.updateDiffViewer()
				}
			case "p":
				if m.hasResolutions() {
					m.mode = ConflictModePreview
				}
			case "a":
				m.applyAllResolutions()
			}

		case ConflictModeResolve:
			current := m.getCurrentConflict()
			key := m.getConflictKey(current)

			switch msg.String() {
			case "esc":
				m.mode = ConflictModeList
			case "o":
				m.resolution[key] = ConflictResolution{Choice: ChoiceOurs}
				m.nextConflict()
			case "t":
				m.resolution[key] = ConflictResolution{Choice: ChoiceTheirs}
				m.nextConflict()
			case "b":
				if current.BaseValue != nil || current.BaseFormula != "" {
					m.resolution[key] = ConflictResolution{Choice: ChoiceBase}
					m.nextConflict()
				}
			case "c":
				// TODO: Implement custom value input
				m.resolution[key] = ConflictResolution{Choice: ChoiceCustom}
				m.nextConflict()
			case "s":
				// Skip this conflict
				delete(m.resolution, key)
				m.nextConflict()
			}

		case ConflictModePreview:
			switch msg.String() {
			case "esc":
				m.mode = ConflictModeList
			case "enter", "a":
				m.applyAllResolutions()
			}
		}
	}

	// Update diff viewer if active
	if m.showDiff && m.diffViewer.View() != "" {
		// Pass through navigation keys
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "tab":
				m.diffViewer.NextMode()
			}
		}
	}

	return m, nil
}

func (m ConflictEnhancedModel) View() string {
	switch m.mode {
	case ConflictModeResolve:
		return m.renderResolveView()
	case ConflictModePreview:
		return m.renderPreviewView()
	default:
		return m.renderListView()
	}
}

func (m ConflictEnhancedModel) renderListView() string {
	containerStyle := lipgloss.NewStyle().
		Padding(2, 4).
		Width(m.width).
		Height(m.height)

	title := styles.TitleStyle.Render("Excel Merge Conflicts")

	// Summary
	resolved := m.countResolved()
	summary := fmt.Sprintf("Total: %d conflicts | Resolved: %d | Remaining: %d",
		len(m.conflicts), resolved, len(m.conflicts)-resolved)
	summaryStyle := styles.MutedStyle
	if resolved == len(m.conflicts) {
		summaryStyle = styles.SuccessStyle
	}
	summaryLine := summaryStyle.Render(summary)

	// Conflict list
	table := components.NewTable([]string{"", "File", "Sheet", "Cell", "Status"})
	if m.height > 15 {
		table.SetHeight(m.height - 15)
	} else {
		table.SetHeight(10) // Default minimum height
	}

	for i, conflict := range m.conflicts {
		status := "Unresolved"
		statusColor := styles.Warning

		if resolution, exists := m.resolution[m.getConflictKey(&conflict)]; exists {
			status = m.getResolutionLabel(resolution)
			statusColor = styles.Success
		}

		selected := ""
		if i == m.currentIdx {
			selected = "▶"
		}

		table.AddRow([]string{
			selected,
			conflict.File,
			conflict.Sheet,
			conflict.Cell,
			lipgloss.NewStyle().Foreground(statusColor).Render(status),
		})
	}

	// Show diff viewer if enabled
	var content string
	if m.showDiff && m.currentIdx < len(m.conflicts) {
		// Split view
		listBox := lipgloss.NewStyle().
			Width(m.width/2 - 2).
			Height(m.height - 10).
			Render(table.View())

		diffBox := lipgloss.NewStyle().
			Width(m.width/2 - 2).
			Height(m.height - 10).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(styles.Muted).
			Padding(1).
			Render(m.renderConflictDiff())

		content = lipgloss.JoinHorizontal(lipgloss.Top, listBox, diffBox)
	} else {
		content = table.View()
	}

	// Help
	help := styles.HelpStyle.Render(
		"[↑/↓] Navigate • [Enter] Resolve • [d] Toggle diff • [p] Preview • [a] Apply all",
	)

	return containerStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			title,
			summaryLine,
			"",
			content,
			"",
			help,
		),
	)
}

func (m ConflictEnhancedModel) renderResolveView() string {
	if m.currentIdx >= len(m.conflicts) {
		return m.renderListView()
	}

	conflict := m.conflicts[m.currentIdx]

	boxStyle := styles.BoxStyle.
		Width(80).
		Padding(2)

	title := styles.TitleStyle.Render(fmt.Sprintf("Resolve Conflict: %s", conflict.Cell))
	location := styles.MutedStyle.Render(fmt.Sprintf("%s - %s", conflict.File, conflict.Sheet))

	// Value boxes
	valueStyle := lipgloss.NewStyle().
		Width(70).
		Padding(1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Muted)

	var sections []string

	// Base version (if exists)
	if conflict.BaseValue != nil || conflict.BaseFormula != "" {
		baseBox := valueStyle.
			BorderForeground(styles.Muted).
			Render(m.formatConflictValue("Base (Original)", conflict.BaseValue, conflict.BaseFormula))
		sections = append(sections, baseBox)
	}

	// Our version
	ourBox := valueStyle.
		BorderForeground(styles.Success).
		Render(m.formatConflictValue("Ours (Current)", conflict.OurValue, conflict.OurFormula))
	sections = append(sections, ourBox)

	// Their version
	theirBox := valueStyle.
		BorderForeground(styles.Primary).
		Render(m.formatConflictValue("Theirs (Incoming)", conflict.TheirValue, conflict.TheirFormula))
	sections = append(sections, theirBox)

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		location,
		"",
		strings.Join(sections, "\n\n"),
		"",
		m.renderResolutionHelp(conflict),
	)

	return styles.Center(m.width, m.height, boxStyle.Render(content))
}

func (m ConflictEnhancedModel) renderPreviewView() string {
	boxStyle := styles.BoxStyle.
		Width(m.width - 10).
		Height(m.height - 5).
		Padding(2)

	title := styles.TitleStyle.Render("Resolution Preview")

	// Group resolutions by file
	fileGroups := make(map[string][]string)
	for _, conflict := range m.conflicts {
		key := m.getConflictKey(&conflict)
		if resolution, exists := m.resolution[key]; exists {
			line := fmt.Sprintf("  %s!%s: %s",
				conflict.Sheet,
				conflict.Cell,
				m.getResolutionLabel(resolution),
			)
			fileGroups[conflict.File] = append(fileGroups[conflict.File], line)
		}
	}

	var preview []string
	for file, resolutions := range fileGroups {
		preview = append(preview, styles.SubtitleStyle.Render(file))
		preview = append(preview, resolutions...)
		preview = append(preview, "")
	}

	help := styles.HelpStyle.Render("[Enter] Apply resolutions • [Esc] Back to list")

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		strings.Join(preview, "\n"),
		"",
		help,
	)

	return styles.Center(m.width, m.height, boxStyle.Render(content))
}

func (m ConflictEnhancedModel) renderConflictDiff() string {
	if m.currentIdx >= len(m.conflicts) {
		return "No conflict selected"
	}

	conflict := m.conflicts[m.currentIdx]

	// Create a simple diff view for the current conflict
	var content []string

	content = append(content,
		styles.SubtitleStyle.Render(fmt.Sprintf("%s!%s", conflict.Sheet, conflict.Cell)),
		"",
	)

	if conflict.HasFormula {
		content = append(content, "Formula Conflict:")
		if conflict.BaseFormula != "" {
			content = append(content, fmt.Sprintf("  Base: %s", styles.MutedStyle.Render(conflict.BaseFormula)))
		}
		content = append(content, fmt.Sprintf("  Ours: %s", styles.SuccessStyle.Render(conflict.OurFormula)))
		content = append(content, fmt.Sprintf("  Theirs: %s", lipgloss.NewStyle().Foreground(styles.Primary).Render(conflict.TheirFormula)))
	} else {
		content = append(content, "Value Conflict:")
		if conflict.BaseValue != nil {
			content = append(content, fmt.Sprintf("  Base: %v", conflict.BaseValue))
		}
		content = append(content, fmt.Sprintf("  Ours: %v", conflict.OurValue))
		content = append(content, fmt.Sprintf("  Theirs: %v", conflict.TheirValue))
	}

	return strings.Join(content, "\n")
}

// Helper methods
func (m ConflictEnhancedModel) formatConflictValue(label string, value interface{}, formula string) string {
	labelStyle := lipgloss.NewStyle().
		Bold(true).
		Underline(true)

	content := labelStyle.Render(label)

	if formula != "" {
		formulaStyle := lipgloss.NewStyle().
			Foreground(styles.Primary)
		content += fmt.Sprintf("\n\nFormula: %s", formulaStyle.Render(formula))
	}

	if value != nil {
		content += fmt.Sprintf("\n\nValue: %v", value)
	} else if formula == "" {
		content += "\n\nValue: <empty>"
	}

	return content
}

func (m ConflictEnhancedModel) renderResolutionHelp(conflict ExcelConflict) string {
	var options []string

	options = append(options, "[o] Use Ours")
	options = append(options, "[t] Use Theirs")

	if conflict.BaseValue != nil || conflict.BaseFormula != "" {
		options = append(options, "[b] Use Base")
	}

	options = append(options, "[c] Custom value")
	options = append(options, "[s] Skip")
	options = append(options, "[Esc] Cancel")

	return styles.HelpStyle.Render(strings.Join(options, " • "))
}

func (m *ConflictEnhancedModel) getCurrentConflict() *ExcelConflict {
	if m.currentIdx < len(m.conflicts) {
		return &m.conflicts[m.currentIdx]
	}
	return nil
}

func (m *ConflictEnhancedModel) getConflictKey(conflict *ExcelConflict) string {
	return fmt.Sprintf("%s:%s:%s", conflict.File, conflict.Sheet, conflict.Cell)
}

func (m *ConflictEnhancedModel) nextConflict() {
	if m.currentIdx < len(m.conflicts)-1 {
		m.currentIdx++
	} else {
		m.mode = ConflictModeList
	}
}

func (m ConflictEnhancedModel) countResolved() int {
	count := 0
	for _, conflict := range m.conflicts {
		if _, exists := m.resolution[m.getConflictKey(&conflict)]; exists {
			count++
		}
	}
	return count
}

func (m ConflictEnhancedModel) hasResolutions() bool {
	return len(m.resolution) > 0
}

func (m ConflictEnhancedModel) getResolutionLabel(resolution ConflictResolution) string {
	switch resolution.Choice {
	case ChoiceOurs:
		return "Use Ours"
	case ChoiceTheirs:
		return "Use Theirs"
	case ChoiceBase:
		return "Use Base"
	case ChoiceCustom:
		return "Custom"
	default:
		return "Unresolved"
	}
}

func (m *ConflictEnhancedModel) updateDiffViewer() {
	// In a real implementation, this would create a diff between versions
	// For now, we'll create a mock diff
	mockDiff := &models.ExcelDiff{
		Summary: models.DiffSummary{
			TotalChanges:   len(m.conflicts),
			ModifiedSheets: 2,
			CellChanges:    len(m.conflicts),
		},
		SheetDiffs: []models.SheetDiff{},
	}

	// Group conflicts by sheet
	sheetConflicts := make(map[string][]models.CellChange)
	for _, conflict := range m.conflicts {
		change := models.CellChange{
			Cell:       conflict.Cell,
			Type:       models.ChangeTypeModify,
			OldValue:   conflict.OurValue,
			NewValue:   conflict.TheirValue,
			OldFormula: conflict.OurFormula,
			NewFormula: conflict.TheirFormula,
		}
		sheetConflicts[conflict.Sheet] = append(sheetConflicts[conflict.Sheet], change)
	}

	for sheet, changes := range sheetConflicts {
		mockDiff.SheetDiffs = append(mockDiff.SheetDiffs, models.SheetDiff{
			SheetName: sheet,
			Changes:   changes,
		})
	}

	m.diffViewer = components.NewDiffViewer(mockDiff)
	m.diffViewer.SetDimensions(m.width/2-4, m.height-10)
}

func (m *ConflictEnhancedModel) applyAllResolutions() {
	// In a real implementation, this would apply the resolutions
	// For now, just clear the mode
	m.mode = ConflictModeList
}
