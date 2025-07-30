package converter

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"

	"github.com/Classic-Homes/sheetsync/pkg/models"
)

func TestExcelToJSON(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel) // Reduce noise in tests
	conv := NewConverter(logger)

	tests := []struct {
		name     string
		file     string
		options  ConvertOptions
		validate func(t *testing.T, doc *models.ExcelDocument)
	}{
		{
			name: "simple Excel file",
			file: "../../test/testdata/sample_files/simple.xlsx",
			options: ConvertOptions{
				PreserveFormulas: true,
				PreserveStyles:   false,
				PreserveComments: false,
				IgnoreEmptyCells: true,
			},
			validate: func(t *testing.T, doc *models.ExcelDocument) {
				assert.Equal(t, "1.0", doc.Version)
				assert.NotEmpty(t, doc.Metadata.Checksum)
				assert.Equal(t, "sheetsync-0.1.0", doc.Metadata.AppVersion)
				assert.Len(t, doc.Sheets, 1)

				sheet := doc.Sheets[0]
				assert.Equal(t, "Sheet1", sheet.Name)
				assert.Equal(t, 0, sheet.Index)

				// Check specific cells
				assert.Equal(t, "Name", sheet.Cells["A1"].Value)
				assert.Equal(t, models.CellTypeString, sheet.Cells["A1"].Type)

				assert.Equal(t, "Age", sheet.Cells["B1"].Value)
				assert.Equal(t, float64(25), sheet.Cells["B2"].Value)
				assert.Equal(t, models.CellTypeNumber, sheet.Cells["B2"].Type)

				// Check formula
				if formulaCell, exists := sheet.Cells["B4"]; exists {
					assert.Equal(t, "=SUM(B2:B3)", formulaCell.Formula)
					assert.Equal(t, models.CellTypeFormula, formulaCell.Type)
				}
			},
		},
		{
			name: "complex Excel file with multiple sheets",
			file: "../../test/testdata/sample_files/complex.xlsx",
			options: ConvertOptions{
				PreserveFormulas: true,
				PreserveStyles:   true,
				PreserveComments: true,
				IgnoreEmptyCells: true,
			},
			validate: func(t *testing.T, doc *models.ExcelDocument) {
				assert.Len(t, doc.Sheets, 2)

				// Check first sheet
				sheet1 := doc.Sheets[0]
				assert.Equal(t, "Sheet1", sheet1.Name)

				// Check merged cells
				assert.NotEmpty(t, sheet1.MergedCells)
				found := false
				for _, mc := range sheet1.MergedCells {
					if mc.Range == "A5:D5" {
						found = true
						break
					}
				}
				assert.True(t, found, "Expected to find merged cell A5:D5")

				// Check formula cells
				assert.Equal(t, "=B2*C2", sheet1.Cells["D2"].Formula)
				assert.Equal(t, models.CellTypeFormula, sheet1.Cells["D2"].Type)

				// Check second sheet
				sheet2 := doc.Sheets[1]
				assert.Equal(t, "Summary", sheet2.Name)
				assert.Equal(t, "=Sheet1.D4", sheet2.Cells["B1"].Formula)
			},
		},
		{
			name: "empty Excel file",
			file: "../../test/testdata/sample_files/empty.xlsx",
			options: ConvertOptions{
				IgnoreEmptyCells: true,
			},
			validate: func(t *testing.T, doc *models.ExcelDocument) {
				assert.Len(t, doc.Sheets, 1)
				sheet := doc.Sheets[0]
				assert.Equal(t, "Sheet1", sheet.Name)
				// Empty file should have no cells when ignoring empty cells
				assert.Empty(t, sheet.Cells)
			},
		},
		{
			name: "with cell limits",
			file: "../../test/testdata/sample_files/simple.xlsx",
			options: ConvertOptions{
				MaxCellsPerSheet: 3,
				IgnoreEmptyCells: true,
			},
			validate: func(t *testing.T, doc *models.ExcelDocument) {
				sheet := doc.Sheets[0]
				// Should only have 3 cells due to limit
				assert.LessOrEqual(t, len(sheet.Cells), 3)
			},
		},
		{
			name: "without ignoring empty cells",
			file: "../../test/testdata/sample_files/empty.xlsx",
			options: ConvertOptions{
				IgnoreEmptyCells: false,
			},
			validate: func(t *testing.T, doc *models.ExcelDocument) {
				// This test may have more cells including empty ones
				// depending on Excel's default structure
				assert.Len(t, doc.Sheets, 1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := conv.ExcelToJSON(tt.file, tt.options)
			require.NoError(t, err)
			require.NotNil(t, doc)

			tt.validate(t, doc)
		})
	}
}

