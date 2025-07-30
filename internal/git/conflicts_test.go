package git

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Classic-Homes/sheetsync/pkg/models"
)

func TestDetectConflicts_NoConflicts(t *testing.T) {
	tempFile := createTempFile(t, "no conflicts here\njust regular content\n")
	defer func() {
		if err := os.Remove(tempFile); err != nil {
			t.Logf("Warning: failed to remove temp file %s: %v", tempFile, err)
		}
	}()

	client := createTestClient(t)
	info, err := client.DetectConflicts(tempFile)

	require.NoError(t, err)
	assert.False(t, info.HasConflicts)
	assert.Empty(t, info.Conflicts)
}

func TestDetectConflicts_SimpleConflict(t *testing.T) {
	content := `line 1
<<<<<<< HEAD
our changes
=======
their changes
>>>>>>> branch
line 2`

	tempFile := createTempFile(t, content)
	defer func() {
		if err := os.Remove(tempFile); err != nil {
			t.Logf("Warning: failed to remove temp file %s: %v", tempFile, err)
		}
	}()

	client := createTestClient(t)
	info, err := client.DetectConflicts(tempFile)

	require.NoError(t, err)
	assert.True(t, info.HasConflicts)
	require.Len(t, info.Conflicts, 1)

	conflict := info.Conflicts[0]
	assert.Equal(t, 2, conflict.StartLine)
	assert.Equal(t, 6, conflict.EndLine)
	assert.Equal(t, []string{"our changes"}, conflict.OurCode)
	assert.Equal(t, []string{"their changes"}, conflict.TheirCode)
}

func TestDetectConflicts_MultipleConflicts(t *testing.T) {
	content := `line 1
<<<<<<< HEAD
first our changes
=======
first their changes
>>>>>>> branch
middle line
<<<<<<< HEAD
second our changes
=======
second their changes
>>>>>>> branch
last line`

	tempFile := createTempFile(t, content)
	defer func() {
		if err := os.Remove(tempFile); err != nil {
			t.Logf("Warning: failed to remove temp file %s: %v", tempFile, err)
		}
	}()

	client := createTestClient(t)
	info, err := client.DetectConflicts(tempFile)

	require.NoError(t, err)
	assert.True(t, info.HasConflicts)
	require.Len(t, info.Conflicts, 2)

	// First conflict
	conflict1 := info.Conflicts[0]
	assert.Equal(t, 2, conflict1.StartLine)
	assert.Equal(t, 6, conflict1.EndLine)
	assert.Equal(t, []string{"first our changes"}, conflict1.OurCode)
	assert.Equal(t, []string{"first their changes"}, conflict1.TheirCode)

	// Second conflict
	conflict2 := info.Conflicts[1]
	assert.Equal(t, 8, conflict2.StartLine)
	assert.Equal(t, 12, conflict2.EndLine)
	assert.Equal(t, []string{"second our changes"}, conflict2.OurCode)
	assert.Equal(t, []string{"second their changes"}, conflict2.TheirCode)
}

func TestResolveConflict_KeepOurs(t *testing.T) {
	content := `line 1
<<<<<<< HEAD
our changes
=======
their changes
>>>>>>> branch
line 2`

	tempFile := createTempFile(t, content)
	defer func() {
		if err := os.Remove(tempFile); err != nil {
			t.Logf("Warning: failed to remove temp file %s: %v", tempFile, err)
		}
	}()

	client := createTestClient(t)
	err := client.ResolveConflict(tempFile, ResolveOurs)

	require.NoError(t, err)

	resolvedContent, err := os.ReadFile(tempFile)
	require.NoError(t, err)

	expected := "line 1\nour changes\nline 2"
	assert.Equal(t, expected, string(resolvedContent))
}

