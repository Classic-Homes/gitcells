package test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"

	"github.com/Classic-Homes/sheetsync/internal/config"
	"github.com/Classic-Homes/sheetsync/internal/converter"
	"github.com/Classic-Homes/sheetsync/pkg/models"
)

func TestFullConversionWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)
	conv := converter.NewConverter(logger)
	
	tempDir := t.TempDir()
	
	// Test Excel -> JSON -> Excel workflow
	originalFile := "testdata/sample_files/complex.xlsx"
	jsonFile := filepath.Join(tempDir, "complex.json")
	recreatedFile := filepath.Join(tempDir, "recreated.xlsx")

	// Step 1: Convert Excel to JSON
	doc, err := conv.ExcelToJSON(originalFile, converter.ConvertOptions{
		PreserveFormulas: true,
		PreserveStyles:   true,
		PreserveComments: true,
		IgnoreEmptyCells: true,
	})
	require.NoError(t, err)
	require.NotNil(t, doc)

	// Step 2: Save JSON to file
	jsonData, err := json.MarshalIndent(doc, "", "  ")
	require.NoError(t, err)
	
	err = os.WriteFile(jsonFile, jsonData, 0644)
	require.NoError(t, err)

	// Step 3: Load JSON from file
	var loadedDoc models.ExcelDocument
	jsonData, err = os.ReadFile(jsonFile)
	require.NoError(t, err)
	
	err = json.Unmarshal(jsonData, &loadedDoc)
	require.NoError(t, err)

	// Step 4: Convert JSON back to Excel
	err = conv.JSONToExcel(&loadedDoc, recreatedFile, converter.ConvertOptions{
		PreserveFormulas: true,
		PreserveStyles:   true,
		PreserveComments: true,
	})
	require.NoError(t, err)

	// Step 5: Verify the recreated file
	f, err := excelize.OpenFile(recreatedFile)
	require.NoError(t, err)
	defer f.Close()

	// Check basic structure
	sheetList := f.GetSheetList()
	assert.Contains(t, sheetList, "Sheet1")
	assert.Contains(t, sheetList, "Summary")

	// Check some cell values
	value, err := f.GetCellValue("Sheet1", "A1")
	require.NoError(t, err)
	assert.Equal(t, "Item", value)

	// Check formulas are preserved
	formula, err := f.GetCellFormula("Sheet1", "D2")
	require.NoError(t, err)
	assert.Contains(t, formula, "B2*C2")
}

func TestConfigurationLoading(t *testing.T) {
	tempDir := t.TempDir()

	// Create a test config file
	configContent := `
version: 1.0
git:
  branch: test-branch
  auto_push: true
  user_name: "Test User"
  user_email: "test@example.com"
watcher:
  debounce_delay: 5s
  file_extensions: [".xlsx", ".xls"]
converter:
  preserve_formulas: true
  max_cells_per_sheet: 50000
`

	configFile := filepath.Join(tempDir, "test-config.yaml")
	err := os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)

	// Load the configuration
	cfg, err := config.Load(configFile)
	require.NoError(t, err)

	// Verify configuration values
	assert.Equal(t, "1.0", cfg.Version)
	assert.Equal(t, "test-branch", cfg.Git.Branch)
	assert.True(t, cfg.Git.AutoPush)
	assert.Equal(t, "Test User", cfg.Git.UserName)
	assert.Equal(t, "test@example.com", cfg.Git.UserEmail)
	assert.Equal(t, "5s", cfg.Watcher.DebounceDelay.String())
	assert.Contains(t, cfg.Watcher.FileExtensions, ".xlsx")
	assert.Contains(t, cfg.Watcher.FileExtensions, ".xls")
	assert.True(t, cfg.Converter.PreserveFormulas)
	assert.Equal(t, 50000, cfg.Converter.MaxCellsPerSheet)
}