func TestExcelToJSON_ErrorCases(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)
	conv := NewConverter(logger)

	tests := []struct {
		name string
		file string
	}{
		{
			name: "non-existent file",
			file: "../../test/testdata/sample_files/nonexistent.xlsx",
		},
		{
			name: "invalid file path",
			file: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := conv.ExcelToJSON(tt.file, ConvertOptions{})
			assert.Error(t, err)
			assert.Nil(t, doc)
		})
	}
}

func TestCalculateChecksum(t *testing.T) {
	logger := logrus.New()
	conv := &converter{logger: logger}

	// Test with a real file
	checksum1, err := conv.calculateChecksum("../../test/testdata/sample_files/simple.xlsx")
	require.NoError(t, err)
	assert.NotEmpty(t, checksum1)
	assert.Len(t, checksum1, 64) // SHA256 produces 64-char hex string

	// Test with same file - should produce same checksum
	checksum2, err := conv.calculateChecksum("../../test/testdata/sample_files/simple.xlsx")
	require.NoError(t, err)
	assert.Equal(t, checksum1, checksum2)

	// Test with different file - should produce different checksum
	checksum3, err := conv.calculateChecksum("../../test/testdata/sample_files/complex.xlsx")
	require.NoError(t, err)
	assert.NotEqual(t, checksum1, checksum3)
}

func TestCalculateChecksum_ErrorCases(t *testing.T) {
	logger := logrus.New()
	conv := &converter{logger: logger}

	// Test with non-existent file
	checksum, err := conv.calculateChecksum("nonexistent.xlsx")
	assert.Error(t, err)
	assert.Empty(t, checksum)
}

func TestExtractProperties(t *testing.T) {
	logger := logrus.New()
	conv := &converter{logger: logger}

	// Test with mock document properties
	props := &excelize.DocProperties{
		Title:       "Test Document",
		Subject:     "Testing",
		Creator:     "Test Author",
		Keywords:    "test,excel",
		Description: "A test document",
		Category:    "Test Company",
	}

	extracted := conv.extractProperties(props)

	assert.Equal(t, "Test Document", extracted.Title)
	assert.Equal(t, "Testing", extracted.Subject)
	assert.Equal(t, "Test Author", extracted.Author)
	assert.Equal(t, "test,excel", extracted.Keywords)
	assert.Equal(t, "A test document", extracted.Description)
	assert.Equal(t, "Test Company", extracted.Company)
}

func TestProcessSheet_ErrorHandling(t *testing.T) {
	logger := logrus.New()
	conv := &converter{logger: logger}

	// This test is more challenging as we need to mock excelize.File
	// For now, we'll test that the function handles errors gracefully
	// by testing with invalid sheet names in a real file

	f, err := excelize.OpenFile("../../test/testdata/sample_files/simple.xlsx")
	require.NoError(t, err)
	defer f.Close()

	// Test with non-existent sheet
	sheet, err := conv.processSheet(f, "NonExistentSheet", 0, ConvertOptions{})
	assert.Error(t, err)
	assert.Nil(t, sheet)
}
