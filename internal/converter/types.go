package converter

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"

	"github.com/Classic-Homes/sheetsync/internal/utils"
	"github.com/Classic-Homes/sheetsync/pkg/models"
	"github.com/xuri/excelize/v2"
)

// detectCellType determines the cell type based on its value and formula
func (c *converter) detectCellType(value interface{}, formula string) models.CellType {
	if value == nil {
		return models.CellTypeString
	}

	// Check for formula first (takes precedence)
	if formula != "" {
		// Check if it's an array formula
		if c.isArrayFormula(formula) {
			return models.CellTypeArrayFormula
		}
		return models.CellTypeFormula
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
		// Check if it starts with = (formula) - backup check
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

// isArrayFormula checks if a formula is an array formula
func (c *converter) isArrayFormula(formula string) bool {
	// Array formulas are typically surrounded by curly braces in Excel
	// or contain specific array patterns that Excel recognizes
	formulaTrimmed := strings.TrimSpace(formula)

	// Check for curly braces (Excel's array formula indicator)
	if strings.HasPrefix(formulaTrimmed, "{") && strings.HasSuffix(formulaTrimmed, "}") {
		return true
	}

	// Check for specific array function patterns
	formulaUpper := strings.ToUpper(formula)

	// Functions that are commonly used in array contexts with multiple ranges
	arrayOnlyFunctions := []string{"TRANSPOSE(", "MMULT(", "SUMPRODUCT("}
	for _, fn := range arrayOnlyFunctions {
		if strings.Contains(formulaUpper, fn) {
			return true
		}
	}

	// More conservative detection for regular functions
	// Only consider it an array formula if it has multiple complex range references
	// or multiple ranges in the same function
	rangeCount := strings.Count(formula, ":")
	commaCount := strings.Count(formula, ",")

	// Array formulas typically have multiple ranges or complex patterns
	if rangeCount > 1 && commaCount > 2 {
		return true
	}

	return false
}

// extractEnhancedFormulaInfo extracts detailed formula information
func (c *converter) extractEnhancedFormulaInfo(f *excelize.File, sheetName, cellRef string) (string, string, *models.ArrayFormula, error) {
	// Get standard A1 formula
	formula, err := f.GetCellFormula(sheetName, cellRef)
	if err != nil {
		return "", "", nil, err
	}

	// TODO: Extract R1C1 formula if supported by excelize
	// Currently excelize doesn't have direct R1C1 support
	formulaR1C1 := ""

	// Check for array formula
	var arrayFormula *models.ArrayFormula
	if formula != "" && c.isArrayFormula(formula) {
		// TODO: Detect array formula range - would need excelize enhancement
		arrayFormula = &models.ArrayFormula{
			Formula: formula,
			Range:   cellRef, // Placeholder - would need actual range detection
			IsCSE:   true,    // Assume CSE for now
		}
	}

	return formula, formulaR1C1, arrayFormula, nil
}

// parseNumber converts a string to a number (float64) if possible
func parseNumber(value string) (float64, error) {
	num, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, utils.WrapError(err, utils.ErrorTypeConverter, "parseNumber",
			fmt.Sprintf("failed to parse number from value: %s", value))
	}
	return num, nil
}

// extractCellStyle extracts style information from Excel style ID
func (c *converter) extractCellStyle(f *excelize.File, styleID int) *models.CellStyle {
	if styleID == 0 {
		return nil // No custom style
	}

	// Get style information from excelize
	// Note: excelize has limited style extraction capabilities
	style := &models.CellStyle{}

	// TODO: Implement comprehensive style extraction
	// Currently excelize doesn't provide direct access to all style properties
	// This would require parsing the underlying XML or using enhanced excelize methods

	// Placeholder implementation - would need actual excelize style API enhancement
	c.logger.Debugf("Style extraction for ID %d - limited implementation", styleID)

	// For now, return basic style structure
	// This would be enhanced when excelize provides better style introspection
	return style
}

// extractFullCellStyle provides comprehensive cell style extraction
func (c *converter) extractFullCellStyle(f *excelize.File, sheetName, cellRef string) *models.CellStyle {
	styleID, err := f.GetCellStyle(sheetName, cellRef)
	if err != nil || styleID == 0 {
		return nil
	}

	style := &models.CellStyle{}

	// Extract number format
	// TODO: excelize doesn't currently provide direct access to number formats
	// This would need to be implemented when excelize supports it

	// Extract font information
	// TODO: Implement font extraction when excelize supports it
	style.Font = &models.Font{
		Name: "Calibri", // Default - would extract actual font
		Size: 11,        // Default - would extract actual size
	}

	// Extract fill information
	// TODO: Implement fill extraction when excelize supports it

	// Extract border information
	// TODO: Implement border extraction when excelize supports it

	// Extract alignment information
	// TODO: Implement alignment extraction when excelize supports it

	c.logger.Debugf("Extracted style for cell %s - placeholder implementation", cellRef)
	return style
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
	if props == nil {
		return models.DocumentProperties{}
	}

	// Type assert to excelize.DocProperties
	if docProps, ok := props.(*excelize.DocProperties); ok {
		return models.DocumentProperties{
			Title:       docProps.Title,
			Subject:     docProps.Subject,
			Author:      docProps.Creator,
			Company:     docProps.Category,
			Keywords:    docProps.Keywords,
			Description: docProps.Description,
		}
	}

	return models.DocumentProperties{}
}

// cellReference converts column and row indices to Excel cell reference (e.g., "A1")
func cellReference(col, row int) (string, error) {
	if col < 1 || row < 1 {
		return "", utils.NewError(utils.ErrorTypeValidation, "cellReference",
			fmt.Sprintf("invalid cell coordinates: col=%d, row=%d", col, row))
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

// MemoryStats tracks memory usage during conversion
type MemoryStats struct {
	InitialMemory uint64
	PeakMemory    uint64
	CurrentMemory uint64
	GCCount       uint32
}

// GetMemoryStats returns current memory statistics
func GetMemoryStats() *MemoryStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return &MemoryStats{
		CurrentMemory: m.Alloc,
		PeakMemory:    m.TotalAlloc,
		GCCount:       m.NumGC,
	}
}

// ForceGC forces garbage collection and returns memory stats
func ForceGC() *MemoryStats {
	runtime.GC()
	return GetMemoryStats()
}

// StreamingConfig defines configuration for streaming processing
type StreamingConfig struct {
	ChunkSize        int    // Number of rows to process at once
	MemoryThreshold  uint64 // Memory threshold in bytes to trigger GC
	ProgressCallback func(processed, total int)
}

// DefaultStreamingConfig returns sensible defaults for streaming
func DefaultStreamingConfig() *StreamingConfig {
	return &StreamingConfig{
		ChunkSize:       1000,      // Process 1000 rows at a time
		MemoryThreshold: 100 << 20, // 100MB threshold
	}
}

// ProgressBar represents a simple console progress bar
type ProgressBar struct {
	current int
	total   int
	width   int
	stage   string
}

// NewProgressBar creates a new progress bar
func NewProgressBar(total int) *ProgressBar {
	return &ProgressBar{
		current: 0,
		total:   total,
		width:   50, // Default width
	}
}

// Update updates the progress bar display
func (pb *ProgressBar) Update(stage string, current, total int) {
	pb.current = current
	pb.total = total
	pb.stage = stage

	// Calculate percentage
	percentage := float64(current) / float64(total) * 100
	if total == 0 {
		percentage = 100
	}

	// Calculate filled portion of bar
	filled := int(float64(current) / float64(total) * float64(pb.width))
	if filled > pb.width {
		filled = pb.width
	}

	// Build progress bar string
	bar := "["
	for i := 0; i < pb.width; i++ {
		if i < filled {
			bar += "="
		} else if i == filled && current < total {
			bar += ">"
		} else {
			bar += " "
		}
	}
	bar += "]"

	// Print progress (use \r to overwrite previous line)
	fmt.Printf("\r%s %s %.1f%% (%d/%d)", stage, bar, percentage, current, total)

	// Print newline when complete
	if current >= total {
		fmt.Println()
	}
}

// DefaultProgressCallback returns a default progress callback that shows a progress bar
func DefaultProgressCallback() func(string, int, int) {
	pb := NewProgressBar(100)
	return pb.Update
}

// extractCharts extracts chart information from Excel sheet
func (c *converter) extractCharts(f interface{}, sheetName string) ([]models.Chart, error) {
	// TODO: Implement chart extraction using excelize
	// Note: excelize has limited chart support, this would need to be enhanced
	// or use a different library for full chart extraction
	charts := []models.Chart{}

	// Placeholder implementation - would need actual excelize chart API
	// This is a framework for when chart extraction is fully supported
	c.logger.Debug("Chart extraction not yet fully implemented - placeholder only")

	return charts, nil
}

// extractPivotTables extracts pivot table information from Excel sheet
func (c *converter) extractPivotTables(f interface{}, sheetName string) ([]models.PivotTable, error) {
	// TODO: Implement pivot table extraction using excelize
	// Note: excelize has limited pivot table support, this would need to be enhanced
	pivotTables := []models.PivotTable{}

	// Placeholder implementation - would need actual excelize pivot table API
	// This is a framework for when pivot table extraction is fully supported
	c.logger.Debug("Pivot table extraction not yet fully implemented - placeholder only")

	return pivotTables, nil
}
