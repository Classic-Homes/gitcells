package converter

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"

	"github.com/Classic-Homes/sheetsync/pkg/models"
)

func TestJSONToExcel(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)
	conv := NewConverter(logger)

	// Create a temporary directory for output files
	tempDir := t.TempDir()

	tests := []struct {
		name     string
		doc      *models.ExcelDocument
		options  ConvertOptions
		validate func(t *testing.T, outputPath string)
	}{
		{
			name: "simple document",
			doc: &models.ExcelDocument{
				Version: "1.0",
				Metadata: models.DocumentMetadata{
					Created:      time.Now(),
					Modified:     time.Now(),
					AppVersion:   "sheetsync-test",
					OriginalFile: "test.xlsx",
					FileSize:     1024,
					Checksum:     "abc123",
				},
				Sheets: []models.Sheet{
					{
						Name:  "Sheet1",
						Index: 0,
						Cells: map[string]models.Cell{
							"A1": {Value: "Hello", Type: models.CellTypeString},
							"B1": {Value: "World", Type: models.CellTypeString},
							"A2": {Value: float64(42), Type: models.CellTypeNumber},
							"B2": {Value: true, Type: models.CellTypeBoolean},
						},
					},
				},
				DefinedNames: make(map[string]string),
			},
			options: ConvertOptions{
				PreserveFormulas: true,
			},
			validate: func(t *testing.T, outputPath string) {
				// Verify file was created
				_, err := os.Stat(outputPath)
				assert.NoError(t, err)

				// Open and verify contents
				f, err := excelize.OpenFile(outputPath)
				require.NoError(t, err)
				defer f.Close()

				// Check cell values
				cellValue, err := f.GetCellValue("Sheet1", "A1")
				assert.NoError(t, err)
				assert.Equal(t, "Hello", cellValue)

				cellValue, err = f.GetCellValue("Sheet1", "B1")
				assert.NoError(t, err)
				assert.Equal(t, "World", cellValue)

				// Check number value
				cellValue, err = f.GetCellValue("Sheet1", "A2")
				assert.NoError(t, err)
				assert.Equal(t, "42", cellValue)

				// Check boolean value
				cellValue, err = f.GetCellValue("Sheet1", "B2")
				assert.NoError(t, err)
				assert.Equal(t, "TRUE", cellValue)
			},
		},
		{
			name: "document with formulas",
			doc: &models.ExcelDocument{
				Version: "1.0",
				Metadata: models.DocumentMetadata{
					Created:    time.Now(),
					Modified:   time.Now(),
					AppVersion: "sheetsync-test",
				},
				Sheets: []models.Sheet{
					{
						Name:  "Sheet1",
						Index: 0,
						Cells: map[string]models.Cell{
							"A1": {Value: float64(10), Type: models.CellTypeNumber},
							"A2": {Value: float64(20), Type: models.CellTypeNumber},
							"A3": {
								Value:   float64(30),
								Formula: "=A1+A2",
								Type:    models.CellTypeFormula,
							},
						},
					},
				},
				DefinedNames: make(map[string]string),
			},
			options: ConvertOptions{
				PreserveFormulas: true,
			},
			validate: func(t *testing.T, outputPath string) {
				f, err := excelize.OpenFile(outputPath)
				require.NoError(t, err)
				defer f.Close()

				// Check formula
				formula, err := f.GetCellFormula("Sheet1", "A3")
				assert.NoError(t, err)
				assert.Equal(t, "A1+A2", formula) // excelize strips the = prefix
			},
		},
		{
			name: "document with merged cells",
			doc: &models.ExcelDocument{
				Version: "1.0",
				Metadata: models.DocumentMetadata{
					Created:    time.Now(),
					Modified:   time.Now(),
					AppVersion: "sheetsync-test",
				},
				Sheets: []models.Sheet{
					{
						Name:  "Sheet1",
						Index: 0,
						Cells: map[string]models.Cell{
							"A1": {Value: "Merged Cell", Type: models.CellTypeString},
						},
						MergedCells: []models.MergedCell{
							{Range: "A1:B2"},
						},
					},
				},
				DefinedNames: make(map[string]string),
			},
			options: ConvertOptions{},
			validate: func(t *testing.T, outputPath string) {
				f, err := excelize.OpenFile(outputPath)
				require.NoError(t, err)
				defer f.Close()

				// Check merged cells
				mergedCells, err := f.GetMergeCells("Sheet1")
				assert.NoError(t, err)
				assert.Len(t, mergedCells, 1)
				
				if len(mergedCells) > 0 {
					startAxis := mergedCells[0].GetStartAxis()
					endAxis := mergedCells[0].GetEndAxis()
					assert.Equal(t, "A1", startAxis)
					assert.Equal(t, "B2", endAxis)
				}
			},
		},
		{
			name: "document with multiple sheets",
			doc: &models.ExcelDocument{
				Version: "1.0",
				Metadata: models.DocumentMetadata{
					Created:    time.Now(),
					Modified:   time.Now(),
					AppVersion: "sheetsync-test",
				},
				Sheets: []models.Sheet{
					{
						Name:  "Sheet1",
						Index: 0,
						Cells: map[string]models.Cell{
							"A1": {Value: "First Sheet", Type: models.CellTypeString},
						},
					},
					{
						Name:  "Data",
						Index: 1,
						Cells: map[string]models.Cell{
							"A1": {Value: "Second Sheet", Type: models.CellTypeString},
						},
					},
				},
				DefinedNames: make(map[string]string),
			},
			options: ConvertOptions{},
			validate: func(t *testing.T, outputPath string) {
				f, err := excelize.OpenFile(outputPath)
				require.NoError(t, err)
				defer f.Close()

				// Check sheet names
				sheetList := f.GetSheetList()
				assert.Contains(t, sheetList, "Sheet1")
				assert.Contains(t, sheetList, "Data")

				// Check cell values in both sheets
				value1, err := f.GetCellValue("Sheet1", "A1")
				assert.NoError(t, err)
				assert.Equal(t, "First Sheet", value1)

				value2, err := f.GetCellValue("Data", "A1")
				assert.NoError(t, err)
				assert.Equal(t, "Second Sheet", value2)
			},
		},
		{
			name: "document with comments",
			doc: &models.ExcelDocument{
				Version: "1.0",
				Metadata: models.DocumentMetadata{
					Created:    time.Now(),
					Modified:   time.Now(),
					AppVersion: "sheetsync-test",
				},
				Sheets: []models.Sheet{
					{
						Name:  "Sheet1",
						Index: 0,
						Cells: map[string]models.Cell{
							"A1": {
								Value: "Cell with comment",
								Type:  models.CellTypeString,
								Comment: &models.Comment{
									Author: "Test Author",
									Text:   "This is a test comment",
								},
							},
						},
					},
				},
				DefinedNames: make(map[string]string),
			},
			options: ConvertOptions{
				PreserveComments: true,
			},
			validate: func(t *testing.T, outputPath string) {
				f, err := excelize.OpenFile(outputPath)
				require.NoError(t, err)
				defer f.Close()

				// Check comments
				comments, err := f.GetComments("Sheet1")
				assert.NoError(t, err)
				
				found := false
				for _, comment := range comments {
					if comment.Cell == "A1" {
						assert.Equal(t, "Test Author", comment.Author)
						assert.Equal(t, "This is a test comment", comment.Text)
						found = true
						break
					}
				}
				assert.True(t, found, "Expected to find comment on cell A1")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outputPath := filepath.Join(tempDir, tt.name+".xlsx")
			err := conv.JSONToExcel(tt.doc, outputPath, tt.options)
			require.NoError(t, err)

			tt.validate(t, outputPath)
		})
	}
}

