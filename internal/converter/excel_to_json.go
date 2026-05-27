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

var (
	// excelOpenTimeout is the budget for retrying a locked Excel file.
	// Overridable in tests.
	excelOpenTimeout = DefaultExcelOpenTimeout
)

func (c *converter) ExcelToJSON(filePath string, options ConvertOptions) (*models.ExcelDocument, error) {
	// Calculate file checksum
	checksum, err := c.calculateChecksum(filePath)
	if err != nil {
		return nil, utils.WrapFileError(err, utils.ErrorTypeFileSystem, "ExcelToJSON", filePath, "failed to calculate checksum")
	}

	// Open Excel file with retry. Excel on Windows holds the file open
	// through its atomic-save rename; without retry, the watcher can lose
	// the conversion if it fires before the rename completes.
	f, err := openExcelWithRetry(filePath, excelOpenTimeout)
	if err != nil {
		return nil, utils.WrapFileError(err, utils.ErrorTypeFileSystem, "ExcelToJSON", filePath, "failed to open Excel file")
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			c.logger.Warnf("failed to close Excel file %s: %v", filePath, closeErr)
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

	// Build a per-sheet comment index once, instead of re-fetching and
	// linearly scanning for every cell. On a dense sheet with thousands
	// of cells, the old code was O(cells × comments). This is O(cells + comments).
	var commentByCell map[string]models.Comment
	if options.PreserveComments {
		commentByCell = buildCommentIndex(f, sheetName, c.logger)
	}

	cellCount := 0
	processedRows := 0
	totalRows := len(rows)
	var cellRefErrors, formulaErrors int

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

			cellRef, refErr := excelize.CoordinatesToCellName(colIndex+1, rowIndex+1)
			if refErr != nil {
				// Coordinate→ref is library-internal; a failure here means
				// excelize changed its validation. Skip the cell rather
				// than produce a malformed map key, and surface the count
				// at end-of-sheet so it's not lost in the noise.
				cellRefErrors++
				continue
			}

			var formula string
			if options.PreserveFormulas {
				f2, fErr := f.GetCellFormula(sheetName, cellRef)
				if fErr != nil {
					formulaErrors++
				} else {
					formula = f2
				}
			}

			// Skip empty cells if option is set, BUT NOT if cell has a formula
			if options.IgnoreEmptyCells && cellValue == "" && formula == "" {
				continue
			}

			var cellVal interface{} = cellValue
			cellType := c.detectCellType(cellValue, formula)

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
				cell.Style = c.extractFullCellStyle(f, sheetName, cellRef)
			}

			if options.PreserveComments {
				if cmt, ok := commentByCell[cellRef]; ok {
					c := cmt
					cell.Comment = &c
				}
			}

			// Hyperlinks: extract target. Type (External vs Location) is
			// not recoverable from excelize's API, so we classify at write
			// time by inspecting the target string.
			if hasLink, target, hErr := f.GetCellHyperLink(sheetName, cellRef); hErr == nil && hasLink {
				cell.Hyperlink = target
			}

			sheet.Cells[cellRef] = cell
			cellCount++
		}
		processedRows++
	}

	if cellRefErrors > 0 {
		c.logger.Warnf("Sheet %s: skipped %d cells due to coordinate conversion errors", sheetName, cellRefErrors)
	}
	if formulaErrors > 0 {
		c.logger.Warnf("Sheet %s: %d cells had formula-read errors (formulas stored as empty)", sheetName, formulaErrors)
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

	// Data validation is always extracted — it round-trips cleanly and is
	// common enough in templates that dropping it would be surprising.
	if err := c.extractDataValidations(f, sheetName, sheet); err != nil {
		c.logger.Warnf("Failed to extract data validations from sheet %s: %v", sheetName, err)
	}

	// Tell the user what we deliberately did NOT preserve from the source.
	c.warnAboutUnsupportedFeatures(f, sheetName)

	c.logger.WithFields(map[string]interface{}{
		"sheet":        sheetName,
		"total_cells":  len(sheet.Cells),
		"merged_cells": len(sheet.MergedCells),
	}).Debug("Sheet processing completed")

	return sheet, nil
}

// buildCommentIndex returns a cell-ref → Comment map for the sheet, or
// an empty map if comments can't be read. Logged as Warnf when read
// fails — comments are not critical to round-trip and shouldn't fail
// the whole conversion.
func buildCommentIndex(f *excelize.File, sheetName string, logger interface {
	Warnf(format string, args ...interface{})
}) map[string]models.Comment {
	idx := make(map[string]models.Comment)
	comments, err := f.GetComments(sheetName)
	if err != nil {
		logger.Warnf("Failed to read comments on sheet %s: %v", sheetName, err)
		return idx
	}
	for _, cmt := range comments {
		idx[cmt.Cell] = models.Comment{
			Author: cmt.Author,
			Text:   cmt.Text,
		}
	}
	return idx
}

func (c *converter) calculateChecksum(filePath string) (string, error) {
	file, err := openFileWithRetry(filePath, excelOpenTimeout)
	if err != nil {
		return "", err
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			c.logger.Warnf("failed to close file %s: %v", filePath, closeErr)
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
