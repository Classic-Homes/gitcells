package converter

import (
	"strings"

	"github.com/Classic-Homes/gitcells/pkg/models"
	"github.com/xuri/excelize/v2"
)

// expandSqref expands an Excel sqref string (space-separated list of ranges,
// e.g. "A1 B2:C3 D5") into the set of cell references it covers. Returns a
// set rather than a slice because the typical use is membership testing.
func expandSqref(sqref string) map[string]struct{} {
	set := make(map[string]struct{})
	for _, group := range strings.Fields(sqref) {
		parts := strings.SplitN(group, ":", 2)
		if len(parts) == 1 {
			set[parts[0]] = struct{}{}
			continue
		}
		startCol, startRow, err := excelize.CellNameToCoordinates(parts[0])
		if err != nil {
			continue
		}
		endCol, endRow, err := excelize.CellNameToCoordinates(parts[1])
		if err != nil {
			continue
		}
		if endCol < startCol {
			startCol, endCol = endCol, startCol
		}
		if endRow < startRow {
			startRow, endRow = endRow, startRow
		}
		for r := startRow; r <= endRow; r++ {
			for c := startCol; c <= endCol; c++ {
				name, err := excelize.CoordinatesToCellName(c, r)
				if err != nil {
					continue
				}
				set[name] = struct{}{}
			}
		}
	}
	return set
}

// extractDataValidations extracts all data validation rules from a sheet
func (c *converter) extractDataValidations(f *excelize.File, sheetName string, sheet *models.Sheet) error {
	// Get all data validations for the sheet
	validations, err := f.GetDataValidations(sheetName)
	if err != nil {
		return err
	}

	for _, validation := range validations {
		// Sqref contains the cell range as a string like "A1:A10" or "A1"
		// For simplicity, we'll apply the validation to all cells that match
		// In a full implementation, we'd parse the range and apply to each cell

		// Create the data validation model
		dv := &models.DataValidation{
			Type:             validation.Type,
			Operator:         validation.Operator,
			Formula1:         validation.Formula1,
			Formula2:         validation.Formula2,
			AllowBlank:       validation.AllowBlank,
			ShowInputMessage: validation.ShowInputMessage,
			ShowErrorMessage: validation.ShowErrorMessage,
		}

		// Handle optional string pointers
		if validation.ErrorTitle != nil {
			dv.ErrorTitle = *validation.ErrorTitle
		}
		if validation.Error != nil {
			dv.Error = *validation.Error
		}
		if validation.PromptTitle != nil {
			dv.PromptTitle = *validation.PromptTitle
		}
		if validation.Prompt != nil {
			dv.Prompt = *validation.Prompt
		}

		// Expand the validation's sqref into the exact set of covered cells
		// and apply only to those. Previously a substring check matched A1
		// inside A10 / A11 etc., spraying validations onto unrelated cells.
		members := expandSqref(validation.Sqref)
		for cellRef := range members {
			if cell, ok := sheet.Cells[cellRef]; ok {
				cell.DataValidation = dv
				sheet.Cells[cellRef] = cell
			}
		}
	}

	return nil
}

// extractConditionalFormats extracts conditional formatting rules from a sheet
func (c *converter) extractConditionalFormats(f *excelize.File, sheetName string) ([]models.ConditionalFormat, error) {
	var formats []models.ConditionalFormat

	// Get conditional formats from excelize
	conditionalFormats, err := f.GetConditionalFormats(sheetName)
	if err != nil {
		return nil, err
	}

	for rangeRef, formatOpts := range conditionalFormats {
		for _, opt := range formatOpts {
			// Convert excelize conditional format to our model
			format := models.ConditionalFormat{
				Range:    rangeRef,
				Type:     opt.Type,
				Criteria: opt.Criteria,
				Value:    opt.Value,
			}

			// Extract format style if present
			if opt.Format != nil {
				format.Format = c.convertConditionalStyle(opt.Format)
			}

			formats = append(formats, format)
		}
	}

	return formats, nil
}

// convertConditionalStyle would convert an excelize conditional format style
// to our CellStyle model. Excelize does not yet expose enough structure to
// implement this round-trip, so we return nil rather than a misleading empty
// style.
func (c *converter) convertConditionalStyle(style interface{}) *models.CellStyle {
	return nil
}

// extractRichText extracts rich text formatting from a cell
func (c *converter) extractRichText(f *excelize.File, sheetName, cellRef string) ([]models.RichTextRun, error) {
	runs, err := f.GetCellRichText(sheetName, cellRef)
	if err != nil || len(runs) == 0 {
		return nil, err
	}

	richTextRuns := make([]models.RichTextRun, 0, len(runs))
	for _, run := range runs {
		rtRun := models.RichTextRun{
			Text: run.Text,
		}

		if run.Font != nil {
			rtRun.Font = &models.Font{
				Name:      run.Font.Family,
				Size:      run.Font.Size,
				Bold:      run.Font.Bold,
				Italic:    run.Font.Italic,
				Underline: run.Font.Underline,
				Color:     run.Font.Color,
			}
		}

		richTextRuns = append(richTextRuns, rtRun)
	}

	return richTextRuns, nil
}

// extractTables extracts Excel tables from a sheet
func (c *converter) extractTables(f *excelize.File, sheetName string) ([]models.Table, error) {
	tables, err := f.GetTables(sheetName)
	if err != nil {
		return nil, err
	}

	excelTables := make([]models.Table, 0, len(tables))
	for _, table := range tables {
		showHeaders := false
		if table.ShowHeaderRow != nil {
			showHeaders = *table.ShowHeaderRow
		}

		excelTable := models.Table{
			Name:            table.Name,
			Range:           table.Range,
			ShowHeaders:     showHeaders,
			ShowTotals:      false, // excelize doesn't provide this
			StyleName:       "",    // excelize doesn't provide this
			ShowFirstColumn: table.ShowFirstColumn,
			ShowLastColumn:  table.ShowLastColumn,
		}

		// Note: Column information would need to be extracted from the table structure
		// This might require additional parsing of the table range

		excelTables = append(excelTables, excelTable)
	}

	return excelTables, nil
}
