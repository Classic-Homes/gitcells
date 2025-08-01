package components

import (
	"fmt"
	"strings"

	"github.com/Classic-Homes/gitcells/internal/tui/styles"
	"github.com/Classic-Homes/gitcells/pkg/models"
	"github.com/charmbracelet/lipgloss"
)

type DiffViewer struct {
	diff         *models.ExcelDiff
	width        int
	height       int
	scrollOffset int
	selectedIdx  int
	viewMode     DiffViewMode
	showDetails  bool
}

type DiffViewMode int

const (
	DiffViewSummary DiffViewMode = iota
	DiffViewBySheet
	DiffViewByCell
	DiffViewSideBySide
)

func NewDiffViewer(diff *models.ExcelDiff) DiffViewer {
	return DiffViewer{
		diff:        diff,
		viewMode:    DiffViewSummary,
		showDetails: false,
		width:       80,
		height:      20,
	}
}

func (d *DiffViewer) SetDimensions(width, height int) {
	d.width = width
	d.height = height
}

func (d *DiffViewer) NextMode() {
	d.viewMode = (d.viewMode + 1) % 4
	d.scrollOffset = 0
	d.selectedIdx = 0
	d.normalizeSelectedIndex()

	// Additional safety check - ensure we never have invalid indices
	if d.selectedIdx < 0 {
		d.selectedIdx = 0
	}
}

func (d *DiffViewer) normalizeSelectedIndex() {
	// Ensure diff is valid first
	if d.diff == nil {
		d.selectedIdx = 0
		return
	}

	maxIdx := d.getMaxSelectableIndex()
	switch {
	case maxIdx < 0:
		// No selectable items in this view mode
		d.selectedIdx = 0
	case d.selectedIdx > maxIdx:
		// Selected index is beyond available items
		d.selectedIdx = maxIdx
	case d.selectedIdx < 0:
		// Selected index is negative
		d.selectedIdx = 0
	}
	// If selectedIdx is within bounds, leave it unchanged
}

func (d *DiffViewer) ToggleDetails() {
	d.showDetails = !d.showDetails
}

func (d *DiffViewer) ScrollUp() {
	if d.scrollOffset > 0 {
		d.scrollOffset--
	}
}

func (d *DiffViewer) ScrollDown() {
	d.scrollOffset++
}

func (d *DiffViewer) SelectNext() {
	maxIdx := d.getMaxSelectableIndex()
	if maxIdx >= 0 && d.selectedIdx < maxIdx {
		d.selectedIdx++
	}
	// Ensure we never exceed bounds
	d.normalizeSelectedIndex()
}

func (d *DiffViewer) SelectPrev() {
	if d.selectedIdx > 0 {
		d.selectedIdx--
	}
	// Ensure we never go below bounds
	d.normalizeSelectedIndex()
}

func (d *DiffViewer) getMaxSelectableIndex() int {
	if d.diff == nil {
		return -1
	}

	switch d.viewMode {
	case DiffViewSummary:
		return -1 // No selection in summary view
	case DiffViewBySheet:
		if len(d.diff.SheetDiffs) == 0 {
			return -1
		}
		return len(d.diff.SheetDiffs) - 1
	case DiffViewByCell, DiffViewSideBySide:
		// Count all cell changes across all sheets
		totalChanges := 0
		for _, sheet := range d.diff.SheetDiffs {
			totalChanges += len(sheet.Changes)
		}
		if totalChanges == 0 {
			return -1
		}
		return totalChanges - 1
	default:
		return -1
	}
}

func (d DiffViewer) View() string {
	if d.diff == nil || !d.diff.HasChanges() {
		return styles.Center(d.width, d.height, styles.SuccessStyle.Render("✓ No changes detected"))
	}

	switch d.viewMode {
	case DiffViewSummary:
		return d.renderSummaryView()
	case DiffViewBySheet:
		return d.renderSheetView()
	case DiffViewByCell:
		return d.renderCellView()
	case DiffViewSideBySide:
		return d.renderSideBySideView()
	default:
		return d.renderSummaryView()
	}
}

