package models

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"
)

const (
	noChangesMsg      = "No changes detected"
	noChangesMsgColor = "\033[32mNo changes detected\033[0m"
	colorGreen        = "\033[32m"
	colorYellow       = "\033[33m"
	colorRed          = "\033[31m"
	colorReset        = "\033[0m"
)

type ExcelDiff struct {
	Timestamp  time.Time   `json:"timestamp"`
	Summary    DiffSummary `json:"summary"`
	SheetDiffs []SheetDiff `json:"sheet_diffs"`
}

type DiffSummary struct {
	TotalChanges   int `json:"total_changes"`
	AddedSheets    int `json:"added_sheets"`
	ModifiedSheets int `json:"modified_sheets"`
	DeletedSheets  int `json:"deleted_sheets"`
	CellChanges    int `json:"cell_changes"`
}

type SheetDiff struct {
	SheetName string       `json:"sheet_name"`
	Action    ChangeType   `json:"action,omitempty"`
	Changes   []CellChange `json:"changes"`
}

type CellChange struct {
	Cell        string      `json:"cell"`
	Type        ChangeType  `json:"type"`
	OldValue    interface{} `json:"old_value,omitempty"`
	NewValue    interface{} `json:"new_value,omitempty"`
	OldFormula  string      `json:"old_formula,omitempty"`
	NewFormula  string      `json:"new_formula,omitempty"`
	Description string      `json:"description,omitempty"`
}

type ChangeType string

const (
	ChangeTypeAdd    ChangeType = "add"
	ChangeTypeModify ChangeType = "modify"
	ChangeTypeDelete ChangeType = "delete"
)

// ComputeDiff computes the differences between two Excel documents
func ComputeDiff(oldDoc, newDoc *ExcelDocument) *ExcelDiff {
	diff := &ExcelDiff{
		Timestamp:  time.Now(),
		SheetDiffs: []SheetDiff{},
	}

	// Create maps for easy lookup
	oldSheets := make(map[string]*Sheet)
	newSheets := make(map[string]*Sheet)

	for i := range oldDoc.Sheets {
		oldSheets[oldDoc.Sheets[i].Name] = &oldDoc.Sheets[i]
	}

	for i := range newDoc.Sheets {
		newSheets[newDoc.Sheets[i].Name] = &newDoc.Sheets[i]
	}

	// Find all unique sheet names
	allSheetNames := make(map[string]bool)
	for name := range oldSheets {
		allSheetNames[name] = true
	}
	for name := range newSheets {
		allSheetNames[name] = true
	}

	// Compare each sheet
	for sheetName := range allSheetNames {
		oldSheet, hasOld := oldSheets[sheetName]
		newSheet, hasNew := newSheets[sheetName]

		sheetDiff := SheetDiff{
			SheetName: sheetName,
			Changes:   []CellChange{},
		}

		switch {
		case !hasOld && hasNew:
			// Sheet added
			sheetDiff.Action = ChangeTypeAdd
			diff.Summary.AddedSheets++

			// Add all cells as new
			for cellRef, cell := range newSheet.Cells {
				sheetDiff.Changes = append(sheetDiff.Changes, CellChange{
					Cell:        cellRef,
					Type:        ChangeTypeAdd,
					NewValue:    cell.Value,
					NewFormula:  cell.Formula,
					Description: "New cell in added sheet",
				})
			}
		case hasOld && !hasNew:
			// Sheet deleted
			sheetDiff.Action = ChangeTypeDelete
			diff.Summary.DeletedSheets++

			// Add all cells as deleted
			for cellRef, cell := range oldSheet.Cells {
				sheetDiff.Changes = append(sheetDiff.Changes, CellChange{
					Cell:        cellRef,
					Type:        ChangeTypeDelete,
					OldValue:    cell.Value,
					OldFormula:  cell.Formula,
					Description: "Cell removed with deleted sheet",
				})
			}
		case hasOld && hasNew:
			// Sheet exists in both, compare cells
			cellChanges := compareCells(oldSheet.Cells, newSheet.Cells)
			if len(cellChanges) > 0 {
				sheetDiff.Changes = cellChanges
				diff.Summary.ModifiedSheets++
			}
		}

		if len(sheetDiff.Changes) > 0 || sheetDiff.Action != "" {
			diff.SheetDiffs = append(diff.SheetDiffs, sheetDiff)
		}
	}

	// Calculate totals
	for _, sheetDiff := range diff.SheetDiffs {
		diff.Summary.CellChanges += len(sheetDiff.Changes)
	}
	diff.Summary.TotalChanges = diff.Summary.AddedSheets + diff.Summary.ModifiedSheets + diff.Summary.DeletedSheets

	return diff
}

