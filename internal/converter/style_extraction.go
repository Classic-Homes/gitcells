package converter

import (
	"github.com/Classic-Homes/gitcells/pkg/models"
	"github.com/xuri/excelize/v2"
)

// extractFullCellStyle extracts comprehensive cell style information using excelize's capabilities
func (c *converter) extractFullCellStyle(f *excelize.File, sheetName, cellRef string) *models.CellStyle {
	// Get the style index for the cell
	styleID, err := f.GetCellStyle(sheetName, cellRef)
	if err != nil || styleID == 0 {
		c.logger.Debugf("No style found for cell %s: %v", cellRef, err)
		return nil
	}

	// Get the actual style details
	excelStyle, err := f.GetStyle(styleID)
	if err != nil {
		c.logger.Debugf("Failed to get style details for cell %s (style ID %d): %v", cellRef, styleID, err)
		return nil
	}

	// Convert excelize style to our model
	cellStyle := &models.CellStyle{}

	// Extract Font information
	if excelStyle.Font != nil {
		cellStyle.Font = &models.Font{
			Name:      excelStyle.Font.Family,
			Size:      excelStyle.Font.Size,
			Bold:      excelStyle.Font.Bold,
			Italic:    excelStyle.Font.Italic,
			Underline: excelStyle.Font.Underline,
			Color:     excelStyle.Font.Color,
		}
	}

	// Extract Fill information
	if excelStyle.Fill.Type != "" || len(excelStyle.Fill.Color) > 0 {
		cellStyle.Fill = &models.Fill{
			Type:    excelStyle.Fill.Type,
			Pattern: convertExcelizePatternToString(excelStyle.Fill.Pattern),
		}

		// Handle color array from excelize
		if len(excelStyle.Fill.Color) > 0 {
			cellStyle.Fill.Color = excelStyle.Fill.Color[0]
			if len(excelStyle.Fill.Color) > 1 {
				cellStyle.Fill.BgColor = excelStyle.Fill.Color[1]
			}
		}
	}

	// Extract Border information
	if len(excelStyle.Border) > 0 {
		cellStyle.Border = &models.Border{}

		for _, border := range excelStyle.Border {
			borderLine := &models.BorderLine{
				Style: convertExcelizeBorderStyleToString(border.Style),
				Color: border.Color,
			}

			switch border.Type {
			case "left":
				cellStyle.Border.Left = borderLine
			case "right":
				cellStyle.Border.Right = borderLine
			case "top":
				cellStyle.Border.Top = borderLine
			case "bottom":
				cellStyle.Border.Bottom = borderLine
			}
		}
	}

	// Extract Alignment information
	if excelStyle.Alignment != nil {
		cellStyle.Alignment = &models.Alignment{
			Horizontal:   excelStyle.Alignment.Horizontal,
			Vertical:     excelStyle.Alignment.Vertical,
			WrapText:     excelStyle.Alignment.WrapText,
			TextRotation: excelStyle.Alignment.TextRotation,
		}
	}

	// Extract Number Format
	if excelStyle.NumFmt != 0 {
		cellStyle.NumberFormat = convertExcelizeNumFmtToString(excelStyle.NumFmt)
	} else if excelStyle.CustomNumFmt != nil && *excelStyle.CustomNumFmt != "" {
		cellStyle.NumberFormat = *excelStyle.CustomNumFmt
	}

	c.logger.Debugf("Extracted style for cell %s - Font: %+v, Fill: %+v", cellRef, cellStyle.Font, cellStyle.Fill)
	return cellStyle
}

// convertExcelizePatternToString converts excelize pattern ID to string
func convertExcelizePatternToString(pattern int) string {
	patterns := map[int]string{
		0:  "",
		1:  "solid",
		2:  "mediumGray",
		3:  "gray75",
		4:  "gray50",
		5:  "gray25",
		6:  "gray125",
		7:  "gray0625",
		8:  "darkGray",
		9:  "lightDown",
		10: "lightUp",
		11: "lightGrid",
		12: "lightTrellis",
		13: "lightHorizontal",
		14: "lightVertical",
		15: "darkDown",
		16: "darkUp",
		17: "darkGrid",
		18: "darkTrellis",
		19: "darkHorizontal",
		20: "darkVertical",
	}

	if val, ok := patterns[pattern]; ok {
		return val
	}
	return ""
}

// convertExcelizeBorderStyleToString converts excelize border style ID to string
func convertExcelizeBorderStyleToString(style int) string {
	styles := map[int]string{
		0:  "",
		1:  "thin",
		2:  "medium",
		3:  "dashed",
		4:  "dotted",
		5:  "thick",
		6:  "double",
		7:  "hair",
		8:  "mediumDashed",
		9:  "dashDot",
		10: "mediumDashDot",
		11: "dashDotDot",
		12: "mediumDashDotDot",
		13: "slantDashDot",
	}

	if val, ok := styles[style]; ok {
		return val
	}
	return "thin" // default
}

// convertExcelizeNumFmtToString converts excelize number format ID to string
func convertExcelizeNumFmtToString(numFmt int) string {
	// Common Excel number format IDs
	formats := map[int]string{
		0:  "General",
		1:  "0",
		2:  "0.00",
		3:  "#,##0",
		4:  "#,##0.00",
		5:  "$#,##0_);($#,##0)",
		6:  "$#,##0_);[Red]($#,##0)",
		7:  "$#,##0.00_);($#,##0.00)",
		8:  "$#,##0.00_);[Red]($#,##0.00)",
		9:  "0%",
		10: "0.00%",
		11: "0.00E+00",
		12: "# ?/?",
		13: "# ??/??",
		14: "mm-dd-yy",
		15: "d-mmm-yy",
		16: "d-mmm",
		17: "mmm-yy",
		18: "h:mm AM/PM",
		19: "h:mm:ss AM/PM",
		20: "h:mm",
		21: "h:mm:ss",
		22: "m/d/yy h:mm",
		37: "#,##0 ;(#,##0)",
		38: "#,##0 ;[Red](#,##0)",
		39: "#,##0.00;(#,##0.00)",
		40: "#,##0.00;[Red](#,##0.00)",
		41: "_(* #,##0_);_(* (#,##0);_(* \"-\"_);_(@_)",
		42: "_($* #,##0_);_($* (#,##0);_($* \"-\"_);_(@_)",
		43: "_(* #,##0.00_);_(* (#,##0.00);_(* \"-\"??_);_(@_)",
		44: "_($* #,##0.00_);_($* (#,##0.00);_($* \"-\"??_);_(@_)",
		45: "mm:ss",
		46: "[h]:mm:ss",
		47: "mmss.0",
		48: "##0.0E+0",
		49: "@",
	}

	if val, ok := formats[numFmt]; ok {
		return val
	}
	// For custom formats, return empty string and let the CustomNumFmt handle it
	return ""
}
