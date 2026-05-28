package converter

import (
	"fmt"
	"os"
	"path/filepath"
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

	// excelize.NewFile() always seeds a "Sheet1". We can't delete it up
	// front (excelize refuses to delete the last sheet) so we postpone
	// until after the requested sheets are created. If the document
	// happens to define a "Sheet1" of its own, the loop below populates
	// it instead of leaving the placeholder.
	const defaultSheetName = "Sheet1"
	docHasDefaultSheet := false
	for _, s := range doc.Sheets {
		if s.Name == defaultSheetName {
			docHasDefaultSheet = true
			break
		}
	}

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
				linkType := classifyHyperlink(cell.Hyperlink)
				err = f.SetCellHyperLink(sheet.Name, cellRef, cell.Hyperlink, linkType)
				if err != nil {
					c.logger.Warnf("Failed to set %s hyperlink for cell %s: %v", linkType, cellRef, err)
				}
			}
		}

		// Apply merged cells. Validate both endpoints before calling
		// MergeCell — excelize swallows malformed ranges silently and a
		// stored range like "A1:B2:C3" would otherwise produce a partial
		// merge with no warning.
		for _, mergedCell := range sheet.MergedCells {
			parts := strings.Split(mergedCell.Range, ":")
			const expectedRangeParts = 2
			if len(parts) != expectedRangeParts {
				c.logger.Warnf("Skipping malformed merged-cell range on sheet %s: %q", sheet.Name, mergedCell.Range)
				continue
			}
			if _, _, err := excelize.CellNameToCoordinates(parts[0]); err != nil {
				c.logger.Warnf("Skipping merged-cell range with invalid start %q on sheet %s: %v", parts[0], sheet.Name, err)
				continue
			}
			if _, _, err := excelize.CellNameToCoordinates(parts[1]); err != nil {
				c.logger.Warnf("Skipping merged-cell range with invalid end %q on sheet %s: %v", parts[1], sheet.Name, err)
				continue
			}
			if err := f.MergeCell(sheet.Name, parts[0], parts[1]); err != nil {
				c.logger.Warnf("Failed to merge cells %s on sheet %s: %v", mergedCell.Range, sheet.Name, err)
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

	// Remove the placeholder default sheet now that real sheets exist
	// (excelize refuses to delete the last sheet, hence the post-loop
	// timing).
	if !docHasDefaultSheet && len(doc.Sheets) > 0 {
		if err := f.DeleteSheet(defaultSheetName); err != nil {
			c.logger.Debugf("DeleteSheet(%s) failed (probably already gone): %v", defaultSheetName, err)
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

	// Set defined names. Iteration is over a sorted slice, so the
	// resulting workbook's defined-name table is also in deterministic
	// order — important for downstream byte-stability checks.
	for _, dn := range doc.DefinedNames {
		err := f.SetDefinedName(&excelize.DefinedName{
			Name:     dn.Name,
			RefersTo: dn.RefersTo,
			Scope:    "Workbook",
		})
		if err != nil {
			c.logger.Warnf("Failed to set defined name %s: %v", dn.Name, err)
		}
	}

	// Save atomically: write to a staging path in the same directory and
	// rename over the target. A crash, lock contention, or concurrent
	// watcher read mid-write must not leave a torn .xlsx replacing a
	// previously-good file. The rename is atomic on every supported OS
	// for files (unlike directories — see chunking.go:atomicReplaceDir).
	//
	// excelize validates the file extension on SaveAs, so we keep the
	// original extension at the end and slot the staging suffix in front
	// of it. The leading dot hides the staging file from typical
	// directory listings while it's in flight.
	ext := filepath.Ext(outputPath)
	base := strings.TrimSuffix(filepath.Base(outputPath), ext)
	stagingPath := filepath.Join(filepath.Dir(outputPath), "."+base+".tmp."+uniqueSuffix()+ext)
	committed := false
	defer func() {
		if !committed {
			if rmErr := os.Remove(stagingPath); rmErr != nil && !os.IsNotExist(rmErr) {
				c.logger.Warnf("Failed to clean up staging file %s: %v", stagingPath, rmErr)
			}
		}
	}()

	if err := f.SaveAs(stagingPath); err != nil {
		return utils.WrapError(err, utils.ErrorTypeConverter, "saveFile", "failed to save Excel file to staging path")
	}

	// fsync the staging file so the rename publishes durable content,
	// not a directory entry pointing at zero bytes after a power loss.
	if sf, err := os.OpenFile(stagingPath, os.O_RDWR, 0); err == nil {
		_ = sf.Sync()
		_ = sf.Close()
	}

	if err := os.Rename(stagingPath, outputPath); err != nil {
		return utils.WrapError(err, utils.ErrorTypeConverter, "saveFile", "failed to install Excel file")
	}
	committed = true

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
		sides := []struct {
			name   string
			border *models.BorderLine
		}{
			{"left", cellStyle.Border.Left},
			{"right", cellStyle.Border.Right},
			{"top", cellStyle.Border.Top},
			{"bottom", cellStyle.Border.Bottom},
		}
		for _, s := range sides {
			if s.border == nil {
				continue
			}
			style.Border = append(style.Border, excelize.Border{
				Type:  s.name,
				Color: s.border.Color,
				Style: convertBorderStyle(s.border.Style),
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

	// Convert Number Format. If the stored format is one of the well-known
	// built-in IDs we use the integer code; otherwise we pass it through
	// as a custom format so the original string survives the round trip.
	// Without this, formats like "[$-409]m/d/yy h:mm AM/PM;@" silently
	// degraded to "General".
	if cellStyle.NumberFormat != "" {
		if id, ok := builtinNumberFormatID(cellStyle.NumberFormat); ok {
			style.NumFmt = id
		} else {
			fmtCopy := cellStyle.NumberFormat
			style.CustomNumFmt = &fmtCopy
		}
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

// builtinNumberFormatIDs maps Excel's built-in number-format strings to
// their stable numeric IDs. Formats not in this table are custom and
// must be passed through via Style.CustomNumFmt so they round-trip.
var builtinNumberFormatIDs = map[string]int{
	"General":                  0,
	"0":                        1,
	"0.00":                     2,
	"#,##0":                    3,
	"#,##0.00":                 4,
	"0%":                       9,
	"0.00%":                    10,
	"0.00E+00":                 11,
	"# ?/?":                    12,
	"# ??/??":                  13,
	"mm-dd-yy":                 14,
	"d-mmm-yy":                 15,
	"d-mmm":                    16,
	"mmm-yy":                   17,
	"h:mm AM/PM":               18,
	"h:mm:ss AM/PM":            19,
	"h:mm":                     20,
	"h:mm:ss":                  21,
	"m/d/yy h:mm":              22,
	"#,##0 ;(#,##0)":           37,
	"#,##0 ;[Red](#,##0)":      38,
	"#,##0.00;(#,##0.00)":      39,
	"#,##0.00;[Red](#,##0.00)": 40,
	"mm:ss":                    45,
	"[h]:mm:ss":                46,
	"mmss.0":                   47,
	"##0.0E+0":                 48,
	"@":                        49,
}

// classifyHyperlink returns the excelize linkType ("External" or
// "Location") for a stored hyperlink target. URL schemes and protocol
// links are External; sheet-relative references like "Sheet2!A1" or
// named ranges are Location. Default is External — a slightly wrong
// type still produces a clickable link, whereas the wrong type with
// a URL target produces no link at all.
func classifyHyperlink(link string) string {
	if strings.Contains(link, "://") {
		return "External"
	}
	if strings.HasPrefix(strings.ToLower(link), "mailto:") || strings.HasPrefix(strings.ToLower(link), "tel:") {
		return "External"
	}
	if strings.Contains(link, "!") {
		return "Location"
	}
	return "External"
}

// builtinNumberFormatID returns the Excel numeric ID for a built-in
// format string. The boolean reports whether the format is a known
// built-in; callers should treat false as "custom format, pass through".
func builtinNumberFormatID(format string) (int, bool) {
	id, ok := builtinNumberFormatIDs[format]
	return id, ok
}