func (d DiffViewer) renderSummaryView() string {
	// Additional safety checks
	if d.diff == nil {
		return styles.Center(d.width, d.height, "No diff data available")
	}

	titleStyle := styles.TitleStyle.
		MarginBottom(1)

	statStyle := lipgloss.NewStyle().
		Bold(true).
		Width(25).
		Padding(1).
		Margin(0, 1).
		Align(lipgloss.Center)

	// Title
	title := titleStyle.Render("Excel Diff Summary")

	// Statistics boxes
	var statBoxes []string

	// Added sheets
	if d.diff.Summary.AddedSheets > 0 {
		box := statStyle.
			Background(lipgloss.Color("28")).
			Foreground(lipgloss.Color("231")).
			Render(fmt.Sprintf("%d\nSheets Added", d.diff.Summary.AddedSheets))
		statBoxes = append(statBoxes, box)
	}

	// Modified sheets
	if d.diff.Summary.ModifiedSheets > 0 {
		box := statStyle.
			Background(lipgloss.Color("214")).
			Foreground(lipgloss.Color("231")).
			Render(fmt.Sprintf("%d\nSheets Modified", d.diff.Summary.ModifiedSheets))
		statBoxes = append(statBoxes, box)
	}

	// Deleted sheets
	if d.diff.Summary.DeletedSheets > 0 {
		box := statStyle.
			Background(lipgloss.Color("196")).
			Foreground(lipgloss.Color("231")).
			Render(fmt.Sprintf("%d\nSheets Deleted", d.diff.Summary.DeletedSheets))
		statBoxes = append(statBoxes, box)
	}

	// Cell changes
	if d.diff.Summary.CellChanges > 0 {
		box := statStyle.
			Background(lipgloss.Color("99")).
			Foreground(lipgloss.Color("231")).
			Render(fmt.Sprintf("%d\nCells Changed", d.diff.Summary.CellChanges))
		statBoxes = append(statBoxes, box)
	}

	stats := lipgloss.JoinHorizontal(lipgloss.Top, statBoxes...)

	// Sheet summary
	sheetSummary := d.renderSheetSummaryList()

	// Help
	help := styles.HelpStyle.Render("[Tab] Change view • [d] Toggle details • [↑/↓] Scroll")

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		"",
		stats,
		"",
		sheetSummary,
		"",
		help,
	)

	return lipgloss.Place(d.width, d.height, lipgloss.Center, lipgloss.Center, content)
}

func (d DiffViewer) renderSheetSummaryList() string {
	if len(d.diff.SheetDiffs) == 0 {
		return ""
	}

	boxStyle := styles.BoxStyle.
		Width(60).
		MaxHeight(10)

	lines := make([]string, 0, len(d.diff.SheetDiffs))
	for _, sheet := range d.diff.SheetDiffs {
		icon := d.getChangeIcon(sheet.Action)
		color := d.getChangeColor(sheet.Action)

		line := fmt.Sprintf("%s %s",
			icon,
			lipgloss.NewStyle().Foreground(color).Render(sheet.SheetName),
		)

		if len(sheet.Changes) > 0 {
			line += styles.MutedStyle.Render(fmt.Sprintf(" (%d changes)", len(sheet.Changes)))
		}

		lines = append(lines, line)
	}

	content := strings.Join(lines, "\n")
	return boxStyle.Render(content)
}

func (d DiffViewer) renderSheetView() string {
	// Additional safety checks
	if d.diff == nil || len(d.diff.SheetDiffs) == 0 {
		return "No sheet changes"
	}

	// Get current sheet - selectedIdx should already be valid
	selectedIdx := d.selectedIdx

	// Ensure we have sheets to work with
	if len(d.diff.SheetDiffs) == 0 {
		return "No sheet changes"
	}

	if selectedIdx >= len(d.diff.SheetDiffs) {
		selectedIdx = len(d.diff.SheetDiffs) - 1
	}
	if selectedIdx < 0 {
		selectedIdx = 0
	}

	// Double-check bounds before array access
	if selectedIdx >= len(d.diff.SheetDiffs) || selectedIdx < 0 {
		return "Invalid sheet selection"
	}

	sheet := d.diff.SheetDiffs[selectedIdx]

	// Header
	headerStyle := styles.TitleStyle.
		MarginBottom(1)

	header := headerStyle.Render(fmt.Sprintf("Sheet: %s", sheet.SheetName))

	// Sheet action
	var action string
	if sheet.Action != "" {
		actionStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(d.getChangeColor(sheet.Action))
		action = actionStyle.Render(fmt.Sprintf("Action: %s", sheet.Action))
	}

	// Changes table
	table := NewTable([]string{"Cell", "Type", "Old Value", "New Value"})
	table.SetHeight(d.height - 10)

	for _, change := range sheet.Changes {
		oldVal := d.formatCellValue(change.OldValue, change.OldFormula)
		newVal := d.formatCellValue(change.NewValue, change.NewFormula)

		table.AddRow([]string{
			change.Cell,
			string(change.Type),
			oldVal,
			newVal,
		})
	}

	// Navigation info
	nav := styles.MutedStyle.Render(
		fmt.Sprintf("Sheet %d of %d", selectedIdx+1, len(d.diff.SheetDiffs)),
	)

	// Help
	help := styles.HelpStyle.Render("[←/→] Navigate sheets • [Tab] Change view • [↑/↓] Scroll")

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		action,
		"",
		table.View(),
		"",
		nav,
		help,
	)

	return lipgloss.NewStyle().
		Padding(1, 2).
		Render(content)
}

