package converter

import (
	"strings"
	
	"github.com/Classic-Homes/gitcells/pkg/models"
	"github.com/xuri/excelize/v2"
)

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
		
		// Apply to cells in the range
		// This is a simplified implementation - a full implementation would
		// parse the Sqref range and apply to all cells in that range
		for cellRef, cell := range sheet.Cells {
			// Check if this cell is within the validation range
			// For now, we'll do a simple string contains check
			if validation.Sqref == cellRef || strings.Contains(validation.Sqref, cellRef) {
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

// extractSheetProtection extracts sheet protection settings
func (c *converter) extractSheetProtection(f *excelize.File, sheetName string) *models.SheetProtection {
	// Note: Excelize doesn't provide a method to read protection settings directly
	// We would need to access the underlying XML structure for full extraction
	// For now, we can detect if a sheet is protected but not the specific settings
	
	// This is a placeholder - actual implementation would require XML parsing
	// or enhancement to the excelize library
	return nil
}

// convertConditionalStyle converts excelize conditional format style to our model
func (c *converter) convertConditionalStyle(style interface{}) *models.CellStyle {
	// This would convert the excelize style format to our CellStyle model
	// Implementation depends on the actual structure provided by excelize
	
	// For now, return a basic style
	return &models.CellStyle{
		// Style properties would be extracted here
	}
}

// extractRichText extracts rich text formatting from a cell
func (c *converter) extractRichText(f *excelize.File, sheetName, cellRef string) ([]models.RichTextRun, error) {
	runs, err := f.GetCellRichText(sheetName, cellRef)
	if err != nil || len(runs) == 0 {
		return nil, err
	}

	var richTextRuns []models.RichTextRun
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

	var excelTables []models.Table
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
			StyleName:       "", // excelize doesn't provide this
			ShowFirstColumn: table.ShowFirstColumn,
			ShowLastColumn:  table.ShowLastColumn,
		}
		
		// Note: Column information would need to be extracted from the table structure
		// This might require additional parsing of the table range
		
		excelTables = append(excelTables, excelTable)
	}
	
	return excelTables, nil
}

// extractAutoFilter extracts AutoFilter settings from a sheet
func (c *converter) extractAutoFilter(f *excelize.File, sheetName string) *models.AutoFilter {
	// Note: Excelize doesn't provide a direct method to read AutoFilter settings
	// This would require XML parsing or library enhancement
	return nil
}