// compareCells compares the cells between two sheets
func compareCells(oldCells, newCells map[string]Cell) []CellChange {
	var changes []CellChange

	// Find all unique cell references
	allCells := make(map[string]bool)
	for cellRef := range oldCells {
		allCells[cellRef] = true
	}
	for cellRef := range newCells {
		allCells[cellRef] = true
	}

	// Compare each cell
	for cellRef := range allCells {
		oldCell, hasOld := oldCells[cellRef]
		newCell, hasNew := newCells[cellRef]

		switch {
		case !hasOld && hasNew:
			// Cell added
			changes = append(changes, CellChange{
				Cell:        cellRef,
				Type:        ChangeTypeAdd,
				NewValue:    newCell.Value,
				NewFormula:  newCell.Formula,
				Description: describeCellChange(nil, &newCell),
			})
		case hasOld && !hasNew:
			// Cell deleted
			changes = append(changes, CellChange{
				Cell:        cellRef,
				Type:        ChangeTypeDelete,
				OldValue:    oldCell.Value,
				OldFormula:  oldCell.Formula,
				Description: describeCellChange(&oldCell, nil),
			})
		case hasOld && hasNew:
			// Cell exists in both, check for changes
			if cellsAreDifferent(&oldCell, &newCell) {
				changes = append(changes, CellChange{
					Cell:        cellRef,
					Type:        ChangeTypeModify,
					OldValue:    oldCell.Value,
					NewValue:    newCell.Value,
					OldFormula:  oldCell.Formula,
					NewFormula:  newCell.Formula,
					Description: describeCellChange(&oldCell, &newCell),
				})
			}
		}
	}

	// Sort changes by cell reference for consistent output
	sort.Slice(changes, func(i, j int) bool {
		return changes[i].Cell < changes[j].Cell
	})

	return changes
}

// cellsAreDifferent checks if two cells are different
func cellsAreDifferent(old, updated *Cell) bool {
	// Compare values
	if !reflect.DeepEqual(old.Value, updated.Value) {
		return true
	}

	// Compare formulas
	if old.Formula != updated.Formula {
		return true
	}

	// Compare types
	if old.Type != updated.Type {
		return true
	}

	// Compare comments
	if (old.Comment == nil) != (updated.Comment == nil) {
		return true
	}
	if old.Comment != nil && updated.Comment != nil && old.Comment.Text != updated.Comment.Text {
		return true
	}

	// Compare hyperlinks
	if old.Hyperlink != updated.Hyperlink {
		return true
	}

	return false
}

// describeCellChange generates a human-readable description of the change
func describeCellChange(old, newCell *Cell) string {
	if old == nil && newCell != nil {
		if newCell.Formula != "" {
			return fmt.Sprintf("Added formula: %s", newCell.Formula)
		}
		return fmt.Sprintf("Added value: %v", newCell.Value)
	}

	if old != nil && newCell == nil {
		if old.Formula != "" {
			return fmt.Sprintf("Removed formula: %s", old.Formula)
		}
		return fmt.Sprintf("Removed value: %v", old.Value)
	}

	if old != nil && newCell != nil {
		var changes []string

		// Check value changes
		if !reflect.DeepEqual(old.Value, newCell.Value) {
			changes = append(changes, fmt.Sprintf("value: %v → %v", old.Value, newCell.Value))
		}

		// Check formula changes
		if old.Formula != newCell.Formula {
			switch {
			case old.Formula == "":
				changes = append(changes, fmt.Sprintf("added formula: %s", newCell.Formula))
			case newCell.Formula == "":
				changes = append(changes, fmt.Sprintf("removed formula: %s", old.Formula))
			default:
				changes = append(changes, fmt.Sprintf("formula: %s → %s", old.Formula, newCell.Formula))
			}
		}

		// Check comment changes
		if (old.Comment == nil) != (newCell.Comment == nil) {
			if old.Comment == nil {
				changes = append(changes, "added comment")
			} else {
				changes = append(changes, "removed comment")
			}
		} else if old.Comment != nil && newCell.Comment != nil && old.Comment.Text != newCell.Comment.Text {
			changes = append(changes, "modified comment")
		}

		// Check hyperlink changes
		if old.Hyperlink != newCell.Hyperlink {
			switch {
			case old.Hyperlink == "":
				changes = append(changes, "added hyperlink")
			case newCell.Hyperlink == "":
				changes = append(changes, "removed hyperlink")
			default:
				changes = append(changes, "modified hyperlink")
			}
		}

		if len(changes) > 0 {
			return "Changed " + strings.Join(changes, ", ")
		}
	}

	return "Modified"
}