func (d DiffViewer) renderCellView() string {
	// Additional safety checks
	if d.diff == nil {
		return "No diff data available"
	}

	// Flatten all cell changes
	var allChanges []struct {
		sheet  string
		change models.CellChange
	}

	for _, sheet := range d.diff.SheetDiffs {
		for _, change := range sheet.Changes {
			allChanges = append(allChanges, struct {
				sheet  string
				change models.CellChange
			}{
				sheet:  sheet.SheetName,
				change: change,
			})
		}
	}

	if len(allChanges) == 0 {
		return "No cell changes"
	}

	// Ensure selected index is valid
	selectedIdx := d.selectedIdx
	if selectedIdx >= len(allChanges) {
		selectedIdx = len(allChanges) - 1
	}
	if selectedIdx < 0 {
		selectedIdx = 0
	}

	// Double-check bounds before array access
	if selectedIdx >= len(allChanges) || selectedIdx < 0 {
		return "Invalid cell selection"
	}

	current := allChanges[selectedIdx]

	// Create a detailed view of the selected cell
	detailBox := styles.BoxStyle.
		Width(70).
		Padding(1)

	icon := d.getChangeIcon(current.change.Type)
	typeStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(d.getChangeColor(current.change.Type))

	header := fmt.Sprintf("%s Cell %s - %s",
		icon,
		styles.TitleStyle.Render(current.change.Cell),
		typeStyle.Render(string(current.change.Type)),
	)

	details := []string{
		header,
		styles.MutedStyle.Render(fmt.Sprintf("Sheet: %s", current.sheet)),
		"",
	}

	// Show values
	if current.change.OldValue != nil || current.change.OldFormula != "" {
		oldContent := d.formatDetailedCellContent("Old", current.change.OldValue, current.change.OldFormula)
		details = append(details, oldContent)
	}

	if current.change.NewValue != nil || current.change.NewFormula != "" {
		newContent := d.formatDetailedCellContent("New", current.change.NewValue, current.change.NewFormula)
		details = append(details, newContent)
	}

	if current.change.Description != "" {
		details = append(details, "", styles.MutedStyle.Render(current.change.Description))
	}

	detailContent := detailBox.Render(strings.Join(details, "\n"))

	// Navigation
	nav := styles.MutedStyle.Render(
		fmt.Sprintf("Change %d of %d", selectedIdx+1, len(allChanges)),
	)

	// Help
	help := styles.HelpStyle.Render("[←/→] Navigate cells • [Tab] Change view")

	return lipgloss.Place(
		d.width, d.height,
		lipgloss.Center, lipgloss.Center,
		lipgloss.JoinVertical(
			lipgloss.Center,
			detailContent,
			"",
			nav,
			help,
		),
	)
}

