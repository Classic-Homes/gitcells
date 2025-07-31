package components

import (
	"fmt"

	"github.com/Classic-Homes/gitcells/internal/tui/styles"
	"github.com/charmbracelet/lipgloss"
)

type Table struct {
	headers      []string
	rows         [][]string
	widths       []int
	selectedRow  int
	showCursor   bool
	height       int
	scrollOffset int
}

func NewTable(headers []string) Table {
	t := Table{
		headers:     headers,
		rows:        [][]string{},
		widths:      make([]int, len(headers)),
		selectedRow: 0,
		showCursor:  true,
		height:      10,
	}

	// Set initial column widths based on headers
	for i, header := range headers {
		t.widths[i] = len(header) + 2
	}

	return t
}

func (t *Table) SetRows(rows [][]string) {
	t.rows = rows
	t.updateColumnWidths()
}

func (t *Table) AddRow(row []string) {
	if len(row) != len(t.headers) {
		return // Row must match header count
	}
	t.rows = append(t.rows, row)
	t.updateColumnWidths()
}

func (t *Table) ClearRows() {
	t.rows = [][]string{}
	t.selectedRow = 0
	t.scrollOffset = 0
}

func (t *Table) SetHeight(height int) {
	t.height = height
}

func (t *Table) SetShowCursor(show bool) {
	t.showCursor = show
}

func (t *Table) MoveUp() {
	if t.selectedRow > 0 {
		t.selectedRow--
		if t.selectedRow < t.scrollOffset {
			t.scrollOffset = t.selectedRow
		}
	}
}

func (t *Table) MoveDown() {
	if t.selectedRow < len(t.rows)-1 {
		t.selectedRow++
		if t.selectedRow >= t.scrollOffset+t.height {
			t.scrollOffset = t.selectedRow - t.height + 1
		}
	}
}

func (t *Table) SelectedIndex() int {
	return t.selectedRow
}

func (t *Table) SelectedRow() []string {
	if t.selectedRow >= 0 && t.selectedRow < len(t.rows) {
		return t.rows[t.selectedRow]
	}
	return nil
}

func (t Table) View() string {
	if len(t.headers) == 0 {
		return ""
	}

	// Styles
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.Primary).
		BorderBottom(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(styles.Muted)

	cellStyle := lipgloss.NewStyle().
		Padding(0, 1)

	selectedStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("236")).
		Foreground(styles.Primary)

	// Build header row
	var headerCells []string
	for i, header := range t.headers {
		cell := cellStyle.Width(t.widths[i]).Render(header)
		headerCells = append(headerCells, cell)
	}
	headerRow := headerStyle.Render(lipgloss.JoinHorizontal(lipgloss.Top, headerCells...))

	// Calculate visible rows
	visibleRows := t.rows
	if len(t.rows) > 0 && t.height > 0 && len(t.rows) > t.height {
		startIdx := t.scrollOffset
		if startIdx < 0 {
			startIdx = 0
		}
		if startIdx >= len(t.rows) {
			startIdx = len(t.rows) - 1
		}

		endIdx := startIdx + t.height
		if endIdx > len(t.rows) {
			endIdx = len(t.rows)
		}

		if startIdx < endIdx {
			visibleRows = t.rows[startIdx:endIdx]
		}
	}

	// Build data rows
	var dataRows []string
	for i, row := range visibleRows {
		actualIndex := i + t.scrollOffset
		var cells []string

		for j, cell := range row {
			if j < len(t.widths) {
				cellContent := truncate(cell, t.widths[j]-2)
				style := cellStyle.Width(t.widths[j])

				if t.showCursor && actualIndex == t.selectedRow {
					style = style.Inherit(selectedStyle)
				}

				cells = append(cells, style.Render(cellContent))
			}
		}

		rowStr := lipgloss.JoinHorizontal(lipgloss.Top, cells...)
		dataRows = append(dataRows, rowStr)
	}

	// Join all rows
	allRows := []string{headerRow}
	allRows = append(allRows, dataRows...)

	table := lipgloss.JoinVertical(lipgloss.Left, allRows...)

	// Add scroll indicators if needed
	if len(t.rows) > t.height {
		scrollInfo := fmt.Sprintf(" %d-%d of %d ",
			t.scrollOffset+1,
			t.scrollOffset+len(visibleRows),
			len(t.rows),
		)
		scrollStyle := styles.MutedStyle.
			Align(lipgloss.Right)

		table += "\n" + scrollStyle.Render(scrollInfo)
	}

	return table
}

func (t *Table) updateColumnWidths() {
	// Reset to header widths
	for i, header := range t.headers {
		t.widths[i] = len(header) + 2
	}

	// Update based on row content
	for _, row := range t.rows {
		for i, cell := range row {
			if i < len(t.widths) {
				cellWidth := len(cell) + 2
				if cellWidth > t.widths[i] {
					t.widths[i] = cellWidth
				}
			}
		}
	}

	// Cap maximum width
	for i := range t.widths {
		if t.widths[i] > 40 {
			t.widths[i] = 40
		}
	}
}

func truncate(s string, maxWidth int) string {
	if len(s) <= maxWidth {
		return s
	}
	if maxWidth <= 3 {
		return s[:maxWidth]
	}
	return s[:maxWidth-3] + "..."
}

// TableWithCommands is a table that can handle keyboard input
type TableWithCommands struct {
	Table
}

func NewTableWithCommands(headers []string) TableWithCommands {
	return TableWithCommands{
		Table: NewTable(headers),
	}
}

func (t TableWithCommands) Update(msg interface{}) TableWithCommands {
	switch msg := msg.(type) {
	case string:
		switch msg {
		case "up", "k":
			t.MoveUp()
		case "down", "j":
			t.MoveDown()
		case "g":
			t.selectedRow = 0
			t.scrollOffset = 0
		case "G":
			t.selectedRow = len(t.rows) - 1
			if t.selectedRow >= t.height {
				t.scrollOffset = t.selectedRow - t.height + 1
			}
		}
	}
	return t
}

// SimpleTable creates a quick table from data without interaction
func SimpleTable(headers []string, rows [][]string) string {
	table := NewTable(headers)
	table.SetRows(rows)
	table.SetShowCursor(false)
	return table.View()
}

// BorderedTable wraps a table in a nice border
func BorderedTable(title string, headers []string, rows [][]string) string {
	table := NewTable(headers)
	table.SetRows(rows)
	table.SetShowCursor(false)

	titleStyle := styles.SubtitleStyle.
		MarginBottom(1)

	boxStyle := styles.BoxStyle

	content := titleStyle.Render(title) + "\n" + table.View()
	return boxStyle.Render(content)
}