// HasChanges returns true if the diff contains any changes
func (d *ExcelDiff) HasChanges() bool {
	return d.Summary.TotalChanges > 0 || d.Summary.CellChanges > 0
}

// formatSummaryParts creates summary parts for the diff
func (d *ExcelDiff) formatSummaryParts(colorized bool) []string {
	var parts []string

	if d.Summary.AddedSheets > 0 {
		if colorized {
			parts = append(parts, fmt.Sprintf("\033[32m+%d sheet(s) added\033[0m", d.Summary.AddedSheets)) // Green
		} else {
			parts = append(parts, fmt.Sprintf("%d sheet(s) added", d.Summary.AddedSheets))
		}
	}

	if d.Summary.ModifiedSheets > 0 {
		if colorized {
			parts = append(parts, fmt.Sprintf("\033[33m~%d sheet(s) modified\033[0m", d.Summary.ModifiedSheets)) // Yellow
		} else {
			parts = append(parts, fmt.Sprintf("%d sheet(s) modified", d.Summary.ModifiedSheets))
		}
	}

	if d.Summary.DeletedSheets > 0 {
		if colorized {
			parts = append(parts, fmt.Sprintf("\033[31m-%d sheet(s) deleted\033[0m", d.Summary.DeletedSheets)) // Red
		} else {
			parts = append(parts, fmt.Sprintf("%d sheet(s) deleted", d.Summary.DeletedSheets))
		}
	}

	if d.Summary.CellChanges > 0 {
		if colorized {
			parts = append(parts, fmt.Sprintf("\033[36m%d cell(s) changed\033[0m", d.Summary.CellChanges)) // Cyan
		} else {
			parts = append(parts, fmt.Sprintf("%d cell(s) changed", d.Summary.CellChanges))
		}
	}

	return parts
}

// String returns a string representation of the diff
func (d *ExcelDiff) String() string {
	if !d.HasChanges() {
		return noChangesMsg
	}

	parts := d.formatSummaryParts(false)
	return strings.Join(parts, ", ")
}

// ToColorizedString returns a colorized string representation of the diff for terminal display
func (d *ExcelDiff) ToColorizedString() string {
	if !d.HasChanges() {
		return noChangesMsgColor
	}

	parts := d.formatSummaryParts(true)
	return strings.Join(parts, ", ")
}

