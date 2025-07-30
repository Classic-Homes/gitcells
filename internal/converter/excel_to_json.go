package converter

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/Classic-Homes/gitcells/pkg/models"
	"github.com/xuri/excelize/v2"
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
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close Excel file: %v\n", closeErr)
		}
	}()

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
			AppVersion:   "gitcells-0.1.0",
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

	// Process each sheet with progress tracking
	sheetList := f.GetSheetList()
	totalSheets := len(sheetList)

	if options.ProgressCallback != nil {
		options.ProgressCallback("Initializing", 0, totalSheets)
	}

	for index, sheetName := range sheetList {
		if options.ProgressCallback != nil {
			options.ProgressCallback("Processing sheets", index, totalSheets)
		}

		sheet, err := c.processSheet(f, sheetName, index, options)
		if err != nil {
			c.logger.Warnf("Failed to process sheet %s: %v", sheetName, err)
			continue
		}
		doc.Sheets = append(doc.Sheets, *sheet)
	}

	if options.ProgressCallback != nil {
		options.ProgressCallback("Processing complete", totalSheets, totalSheets)
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
	processedRows := 0
	totalRows := len(rows)

	for rowIndex, row := range rows {
		// Report progress every 100 rows to avoid performance impact
		if options.ProgressCallback != nil && rowIndex%100 == 0 {
			options.ProgressCallback(fmt.Sprintf("Processing sheet %s", sheetName), processedRows, totalRows)
		}

		for colIndex, cellValue := range row {
			if options.MaxCellsPerSheet > 0 && cellCount >= options.MaxCellsPerSheet {
				c.logger.Warnf("Sheet %s exceeded max cells limit (%d)", sheetName, options.MaxCellsPerSheet)
				break
			}

			cellRef, _ := excelize.CoordinatesToCellName(colIndex+1, rowIndex+1)

			// Get enhanced formula information if requested
			var formula, formulaR1C1 string
			var arrayFormula *models.ArrayFormula
			if options.PreserveFormulas {
				var err error
				formula, formulaR1C1, arrayFormula, err = c.extractEnhancedFormulaInfo(f, sheetName, cellRef)
				if err != nil {
					c.logger.Debugf("Failed to extract formula info for %s: %v", cellRef, err)
					// Fallback to basic formula extraction
					formula, _ = f.GetCellFormula(sheetName, cellRef)
				}
			}

			// Skip empty cells if option is set, BUT NOT if cell has a formula
			if options.IgnoreEmptyCells && cellValue == "" && formula == "" {
				continue
			}

			// Start with string value from rows
			var cellVal interface{} = cellValue
			cellType := c.detectCellType(cellValue, formula)

			// Convert numeric strings to actual numbers if it's a number type
			if cellType == models.CellTypeNumber && formula == "" {
				if numVal, err := parseNumber(cellValue); err == nil {
					cellVal = numVal
				}
			}

			cell := models.Cell{
				Value:        cellVal,
				Type:         cellType,
				Formula:      formula,
				FormulaR1C1:  formulaR1C1,
				ArrayFormula: arrayFormula,
			}

			// Get style if requested
			if options.PreserveStyles {
				cell.Style = c.extractFullCellStyle(f, sheetName, cellRef)
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
		processedRows++
	}

	c.logger.WithFields(map[string]interface{}{
		"sheet":          sheetName,
		"processed_rows": processedRows,
		"total_cells":    cellCount,
	}).Debug("Completed sheet cell processing")

	// Get merged cells
	mergedCells, _ := f.GetMergeCells(sheetName)
	for _, mc := range mergedCells {
		sheet.MergedCells = append(sheet.MergedCells, models.MergedCell{
			Range: mc.GetStartAxis() + ":" + mc.GetEndAxis(),
		})
	}

	// Extract charts if requested
	if options.PreserveCharts {
		charts, err := c.extractCharts(f, sheetName)
		if err != nil {
			c.logger.Warnf("Failed to extract charts from sheet %s: %v", sheetName, err)
		} else {
			sheet.Charts = charts
		}
	}

	// Extract pivot tables if requested
	if options.PreservePivotTables {
		pivotTables, err := c.extractPivotTables(f, sheetName)
		if err != nil {
			c.logger.Warnf("Failed to extract pivot tables from sheet %s: %v", sheetName, err)
		} else {
			sheet.PivotTables = pivotTables
		}
	}

	c.logger.WithFields(map[string]interface{}{
		"sheet":        sheetName,
		"total_cells":  len(sheet.Cells),
		"merged_cells": len(sheet.MergedCells),
		"charts":       len(sheet.Charts),
		"pivot_tables": len(sheet.PivotTables),
	}).Debug("Sheet processing completed")

	return sheet, nil
}

func (c *converter) calculateChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath) // #nosec G304 - file path is user input
	if err != nil {
		return "", err
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close file: %v\n", closeErr)
		}
	}()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}
