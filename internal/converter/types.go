package converter

import (
	"fmt"
	"strconv"
	"strings"
	"github.com/Classic-Homes/sheetsync/pkg/models"
)

// detectCellType determines the cell type based on its value
func (c *converter) detectCellType(value interface{}) models.CellType {
	if value == nil {
		return models.CellTypeString
	}

	switch v := value.(type) {
	case string:
		// Check if it's a number stored as string
		if _, err := strconv.ParseFloat(v, 64); err == nil {
			return models.CellTypeNumber
		}
		// Check if it's a boolean
		if strings.ToLower(v) == "true" || strings.ToLower(v) == "false" {
			return models.CellTypeBoolean
		}
		// Check if it starts with = (formula)
		if strings.HasPrefix(v, "=") {
			return models.CellTypeFormula
		}
		return models.CellTypeString
	case float64, float32, int, int32, int64:
		return models.CellTypeNumber
	case bool:
		return models.CellTypeBoolean
	default:
		return models.CellTypeString
	}
}

// parseNumber converts a string to a number (float64) if possible
func parseNumber(value string) (float64, error) {
	return strconv.ParseFloat(value, 64)
}

// parseStyleID extracts style information from Excel style ID
func (c *converter) extractCellStyle(f interface{}, styleID int) *models.CellStyle {
	// TODO: Implement style extraction using excelize
	// This is a placeholder that will be implemented when we have the full converter
	return nil
}

// formatCommitMessage formats a commit message with variable substitution
func formatCommitMessage(template string, replacements map[string]string) string {
	result := template
	for key, value := range replacements {
		result = strings.ReplaceAll(result, "{"+key+"}", value)
	}
	return result
}

// extractProperties converts excelize document properties to our model
func (c *converter) extractProperties(props interface{}) models.DocumentProperties {
	// TODO: Implement property extraction
	// This is a placeholder that will be implemented when we have the full converter
	return models.DocumentProperties{}
}

// cellReference converts column and row indices to Excel cell reference (e.g., "A1")
func cellReference(col, row int) (string, error) {
	if col < 1 || row < 1 {
		return "", fmt.Errorf("invalid cell coordinates: col=%d, row=%d", col, row)
	}
	
	// Convert column number to letter(s)
	colName := ""
	for col > 0 {
		col--
		colName = string(rune('A'+col%26)) + colName
		col /= 26
	}
	
	return fmt.Sprintf("%s%d", colName, row), nil
}