func (d DiffViewer) renderSideBySideView() string {
	// Additional safety checks
	if d.diff == nil {
		return d.renderEmptySideBySideView()
	}

	// Flatten all cell changes for navigation
	var allChanges []struct {
		sheet  string
		change models.CellChange
	}

	for _, sheet := range d.diff.SheetDiffs {
		for _, change := range sheet.Changes {
			allChanges = append(allChanges, struct {
				sheet  string
				change models.CellChange
			}{
				sheet:  sheet.SheetName,
				change: change,
			})
		}
	}

	if len(allChanges) == 0 {
		return d.renderEmptySideBySideView()
	}

	// Ensure selected index is valid
	selectedIdx := d.selectedIdx
	if selectedIdx >= len(allChanges) {
		selectedIdx = len(allChanges) - 1
	}
	if selectedIdx < 0 {
		selectedIdx = 0
	}

	// Double-check bounds before array access
	if selectedIdx >= len(allChanges) || selectedIdx < 0 {
		return d.renderEmptySideBySideView()
	}

	current := allChanges[selectedIdx]

	// Calculate pane dimensions with minimum width safeguards
	paneWidth := (d.width - 12) / 2 // Account for spacing and borders
	if paneWidth < 20 {             // Minimum usable width
		paneWidth = 20
	}
	paneHeight := d.height - 10 // Account for title and help
	if paneHeight < 5 {         // Minimum usable height
		paneHeight = 5
	}

	// Create left pane (old version)
	leftContent := d.renderSideBySidePane("Old Version", current.change.OldValue, current.change.OldFormula, paneWidth, paneHeight, current.change.Type == models.ChangeTypeAdd)

	// Create right pane (new version)
	rightContent := d.renderSideBySidePane("New Version", current.change.NewValue, current.change.NewFormula, paneWidth, paneHeight, current.change.Type == models.ChangeTypeDelete)

	// Header with cell info
	headerStyle := styles.TitleStyle.MarginBottom(1)
	cellIcon := d.getChangeIcon(current.change.Type)
	cellTypeStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(d.getChangeColor(current.change.Type))

	header := headerStyle.Render(fmt.Sprintf("%s Cell %s (%s) - %s",
		cellIcon,
		current.change.Cell,
		current.sheet,
		cellTypeStyle.Render(string(current.change.Type)),
	))

	// Join panes horizontally
	content := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftContent,
		"  ",
		rightContent,
	)

	// Navigation info
	nav := styles.MutedStyle.Render(
		fmt.Sprintf("Change %d of %d", selectedIdx+1, len(allChanges)),
	)

	// Help text
	help := styles.HelpStyle.Render("[←/→] Navigate changes • [Tab] Change view • [d] Toggle details")

	// Combine everything
	fullContent := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		content,
		"",
		nav,
		help,
	)

	return lipgloss.NewStyle().
		Padding(1, 2).
		Render(fullContent)
}

func (d DiffViewer) renderEmptySideBySideView() string {
	// Calculate safe dimensions
	boxWidth := d.width - 10
	if boxWidth < 20 {
		boxWidth = 20
	}
	boxHeight := d.height - 5
	if boxHeight < 10 {
		boxHeight = 10
	}

	boxStyle := styles.BoxStyle.
		Width(boxWidth).
		Height(boxHeight)

	paneWidth := d.width/2 - 6
	if paneWidth < 15 {
		paneWidth = 15
	}
	paneHeight := d.height - 8
	if paneHeight < 5 {
		paneHeight = 5
	}

	leftPane := lipgloss.NewStyle().
		Width(paneWidth).
		Height(paneHeight).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Muted).
		Padding(1).
		Render("Old Version\n\nNo changes to display")

	rightPane := lipgloss.NewStyle().
		Width(paneWidth).
		Height(paneHeight).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Muted).
		Padding(1).
		Render("New Version\n\nNo changes to display")

	content := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftPane,
		"  ",
		rightPane,
	)

	help := styles.HelpStyle.Render("[Tab] Change view")

	return boxStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			styles.TitleStyle.Render("Side-by-Side Comparison"),
			"",
			content,
			"",
			help,
		),
	)
}

func (d DiffViewer) renderSideBySidePane(title string, value interface{}, formula string, width, height int, isEmpty bool) string {
	// Determine border color based on content
	borderColor := styles.Muted
	if isEmpty {
		borderColor = lipgloss.Color("240") // Dimmer for empty panes
	}

	paneStyle := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(1)

	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Underline(true).
		MarginBottom(1)

	content := []string{titleStyle.Render(title)}

	if isEmpty {
		content = append(content, "", styles.MutedStyle.Render("<empty>"))
	} else {
		// Show formula if present
		if formula != "" {
			formulaStyle := lipgloss.NewStyle().
				Foreground(styles.Primary).
				Bold(true)
			content = append(content, "", "Formula:")
			content = append(content, formulaStyle.Render(fmt.Sprintf("=%s", formula)))
		}

		// Show value
		if value != nil {
			valueStr := fmt.Sprintf("%v", value)
			content = append(content, "", "Value:")

			// Wrap long values
			wrapWidth := width - 6
			if wrapWidth < 10 { // Ensure minimum wrap width
				wrapWidth = 10
			}
			if len(valueStr) > wrapWidth {
				wrapped := d.wrapText(valueStr, wrapWidth)
				content = append(content, wrapped)
			} else {
				content = append(content, valueStr)
			}
		} else if formula == "" {
			content = append(content, "", styles.MutedStyle.Render("<empty>"))
		}

		// Show type information if detailed view is enabled
		if d.showDetails && value != nil {
			content = append(content, "", styles.MutedStyle.Render(fmt.Sprintf("Type: %T", value)))
		}
	}

	return paneStyle.Render(strings.Join(content, "\n"))
}

