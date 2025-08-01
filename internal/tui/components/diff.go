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
	d.selectedIdx++
}

func (d *DiffViewer) SelectPrev() {
	if d.selectedIdx > 0 {
		d.selectedIdx--
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
	if len(d.diff.SheetDiffs) == 0 {
		return "No sheet changes"
	}

	// Get current sheet
	if d.selectedIdx >= len(d.diff.SheetDiffs) {
		d.selectedIdx = len(d.diff.SheetDiffs) - 1
	}
	sheet := d.diff.SheetDiffs[d.selectedIdx]

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
		fmt.Sprintf("Sheet %d of %d", d.selectedIdx+1, len(d.diff.SheetDiffs)),
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
	if d.selectedIdx >= len(allChanges) {
		d.selectedIdx = len(allChanges) - 1
	}

	current := allChanges[d.selectedIdx]

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
		fmt.Sprintf("Change %d of %d", d.selectedIdx+1, len(allChanges)),
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
	// This would show a side-by-side comparison of specific cells or sheets
	// For now, show a placeholder
	boxStyle := styles.BoxStyle.
		Width(d.width - 10).
		Height(d.height - 5)

	leftPane := lipgloss.NewStyle().
		Width(d.width/2 - 6).
		Height(d.height - 8).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Muted).
		Padding(1).
		Render("Old Version\n\nSelect a cell to compare")

	rightPane := lipgloss.NewStyle().
		Width(d.width/2 - 6).
		Height(d.height - 8).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Muted).
		Padding(1).
		Render("New Version\n\nSelect a cell to compare")

	content := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftPane,
		"  ",
		rightPane,
	)

	help := styles.HelpStyle.Render("[Tab] Change view • [↑/↓] Select cell • [Enter] Compare")

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