func TestResolveConflict_KeepTheirs(t *testing.T) {
	content := `line 1
<<<<<<< HEAD
our changes
=======
their changes
>>>>>>> branch
line 2`

	tempFile := createTempFile(t, content)
	defer func() {
		if err := os.Remove(tempFile); err != nil {
			t.Logf("Warning: failed to remove temp file %s: %v", tempFile, err)
		}
	}()

	client := createTestClient(t)
	err := client.ResolveConflict(tempFile, ResolveTheirs)

	require.NoError(t, err)

	resolvedContent, err := os.ReadFile(tempFile)
	require.NoError(t, err)

	expected := "line 1\ntheir changes\nline 2"
	assert.Equal(t, expected, string(resolvedContent))
}

func TestResolveConflict_KeepBoth(t *testing.T) {
	content := `line 1
<<<<<<< HEAD
our changes
=======
their changes
>>>>>>> branch
line 2`

	tempFile := createTempFile(t, content)
	defer func() {
		if err := os.Remove(tempFile); err != nil {
			t.Logf("Warning: failed to remove temp file %s: %v", tempFile, err)
		}
	}()

	client := createTestClient(t)
	err := client.ResolveConflict(tempFile, ResolveBoth)

	require.NoError(t, err)

	resolvedContent, err := os.ReadFile(tempFile)
	require.NoError(t, err)

	expected := "line 1\nour changes\ntheir changes\nline 2"
	assert.Equal(t, expected, string(resolvedContent))
}

func TestSmartMergeExcelJSON_ValidDocuments(t *testing.T) {
	client := createTestClient(t)

	// Create two test documents
	doc1 := &models.ExcelDocument{
		Version: "1.0",
		Metadata: models.DocumentMetadata{
			Created:  time.Now().Add(-time.Hour),
			Modified: time.Now().Add(-time.Hour),
		},
		Sheets: []models.Sheet{
			{
				Name:  "Sheet1",
				Index: 0,
				Cells: map[string]models.Cell{
					"A1": {Value: "old_value", Type: models.CellTypeString},
					"B1": {Value: "shared_value", Type: models.CellTypeString},
				},
			},
		},
		DefinedNames: map[string]string{"name1": "Sheet1!A1"},
	}

	doc2 := &models.ExcelDocument{
		Version: "1.0",
		Metadata: models.DocumentMetadata{
			Created:  time.Now().Add(-time.Hour),
			Modified: time.Now(), // Newer
		},
		Sheets: []models.Sheet{
			{
				Name:  "Sheet1",
				Index: 0,
				Cells: map[string]models.Cell{
					"A1": {Value: "new_value", Type: models.CellTypeString},
					"B1": {Value: "shared_value", Type: models.CellTypeString},
					"C1": {Value: "additional_value", Type: models.CellTypeString},
				},
			},
		},
		DefinedNames: map[string]string{
			"name1": "Sheet1!A1",
			"name2": "Sheet1!C1",
		},
	}

	// Convert to JSON lines
	ourJSON, _ := json.MarshalIndent(doc1, "", "  ")
	theirJSON, _ := json.MarshalIndent(doc2, "", "  ")

	ourCode := strings.Split(string(ourJSON), "\n")
	theirCode := strings.Split(string(theirJSON), "\n")

	// Test smart merge
	result := client.smartMergeExcelJSON(ourCode, theirCode)

	// Parse the result back to verify it's valid JSON
	resultJSON := strings.Join(result, "\n")
	var mergedDoc models.ExcelDocument
	err := json.Unmarshal([]byte(resultJSON), &mergedDoc)
	require.NoError(t, err)

	// Verify merge results
	assert.True(t, mergedDoc.Metadata.Modified.After(doc1.Metadata.Modified) || mergedDoc.Metadata.Modified.Equal(doc2.Metadata.Modified)) // Should use newer timestamp
	require.Len(t, mergedDoc.Sheets, 1)

	// Check merged cells
	cells := mergedDoc.Sheets[0].Cells
	// Note: Our merge logic keeps "ours" when both have values, "theirs" only when ours is empty
	assert.Equal(t, "old_value", cells["A1"].Value)        // Ours kept (both non-empty)
	assert.Equal(t, "shared_value", cells["B1"].Value)     // Shared value
	assert.Equal(t, "additional_value", cells["C1"].Value) // Additional from theirs

	// Check merged defined names
	assert.Equal(t, "Sheet1!A1", mergedDoc.DefinedNames["name1"])
	assert.Equal(t, "Sheet1!C1", mergedDoc.DefinedNames["name2"])
}

