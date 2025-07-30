package converter

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Classic-Homes/sheetsync/internal/utils"
	"github.com/Classic-Homes/sheetsync/pkg/models"
	"github.com/xuri/excelize/v2"
)

func (c *converter) JSONToExcel(doc *models.ExcelDocument, outputPath string, options ConvertOptions) error {
	if doc == nil {
		return utils.WrapError(fmt.Errorf("document cannot be nil"), utils.ErrorTypeConverter, "JSONToExcel", "nil document provided")
	}

	// Create new Excel file
	f := excelize.NewFile()
	defer f.Close()

	// Remove default sheet
	f.DeleteSheet("Sheet1")

	// Process each sheet
	for _, sheet := range doc.Sheets {
		// Create sheet
		sheetIndex, err := f.NewSheet(sheet.Name)
		if err != nil {
			return utils.WrapError(err, utils.ErrorTypeConverter, "createSheet",
				fmt.Sprintf("failed to create sheet %s", sheet.Name))
		}

		// Set as active sheet if it's the first one
		if sheet.Index == 0 {
			f.SetActiveSheet(sheetIndex)
		}

		// Set sheet visibility
		if sheet.Hidden {
			err = f.SetSheetVisible(sheet.Name, false)
			if err != nil {
				c.logger.Warnf("Failed to hide sheet %s: %v", sheet.Name, err)
			}
		}

		// Process cells
		for cellRef, cell := range sheet.Cells {
			// Set cell value
			if cell.Formula != "" && options.PreserveFormulas {
				err = f.SetCellFormula(sheet.Name, cellRef, cell.Formula)
			} else {
				err = f.SetCellValue(sheet.Name, cellRef, cell.Value)
			}

			if err != nil {
				c.logger.Warnf("Failed to set cell %s in sheet %s: %v", cellRef, sheet.Name, err)
				continue
			}

			// Set cell style if requested
			if options.PreserveStyles && cell.Style != nil {
				// TODO: Apply cell style
				// This requires creating style from our model
			}

			// Set comment if requested
			if options.PreserveComments && cell.Comment != nil {
				err = f.AddComment(sheet.Name, excelize.Comment{
					Cell:   cellRef,
					Author: cell.Comment.Author,
					Text:   cell.Comment.Text,
				})
				if err != nil {
					c.logger.Warnf("Failed to add comment to cell %s: %v", cellRef, err)
				}
			}

			// Set hyperlink if exists
			if cell.Hyperlink != "" {
				err = f.SetCellHyperLink(sheet.Name, cellRef, cell.Hyperlink, "External")
				if err != nil {
					c.logger.Warnf("Failed to set hyperlink for cell %s: %v", cellRef, err)
				}
			}
		}

		// Apply merged cells
		for _, mergedCell := range sheet.MergedCells {
			parts := strings.Split(mergedCell.Range, ":")
			if len(parts) == 2 {
				err = f.MergeCell(sheet.Name, parts[0], parts[1])
				if err != nil {
					c.logger.Warnf("Failed to merge cells %s: %v", mergedCell.Range, err)
				}
			}
		}

		// Set column widths
		for col, width := range sheet.ColumnWidths {
			err = f.SetColWidth(sheet.Name, col, col, width)
			if err != nil {
				c.logger.Warnf("Failed to set column width for %s: %v", col, err)
			}
		}

		// Set row heights
		for row, height := range sheet.RowHeights {
			err = f.SetRowHeight(sheet.Name, row, height)
			if err != nil {
				c.logger.Warnf("Failed to set row height for %d: %v", row, err)
			}
		}
	}

	// Set document properties if available
	if doc.Properties != (models.DocumentProperties{}) {
		err := f.SetDocProps(&excelize.DocProperties{
			Title:       doc.Properties.Title,
			Subject:     doc.Properties.Subject,
			Creator:     doc.Properties.Author,
			Keywords:    doc.Properties.Keywords,
			Description: doc.Properties.Description,
			// Note: Company field is not available in excelize.DocProperties
			// It's stored in our model but cannot be set back in Excel
		})
		if err != nil {
			c.logger.Warnf("Failed to set document properties: %v", err)
		}
	}

	// Set defined names
	for name, refersTo := range doc.DefinedNames {
		err := f.SetDefinedName(&excelize.DefinedName{
			Name:     name,
			RefersTo: refersTo,
			Scope:    "Workbook",
		})
		if err != nil {
			c.logger.Warnf("Failed to set defined name %s: %v", name, err)
		}
	}

	// Save the file
	if err := f.SaveAs(outputPath); err != nil {
		return utils.WrapError(err, utils.ErrorTypeConverter, "saveFile", "failed to save Excel file")
	}

	return nil
}

// Helper function to convert string column to int
func columnNameToNumber(name string) (int, error) {
	if name == "" {
		return 0, utils.NewError(utils.ErrorTypeValidation, "columnNameToNumber", "column name cannot be empty")
	}

	col := 0
	for i := 0; i < len(name); i++ {
		if name[i] < 'A' || name[i] > 'Z' {
			return 0, utils.NewError(utils.ErrorTypeValidation, "columnNameToNumber",
				fmt.Sprintf("invalid column name: %s", name))
		}
		col = col*26 + int(name[i]-'A'+1)
	}

	return col, nil
}

// Helper function to parse cell reference
func parseCellReference(ref string) (col, row int, err error) {
	// Find where letters end and numbers begin
	i := 0
	for i < len(ref) && (ref[i] >= 'A' && ref[i] <= 'Z') {
		i++
	}

	if i == 0 || i == len(ref) {
		return 0, 0, utils.NewError(utils.ErrorTypeValidation, "parseCellReference",
			fmt.Sprintf("invalid cell reference: %s", ref))
	}

	colName := ref[:i]
	rowStr := ref[i:]

	col, err = columnNameToNumber(colName)
	if err != nil {
		return 0, 0, err
	}

	row, err = strconv.Atoi(rowStr)
	if err != nil {
		return 0, 0, utils.WrapError(err, utils.ErrorTypeValidation, "parseCellReference",
			fmt.Sprintf("invalid row number in reference %s", ref))
	}

	return col, row, nil
}
