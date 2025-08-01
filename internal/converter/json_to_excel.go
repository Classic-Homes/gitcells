package converter

import (
	"fmt"
	"strings"

	"github.com/Classic-Homes/gitcells/internal/utils"
	"github.com/Classic-Homes/gitcells/pkg/models"
	"github.com/xuri/excelize/v2"
)

func (c *converter) JSONToExcel(doc *models.ExcelDocument, outputPath string, options ConvertOptions) error {
	if doc == nil {
		return utils.NewError(utils.ErrorTypeConverter, "JSONToExcel", "document cannot be nil")
	}

	// Create new Excel file
	f := excelize.NewFile()
	defer func() { _ = f.Close() }()

	// Remove default sheet
	_ = f.DeleteSheet("Sheet1")

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
				excelizeStyle := c.convertCellStyleToExcelizeStyle(cell.Style)
				if excelizeStyle != nil {
					styleID, err := f.NewStyle(excelizeStyle)
					if err != nil {
						c.logger.Warnf("Failed to create style for cell %s: %v", cellRef, err)
					} else {
						err = f.SetCellStyle(sheet.Name, cellRef, cellRef, styleID)
						if err != nil {
							c.logger.Warnf("Failed to apply style to cell %s: %v", cellRef, err)
						}
					}
				}
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
			const expectedRangeParts = 2
			if len(parts) == expectedRangeParts {
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

// convertCellStyleToExcelizeStyle converts our CellStyle model to excelize.Style
func (c *converter) convertCellStyleToExcelizeStyle(cellStyle *models.CellStyle) *excelize.Style {
	if cellStyle == nil {
		return nil
	}

	style := &excelize.Style{}

	// Convert Font
	if cellStyle.Font != nil {
		style.Font = &excelize.Font{
			Bold:   cellStyle.Font.Bold,
			Italic: cellStyle.Font.Italic,
		}

		if cellStyle.Font.Name != "" {
			style.Font.Family = cellStyle.Font.Name
		}

		if cellStyle.Font.Size > 0 {
			style.Font.Size = cellStyle.Font.Size
		}

		if cellStyle.Font.Color != "" {
			style.Font.Color = cellStyle.Font.Color
		}

		if cellStyle.Font.Underline != "" {
			style.Font.Underline = cellStyle.Font.Underline
		}
	}

	// Convert Fill
	if cellStyle.Fill != nil {
		style.Fill = excelize.Fill{}

		if cellStyle.Fill.Type != "" {
			style.Fill.Type = cellStyle.Fill.Type
		}

		if cellStyle.Fill.Pattern != "" {
			style.Fill.Pattern = convertFillPattern(cellStyle.Fill.Pattern)
		}

		// Handle color array for excelize
		var colors []string
		if cellStyle.Fill.Color != "" {
			colors = append(colors, cellStyle.Fill.Color)
		}
		if cellStyle.Fill.BgColor != "" {
			colors = append(colors, cellStyle.Fill.BgColor)
		}
		if len(colors) > 0 {
			style.Fill.Color = colors
		}
	}

	// Convert Border
	if cellStyle.Border != nil {
		style.Border = []excelize.Border{}

		if cellStyle.Border.Left != nil {
			style.Border = append(style.Border, excelize.Border{
				Type:  "left",
				Color: cellStyle.Border.Left.Color,
				Style: convertBorderStyle(cellStyle.Border.Left.Style),
			})
		}

		if cellStyle.Border.Right != nil {
			style.Border = append(style.Border, excelize.Border{
				Type:  "right",
				Color: cellStyle.Border.Right.Color,
				Style: convertBorderStyle(cellStyle.Border.Right.Style),
			})
		}

		if cellStyle.Border.Top != nil {
			style.Border = append(style.Border, excelize.Border{
				Type:  "top",
				Color: cellStyle.Border.Top.Color,
				Style: convertBorderStyle(cellStyle.Border.Top.Style),
			})
		}

		if cellStyle.Border.Bottom != nil {
			style.Border = append(style.Border, excelize.Border{
				Type:  "bottom",
				Color: cellStyle.Border.Bottom.Color,
				Style: convertBorderStyle(cellStyle.Border.Bottom.Style),
			})
		}
	}

	// Convert Alignment
	if cellStyle.Alignment != nil {
		style.Alignment = &excelize.Alignment{
			Horizontal:   cellStyle.Alignment.Horizontal,
			Vertical:     cellStyle.Alignment.Vertical,
			WrapText:     cellStyle.Alignment.WrapText,
			TextRotation: cellStyle.Alignment.TextRotation,
		}
	}

	// Convert Number Format
	if cellStyle.NumberFormat != "" {
		style.NumFmt = convertNumberFormat(cellStyle.NumberFormat)
	}

	return style
}

// convertBorderStyle converts our border style to excelize border style
func convertBorderStyle(borderStyle string) int {
	switch borderStyle {
	case "thin":
		return 1
	case "medium":
		return 2
	case "thick":
		return 5
	case "double":
		return 6
	case "dotted":
		return 3
	case "dashed":
		return 4
	case "hair":
		return 7
	case "medium_dashed":
		return 8
	case "dash_dot":
		return 9
	case "medium_dash_dot":
		return 10
	case "dash_dot_dot":
		return 11
	case "medium_dash_dot_dot":
		return 12
	case "slanted_dash_dot":
		return 13
	default:
		return 1 // Default to thin
	}
}

// convertFillPattern converts our fill pattern to excelize fill pattern
func convertFillPattern(pattern string) int {
	switch pattern {
	case "solid":
		return 1
	case "gray75":
		return 3
	case "gray50":
		return 4
	case "gray25":
		return 5
	case "gray125":
		return 6
	case "gray0625":
		return 7
	case "lightDown":
		return 9
	case "lightUp":
		return 10
	case "lightGrid":
		return 11
	case "lightTrellis":
		return 12
	case "lightHorizontal":
		return 13
	case "lightVertical":
		return 14
	case "darkDown":
		return 15
	case "darkUp":
		return 16
	case "darkGrid":
		return 17
	case "darkTrellis":
		return 18
	case "darkHorizontal":
		return 19
	case "darkVertical":
		return 20
	default:
		return 1 // Default to solid
	}
}

// convertNumberFormat converts our number format to excelize number format ID
func convertNumberFormat(format string) int {
	// Common number formats with their Excel IDs
	switch format {
	case "General":
		return 0
	case "0":
		return 1
	case "0.00":
		return 2
	case "#,##0":
		return 3
	case "#,##0.00":
		return 4
	case "0%":
		return 9
	case "0.00%":
		return 10
	case "0.00E+00":
		return 11
	case "# ?/?":
		return 12
	case "# ??/??":
		return 13
	case "mm-dd-yy":
		return 14
	case "d-mmm-yy":
		return 15
	case "d-mmm":
		return 16
	case "mmm-yy":
		return 17
	case "h:mm AM/PM":
		return 18
	case "h:mm:ss AM/PM":
		return 19
	case "h:mm":
		return 20
	case "h:mm:ss":
		return 21
	case "m/d/yy h:mm":
		return 22
	case "#,##0 ;(#,##0)":
		return 37
	case "#,##0 ;[Red](#,##0)":
		return 38
	case "#,##0.00;(#,##0.00)":
		return 39
	case "#,##0.00;[Red](#,##0.00)":
		return 40
	case "mm:ss":
		return 45
	case "[h]:mm:ss":
		return 46
	case "mmss.0":
		return 47
	case "##0.0E+0":
		return 48
	case "@":
		return 49
	default:
		// For custom formats, return a default number format
		return 0 // General
	}
}
