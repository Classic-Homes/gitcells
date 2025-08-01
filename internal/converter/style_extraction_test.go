package converter

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"
)

func TestStyleExtraction(t *testing.T) {
	// Create a test Excel file with various styles
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Sheet1"
	
	// Create a style with font formatting
	style1, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:      true,
			Italic:    true,
			Family:    "Arial",
			Size:      14,
			Color:     "#FF0000",
			Underline: "single",
		},
	})
	require.NoError(t, err)
	
	// Create a style with fill
	style2, err := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{
			Type:    "pattern",
			Pattern: 1, // solid
			Color:   []string{"#FFFF00"},
		},
	})
	require.NoError(t, err)
	
	// Create a style with borders
	style3, err := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 2},
			{Type: "top", Color: "#0000FF", Style: 3},
			{Type: "bottom", Color: "#0000FF", Style: 4},
		},
	})
	require.NoError(t, err)
	
	// Create a style with alignment
	style4, err := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			Horizontal:   "center",
			Vertical:     "middle",
			WrapText:     true,
			TextRotation: 45,
		},
	})
	require.NoError(t, err)
	
	// Create a style with number format
	style5, err := f.NewStyle(&excelize.Style{
		NumFmt: 4, // #,##0.00
	})
	require.NoError(t, err)

	// Apply styles to cells
	f.SetCellValue(sheetName, "A1", "Bold Italic Red Arial")
	f.SetCellStyle(sheetName, "A1", "A1", style1)
	
	f.SetCellValue(sheetName, "B1", "Yellow Background")
	f.SetCellStyle(sheetName, "B1", "B1", style2)
	
	f.SetCellValue(sheetName, "C1", "Borders")
	f.SetCellStyle(sheetName, "C1", "C1", style3)
	
	f.SetCellValue(sheetName, "D1", "Centered & Rotated")
	f.SetCellStyle(sheetName, "D1", "D1", style4)
	
	f.SetCellValue(sheetName, "E1", 1234.56)
	f.SetCellStyle(sheetName, "E1", "E1", style5)

	// Save file
	testFile := "test_styles.xlsx"
	err = f.SaveAs(testFile)
	require.NoError(t, err)
	defer os.Remove(testFile)

	// Now test extraction
	logger := logrus.New()
	conv := NewConverter(logger)
	
	// Open the file for reading
	readFile, err := excelize.OpenFile(testFile)
	require.NoError(t, err)
	defer readFile.Close()

	// Test font extraction
	t.Run("Font extraction", func(t *testing.T) {
		style := conv.(*converter).extractFullCellStyle(readFile, sheetName, "A1")
		require.NotNil(t, style)
		require.NotNil(t, style.Font)
		assert.Equal(t, "Arial", style.Font.Name)
		assert.Equal(t, float64(14), style.Font.Size)
		assert.True(t, style.Font.Bold)
		assert.True(t, style.Font.Italic)
		assert.Equal(t, "FF0000", style.Font.Color)
		assert.Equal(t, "single", style.Font.Underline)
	})

	// Test fill extraction
	t.Run("Fill extraction", func(t *testing.T) {
		style := conv.(*converter).extractFullCellStyle(readFile, sheetName, "B1")
		require.NotNil(t, style)
		require.NotNil(t, style.Fill)
		assert.Equal(t, "pattern", style.Fill.Type)
		assert.Equal(t, "solid", style.Fill.Pattern)
		assert.Equal(t, "FFFF00", style.Fill.Color)
	})

	// Test border extraction
	t.Run("Border extraction", func(t *testing.T) {
		style := conv.(*converter).extractFullCellStyle(readFile, sheetName, "C1")
		require.NotNil(t, style)
		require.NotNil(t, style.Border)
		assert.NotNil(t, style.Border.Left)
		assert.Equal(t, "thin", style.Border.Left.Style)
		assert.NotNil(t, style.Border.Right)
		assert.Equal(t, "medium", style.Border.Right.Style)
		assert.NotNil(t, style.Border.Top)
		assert.Equal(t, "dashed", style.Border.Top.Style)
		assert.NotNil(t, style.Border.Bottom)
		assert.Equal(t, "dotted", style.Border.Bottom.Style)
	})

	// Test alignment extraction
	t.Run("Alignment extraction", func(t *testing.T) {
		style := conv.(*converter).extractFullCellStyle(readFile, sheetName, "D1")
		require.NotNil(t, style)
		require.NotNil(t, style.Alignment)
		assert.Equal(t, "center", style.Alignment.Horizontal)
		assert.Equal(t, "middle", style.Alignment.Vertical)
		assert.True(t, style.Alignment.WrapText)
		assert.Equal(t, 45, style.Alignment.TextRotation)
	})

	// Test number format extraction
	t.Run("Number format extraction", func(t *testing.T) {
		style := conv.(*converter).extractFullCellStyle(readFile, sheetName, "E1")
		require.NotNil(t, style)
		assert.Equal(t, "#,##0.00", style.NumberFormat)
	})
}