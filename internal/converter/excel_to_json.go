package converter

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/Classic-Homes/gitcells/internal/utils"
	"github.com/Classic-Homes/gitcells/pkg/models"
	"github.com/xuri/excelize/v2"
)

func (c *converter) ExcelToJSON(filePath string, options ConvertOptions) (*models.ExcelDocument, error) {
	// Calculate file checksum
	checksum, err := c.calculateChecksum(filePath)
	if err != nil {
		return nil, utils.WrapFileError(err, utils.ErrorTypeFileSystem, "ExcelToJSON", filePath, "failed to calculate checksum")
	}

	// Open Excel file
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, utils.WrapFileError(err, utils.ErrorTypeFileSystem, "ExcelToJSON", filePath, "failed to open Excel file")
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close Excel file: %v\n", closeErr)
		}
	}()

	// Get file info
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, utils.WrapFileError(err, utils.ErrorTypeFileSystem, "ExcelToJSON", filePath, "failed to get file info")
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
	sheetsToProcess := c.filterSheets(sheetList, options)
	totalSheets := len(sheetsToProcess)

	if options.ProgressCallback != nil {
		options.ProgressCallback("Initializing", 0, totalSheets)
	}

	processedIndex := 0
	for originalIndex, sheetName := range sheetList {
		// Skip sheets not in our filtered list
		if !c.shouldProcessSheet(sheetName, originalIndex, options) {
			continue
		}

		if options.ProgressCallback != nil {
			options.ProgressCallback("Processing sheets", processedIndex, totalSheets)
		}

		sheet, err := c.processSheet(f, sheetName, originalIndex, options)
		if err != nil {
			c.logger.Warnf("Failed to process sheet %s: %v", sheetName, err)
			continue
		}
		doc.Sheets = append(doc.Sheets, *sheet)
		processedIndex++
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

	// Extract data validations if requested
	if options.PreserveDataValidation {
		if err := c.extractDataValidations(f, sheetName, sheet); err != nil {
			c.logger.Warnf("Failed to extract data validations from sheet %s: %v", sheetName, err)
		}
	}

	// Extract conditional formatting if requested
	if options.PreserveConditionalFormats {
		if conditionalFormats, err := c.extractConditionalFormats(f, sheetName); err != nil {
			c.logger.Warnf("Failed to extract conditional formats from sheet %s: %v", sheetName, err)
		} else {
			sheet.ConditionalFormats = conditionalFormats
		}
	}

	// Extract tables if requested
	if options.PreserveTables {
		if tables, err := c.extractTables(f, sheetName); err != nil {
			c.logger.Warnf("Failed to extract tables from sheet %s: %v", sheetName, err)
		} else {
			sheet.Tables = tables
		}
	}

	// Extract rich text for cells if requested
	if options.PreserveRichText {
		for cellRef, cell := range sheet.Cells {
			if richText, err := c.extractRichText(f, sheetName, cellRef); err == nil && len(richText) > 0 {
				cell.RichText = richText
				sheet.Cells[cellRef] = cell
			}
		}
	}

	c.logger.WithFields(map[string]interface{}{
		"sheet":               sheetName,
		"total_cells":         len(sheet.Cells),
		"merged_cells":        len(sheet.MergedCells),
		"charts":              len(sheet.Charts),
		"pivot_tables":        len(sheet.PivotTables),
		"conditional_formats": len(sheet.ConditionalFormats),
		"tables":              len(sheet.Tables),
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

// filterSheets returns the list of sheets that should be processed based on options
func (c *converter) filterSheets(allSheets []string, options ConvertOptions) []string {
	var filtered []string

	for index, sheetName := range allSheets {
		if c.shouldProcessSheet(sheetName, index, options) {
			filtered = append(filtered, sheetName)
		}
	}

	return filtered
}

// shouldProcessSheet determines if a sheet should be processed based on filtering options
func (c *converter) shouldProcessSheet(sheetName string, index int, options ConvertOptions) bool {
	// If ExcludeSheets is specified, check if this sheet should be excluded
	for _, excludeSheet := range options.ExcludeSheets {
		if sheetName == excludeSheet {
			return false
		}
	}

	// If SheetsToConvert is specified, only process sheets in this list
	if len(options.SheetsToConvert) > 0 {
		for _, includeSheet := range options.SheetsToConvert {
			if sheetName == includeSheet {
				return true
			}
		}
		return false
	}

	// If SheetIndices is specified, only process sheets at these indices
	if len(options.SheetIndices) > 0 {
		for _, includeIndex := range options.SheetIndices {
			if index == includeIndex {
				return true
			}
		}
		return false
	}

	// If no specific filters are set, process all sheets (except those excluded)
	return true
}