func TestDiffGeneration(t *testing.T) {
	// Create two similar documents with differences
	doc1 := &models.ExcelDocument{
		Version: "1.0",
		Sheets: []models.Sheet{
			{
				Name:  "Sheet1",
				Index: 0,
				Cells: map[string]models.Cell{
					"A1": {Value: "Hello", Type: models.CellTypeString},
					"A2": {Value: float64(10), Type: models.CellTypeNumber},
					"A3": {Value: "Old Value", Type: models.CellTypeString},
				},
			},
		},
	}

	doc2 := &models.ExcelDocument{
		Version: "1.0",
		Sheets: []models.Sheet{
			{
				Name:  "Sheet1",
				Index: 0,
				Cells: map[string]models.Cell{
					"A1": {Value: "Hello", Type: models.CellTypeString},
					"A2": {Value: float64(20), Type: models.CellTypeNumber}, // Changed value
					"A3": {Value: "New Value", Type: models.CellTypeString}, // Changed value
					"A4": {Value: "Added Cell", Type: models.CellTypeString}, // New cell
				},
			},
		},
	}

	// Generate diff
	diff := models.ComputeDiff(doc1, doc2)

	// Verify diff results
	assert.NotNil(t, diff)
	assert.True(t, diff.Summary.TotalChanges > 0)
	assert.Len(t, diff.SheetDiffs, 1)

	sheetDiff := diff.SheetDiffs[0]
	assert.Equal(t, "Sheet1", sheetDiff.SheetName)
	assert.True(t, len(sheetDiff.Changes) >= 3) // At least 3 changes

	// Check for specific changes
	changedCells := make(map[string]models.CellChange)
	for _, change := range sheetDiff.Changes {
		changedCells[change.Cell] = change
	}

	// A2 should be modified
	if change, exists := changedCells["A2"]; exists {
		assert.Equal(t, models.ChangeTypeModify, change.Type)
		assert.Equal(t, float64(10), change.OldValue)
		assert.Equal(t, float64(20), change.NewValue)
	}

	// A4 should be added
	if change, exists := changedCells["A4"]; exists {
		assert.Equal(t, models.ChangeTypeAdd, change.Type)
		assert.Equal(t, "Added Cell", change.NewValue)
	}
}

func TestLargeFileHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)
	conv := converter.NewConverter(logger)

	tempDir := t.TempDir()
	largeFile := filepath.Join(tempDir, "large.xlsx")

	// Create a file with many cells
	f := excelize.NewFile()
	defer f.Close()

	// Add data to create a reasonably large file
	for row := 1; row <= 100; row++ {
		for col := 1; col <= 26; col++ { // A-Z columns
			cellRef, _ := excelize.CoordinatesToCellName(col, row)
			value := fmt.Sprintf("Cell_%d_%d", row, col)
			f.SetCellValue("Sheet1", cellRef, value)
		}
	}

	err := f.SaveAs(largeFile)
	require.NoError(t, err)

	// Test conversion with cell limit
	doc, err := conv.ExcelToJSON(largeFile, converter.ConvertOptions{
		MaxCellsPerSheet: 1000, // Limit to 1000 cells
		IgnoreEmptyCells: true,
	})
	require.NoError(t, err)
	require.NotNil(t, doc)

	// Should have at most 1000 cells
	totalCells := 0
	for _, sheet := range doc.Sheets {
		totalCells += len(sheet.Cells)
	}
	assert.LessOrEqual(t, totalCells, 1000)
}

func TestErrorRecovery(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)
	conv := converter.NewConverter(logger)

	// Test with non-existent file
	doc, err := conv.ExcelToJSON("non-existent-file.xlsx", converter.ConvertOptions{})
	assert.Error(t, err)
	assert.Nil(t, doc)

	// Test with invalid output path
	validDoc := &models.ExcelDocument{
		Version: "1.0",
		Sheets: []models.Sheet{
			{
				Name:  "Sheet1",
				Index: 0,
				Cells: map[string]models.Cell{
					"A1": {Value: "Test", Type: models.CellTypeString},
				},
			},
		},
	}

	err = conv.JSONToExcel(validDoc, "/invalid/path/output.xlsx", converter.ConvertOptions{})
	assert.Error(t, err)
}

func TestMemoryUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory test in short mode")
	}

	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)
	conv := converter.NewConverter(logger)

	// Test with empty cells options to ensure we don't load unnecessary data
	doc, err := conv.ExcelToJSON("testdata/sample_files/complex.xlsx", converter.ConvertOptions{
		IgnoreEmptyCells: true,
		MaxCellsPerSheet: 100,
	})
	require.NoError(t, err)
	require.NotNil(t, doc)

	// Verify that the converter respects limits and options
	for _, sheet := range doc.Sheets {
		assert.LessOrEqual(t, len(sheet.Cells), 100, "Sheet %s has too many cells", sheet.Name)
	}
}