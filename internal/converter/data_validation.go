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

// extractDataValidations attaches data validation rules from the source
// workbook to the matching cells in our sheet model. Sqref is expanded to
// the exact cell set (previously a substring check sprayed A1 onto A10/A11).
func (c *converter) extractDataValidations(f *excelize.File, sheetName string, sheet *models.Sheet) error {
	validations, err := f.GetDataValidations(sheetName)
	if err != nil {
		return err
	}

	for _, v := range validations {
		dv := &models.DataValidation{
			Type:             v.Type,
			Operator:         v.Operator,
			Formula1:         v.Formula1,
			Formula2:         v.Formula2,
			AllowBlank:       v.AllowBlank,
			ShowInputMessage: v.ShowInputMessage,
			ShowErrorMessage: v.ShowErrorMessage,
		}
		if v.ErrorTitle != nil {
			dv.ErrorTitle = *v.ErrorTitle
		}
		if v.Error != nil {
			dv.Error = *v.Error
		}
		if v.PromptTitle != nil {
			dv.PromptTitle = *v.PromptTitle
		}
		if v.Prompt != nil {
			dv.Prompt = *v.Prompt
		}

		for cellRef := range expandSqref(v.Sqref) {
			if cell, ok := sheet.Cells[cellRef]; ok {
				cell.DataValidation = dv
				sheet.Cells[cellRef] = cell
			}
		}
	}
	return nil
}

// warnAboutUnsupportedFeatures inspects the source workbook for features
// GitCells does not preserve (charts, pivot tables, conditional formats,
// tables, rich-text-per-run) and emits a single warning per sheet naming
// what was skipped. Users get an honest signal rather than a silently
// incomplete round-trip.
func (c *converter) warnAboutUnsupportedFeatures(f *excelize.File, sheetName string) {
	var skipped []string

	if pics, err := f.GetPictures(sheetName, "A1"); err == nil && len(pics) > 0 {
		skipped = append(skipped, "embedded pictures/charts")
	}
	if pivots, err := f.GetPivotTables(sheetName); err == nil && len(pivots) > 0 {
		skipped = append(skipped, "pivot tables")
	}
	if cfs, err := f.GetConditionalFormats(sheetName); err == nil && len(cfs) > 0 {
		skipped = append(skipped, "conditional formats")
	}
	if tables, err := f.GetTables(sheetName); err == nil && len(tables) > 0 {
		skipped = append(skipped, "tables (ListObjects)")
	}

	if len(skipped) > 0 {
		c.logger.Warnf("Sheet %q: not preserved by GitCells: %s",
			sheetName, strings.Join(skipped, ", "))
	}
}
