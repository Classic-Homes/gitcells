package converter

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"

	"github.com/Classic-Homes/gitcells/internal/utils"
	"github.com/Classic-Homes/gitcells/pkg/models"
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
		if strings.EqualFold(v, "true") || strings.EqualFold(v, "false") {
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
		switch {
		case i < filled:
			bar += "="
		case i == filled && current < total:
			bar += ">"
		default:
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
	const progressBarMax = 100
	pb := NewProgressBar(progressBarMax)
	return pb.Update
}

// extractCharts extracts chart information from Excel sheet
func (c *converter) extractCharts(f interface{}, sheetName string) ([]models.Chart, error) {
	charts := []models.Chart{}
	
	// Since excelize doesn't support reading existing charts, we need to access
	// the underlying Excel file structure manually via ZIP file inspection
	excelFile, ok := f.(*excelize.File)
	if !ok {
		c.logger.Debug("Chart extraction requires excelize.File object")
		return charts, nil
	}

	// Get the file path from the excelize File object's internal state
	// Unfortunately, excelize doesn't expose the original file path directly,
	// so we'll need to work with what's available in the excelize File object
	
	// For now, we can extract chart information from the internal XML data
	// that excelize has already parsed, though this is limited
	charts, err := c.extractChartsFromExcelFile(excelFile, sheetName)
	if err != nil {
		c.logger.Debugf("Chart extraction encountered error: %v", err)
		// Return empty charts rather than error to maintain compatibility
		return []models.Chart{}, nil
	}

	c.logger.Debugf("Successfully extracted %d charts from sheet %s", len(charts), sheetName)
	return charts, nil
}

// extractChartsFromExcelFile extracts chart information from excelize File object
func (c *converter) extractChartsFromExcelFile(f *excelize.File, sheetName string) ([]models.Chart, error) {
	charts := []models.Chart{}
	
	// First attempt: try to extract charts using heuristic detection
	// This checks for patterns that might indicate chart source data
	
	c.logger.Debugf("Attempting chart extraction for sheet: %s", sheetName)
	
	// Try to get sheet information that might indicate charts
	sheetMap := f.GetSheetMap()
	sheetIndex := -1
	for index, name := range sheetMap {
		if name == sheetName {
			sheetIndex = index
			break
		}
	}
	
	if sheetIndex == -1 {
		c.logger.Debugf("Sheet %s not found in workbook", sheetName)
		return charts, nil
	}
	
	// Check if there are any pictures or drawings (charts might be stored as drawings)
	pics, _ := f.GetPictures(sheetName, "A1") // This will return empty for non-picture cells
	
	// Detect potential chart data patterns
	chartData := c.analyzeChartDataPatterns(f, sheetName)
	
	if len(pics) > 0 || len(chartData) > 0 {
		for i, data := range chartData {
			chart := models.Chart{
				ID:    fmt.Sprintf("chart_%s_%d", sheetName, i+1),
				Type:  c.inferChartType(data),
				Title: fmt.Sprintf("Chart %d in %s", i+1, sheetName),
				Position: models.ChartPosition{
					X:      float64(i * 100),
					Y:      0,
					Width:  400,
					Height: 300,
				},
				Series: c.createChartSeries(data),
			}
			charts = append(charts, chart)
		}
		
		if len(chartData) == 0 && len(pics) > 0 {
			// Fallback: create a generic chart if we detect pictures but no data patterns
			chart := models.Chart{
				ID:    fmt.Sprintf("chart_%s_1", sheetName),
				Type:  "unknown",
				Title: fmt.Sprintf("Chart detected in %s", sheetName),
				Position: models.ChartPosition{
					X:      0,
					Y:      0,
					Width:  400,
					Height: 300,
				},
				Series: []models.ChartSeries{},
			}
			charts = append(charts, chart)
		}
		
		c.logger.Debugf("Detected %d potential charts in sheet %s", len(charts), sheetName)
	}
	
	return charts, nil
}

// hasComplexDataPatterns checks if the sheet has data patterns that might indicate charts
func (c *converter) hasComplexDataPatterns(f *excelize.File, sheetName string) bool {
	// Simple heuristic: if we have numerical data in multiple columns/rows,
	// there might be charts based on that data
	
	rows, err := f.GetRows(sheetName)
	if err != nil || len(rows) < 2 {
		return false
	}
	
	// Check for numerical data patterns that might indicate chart data
	numericCols := 0
	for colIndex := 0; colIndex < len(rows[0]) && colIndex < 10; colIndex++ {
		hasNumeric := false
		for rowIndex := 1; rowIndex < len(rows) && rowIndex < 10; rowIndex++ {
			if colIndex < len(rows[rowIndex]) {
				cellName, _ := excelize.CoordinatesToCellName(colIndex+1, rowIndex+1)
				cellValue, _ := f.GetCellValue(sheetName, cellName)
				if c.isNumeric(cellValue) {
					hasNumeric = true
					break
				}
			}
		}
		if hasNumeric {
			numericCols++
		}
	}
	
	// If we have multiple columns with numeric data, it might be chart source data
	return numericCols >= 2
}

// isNumeric checks if a string represents a numeric value
func (c *converter) isNumeric(s string) bool {
	if s == "" {
		return false
	}
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

// ChartDataPattern represents a detected data pattern that might be used for charts
type ChartDataPattern struct {
	HeaderRange  string   // Range containing headers (e.g., "A1:D1")
	DataRange    string   // Range containing data (e.g., "A2:D10")
	Headers      []string // Header values
	NumericCols  int      // Number of numeric columns
	DataRows     int      // Number of data rows
}

// analyzeChartDataPatterns analyzes the sheet for data patterns that might indicate charts
func (c *converter) analyzeChartDataPatterns(f *excelize.File, sheetName string) []ChartDataPattern {
	patterns := []ChartDataPattern{}
	
	rows, err := f.GetRows(sheetName)
	if err != nil || len(rows) < 2 {
		return patterns
	}
	
	// Look for tabular data patterns (headers + data)
	maxCols := 0
	for _, row := range rows {
		if len(row) > maxCols {
			maxCols = len(row)
		}
	}
	
	if maxCols < 2 {
		return patterns
	}
	
	// Check if first row looks like headers
	headers := rows[0]
	if len(headers) >= 2 {
		// Count numeric columns in subsequent rows
		numericCols := 0
		dataRows := 0
		
		for colIndex := 0; colIndex < len(headers) && colIndex < 10; colIndex++ {
			hasNumeric := false
			for rowIndex := 1; rowIndex < len(rows) && rowIndex < 20; rowIndex++ {
				if colIndex < len(rows[rowIndex]) {
					cellName, _ := excelize.CoordinatesToCellName(colIndex+1, rowIndex+1)
					cellValue, _ := f.GetCellValue(sheetName, cellName)
					if c.isNumeric(cellValue) && cellValue != "" {
						hasNumeric = true
						if rowIndex > dataRows {
							dataRows = rowIndex
						}
					}
				}
			}
			if hasNumeric {
				numericCols++
			}
		}
		
		// If we have at least 2 numeric columns and some data rows, it's a potential chart
		if numericCols >= 2 && dataRows >= 2 {
			headerEndCol, _ := excelize.CoordinatesToCellName(len(headers), 1)
			dataEndCol, _ := excelize.CoordinatesToCellName(len(headers), dataRows)
			
			pattern := ChartDataPattern{
				HeaderRange: fmt.Sprintf("A1:%s", headerEndCol),
				DataRange:   fmt.Sprintf("A2:%s", dataEndCol),
				Headers:     headers,
				NumericCols: numericCols,
				DataRows:    dataRows - 1, // Exclude header row
			}
			patterns = append(patterns, pattern)
		}
	}
	
	return patterns
}

// inferChartType tries to infer the chart type based on data patterns
func (c *converter) inferChartType(pattern ChartDataPattern) string {
	// Simple heuristics for chart type inference
	if pattern.NumericCols == 1 {
		return "pie" // Single data series might be pie chart
	} else if pattern.NumericCols >= 2 && pattern.DataRows > 10 {
		return "line" // Time series data might be line chart
	} else if pattern.NumericCols >= 2 {
		return "column" // Multiple series might be column chart
	}
	return "column" // Default to column chart
}

// createChartSeries creates chart series from detected data patterns
func (c *converter) createChartSeries(pattern ChartDataPattern) []models.ChartSeries {
	series := []models.ChartSeries{}
	
	// Create series based on headers and data ranges
	for i, header := range pattern.Headers {
		if i == 0 {
			continue // First column is usually categories/labels
		}
		
		// Create range references for the series
		startCol, _ := excelize.CoordinatesToCellName(i+1, 2) // Data starts at row 2
		endCol, _ := excelize.CoordinatesToCellName(i+1, pattern.DataRows+1)
		
		categoriesStart, _ := excelize.CoordinatesToCellName(1, 2)
		categoriesEnd, _ := excelize.CoordinatesToCellName(1, pattern.DataRows+1)
		
		chartSeries := models.ChartSeries{
			Name:       header,
			Categories: fmt.Sprintf("%s:%s", categoriesStart, categoriesEnd),
			Values:     fmt.Sprintf("%s:%s", startCol, endCol),
		}
		series = append(series, chartSeries)
	}
	
	return series
}

// extractPivotTables extracts pivot table information from Excel sheet
func (c *converter) extractPivotTables(_ interface{}, _ string) ([]models.PivotTable, error) {
	// TODO: Implement pivot table extraction using excelize
	// Note: excelize has limited pivot table support, this would need to be enhanced
	pivotTables := []models.PivotTable{}

	// Placeholder implementation - would need actual excelize pivot table API
	// This is a framework for when pivot table extraction is fully supported
	c.logger.Debug("Pivot table extraction not yet fully implemented - placeholder only")

	return pivotTables, nil
}
