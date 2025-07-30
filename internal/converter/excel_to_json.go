package converter

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"time"
	
	"github.com/xuri/excelize/v2"
	"github.com/Classic-Homes/sheetsync/pkg/models"
)

func (c *converter) ExcelToJSON(filePath string, options ConvertOptions) (*models.ExcelDocument, error) {
	// Calculate file checksum
	checksum, err := c.calculateChecksum(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate checksum: %w", err)
	}

	// Open Excel file
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open Excel file: %w", err)
	}
	defer f.Close()

	// Get file info
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	doc := &models.ExcelDocument{
		Version: "1.0",
		Metadata: models.DocumentMetadata{
			Created:      time.Now(),
			Modified:     fileInfo.ModTime(),
			AppVersion:   "sheetsync-0.1.0",
			OriginalFile: filePath,
			FileSize:     fileInfo.Size(),
			Checksum:     checksum,
		},
		Sheets:       []models.Sheet{},
		DefinedNames: make(map[string]string),
	}

	// Extract document properties
	props, err := f.GetDocProps()
	if err == nil && props != nil {
		doc.Properties = c.extractProperties(props)
	}

	// Process each sheet
	for index, sheetName := range f.GetSheetList() {
		sheet, err := c.processSheet(f, sheetName, index, options)
		if err != nil {
			c.logger.Warnf("Failed to process sheet %s: %v", sheetName, err)
			continue
		}
		doc.Sheets = append(doc.Sheets, *sheet)
	}

	// Extract defined names
	for _, definedName := range f.GetDefinedName() {
		doc.DefinedNames[definedName.Name] = definedName.RefersTo
	}

	return doc, nil
}

func (c *converter) processSheet(f *excelize.File, sheetName string, index int, options ConvertOptions) (*models.Sheet, error) {
	sheet := &models.Sheet{
		Name:         sheetName,
		Index:        index,
		Cells:        make(map[string]models.Cell),
		MergedCells:  []models.MergedCell{},
		RowHeights:   make(map[int]float64),
		ColumnWidths: make(map[string]float64),
	}

	// Get all cells in sheet
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, err
	}

	cellCount := 0
	for rowIndex, row := range rows {
		for colIndex, cellValue := range row {
			if options.MaxCellsPerSheet > 0 && cellCount >= options.MaxCellsPerSheet {
				c.logger.Warnf("Sheet %s exceeded max cells limit (%d)", sheetName, options.MaxCellsPerSheet)
				break
			}

			cellRef, _ := excelize.CoordinatesToCellName(colIndex+1, rowIndex+1)
			
			// Get formula if exists - check this first
			var formula string
			if options.PreserveFormulas {
				formula, _ = f.GetCellFormula(sheetName, cellRef)
			}

			// Skip empty cells if option is set, BUT NOT if cell has a formula
			if options.IgnoreEmptyCells && cellValue == "" && formula == "" {
				continue
			}

			// Start with string value from rows
			var cellVal interface{} = cellValue
			cellType := c.detectCellType(cellValue)

			// Formula takes precedence over detected type
			if formula != "" {
				cellType = models.CellTypeFormula
			}

			// Convert numeric strings to actual numbers if it's a number type
			if cellType == models.CellTypeNumber && formula == "" {
				if numVal, err := parseNumber(cellValue); err == nil {
					cellVal = numVal
				}
			}

			cell := models.Cell{
				Value:   cellVal,
				Type:    cellType,
				Formula: formula,
			}

			// Get style if requested
			if options.PreserveStyles {
				styleID, _ := f.GetCellStyle(sheetName, cellRef)
				if styleID > 0 {
					cell.Style = c.extractCellStyle(f, styleID)
				}
			}

			// Get comment if requested
			if options.PreserveComments {
				comments, _ := f.GetComments(sheetName)
				for _, comment := range comments {
					if comment.Cell == cellRef {
						cell.Comment = &models.Comment{
							Author: comment.Author,
							Text:   comment.Text,
						}
						break
					}
				}
			}

			sheet.Cells[cellRef] = cell
			cellCount++
		}
	}

	// Get merged cells
	mergedCells, _ := f.GetMergeCells(sheetName)
	for _, mc := range mergedCells {
		sheet.MergedCells = append(sheet.MergedCells, models.MergedCell{
			Range: mc.GetStartAxis() + ":" + mc.GetEndAxis(),
		})
	}

	return sheet, nil
}

func (c *converter) calculateChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}