func TestSmartMergeExcelJSON_InvalidJSON(t *testing.T) {
	client := createTestClient(t)

	ourCode := []string{"invalid json"}
	theirCode := []string{`{"valid": "json"}`}

	// Should fall back to timestamp resolution
	result := client.smartMergeExcelJSON(ourCode, theirCode)

	// Since theirs is valid JSON and ours is not, should choose theirs
	assert.Equal(t, theirCode, result)
}

func TestResolveByTimestamp_ValidDocuments(t *testing.T) {
	client := createTestClient(t)

	// Create documents with different timestamps
	olderDoc := &models.ExcelDocument{
		Metadata: models.DocumentMetadata{
			Modified: time.Now().Add(-time.Hour),
		},
	}

	newerDoc := &models.ExcelDocument{
		Metadata: models.DocumentMetadata{
			Modified: time.Now(),
		},
	}

	ourJSON, _ := json.MarshalIndent(olderDoc, "", "  ")
	theirJSON, _ := json.MarshalIndent(newerDoc, "", "  ")

	ourCode := strings.Split(string(ourJSON), "\n")
	theirCode := strings.Split(string(theirJSON), "\n")

	result := client.resolveByTimestamp(ourCode, theirCode)

	// Should choose the newer one (theirs)
	assert.Equal(t, theirCode, result)
}

func TestResolveByTimestamp_InvalidJSON(t *testing.T) {
	client := createTestClient(t)

	ourCode := []string{"invalid json"}
	theirCode := []string{"also invalid"}

	result := client.resolveByTimestamp(ourCode, theirCode)

	// Should default to ours when both are invalid
	assert.Equal(t, ourCode, result)
}

func TestMergeSheetsIntelligently(t *testing.T) {
	client := createTestClient(t)

	ourSheet := &models.Sheet{
		Name:  "TestSheet",
		Index: 0,
		Cells: map[string]models.Cell{
			"A1": {Value: "our_value", Type: models.CellTypeString},
			"B1": {Value: "shared", Type: models.CellTypeString},
			"C1": {Value: "", Type: models.CellTypeString}, // Empty value
		},
		RowHeights:   map[int]float64{1: 20.0},
		ColumnWidths: map[string]float64{"A": 100.0},
		MergedCells: []models.MergedCell{
			{Range: "A1:A2"},
		},
	}

	theirSheet := &models.Sheet{
		Name:  "TestSheet",
		Index: 0,
		Cells: map[string]models.Cell{
			"A1": {Value: "their_value", Type: models.CellTypeString},
			"B1": {Value: "shared", Type: models.CellTypeString},
			"C1": {Value: "filled_value", Type: models.CellTypeString}, // Filled value
			"D1": {Value: "new_cell", Type: models.CellTypeString},     // New cell
		},
		RowHeights:   map[int]float64{1: 25.0, 2: 15.0},         // Larger height + new row
		ColumnWidths: map[string]float64{"A": 80.0, "B": 120.0}, // Smaller A, new B
		MergedCells: []models.MergedCell{
			{Range: "A1:A2"},
			{Range: "B1:B2"}, // New merged cell
		},
	}

	merged := client.mergeSheetsIntelligently(ourSheet, theirSheet)

	// Check cells - should prefer non-empty values and formulas
	assert.Equal(t, "our_value", merged.Cells["A1"].Value)    // Keep ours (both non-empty)
	assert.Equal(t, "shared", merged.Cells["B1"].Value)       // Same value
	assert.Equal(t, "filled_value", merged.Cells["C1"].Value) // Theirs has value, ours empty
	assert.Equal(t, "new_cell", merged.Cells["D1"].Value)     // New cell from theirs

	// Check row heights - should use larger values
	assert.Equal(t, 25.0, merged.RowHeights[1]) // Larger from theirs
	assert.Equal(t, 15.0, merged.RowHeights[2]) // New from theirs

	// Check column widths - should use larger values
	assert.Equal(t, 100.0, merged.ColumnWidths["A"]) // Larger from ours
	assert.Equal(t, 120.0, merged.ColumnWidths["B"]) // New from theirs

	// Check merged cells - should combine unique ranges
	assert.Len(t, merged.MergedCells, 2)
	ranges := make(map[string]bool)
	for _, mc := range merged.MergedCells {
		ranges[mc.Range] = true
	}
	assert.True(t, ranges["A1:A2"])
	assert.True(t, ranges["B1:B2"])
}

