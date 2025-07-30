package models

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"
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

		if !hasOld && hasNew {
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
		} else if hasOld && !hasNew {
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
		} else if hasOld && hasNew {
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

		if !hasOld && hasNew {
			// Cell added
			changes = append(changes, CellChange{
				Cell:        cellRef,
				Type:        ChangeTypeAdd,
				NewValue:    newCell.Value,
				NewFormula:  newCell.Formula,
				Description: describeCellChange(nil, &newCell),
			})
		} else if hasOld && !hasNew {
			// Cell deleted
			changes = append(changes, CellChange{
				Cell:        cellRef,
				Type:        ChangeTypeDelete,
				OldValue:    oldCell.Value,
				OldFormula:  oldCell.Formula,
				Description: describeCellChange(&oldCell, nil),
			})
		} else if hasOld && hasNew {
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
func cellsAreDifferent(old, new *Cell) bool {
	// Compare values
	if !reflect.DeepEqual(old.Value, new.Value) {
		return true
	}

	// Compare formulas
	if old.Formula != new.Formula {
		return true
	}

	// Compare types
	if old.Type != new.Type {
		return true
	}

	// Compare comments
	if (old.Comment == nil) != (new.Comment == nil) {
		return true
	}
	if old.Comment != nil && new.Comment != nil && old.Comment.Text != new.Comment.Text {
		return true
	}

	// Compare hyperlinks
	if old.Hyperlink != new.Hyperlink {
		return true
	}

	return false
}

// describeCellChange generates a human-readable description of the change
func describeCellChange(old, new *Cell) string {
	if old == nil && new != nil {
		if new.Formula != "" {
			return fmt.Sprintf("Added formula: %s", new.Formula)
		}
		return fmt.Sprintf("Added value: %v", new.Value)
	}

	if old != nil && new == nil {
		if old.Formula != "" {
			return fmt.Sprintf("Removed formula: %s", old.Formula)
		}
		return fmt.Sprintf("Removed value: %v", old.Value)
	}

	if old != nil && new != nil {
		var changes []string

		// Check value changes
		if !reflect.DeepEqual(old.Value, new.Value) {
			changes = append(changes, fmt.Sprintf("value: %v → %v", old.Value, new.Value))
		}

		// Check formula changes
		if old.Formula != new.Formula {
			if old.Formula == "" {
				changes = append(changes, fmt.Sprintf("added formula: %s", new.Formula))
			} else if new.Formula == "" {
				changes = append(changes, fmt.Sprintf("removed formula: %s", old.Formula))
			} else {
				changes = append(changes, fmt.Sprintf("formula: %s → %s", old.Formula, new.Formula))
			}
		}

		// Check comment changes
		if (old.Comment == nil) != (new.Comment == nil) {
			if old.Comment == nil {
				changes = append(changes, "added comment")
			} else {
				changes = append(changes, "removed comment")
			}
		} else if old.Comment != nil && new.Comment != nil && old.Comment.Text != new.Comment.Text {
			changes = append(changes, "modified comment")
		}

		// Check hyperlink changes
		if old.Hyperlink != new.Hyperlink {
			if old.Hyperlink == "" {
				changes = append(changes, "added hyperlink")
			} else if new.Hyperlink == "" {
				changes = append(changes, "removed hyperlink")
			} else {
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

// String returns a string representation of the diff
func (d *ExcelDiff) String() string {
	if !d.HasChanges() {
		return "No changes detected"
	}

	var parts []string
	
	if d.Summary.AddedSheets > 0 {
		parts = append(parts, fmt.Sprintf("%d sheet(s) added", d.Summary.AddedSheets))
	}
	
	if d.Summary.ModifiedSheets > 0 {
		parts = append(parts, fmt.Sprintf("%d sheet(s) modified", d.Summary.ModifiedSheets))
	}
	
	if d.Summary.DeletedSheets > 0 {
		parts = append(parts, fmt.Sprintf("%d sheet(s) deleted", d.Summary.DeletedSheets))
	}

	if d.Summary.CellChanges > 0 {
		parts = append(parts, fmt.Sprintf("%d cell(s) changed", d.Summary.CellChanges))
	}

	return strings.Join(parts, ", ")
}