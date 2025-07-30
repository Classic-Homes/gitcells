package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComputeDiff_NoChanges(t *testing.T) {
	doc1 := createTestDocument()
	doc2 := createTestDocument()

	diff := ComputeDiff(doc1, doc2)

	assert.False(t, diff.HasChanges())
	assert.Equal(t, 0, diff.Summary.TotalChanges)
	assert.Equal(t, 0, diff.Summary.CellChanges)
	assert.Equal(t, "No changes detected", diff.String())
}

func TestComputeDiff_CellValueChange(t *testing.T) {
	doc1 := createTestDocument()
	doc2 := createTestDocument()

	// Modify a cell value in doc2
	doc2.Sheets[0].Cells["A1"] = Cell{
		Value: "Modified Value",
		Type:  CellTypeString,
	}

	diff := ComputeDiff(doc1, doc2)

	assert.True(t, diff.HasChanges())
	assert.Equal(t, 1, diff.Summary.ModifiedSheets)
	assert.Equal(t, 1, diff.Summary.CellChanges)
	require.Len(t, diff.SheetDiffs, 1)
	require.Len(t, diff.SheetDiffs[0].Changes, 1)

	change := diff.SheetDiffs[0].Changes[0]
	assert.Equal(t, "A1", change.Cell)
	assert.Equal(t, ChangeTypeModify, change.Type)
	assert.Equal(t, "Test Value", change.OldValue)
	assert.Equal(t, "Modified Value", change.NewValue)
}

func TestComputeDiff_AddedCell(t *testing.T) {
	doc1 := createTestDocument()
	doc2 := createTestDocument()

	// Add a new cell in doc2
	doc2.Sheets[0].Cells["B2"] = Cell{
		Value: "New Cell",
		Type:  CellTypeString,
	}

	diff := ComputeDiff(doc1, doc2)

	assert.True(t, diff.HasChanges())
	assert.Equal(t, 1, diff.Summary.ModifiedSheets)
	assert.Equal(t, 1, diff.Summary.CellChanges)
	require.Len(t, diff.SheetDiffs, 1)
	require.Len(t, diff.SheetDiffs[0].Changes, 1)

	change := diff.SheetDiffs[0].Changes[0]
	assert.Equal(t, "B2", change.Cell)
	assert.Equal(t, ChangeTypeAdd, change.Type)
	assert.Equal(t, nil, change.OldValue)
	assert.Equal(t, "New Cell", change.NewValue)
}

func TestComputeDiff_DeletedCell(t *testing.T) {
	doc1 := createTestDocument()
	doc2 := createTestDocument()

	// Remove a cell from doc2
	delete(doc2.Sheets[0].Cells, "A1")

	diff := ComputeDiff(doc1, doc2)

	assert.True(t, diff.HasChanges())
	assert.Equal(t, 1, diff.Summary.ModifiedSheets)
	assert.Equal(t, 1, diff.Summary.CellChanges)
	require.Len(t, diff.SheetDiffs, 1)
	require.Len(t, diff.SheetDiffs[0].Changes, 1)

	change := diff.SheetDiffs[0].Changes[0]
	assert.Equal(t, "A1", change.Cell)
	assert.Equal(t, ChangeTypeDelete, change.Type)
	assert.Equal(t, "Test Value", change.OldValue)
	assert.Equal(t, nil, change.NewValue)
}

func TestComputeDiff_AddedSheet(t *testing.T) {
	doc1 := createTestDocument()
	doc2 := createTestDocument()

	// Add a new sheet to doc2
	newSheet := Sheet{
		Name:  "New Sheet",
		Index: 1,
		Cells: map[string]Cell{
			"A1": {
				Value: "New Sheet Data",
				Type:  CellTypeString,
			},
		},
	}
	doc2.Sheets = append(doc2.Sheets, newSheet)

	diff := ComputeDiff(doc1, doc2)

	assert.True(t, diff.HasChanges())
	assert.Equal(t, 1, diff.Summary.AddedSheets)
	assert.Equal(t, 1, diff.Summary.CellChanges)
	require.Len(t, diff.SheetDiffs, 1)

	sheetDiff := diff.SheetDiffs[0]
	assert.Equal(t, "New Sheet", sheetDiff.SheetName)
	assert.Equal(t, ChangeTypeAdd, sheetDiff.Action)
	require.Len(t, sheetDiff.Changes, 1)

	change := sheetDiff.Changes[0]
	assert.Equal(t, "A1", change.Cell)
	assert.Equal(t, ChangeTypeAdd, change.Type)
	assert.Equal(t, "New Sheet Data", change.NewValue)
}