func TestApplyResolutionStrategy(t *testing.T) {
	client := createTestClient(t)

	ourCode := []string{"our line 1", "our line 2"}
	theirCode := []string{"their line 1", "their line 2"}

	tests := []struct {
		name     string
		strategy ConflictResolutionStrategy
		expected []string
	}{
		{
			name:     "resolve ours",
			strategy: ResolveOurs,
			expected: ourCode,
		},
		{
			name:     "resolve theirs",
			strategy: ResolveTheirs,
			expected: theirCode,
		},
		{
			name:     "resolve both",
			strategy: ResolveBoth,
			expected: []string{"our line 1", "our line 2", "their line 1", "their line 2"},
		},
		{
			name:     "manual resolution",
			strategy: ResolveManual,
			expected: []string{
				"<<<<<<< HEAD",
				"our line 1",
				"our line 2",
				"=======",
				"their line 1",
				"their line 2",
				">>>>>>> incoming",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.applyResolutionStrategy(ourCode, theirCode, tt.strategy)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Helper functions

func createTestClient(t *testing.T) *Client {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	logger.SetOutput(os.Stdout) // For test visibility

	return &Client{
		logger: logger,
		config: &Config{
			UserName:  "Test User",
			UserEmail: "test@example.com",
		},
	}
}

func createTempFile(t *testing.T, content string) string {
	tempFile, err := os.CreateTemp("", "test_conflict_*.txt")
	require.NoError(t, err)

	_, err = tempFile.WriteString(content)
	require.NoError(t, err)

	err = tempFile.Close()
	require.NoError(t, err)

	return tempFile.Name()
}

// Mock worktree for testing would go here if needed
// Currently not implementing the full git interface for these tests

// Benchmark tests

func BenchmarkComputeDiff_LargeDocument(b *testing.B) {
	// Create large documents for benchmarking
	doc1 := createLargeTestDocument(1000, 100) // 1000 sheets, 100 cells each
	doc2 := createLargeTestDocument(1000, 100)

	// Modify some cells in doc2
	doc2.Sheets[0].Cells["A1"] = models.Cell{
		Value: "modified",
		Type:  models.CellTypeString,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		models.ComputeDiff(doc1, doc2)
	}
}

func createLargeTestDocument(numSheets, cellsPerSheet int) *models.ExcelDocument {
	doc := &models.ExcelDocument{
		Version: "1.0",
		Metadata: models.DocumentMetadata{
			Created:  time.Now(),
			Modified: time.Now(),
		},
		Sheets:       make([]models.Sheet, numSheets),
		DefinedNames: make(map[string]string),
	}

	for i := 0; i < numSheets; i++ {
		sheet := models.Sheet{
			Name:  fmt.Sprintf("Sheet%d", i+1),
			Index: i,
			Cells: make(map[string]models.Cell),
		}

		for j := 0; j < cellsPerSheet; j++ {
			cellRef := fmt.Sprintf("A%d", j+1)
			sheet.Cells[cellRef] = models.Cell{
				Value: fmt.Sprintf("value_%d_%d", i, j),
				Type:  models.CellTypeString,
			}
		}

		doc.Sheets[i] = sheet
	}

	return doc
}