// ToDetailedString returns a detailed string representation showing all changes
func (d *ExcelDiff) ToDetailedString() string {
	if !d.HasChanges() {
		return noChangesMsg
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Excel Diff Summary (%s):\n", d.Timestamp.Format("2006-01-02 15:04:05")))
	result.WriteString(fmt.Sprintf("  Total Changes: %d\n", d.Summary.TotalChanges))
	result.WriteString(fmt.Sprintf("  Cell Changes: %d\n", d.Summary.CellChanges))
	result.WriteString("\n")

	for _, sheetDiff := range d.SheetDiffs {
		result.WriteString(fmt.Sprintf("Sheet: %s\n", sheetDiff.SheetName))

		if sheetDiff.Action != "" {
			result.WriteString(fmt.Sprintf("  Action: %s\n", sheetDiff.Action))
		}

		if len(sheetDiff.Changes) > 0 {
			result.WriteString(fmt.Sprintf("  Changes (%d):\n", len(sheetDiff.Changes)))
			for _, change := range sheetDiff.Changes {
				result.WriteString(fmt.Sprintf("    %s [%s]: %s\n", change.Cell, change.Type, change.Description))
			}
		}
		result.WriteString("\n")
	}

	return result.String()
}

// ToColorizedDetailedString returns a detailed colorized string representation
func (d *ExcelDiff) ToColorizedDetailedString() string {
	if !d.HasChanges() {
		return "\033[32mNo changes detected\033[0m"
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("\033[1mExcel Diff Summary (%s):\033[0m\n", d.Timestamp.Format("2006-01-02 15:04:05")))
	result.WriteString(fmt.Sprintf("  Total Changes: \033[36m%d\033[0m\n", d.Summary.TotalChanges))
	result.WriteString(fmt.Sprintf("  Cell Changes: \033[36m%d\033[0m\n", d.Summary.CellChanges))
	result.WriteString("\n")

	for _, sheetDiff := range d.SheetDiffs {
		result.WriteString(fmt.Sprintf("\033[1mSheet: %s\033[0m\n", sheetDiff.SheetName))

		if sheetDiff.Action != "" {
			var color string
			switch sheetDiff.Action {
			case ChangeTypeAdd:
				color = colorGreen
			case ChangeTypeModify:
				color = colorYellow
			case ChangeTypeDelete:
				color = colorRed
			default:
				color = colorReset
			}
			result.WriteString(fmt.Sprintf("  Action: %s%s\033[0m\n", color, sheetDiff.Action))
		}

		if len(sheetDiff.Changes) > 0 {
			result.WriteString(fmt.Sprintf("  Changes (\033[36m%d\033[0m):\n", len(sheetDiff.Changes)))
			for _, change := range sheetDiff.Changes {
				var typeColor string
				switch change.Type {
				case ChangeTypeAdd:
					typeColor = "\033[32m" // Green
				case ChangeTypeModify:
					typeColor = "\033[33m" // Yellow
				case ChangeTypeDelete:
					typeColor = "\033[31m" // Red
				default:
					typeColor = "\033[0m"
				}
				result.WriteString(fmt.Sprintf("    \033[1m%s\033[0m [%s%s\033[0m]: %s\n",
					change.Cell, typeColor, change.Type, change.Description))
			}
		}
		result.WriteString("\n")
	}

	return result.String()
}

// ToHTML returns an HTML representation of the diff for web display
func (d *ExcelDiff) ToHTML() string {
	if !d.HasChanges() {
		return "<div class='no-changes'>No changes detected</div>"
	}

	var result strings.Builder
	result.WriteString("<div class='excel-diff'>")
	result.WriteString(fmt.Sprintf("<h2>Excel Diff Summary (%s)</h2>", d.Timestamp.Format("2006-01-02 15:04:05")))
	result.WriteString(fmt.Sprintf("<p>Total Changes: <span class='count'>%d</span></p>", d.Summary.TotalChanges))
	result.WriteString(fmt.Sprintf("<p>Cell Changes: <span class='count'>%d</span></p>", d.Summary.CellChanges))

	for _, sheetDiff := range d.SheetDiffs {
		result.WriteString(fmt.Sprintf("<div class='sheet-diff'><h3>Sheet: %s</h3>", sheetDiff.SheetName))

		if sheetDiff.Action != "" {
			result.WriteString(fmt.Sprintf("<p>Action: <span class='action %s'>%s</span></p>", sheetDiff.Action, sheetDiff.Action))
		}

		if len(sheetDiff.Changes) > 0 {
			result.WriteString(fmt.Sprintf("<h4>Changes (%d)</h4><ul class='changes'>", len(sheetDiff.Changes)))
			for _, change := range sheetDiff.Changes {
				result.WriteString(fmt.Sprintf("<li class='change %s'><strong>%s</strong> [%s]: %s</li>",
					change.Type, change.Cell, change.Type, change.Description))
			}
			result.WriteString("</ul>")
		}
		result.WriteString("</div>")
	}

	result.WriteString("</div>")
	return result.String()
}

// GetDiffCSS returns CSS styles for HTML diff display
func GetDiffCSS() string {
	return `
.excel-diff {
	font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
	background: #f8f9fa;
	padding: 20px;
	border-radius: 8px;
	border: 1px solid #e9ecef;
}

.excel-diff h2, .excel-diff h3, .excel-diff h4 {
	color: #343a40;
	margin-top: 0;
}

.excel-diff .count {
	font-weight: bold;
	color: #007bff;
}

.excel-diff .sheet-diff {
	background: white;
	margin: 15px 0;
	padding: 15px;
	border-radius: 4px;
	border-left: 4px solid #007bff;
}

.excel-diff .action.add {
	color: #28a745;
}

.excel-diff .action.modify {
	color: #ffc107;
}

.excel-diff .action.delete {
	color: #dc3545;
}

.excel-diff .changes {
	list-style: none;
	padding: 0;
}

.excel-diff .change {
	padding: 8px 12px;
	margin: 4px 0;
	border-radius: 4px;
	border-left: 3px solid;
}

.excel-diff .change.add {
	background: #d4edda;
	border-left-color: #28a745;
}

.excel-diff .change.modify {
	background: #fff3cd;
	border-left-color: #ffc107;
}

.excel-diff .change.delete {
	background: #f8d7da;
	border-left-color: #dc3545;
}

.excel-diff .no-changes {
	color: #28a745;
	font-weight: bold;
	text-align: center;
	padding: 20px;
}
`
}