func TestComputeDiff_DeletedSheet(t *testing.T) {
	doc1 := createTestDocument()
	doc2 := &ExcelDocument{
		Version:  "1.0",
		Metadata: DocumentMetadata{},
		Sheets:   []Sheet{}, // Empty sheets
	}

	diff := ComputeDiff(doc1, doc2)

	assert.True(t, diff.HasChanges())
	assert.Equal(t, 1, diff.Summary.DeletedSheets)
	assert.Equal(t, 1, diff.Summary.CellChanges) // The deleted cell
	require.Len(t, diff.SheetDiffs, 1)

	sheetDiff := diff.SheetDiffs[0]
	assert.Equal(t, "Test Sheet", sheetDiff.SheetName)
	assert.Equal(t, ChangeTypeDelete, sheetDiff.Action)
	require.Len(t, sheetDiff.Changes, 1)

	change := sheetDiff.Changes[0]
	assert.Equal(t, "A1", change.Cell)
	assert.Equal(t, ChangeTypeDelete, change.Type)
	assert.Equal(t, "Test Value", change.OldValue)
}

func TestComputeDiff_FormulaChange(t *testing.T) {
	doc1 := createTestDocument()
	doc2 := createTestDocument()

	// Add formula to a cell in doc2
	doc2.Sheets[0].Cells["A1"] = Cell{
		Value:   "Test Value",
		Formula: "=SUM(B1:B10)",
		Type:    CellTypeFormula,
	}

	diff := ComputeDiff(doc1, doc2)

	assert.True(t, diff.HasChanges())
	require.Len(t, diff.SheetDiffs, 1)
	require.Len(t, diff.SheetDiffs[0].Changes, 1)

	change := diff.SheetDiffs[0].Changes[0]
	assert.Equal(t, "A1", change.Cell)
	assert.Equal(t, ChangeTypeModify, change.Type)
	assert.Equal(t, "", change.OldFormula)
	assert.Equal(t, "=SUM(B1:B10)", change.NewFormula)
	assert.Contains(t, change.Description, "added formula")
}

func TestComputeDiff_CommentChange(t *testing.T) {
	doc1 := createTestDocument()
	doc2 := createTestDocument()

	// Add comment to a cell in doc2
	doc2.Sheets[0].Cells["A1"] = Cell{
		Value: "Test Value",
		Type:  CellTypeString,
		Comment: &Comment{
			Text: "This is a comment",
		},
	}

	diff := ComputeDiff(doc1, doc2)

	assert.True(t, diff.HasChanges())
	require.Len(t, diff.SheetDiffs, 1)
	require.Len(t, diff.SheetDiffs[0].Changes, 1)

	change := diff.SheetDiffs[0].Changes[0]
	assert.Equal(t, "A1", change.Cell)
	assert.Equal(t, ChangeTypeModify, change.Type)
	assert.Contains(t, change.Description, "added comment")
}