func TestJSONToExcel_ErrorCases(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)
	conv := NewConverter(logger)

	tests := []struct {
		name       string
		doc        *models.ExcelDocument
		outputPath string
	}{
		{
			name:       "nil document",
			doc:        nil,
			outputPath: "/tmp/test.xlsx",
		},
		{
			name: "invalid output path",
			doc: &models.ExcelDocument{
				Version: "1.0",
				Sheets:  []models.Sheet{},
			},
			outputPath: "/invalid/path/that/does/not/exist/test.xlsx",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := conv.JSONToExcel(tt.doc, tt.outputPath, ConvertOptions{})
			assert.Error(t, err)
		})
	}
}

func TestRoundTripConversion(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)
	conv := NewConverter(logger)

	// Test round-trip conversion: Excel -> JSON -> Excel
	originalFile := "../../test/testdata/sample_files/simple.xlsx"
	tempDir := t.TempDir()
	
	// Convert Excel to JSON
	doc, err := conv.ExcelToJSON(originalFile, ConvertOptions{
		PreserveFormulas: true,
		PreserveComments: true,
	})
	require.NoError(t, err)

	// Convert JSON back to Excel
	outputFile := filepath.Join(tempDir, "roundtrip.xlsx")
	err = conv.JSONToExcel(doc, outputFile, ConvertOptions{
		PreserveFormulas: true,
		PreserveComments: true,
	})
	require.NoError(t, err)

	// Verify the output file exists and can be opened
	f, err := excelize.OpenFile(outputFile)
	require.NoError(t, err)
	defer f.Close()

	// Basic verification - check that we have at least one sheet with some data
	sheetList := f.GetSheetList()
	assert.NotEmpty(t, sheetList)

	// Check some basic cell values
	cellValue, err := f.GetCellValue(sheetList[0], "A1")
	assert.NoError(t, err)
	assert.NotEmpty(t, cellValue)
}