func (d DiffViewer) wrapText(text string, width int) string {
	// Safety check for negative or too small widths
	if width <= 0 {
		return text // Return original text if width is invalid
	}
	if len(text) <= width {
		return text
	}

	var lines []string
	for len(text) > 0 {
		if len(text) <= width {
			lines = append(lines, text)
			break
		}

		// Find last space before width
		cutPoint := width
		for i := width; i > 0; i-- {
			if text[i-1] == ' ' {
				cutPoint = i
				break
			}
		}

		lines = append(lines, text[:cutPoint])
		text = text[cutPoint:]
		if len(text) > 0 && text[0] == ' ' {
			text = text[1:] // Remove leading space
		}
	}

	return strings.Join(lines, "\n")
}

// Helper methods
func (d DiffViewer) getChangeIcon(changeType interface{}) string {
	if ct, ok := changeType.(models.ChangeType); ok {
		switch ct {
		case models.ChangeTypeAdd:
			return "+"
		case models.ChangeTypeModify:
			return "~"
		case models.ChangeTypeDelete:
			return "-"
		}
	}
	return "•"
}

func (d DiffViewer) getChangeColor(changeType interface{}) lipgloss.Color {
	if ct, ok := changeType.(models.ChangeType); ok {
		switch ct {
		case models.ChangeTypeAdd:
			return styles.Success
		case models.ChangeTypeModify:
			return styles.Warning
		case models.ChangeTypeDelete:
			return styles.Error
		}
	}
	return styles.Muted
}

func (d DiffViewer) formatCellValue(value interface{}, formula string) string {
	if formula != "" {
		return fmt.Sprintf("=%s", formula)
	}
	if value == nil {
		return "<empty>"
	}

	// Truncate long values
	str := fmt.Sprintf("%v", value)
	if len(str) > 30 {
		return str[:27] + "..."
	}
	return str
}

func (d DiffViewer) formatDetailedCellContent(label string, value interface{}, formula string) string {
	labelStyle := lipgloss.NewStyle().
		Bold(true).
		Underline(true)

	content := labelStyle.Render(label + ":")

	if formula != "" {
		formulaStyle := lipgloss.NewStyle().
			Foreground(styles.Primary)
		content += fmt.Sprintf("\n  Formula: %s", formulaStyle.Render(formula))
	}

	if value != nil {
		valueStr := fmt.Sprintf("%v", value)
		// For long values, wrap them
		if len(valueStr) > 50 {
			wrapped := wordWrap(valueStr, 50)
			content += fmt.Sprintf("\n  Value: %s", wrapped)
		} else {
			content += fmt.Sprintf("\n  Value: %s", valueStr)
		}
	} else if formula == "" {
		content += "\n  Value: <empty>"
	}

	return content
}

func wordWrap(text string, width int) string {
	// Safety check for negative or too small widths
	if width <= 0 {
		return text // Return original text if width is invalid
	}
	if len(text) <= width {
		return text
	}

	var lines []string
	for len(text) > 0 {
		if len(text) <= width {
			lines = append(lines, text)
			break
		}

		// Find last space before width
		cutPoint := width
		for i := width; i > 0; i-- {
			if text[i-1] == ' ' {
				cutPoint = i
				break
			}
		}

		lines = append(lines, text[:cutPoint])
		text = text[cutPoint:]
		if len(text) > 0 && text[0] == ' ' {
			text = text[1:] // Remove leading space
		}
	}

	return strings.Join(lines, "\n         ")
}

// SimpleDiff creates a quick diff view from two Excel documents
func SimpleDiff(oldDoc, newDoc *models.ExcelDocument) string {
	diff := models.ComputeDiff(oldDoc, newDoc)
	viewer := NewDiffViewer(diff)
	viewer.SetDimensions(80, 30)
	return viewer.View()
}