func TestCellsAreDifferent(t *testing.T) {
	tests := []struct {
		name     string
		old      Cell
		new      Cell
		expected bool
	}{
		{
			name: "identical cells",
			old: Cell{
				Value: "test",
				Type:  CellTypeString,
			},
			new: Cell{
				Value: "test",
				Type:  CellTypeString,
			},
			expected: false,
		},
		{
			name: "different values",
			old: Cell{
				Value: "old",
				Type:  CellTypeString,
			},
			new: Cell{
				Value: "new",
				Type:  CellTypeString,
			},
			expected: true,
		},
		{
			name: "different formulas",
			old: Cell{
				Value:   10,
				Formula: "=SUM(A1:A10)",
				Type:    CellTypeFormula,
			},
			new: Cell{
				Value:   10,
				Formula: "=SUM(B1:B10)",
				Type:    CellTypeFormula,
			},
			expected: true,
		},
		{
			name: "different types",
			old: Cell{
				Value: "10",
				Type:  CellTypeString,
			},
			new: Cell{
				Value: 10,
				Type:  CellTypeNumber,
			},
			expected: true,
		},
		{
			name: "different hyperlinks",
			old: Cell{
				Value:     "Link",
				Type:      CellTypeString,
				Hyperlink: "http://old.com",
			},
			new: Cell{
				Value:     "Link",
				Type:      CellTypeString,
				Hyperlink: "http://new.com",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cellsAreDifferent(&tt.old, &tt.new)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDescribeCellChange(t *testing.T) {
	tests := []struct {
		name     string
		old      *Cell
		new      *Cell
		expected string
	}{
		{
			name: "added cell with value",
			old:  nil,
			new: &Cell{
				Value: "new value",
				Type:  CellTypeString,
			},
			expected: "Added value: new value",
		},
		{
			name: "added cell with formula",
			old:  nil,
			new: &Cell{
				Value:   10,
				Formula: "=SUM(A1:A10)",
				Type:    CellTypeFormula,
			},
			expected: "Added formula: =SUM(A1:A10)",
		},
		{
			name: "removed cell",
			old: &Cell{
				Value: "old value",
				Type:  CellTypeString,
			},
			new:      nil,
			expected: "Removed value: old value",
		},
		{
			name: "value change",
			old: &Cell{
				Value: "old",
				Type:  CellTypeString,
			},
			new: &Cell{
				Value: "new",
				Type:  CellTypeString,
			},
			expected: "Changed value: old → new",
		},
		{
			name: "formula change",
			old: &Cell{
				Value:   10,
				Formula: "=SUM(A1:A10)",
				Type:    CellTypeFormula,
			},
			new: &Cell{
				Value:   15,
				Formula: "=SUM(B1:B10)",
				Type:    CellTypeFormula,
			},
			expected: "Changed value: 10 → 15, formula: =SUM(A1:A10) → =SUM(B1:B10)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := describeCellChange(tt.old, tt.new)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDiffString(t *testing.T) {
	tests := []struct {
		name     string
		diff     *ExcelDiff
		expected string
	}{
		{
			name: "no changes",
			diff: &ExcelDiff{
				Summary: DiffSummary{},
			},
			expected: "No changes detected",
		},
		{
			name: "sheet changes only",
			diff: &ExcelDiff{
				Summary: DiffSummary{
					TotalChanges:   2,
					AddedSheets:    1,
					ModifiedSheets: 1,
				},
			},
			expected: "1 sheet(s) added, 1 sheet(s) modified",
		},
		{
			name: "cell changes only",
			diff: &ExcelDiff{
				Summary: DiffSummary{
					CellChanges: 5,
				},
			},
			expected: "5 cell(s) changed",
		},
		{
			name: "mixed changes",
			diff: &ExcelDiff{
				Summary: DiffSummary{
					TotalChanges: 1,
					AddedSheets:  1,
					CellChanges:  3,
				},
			},
			expected: "1 sheet(s) added, 3 cell(s) changed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.diff.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Helper function to create a test document
func createTestDocument() *ExcelDocument {
	return &ExcelDocument{
		Version: "1.0",
		Metadata: DocumentMetadata{
			Created:      time.Now(),
			Modified:     time.Now(),
			AppVersion:   "test",
			OriginalFile: "test.xlsx",
		},
		Sheets: []Sheet{
			{
				Name:  "Test Sheet",
				Index: 0,
				Cells: map[string]Cell{
					"A1": {
						Value: "Test Value",
						Type:  CellTypeString,
					},
				},
			},
		},
		DefinedNames: make(map[string]string),
	}
}
